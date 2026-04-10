package api

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"time"

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

// Handler returns the HTTP handler with logging and CORS middleware applied.
func (s *Server) Handler() http.Handler {
	return logMiddleware(corsMiddleware(s.mux))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Microsecond))
	})
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/transactions", s.handleListTransactions)
	s.mux.HandleFunc("GET /api/categories", s.handleListCategories)
	s.mux.HandleFunc("PATCH /api/categories/{id}", s.handleUpdateCategoryIcon)
	s.mux.HandleFunc("GET /api/trends", s.handleTrends)
	s.mux.HandleFunc("GET /api/budgets", s.handleBudgets)
	s.mux.HandleFunc("GET /api/accounts", s.handleListAccounts)
	s.mux.HandleFunc("GET /api/reports/latest", s.handleLatestReport)
	s.mux.HandleFunc("GET /api/merchant-icons", s.handleListMerchantIcons)
	s.mux.HandleFunc("PUT /api/merchant-icons", s.handleSetMerchantIcon)
	s.mux.HandleFunc("DELETE /api/merchant-icons/{name}", s.handleDeleteMerchantIcon)

	if s.static != nil {
		s.mux.Handle("/", http.FileServerFS(s.static))
	}
}

// corsMiddleware allows requests from the Svelte dev server on :5173.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, PUT, DELETE, OPTIONS")
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
