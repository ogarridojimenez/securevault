package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ogarridojimenez/securevault/internal/middleware"
	"github.com/ogarridojimenez/securevault/internal/model"
	"github.com/ogarridojimenez/securevault/internal/service"
)

type SecretHandler struct {
	secretService *service.SecretService
}

func NewSecretHandler(secretService *service.SecretService) *SecretHandler {
	return &SecretHandler{secretService: secretService}
}

func (h *SecretHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")

	resp, err := h.secretService.List(r.Context(), userID, vaultID)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SecretHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")

	var req model.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	secret, err := h.secretService.Create(r.Context(), userID, vaultID, req.Name, req.Value, req.Metadata)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, secret)
}

func (h *SecretHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")
	secretID := chi.URLParam(r, "secretID")

	secret, err := h.secretService.Get(r.Context(), userID, vaultID, secretID)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, secret)
}

func (h *SecretHandler) GetByName(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")
	name := chi.URLParam(r, "name")

	secret, err := h.secretService.GetByName(r.Context(), userID, vaultID, name)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, secret)
}

func (h *SecretHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")
	secretID := chi.URLParam(r, "secretID")

	if err := h.secretService.Delete(r.Context(), userID, vaultID, secretID); err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
