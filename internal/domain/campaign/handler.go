package campaign

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	domainboleto "github.com/websterdev/cred-master/internal/domain/boleto"
	domainclient "github.com/websterdev/cred-master/internal/domain/client"
	domainwhatsapp "github.com/websterdev/cred-master/internal/domain/whatsapp"
	appmiddleware "github.com/websterdev/cred-master/internal/middleware"
)

type Handler struct {
	repo       Repository
	clientRepo domainclient.Repository
	boletoRepo domainboleto.Repository
	whatsRepo  domainwhatsapp.Repository
	webhookURL string
	baseURL    string
}

func NewHandler(repo Repository, clientRepo domainclient.Repository, boletoRepo domainboleto.Repository, whatsRepo domainwhatsapp.Repository, webhookURL string, baseURL string) *Handler {
	return &Handler{repo: repo, clientRepo: clientRepo, boletoRepo: boletoRepo, whatsRepo: whatsRepo, webhookURL: webhookURL, baseURL: baseURL}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"success": false, "message": msg})
}

type createRequest struct {
	Name         string `json:"name"`
	Tipo         string `json:"tipo"`
	TemplateName string `json:"templateName"`
	Descricao    string `json:"descricao"`
}

type dispatchRequest struct {
	ClientIDs []uint `json:"clientIds"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())

	campaigns, err := h.repo.FindAllByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch campaigns")
		return
	}

	if campaigns == nil {
		campaigns = []Campaign{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": campaigns})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())

	var body createRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name é obrigatório")
		return
	}

	c := &Campaign{
		CompanyID:    companyID,
		Name:         body.Name,
		Tipo:         Tipo(body.Tipo),
		TemplateName: body.TemplateName,
		Descricao:    body.Descricao,
		Status:       StatusRascunho,
	}

	if err := h.repo.Create(r.Context(), c); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create campaign")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": c})
}

func (h *Handler) Dispatch(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	var body dispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(body.ClientIDs) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "clientIds não pode ser vazio")
		return
	}

	campaign, err := h.repo.FindByID(r.Context(), uint(id), companyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}

	clients, err := h.clientRepo.FindByIDs(r.Context(), body.ClientIDs, companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch clients")
		return
	}

	if h.webhookURL != "" {
		go func() {
			cpfs := make([]string, 0, len(clients))
			for _, c := range clients {
				if c.CPF != "" {
					cpfs = append(cpfs, c.CPF)
				}
			}

			var boletos []domainboleto.Boleto
			if len(cpfs) > 0 {
				boletos, _ = h.boletoRepo.FindByCPFs(context.Background(), cpfs, companyID)
			}

			boletosByCPF := make(map[string][]domainboleto.Boleto)
			for _, b := range boletos {
				boletosByCPF[b.ClienteCPF] = append(boletosByCPF[b.ClienteCPF], b)
			}

			whatsConfig, _ := h.whatsRepo.FindByCompanyID(context.Background(), companyID)
			wabaID := ""
			token := ""
			if whatsConfig != nil {
				wabaID = whatsConfig.WabaID
				token = whatsConfig.UserAccessToken
				if token == "" {
					token = whatsConfig.AccessToken
				}
			}

			var enrichedClients []map[string]any
			for _, c := range clients {
				clientMap := map[string]any{
					"id":        c.ID,
					"nome":      c.Nome,
					"cpf":       c.CPF,
					"telefone":  c.Telefone,
					"whatsapp":  c.Whatsapp,
					"email":     c.Email,
					"boletos":   []domainboleto.Boleto{},
				}
				if bs, ok := boletosByCPF[c.CPF]; ok {
					clientMap["boletos"] = bs
				}
				enrichedClients = append(enrichedClients, clientMap)
			}

			payload := map[string]any{
				"campaign": campaign,
				"clients":  enrichedClients,
				"waba_id":  wabaID,
				"token":    token,
				"base_url": h.baseURL,
			}
			b, _ := json.Marshal(payload)
			resp, werr := http.Post(h.webhookURL, "application/json", bytes.NewReader(b))
			if werr != nil {
				log.Printf("webhook dispatch error: %v", werr)
				return
			}
			defer resp.Body.Close()
			log.Printf("webhook dispatch response: %d", resp.StatusCode)
		}()
	}

	if err := h.repo.Dispatch(r.Context(), uint(id), companyID, len(clients)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to dispatch campaign")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "disparo iniciado"})
}
