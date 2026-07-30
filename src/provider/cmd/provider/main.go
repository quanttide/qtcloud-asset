// Command provider is the entry point for the QtCloud Asset Provider server.
//
// Usage:
//
//	export PROVIDER_PORT=9000
//	provider
package main

import (
	"log"
	"net/http"

	"github.com/quanttide/qtcloud-asset/provider/internal/api"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
)

func main() {
	cfg := config.Load()

	handler := api.New(cfg)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Wrap with CORS middleware
	srv := api.CORSMiddleware(cfg.StudioOrigins, mux)

	addr := ":" + cfg.Port
	log.Printf("qtcloud-asset-provider starting on %s", addr)
	log.Printf("base URL: %s", cfg.BaseURL)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
