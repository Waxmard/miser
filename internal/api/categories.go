package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type categoryResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	ParentID         *string `json:"parent_id"`
	Icon             *string `json:"icon,omitempty"`
	TransactionCount int     `json:"transaction_count"`
	TotalAmount      float64 `json:"total_amount"`
}

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	now := time.Now().UTC()
	year, month, _ := now.Date()
	from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	cats, err := s.repo.Categories().ListWithCounts(ctx, from, to)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	resp := make([]categoryResponse, len(cats))
	for i := range cats {
		c := &cats[i]
		resp[i] = categoryResponse{
			ID:               c.ID,
			Name:             c.Name,
			ParentID:         c.ParentID,
			Icon:             c.Icon,
			TransactionCount: c.TransactionCount,
			TotalAmount:      c.TotalAmount,
		}
	}

	jsonOK(w, resp)
}

func (s *Server) handleUpdateCategoryIcon(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	var body struct {
		Icon *string `json:"icon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := s.repo.Categories().GetByID(ctx, id)
	if err != nil {
		jsonError(w, http.StatusNotFound, "category not found")
		return
	}

	cat.Icon = body.Icon
	if err := s.repo.Categories().Update(ctx, cat); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to update category")
		return
	}

	jsonOK(w, map[string]any{"id": cat.ID, "icon": cat.Icon})
}
