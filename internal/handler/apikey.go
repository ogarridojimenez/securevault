package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ogarridojimenez/securevault/internal/middleware"
	"github.com/ogarridojimenez/securevault/internal/model"
	"github.com/ogarridojimenez/securevault/internal/service"
)

type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req model.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.apiKeyService.Create(r.Context(), userID, req.Name)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	keys, err := h.apiKeyService.List(r.Context(), userID)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, keys)
}

func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	keyID := chi.URLParam(r, "keyID")

	if err := h.apiKeyService.Revoke(r.Context(), userID, keyID); err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "api key revoked"})
}
