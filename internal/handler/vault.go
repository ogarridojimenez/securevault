package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ogarridojimenez/securevault/internal/middleware"
	"github.com/ogarridojimenez/securevault/internal/model"
	"github.com/ogarridojimenez/securevault/internal/service"
)

type VaultHandler struct {
	vaultService *service.VaultService
}

func NewVaultHandler(vaultService *service.VaultService) *VaultHandler {
	return &VaultHandler{vaultService: vaultService}
}

func (h *VaultHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	resp, err := h.vaultService.List(r.Context(), userID)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *VaultHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req model.CreateVaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	vault, err := h.vaultService.Create(r.Context(), userID, req.Name, req.Description)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, vault)
}

func (h *VaultHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")

	vault, err := h.vaultService.Get(r.Context(), userID, vaultID)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, vault)
}

func (h *VaultHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")

	var req model.UpdateVaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	vault, err := h.vaultService.Update(r.Context(), userID, vaultID, req.Name, req.Description)
	if err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, vault)
}

func (h *VaultHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	vaultID := chi.URLParam(r, "vaultID")

	if err := h.vaultService.Delete(r.Context(), userID, vaultID, false); err != nil {
		resolveError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
