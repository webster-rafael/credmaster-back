package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	appmiddleware "github.com/websterdev/cred-master/internal/middleware"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"success": false, "message": msg})
}

type createClientRequest struct {
	Nome     string `json:"nome"`
	CPF      string `json:"cpf"`
	Telefone string `json:"telefone"`
	Whatsapp string `json:"whatsapp"`
	Email    string `json:"email"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	clients, err := h.repo.FindAllByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch clients")
		return
	}

	if clients == nil {
		clients = []Client{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": clients})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	var body createClientRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Nome == "" || body.CPF == "" {
		writeError(w, http.StatusUnprocessableEntity, "nome e cpf são obrigatórios")
		return
	}

	client := &Client{
		CompanyID: companyID,
		Nome:      body.Nome,
		CPF:       body.CPF,
		Telefone:  body.Telefone,
		Whatsapp:  body.Whatsapp,
		Email:     body.Email,
	}

	if err := h.repo.Create(r.Context(), client); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": client})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid client id")
		return
	}

	if err := h.repo.Delete(r.Context(), uint(id), companyID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "client not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete client")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "cliente removido"})
}
