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
	"github.com/quanttide/qtcloud-asset/provider/internal/repository"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
)

func main() {
	cfg := config.Load()

	// Build the OSS adapter + bucket service (read-only discovery).
	// If credentials are not configured, the service stays nil and
	// /buckets responds 503 instead of failing to start.
	var bucketService *service.BucketService
	if cfg.OSSAccessKeyID != "" && cfg.OSSAccessKeySecret != "" {
		adapter := repository.NewOssAdapter(
			cfg.OSSEndpoint,
			cfg.OSSAccessKeyID,
			cfg.OSSAccessKeySecret,
			cfg.OSSSecurityToken,
		)
		bucketService = service.NewBucketService(adapter, adapter, adapter)
		log.Printf("OSS adapter configured (endpoint: %s)", cfg.OSSEndpoint)
	} else {
		log.Printf("OSS adapter NOT configured: OSS credentials missing")
	}

	handler := api.New(cfg, bucketService)

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
