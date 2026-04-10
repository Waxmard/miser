package api

import "net/http"

type accountResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Institution string `json:"institution"`
	AccountType string `json:"account_type"`
	Source      string `json:"source"`
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accounts, err := s.repo.Accounts().List(ctx)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	resp := make([]accountResponse, len(accounts))
	for i := range accounts {
		a := &accounts[i]
		resp[i] = accountResponse{
			ID:          a.ID,
			Name:        a.Name,
			Institution: a.Institution,
			AccountType: a.AccountType,
			Source:      a.Source,
		}
	}

	jsonOK(w, resp)
}
