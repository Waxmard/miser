package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

type merchantIconResponse struct {
	MerchantName string `json:"merchant_name"`
	IconSlug     string `json:"icon_slug"`
	UpdatedAt    string `json:"updated_at"`
}

func toMerchantIconResponse(m *repository.MerchantIcon) merchantIconResponse {
	return merchantIconResponse{
		MerchantName: m.MerchantName,
		IconSlug:     m.IconSlug,
		UpdatedAt:    m.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *Server) handleListMerchantIcons(w http.ResponseWriter, r *http.Request) {
	icons, err := s.repo.MerchantIcons().List(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list merchant icons")
		return
	}
	resp := make([]merchantIconResponse, len(icons))
	for i := range icons {
		resp[i] = toMerchantIconResponse(&icons[i])
	}
	jsonOK(w, resp)
}

func (s *Server) handleSetMerchantIcon(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MerchantName string `json:"merchant_name"`
		IconSlug     string `json:"icon_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.MerchantName == "" || body.IconSlug == "" {
		jsonError(w, http.StatusBadRequest, "merchant_name and icon_slug are required")
		return
	}

	m := &repository.MerchantIcon{
		MerchantName: body.MerchantName,
		IconSlug:     body.IconSlug,
	}
	if err := s.repo.MerchantIcons().Set(r.Context(), m); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to save merchant icon")
		return
	}
	jsonOK(w, toMerchantIconResponse(m))
}

func (s *Server) handleDeleteMerchantIcon(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.repo.MerchantIcons().Delete(r.Context(), name); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to delete merchant icon")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
