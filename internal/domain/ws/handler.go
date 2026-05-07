package ws

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	"github.com/websterdev/cred-master/internal/hub"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub       *hub.Hub
	jwtSecret string
}

func NewHandler(h *hub.Hub, jwtSecret string) *Handler {
	return &Handler{hub: h, jwtSecret: jwtSecret}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	companyID := h.companyIDFromToken(r)
	if companyID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := hub.NewClient(h.hub, conn, companyID)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func (h *Handler) companyIDFromToken(r *http.Request) uint {
	raw := r.URL.Query().Get("token")
	if raw == "" {
		// fallback: Authorization header
		auth := r.Header.Get("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 {
			raw = parts[1]
		}
	}
	if raw == "" {
		return 0
	}

	token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0
	}

	if v, ok := claims["company_id"].(float64); ok {
		return uint(v)
	}
	return 0
}
