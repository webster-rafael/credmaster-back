package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"message": message,
	})
}

type userView struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type loginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequestBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordRequestBody struct {
	Email string `json:"email"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, pair, err := h.service.Login(r.Context(), Credentials{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "login successful",
		"data": map[string]any{
			"user":   userView{ID: user.ID, Name: user.Name, Email: user.Email},
			"tokens": pair,
		},
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user := &User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.service.Register(r.Context(), user); err != nil {
		if errors.Is(err, ErrEmailAlreadyRegistered) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "account created",
		"data": map[string]any{
			"user": userView{ID: user.ID, Name: user.Name, Email: user.Email},
		},
	})
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "se o email existir, as instruções foram enviadas",
		"data": map[string]any{
			"resetToken": result.ResetToken,
		},
	})
}
