// Package api provides HTTP handlers for the provider API.
//
// API Layer: request routing, CORS, response formatting.
package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	cfg *config.Config
}

// New creates a new Handler.
func New(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /", h.root)
	mux.HandleFunc("GET /config", h.config)
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, schema.ErrorResponse{Error: http.StatusText(status), Message: msg})
}

// CORSMiddleware wraps a handler with CORS headers.
func CORSMiddleware(origins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowed := range origins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// health responds with service health status.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, schema.HealthResponse{
		Status:  "ok",
		Service: "qtcloud-asset-provider",
	})
}

// root responds with service information.
func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, schema.RootResponse{
		Name:        "qtcloud-asset-provider",
		Description: "QtCloud Asset API provider",
		Status:      "ready",
	})
}

// config responds with provider configuration.
func (h *Handler) config(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, schema.ConfigResponse{
		ProviderBaseURL: h.cfg.BaseURL,
		StudioOrigin:    "https://asset.quanttide.com",
		CORS:            "enabled",
	})
}
