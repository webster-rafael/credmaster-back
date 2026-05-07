package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	appmiddleware "github.com/websterdev/cred-master/internal/middleware"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// ─── Meta API structs ────────────────────────────────────────────────────────

type metaComponent struct {
	Type   string `json:"type"`
	Format string `json:"format"`
	Text   string `json:"text"`
}

type metaTemplate struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Status     string          `json:"status"`
	Category   string          `json:"category"`
	Language   string          `json:"language"`
	Components []metaComponent `json:"components"`
}

type metaTemplatesResp struct {
	Data []metaTemplate `json:"data"`
}

// ─── Output struct sent to frontend ─────────────────────────────────────────

type TemplateOut struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Category string `json:"category"`
	Language string `json:"language"`
	Header   string `json:"header,omitempty"`
	Body     string `json:"body"`
	Footer   string `json:"footer,omitempty"`
}

// ─── Handler ─────────────────────────────────────────────────────────────────

func (h *Handler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())

	cfg, err := h.repo.FindByCompanyID(r.Context(), companyID)
	if err != nil || cfg.UserAccessToken == "" {
		http.Error(w, `{"error":"whatsapp not configured"}`, http.StatusBadRequest)
		return
	}

	wabaID := cfg.WabaID
	if wabaID == "" {
		wabaID, err = discoverWabaID(cfg.UserAccessToken)
		if err != nil {
			http.Error(w, `{"error":"could not discover waba id: `+err.Error()+`"}`, http.StatusBadGateway)
			return
		}
		_ = h.repo.UpdateWabaID(r.Context(), cfg.ID, wabaID)
	}

	url := fmt.Sprintf(
		"https://graph.facebook.com/v22.0/%s/message_templates?access_token=%s&limit=100",
		wabaID, cfg.UserAccessToken,
	)

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		http.Error(w, `{"error":"meta api unreachable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var metaResp metaTemplatesResp
	if err := json.Unmarshal(body, &metaResp); err != nil {
		http.Error(w, `{"error":"invalid meta response"}`, http.StatusBadGateway)
		return
	}

	out := make([]TemplateOut, 0, len(metaResp.Data))
	for _, t := range metaResp.Data {
		item := TemplateOut{
			ID:       t.ID,
			Name:     t.Name,
			Status:   strings.ToLower(t.Status),
			Category: strings.ToLower(t.Category),
			Language: t.Language,
		}
		for _, c := range t.Components {
			switch c.Type {
			case "HEADER":
				if c.Format == "TEXT" {
					item.Header = c.Text
				}
			case "BODY":
				item.Body = c.Text
			case "FOOTER":
				item.Footer = c.Text
			}
		}
		out = append(out, item)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())

	cfg, err := h.repo.FindByCompanyID(r.Context(), companyID)
	if err != nil || cfg.UserAccessToken == "" {
		http.Error(w, `{"error":"whatsapp not configured"}`, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/message_templates", cfg.WabaID)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.UserAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"meta api unreachable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

// ─── WABA discovery ──────────────────────────────────────────────────────────

func discoverWabaID(token string) (string, error) {
	type wabaEntry struct {
		ID string `json:"id"`
	}
	type wabaList struct {
		Data []wabaEntry `json:"data"`
	}
	type bizEntry struct {
		OwnedWABA wabaList `json:"owned_whatsapp_business_accounts"`
	}
	type bizList struct {
		Data []bizEntry `json:"data"`
	}
	type meResp struct {
		Businesses bizList `json:"businesses"`
	}

	url := fmt.Sprintf(
		"https://graph.facebook.com/v22.0/me?fields=businesses{owned_whatsapp_business_accounts}&access_token=%s",
		token,
	)

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var me meResp
	if err := json.Unmarshal(body, &me); err != nil {
		return "", err
	}

	for _, biz := range me.Businesses.Data {
		for _, waba := range biz.OwnedWABA.Data {
			if waba.ID != "" {
				return waba.ID, nil
			}
		}
	}

	return "", fmt.Errorf("no waba account found for this token")
}
