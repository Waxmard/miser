package api

import (
	"net/http"

	"github.com/Waxmard/miser/internal/process"
)

func (s *Server) handleTrends(w http.ResponseWriter, r *http.Request) {
	out, err := process.GetTrends(r.Context(), s.repo)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to compute trends")
		return
	}
	jsonOK(w, out)
}

func (s *Server) handleBudgets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	budgets, err := s.repo.Budgets().List(ctx)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list budgets")
		return
	}

	type budgetResponse struct {
		ID            string  `json:"id"`
		CategoryID    string  `json:"category_id"`
		CategoryName  string  `json:"category_name"`
		MonthlyAmount float64 `json:"monthly_amount"`
	}

	resp := make([]budgetResponse, len(budgets))
	for i := range budgets {
		b := &budgets[i]
		resp[i] = budgetResponse{
			ID:            b.ID,
			CategoryID:    b.CategoryID,
			CategoryName:  b.CategoryName,
			MonthlyAmount: b.MonthlyAmount,
		}
	}

	jsonOK(w, resp)
}
