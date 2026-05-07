package hub

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

// WsMessage is the envelope sent to all connected frontends.
type WsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// notifyPayload mirrors the JSON emitted by the PostgreSQL trigger.
type notifyPayload struct {
	ID          string `json:"id"`
	CompanyID   uint   `json:"companyId"`
	ClientID    *uint  `json:"clientId"`
	ContactName string `json:"contactName"`
	From        string `json:"from"`
	FromUserID  string `json:"fromUserId"`
	WaTimestamp int64  `json:"waTimestamp"`
	Body        string `json:"body"`
	MessageType string `json:"messageType"`
	Direction   string `json:"direction"`
	Status      string `json:"status"`
}

// Client represents a single WebSocket connection.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	CompanyID uint
}

// Hub manages all active WebSocket clients grouped by company.
type Hub struct {
	mu         sync.RWMutex
	rooms      map[uint]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

type broadcastMsg struct {
	companyID uint
	data      []byte
}

func New() *Hub {
	return &Hub{
		rooms:      make(map[uint]map[*Client]struct{}),
		register:   make(chan *Client, 32),
		unregister: make(chan *Client, 32),
		broadcast:  make(chan broadcastMsg, 512),
	}
}

// Run processes registrations and broadcasts. Call in a dedicated goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.rooms[c.CompanyID] == nil {
				h.rooms[c.CompanyID] = make(map[*Client]struct{})
			}
			h.rooms[c.CompanyID][c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[c.CompanyID]; ok {
				delete(room, c)
				if len(room) == 0 {
					delete(h.rooms, c.CompanyID)
				}
			}
			h.mu.Unlock()
			close(c.send)

		case msg := <-h.broadcast:
			h.mu.RLock()
			room := h.rooms[msg.companyID]
			h.mu.RUnlock()
			for c := range room {
				select {
				case c.send <- msg.data:
				default:
					h.mu.Lock()
					delete(h.rooms[c.CompanyID], c)
					h.mu.Unlock()
					close(c.send)
				}
			}
		}
	}
}

// Broadcast sends data to all clients of a company.
func (h *Hub) Broadcast(companyID uint, data []byte) {
	h.broadcast <- broadcastMsg{companyID: companyID, data: data}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) { h.register <- c }

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

// WritePump sends queued messages to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump reads from the WebSocket to detect disconnection.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// NewClient creates a Client and registers it in the hub.
func NewClient(h *Hub, conn *websocket.Conn, companyID uint) *Client {
	return &Client{
		hub:       h,
		conn:      conn,
		send:      make(chan []byte, 256),
		CompanyID: companyID,
	}
}

// StartPGListener connects to PostgreSQL and listens for new message notifications,
// broadcasting them to the appropriate company room in the hub.
func StartPGListener(ctx context.Context, databaseURL string, h *Hub) {
	go func() {
		for {
			if err := listen(ctx, databaseURL, h); err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[ws] pg listener error: %v — reconnecting in 5s", err)
				select {
				case <-time.After(5 * time.Second):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
}

func listen(ctx context.Context, databaseURL string, h *Hub) error {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "LISTEN new_message"); err != nil {
		return err
	}

	log.Println("[ws] pg listener ready")

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}

		var payload notifyPayload
		if err := json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
			log.Printf("[ws] invalid notify payload: %v", err)
			continue
		}

		data, err := json.Marshal(payload)
		if err != nil {
			continue
		}

		envelope, err := json.Marshal(WsMessage{Type: "new_message", Data: data})
		if err != nil {
			continue
		}

		h.Broadcast(payload.CompanyID, envelope)
	}
}
