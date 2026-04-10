package api

import (
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
