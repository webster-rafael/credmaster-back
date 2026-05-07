package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	domainclient "github.com/websterdev/cred-master/internal/domain/client"
	domainwhatsapp "github.com/websterdev/cred-master/internal/domain/whatsapp"
	appmiddleware "github.com/websterdev/cred-master/internal/middleware"
)

type Handler struct {
	repo           Repository
	whatsRepo      domainwhatsapp.Repository
	clientRepo     domainclient.Repository
	webhookSendURL string
	baseURL        string
}

func NewHandler(repo Repository, whatsRepo domainwhatsapp.Repository, clientRepo domainclient.Repository, webhookSendURL, baseURL string) *Handler {
	return &Handler{repo: repo, whatsRepo: whatsRepo, clientRepo: clientRepo, webhookSendURL: webhookSendURL, baseURL: baseURL}
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

// DeleteConversation removes all messages for a contact (by from_user_id).
func (h *Handler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
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

	if err := h.repo.DeleteByWaID(r.Context(), waID, companyID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete conversation")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// SendDocument receives a file, encodes it as base64 and forwards to the n8n webhook.
func (h *Handler) SendDocument(w http.ResponseWriter, r *http.Request) {
	if h.webhookSendURL == "" {
		writeError(w, http.StatusServiceUnavailable, "webhook not configured")
		return
	}

	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse form (max 20MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	to := r.FormValue("to")
	caption := r.FormValue("caption")
	if to == "" {
		writeError(w, http.StatusBadRequest, "to field is required")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	// Save file to disk and generate a public URL
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create uploads dir")
		return
	}
	ext := filepath.Ext(header.Filename)
	storedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	storedPath := filepath.Join(uploadsDir, storedName)
	if err := os.WriteFile(storedPath, fileBytes, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	mediaURL := fmt.Sprintf("%s/uploads/%s", h.baseURL, storedName)

	whatsConfig, err := h.whatsRepo.FindByCompanyID(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "whatsapp config not found")
		return
	}

	resolvedTo, clientID, fromUserID, contactName := h.resolveClient(r, to, companyID)

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                resolvedTo,
		"type":              "document",
		"document": map[string]any{
			"data":      fileBytes,
			"filename":  header.Filename,
			"mime_type": header.Header.Get("Content-Type"),
			"caption":   caption,
			"media_url": mediaURL,
		},
		"phone_number_id": whatsConfig.PhoneNumberID,
		"from":            whatsConfig.DisplayPhoneNumber,
		"company_id":      companyID,
		"client_id":       clientID,
		"from_user_id":    fromUserID,
		"contact_name":    contactName,
		"timestamp":       time.Now().Unix(),
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

func (h *Handler) resolveClient(r *http.Request, to string, companyID uint) (resolvedTo string, clientID *uint, fromUserID string, contactName string) {
	resolvedTo = to
	if client, err := h.clientRepo.FindByWaID(r.Context(), to, companyID); err == nil {
		clientID = &client.ID
		fromUserID = client.ContactUserID
		contactName = client.Nome
		if client.WaID != "" {
			resolvedTo = client.WaID
		} else if client.Whatsapp != "" {
			resolvedTo = client.Whatsapp
		}
	}
	return
}

// Send forwards a text or document message to the n8n webhook.
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

	to, clientID, fromUserID, contactName := h.resolveClient(r, req.To, companyID)

	payload := whatsappTextPayload{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text:             whatsappTextBody{PreviewURL: false, Body: req.Body},
		PhoneNumberID:    whatsConfig.PhoneNumberID,
		From:             whatsConfig.DisplayPhoneNumber,
		CompanyID:        companyID,
		ClientID:         clientID,
		FromUserID:       fromUserID,
		ContactName:      contactName,
		Timestamp:        time.Now().Unix(),
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
