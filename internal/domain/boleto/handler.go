package boleto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appmiddleware "github.com/websterdev/cred-master/internal/middleware"
)

type Handler struct {
	repo    Repository
	baseURL string
}

func NewHandler(repo Repository, baseURL string) *Handler {
	return &Handler{repo: repo, baseURL: baseURL}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"success": false, "message": msg})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	boletos, err := h.repo.FindAllByCompany(r.Context(), companyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch boletos")
		return
	}
	if boletos == nil {
		boletos = []Boleto{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": boletos})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	b := Boleto{CompanyID: companyID, DisparoStatus: "nao_enviado"}

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse form")
			return
		}
		b.ClienteNome = r.FormValue("clienteNome")
		b.ClienteCPF = r.FormValue("clienteCPF")
		b.Valor, _ = strconv.ParseFloat(r.FormValue("valor"), 64)
		b.Vencimento = r.FormValue("vencimento")
		b.Status = r.FormValue("status")
		if b.Status == "" {
			b.Status = "pendente"
		}

		if fhs := r.MultipartForm.File["arquivo"]; len(fhs) > 0 {
			fh := fhs[0]
			src, err := fh.Open()
			if err == nil {
				defer src.Close()
				safeFile := filepath.Base(fh.Filename)
				dir := filepath.Join("uploads", "boletos", strconv.Itoa(int(companyID)))
				_ = os.MkdirAll(dir, 0755)
				storedName := fmt.Sprintf("%d_%s", companyID, safeFile)
				dst, err2 := os.Create(filepath.Join(dir, storedName))
				if err2 == nil {
					defer dst.Close()
					sz, _ := io.Copy(dst, src)
					b.Arquivo = fh.Filename
					b.Tamanho = sz
					b.ArquivoURL = fmt.Sprintf("%s/uploads/boletos/%d/%s", h.baseURL, companyID, storedName)
				}
			}
		}
	} else {
		var req struct {
			ClienteNome string  `json:"clienteNome"`
			ClienteCPF  string  `json:"clienteCPF"`
			Valor       float64 `json:"valor"`
			Vencimento  string  `json:"vencimento"`
			Status      string  `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		b.ClienteNome = req.ClienteNome
		b.ClienteCPF = req.ClienteCPF
		b.Valor = req.Valor
		b.Vencimento = req.Vencimento
		b.Status = req.Status
		if b.Status == "" {
			b.Status = "pendente"
		}
	}

	if err := h.repo.Create(r.Context(), &b); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create boleto")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": b})
}

func (h *Handler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	var reqs []struct {
		ClienteNome string  `json:"clienteNome"`
		ClienteCPF  string  `json:"clienteCPF"`
		Valor       float64 `json:"valor"`
		Vencimento  string  `json:"vencimento"`
		Arquivo     string  `json:"arquivo"`
		Tamanho     int64   `json:"tamanho"`
		Status      string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil || len(reqs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	boletos := make([]Boleto, len(reqs))
	for i, req := range reqs {
		status := req.Status
		if status == "" {
			status = "pendente"
		}
		boletos[i] = Boleto{
			CompanyID:     companyID,
			ClienteNome:   req.ClienteNome,
			ClienteCPF:    req.ClienteCPF,
			Valor:         req.Valor,
			Vencimento:    req.Vencimento,
			Arquivo:       req.Arquivo,
			Tamanho:       req.Tamanho,
			Status:        status,
			DisparoStatus: "nao_enviado",
		}
	}

	if err := h.repo.BulkCreate(r.Context(), boletos); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create boletos")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "count": len(boletos)})
}

func (h *Handler) DeleteBulk(w http.ResponseWriter, r *http.Request) {
	companyID := appmiddleware.GetCompanyID(r.Context())
	if companyID == 0 {
		writeError(w, http.StatusUnauthorized, "missing company context")
		return
	}

	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.repo.DeleteByIDs(r.Context(), req.IDs, companyID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete boletos")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
