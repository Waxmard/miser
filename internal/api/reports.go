package api

import (
	"database/sql"
	"errors"
	"net/http"
)

type reportResponse struct {
	ID        string `json:"id"`
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Narrative string `json:"narrative"`
}

func (s *Server) handleLatestReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	report, err := s.repo.Reports().GetLatest(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonOK(w, nil)
			return
		}
		jsonError(w, http.StatusInternalServerError, "failed to get report")
		return
	}

	jsonOK(w, reportResponse{
		ID:        report.ID,
		Year:      report.Year,
		Month:     report.Month,
		Narrative: report.Narrative,
	})
}
