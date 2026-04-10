package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

type transactionResponse struct {
	ID            string  `json:"id"`
	AccountID     string  `json:"account_id"`
	AccountName   string  `json:"account_name"`
	CategoryID    *string `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	Amount        float64 `json:"amount"`
	Merchant      string  `json:"merchant"`
	MerchantClean *string `json:"merchant_clean,omitempty"`
	Description   *string `json:"description,omitempty"`
	Date          string  `json:"date"`
	Source        string  `json:"source"`
	Status        string  `json:"status"`
	Tags          *string `json:"tags,omitempty"`
	Owner         *string `json:"owner,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

func (s *Server) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	filters := &repository.TransactionFilters{Limit: 50}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filters.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filters.Offset = n
		}
	}
	if v := q.Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "invalid from date, use YYYY-MM-DD")
			return
		}
		filters.From = &t
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "invalid to date, use YYYY-MM-DD")
			return
		}
		filters.To = &t
	}
	if v := q.Get("category"); v != "" {
		cat, err := s.repo.Categories().GetByName(ctx, v)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "category not found")
			return
		}
		filters.CategoryID = &cat.ID
	}
	if v := q.Get("account"); v != "" {
		acct, err := s.repo.Accounts().GetByName(ctx, v)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "account not found")
			return
		}
		filters.AccountID = &acct.ID
	}
	if v := q.Get("tag"); v != "" {
		filters.Tag = &v
	}
	if v := q.Get("owner"); v != "" {
		filters.Owner = &v
	}
	if v := q.Get("q"); v != "" {
		filters.Merchant = &v
	}

	txns, err := s.repo.Transactions().List(ctx, filters)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	resp := make([]transactionResponse, len(txns))
	for i := range txns {
		t := &txns[i]
		resp[i] = transactionResponse{
			ID:            t.ID,
			AccountID:     t.AccountID,
			AccountName:   t.AccountName,
			CategoryID:    t.CategoryID,
			CategoryName:  t.CategoryName,
			Amount:        t.Amount,
			Merchant:      t.Merchant,
			MerchantClean: t.MerchantClean,
			Description:   t.Description,
			Date:          t.Date.Format("2006-01-02"),
			Source:        t.Source,
			Status:        t.Status,
			Tags:          t.Tags,
			Owner:         t.Owner,
			Notes:         t.Notes,
		}
	}

	jsonOK(w, resp)
}
