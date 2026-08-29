// Command provider is the entry point for the QtCloud Asset Provider server.
//
// Usage:
//
//	export PROVIDER_PORT=9000
//	provider
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/api"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/repository"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
	"github.com/quanttide/qtcloud-asset/provider/internal/storage"
)

func main() {
	cfg := config.Load()
	if cfg.UserMigration != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := storage.ApplyUserMigration(ctx, cfg)
		cancel()
		if err != nil {
			log.Fatalf("user schema migration %q failed: %v", cfg.UserMigration, err)
		}
		log.Printf("user schema migration %q completed", cfg.UserMigration)
	}
	if cfg.ShareMigration != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := storage.ApplyShareMigration(ctx, cfg)
		cancel()
		if err != nil {
			log.Fatalf("share schema migration %q failed: %v", cfg.ShareMigration, err)
		}
		log.Printf("share schema migration %q completed", cfg.ShareMigration)
	}

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

	shareStore, closeShareStore, err := storage.OpenShareStore(cfg)
	if err != nil {
		log.Printf("share store setup failed: %v", err)
	}
	if closeShareStore != nil {
		defer func() {
			if err := closeShareStore(); err != nil {
				log.Printf("close share store: %v", err)
			}
		}()
	}
	if cfg.ShareStoreMode == "memory" {
		log.Printf("WARNING: in-memory shares enabled; links are not durable across restarts")
	} else if cfg.RDSConnectionString == "" {
		log.Printf("share persistence NOT configured: RDS_CONNECTION_STRING missing")
	}

	userStore, closeUserStore, userStoreErr := storage.OpenUserStore(cfg)
	if closeUserStore != nil {
		defer func() {
			if err := closeUserStore(); err != nil {
				log.Printf("close user store: %v", err)
			}
		}()
	}
	if userStoreErr != nil {
		if cfg.RDSConnectionString != "" && cfg.UserStoreMode != "memory" {
			log.Fatalf("user store setup failed: %v", userStoreErr)
		}
		log.Printf("user persistence NOT configured: %v", userStoreErr)
	} else if cfg.UserStoreMode == "memory" {
		log.Printf("WARNING: in-memory users enabled; user management is not durable across restarts")
	}

	handler := api.NewWithStoresAndShares(cfg, bucketService, nil, nil, userStore, nil, shareStore)

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
