package message

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	domainclient "github.com/websterdev/cred-master/internal/domain/client"
	domainwhatsapp "github.com/websterdev/cred-master/internal/domain/whatsapp"
	appmiddleware "github.com/websterdev/cred-master/internal/middleware"
)

type Handler struct {
	repo           Repository
	whatsRepo      domainwhatsapp.Repository
	clientRepo     domainclient.Repository
	webhookSendURL string
}

func NewHandler(repo Repository, whatsRepo domainwhatsapp.Repository, clientRepo domainclient.Repository, webhookSendURL string) *Handler {
	return &Handler{repo: repo, whatsRepo: whatsRepo, clientRepo: clientRepo, webhookSendURL: webhookSendURL}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"success": false, "message": msg})
}

// GetConversations returns the latest message per contact for the authenticated company.
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	conversations, err := h.repo.GetConversations(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch conversations")
		return
	}

	if conversations == nil {
		conversations = []ConversationSummary{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    conversations,
	})
}

// GetThread returns all messages for a specific contact (by from_user_id).
func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	waID := r.URL.Query().Get("wa_id")
	if waID == "" {
		writeError(w, http.StatusBadRequest, "wa_id query param is required")
		return
	}

	messages, err := h.repo.FindByWaID(r.Context(), waID, companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch thread")
		return
	}

	if messages == nil {
		messages = []Message{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    messages,
	})
}

type sendMessageRequest struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

type whatsappTextPayload struct {
	MessagingProduct string           `json:"messaging_product"`
	RecipientType    string           `json:"recipient_type"`
	To               string           `json:"to"`
	Type             string           `json:"type"`
	Text             whatsappTextBody `json:"text"`
	PhoneNumberID    string           `json:"phone_number_id"`
	From             string           `json:"from"`
	CompanyID        uint             `json:"company_id"`
	ClientID         *uint            `json:"client_id"`
	FromUserID       string           `json:"from_user_id"`
	ContactName      string           `json:"contact_name"`
	Timestamp        int64            `json:"timestamp"`
}

type whatsappTextBody struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

// Send forwards a text message to the n8n webhook which delivers it via WhatsApp.
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	if h.webhookSendURL == "" {
		writeError(w, http.StatusServiceUnavailable, "webhook not configured")
		return
	}

	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.To == "" || req.Body == "" {
		writeError(w, http.StatusUnprocessableEntity, "to e body são obrigatórios")
		return
	}

	whatsConfig, err := h.whatsRepo.FindByCompanyID(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "whatsapp config not found for company")
		return
	}

	var clientID *uint
	var fromUserID, contactName string
	to := req.To
	if client, err := h.clientRepo.FindByWaID(r.Context(), req.To, companyID); err == nil {
		clientID = &client.ID
		fromUserID = client.ContactUserID
		contactName = client.Nome
		if client.WaID != "" {
			to = client.WaID
		} else if client.Whatsapp != "" {
			to = client.Whatsapp
		}
	}

	payload := whatsappTextPayload{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text: whatsappTextBody{
			PreviewURL: false,
			Body:       req.Body,
		},
		PhoneNumberID: whatsConfig.PhoneNumberID,
		From:          whatsConfig.DisplayPhoneNumber,
		CompanyID:     companyID,
		ClientID:      clientID,
		FromUserID:    fromUserID,
		ContactName:   contactName,
		Timestamp:     time.Now().Unix(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode payload")
		return
	}

	resp, err := http.Post(h.webhookSendURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to reach webhook")
		return
	}
	defer resp.Body.Close()

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
