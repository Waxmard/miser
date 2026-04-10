package api

import (
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/Waxmard/miser/internal/repository"
)

// Server holds the HTTP mux and repository handle.
type Server struct {
	repo   repository.Repository
	mux    *http.ServeMux
	static fs.FS // nil means no static file serving
}

// New creates a Server with all API routes registered.
// Pass a non-nil static FS to serve frontend assets at non-/api routes (production).
func New(repo repository.Repository, static fs.FS) *Server {
	s := &Server{repo: repo, static: static}
	s.mux = http.NewServeMux()
	s.registerRoutes()
	return s
}

// Handler returns the HTTP handler with CORS middleware applied.
func (s *Server) Handler() http.Handler {
	return corsMiddleware(s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/transactions", s.handleListTransactions)
	s.mux.HandleFunc("GET /api/categories", s.handleListCategories)
	s.mux.HandleFunc("GET /api/trends", s.handleTrends)
	s.mux.HandleFunc("GET /api/budgets", s.handleBudgets)
	s.mux.HandleFunc("GET /api/accounts", s.handleListAccounts)
	s.mux.HandleFunc("GET /api/reports/latest", s.handleLatestReport)

	if s.static != nil {
		s.mux.Handle("/", http.FileServerFS(s.static))
	}
}

// corsMiddleware allows requests from the Svelte dev server on :5173.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
