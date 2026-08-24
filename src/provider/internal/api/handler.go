// Package api provides HTTP handlers for the provider API.
//
// API Layer: request routing, CORS, response formatting.
package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	cfg     *config.Config
	buckets *service.BucketService
}

// New creates a new Handler.
func New(cfg *config.Config, buckets *service.BucketService) *Handler {
	return &Handler{cfg: cfg, buckets: buckets}
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /", h.root)
	mux.HandleFunc("GET /config", h.config)
	mux.HandleFunc("GET /buckets", h.bucketsList)
	mux.HandleFunc("GET /buckets/{name}/objects", h.bucketObjectsList)
	mux.HandleFunc("GET /buckets/{name}/object-url", h.objectURL)
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
		StudioOrigin:    h.cfg.StudioOrigin,
		CORS:            "enabled",
	})
}

// bucketsList responds with the discovered OSS buckets (read-only).
// Query params: sort (name/created), order (asc/desc).
func (h *Handler) bucketsList(w http.ResponseWriter, r *http.Request) {
	if h.buckets == nil {
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}

	buckets, err := h.buckets.ListBuckets()
	if err != nil {
		log.Printf("list buckets error: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to list OSS buckets")
		return
	}

	sortBuckets(buckets, r.URL.Query().Get("sort"), r.URL.Query().Get("order"))

	respondJSON(w, http.StatusOK, schema.BucketListResponse{
		Buckets: buckets,
		Total:   len(buckets),
	})
}

// sortBuckets sorts buckets in memory by the given field and order.
func sortBuckets(buckets []schema.Bucket, sortKey, order string) {
	if sortKey == "" {
		return
	}
	desc := order == "desc"
	switch sortKey {
	case "name":
		sort.Slice(buckets, func(i, j int) bool {
			if desc {
				return buckets[i].Name > buckets[j].Name
			}
			return buckets[i].Name < buckets[j].Name
		})
	case "created":
		sort.Slice(buckets, func(i, j int) bool {
			if desc {
				return buckets[i].CreatedAt > buckets[j].CreatedAt
			}
			return buckets[i].CreatedAt < buckets[j].CreatedAt
		})
	}
}

// bucketObjectsList responds with objects inside a bucket (read-only).
// Query params: prefix, sort (key/size/date), order (asc/desc), limit, marker.
func (h *Handler) bucketObjectsList(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "bucket name is required")
		return
	}
	if h.buckets == nil {
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}

	q := r.URL.Query()
	params := schema.ListObjectsParams{
		Prefix: q.Get("prefix"),
		Sort:   q.Get("sort"),
		Order:  q.Get("order"),
		Marker: q.Get("marker"),
	}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			respondError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		params.Limit = n
	}

	objects, nextMarker, truncated, err := h.buckets.ListObjects(name, params)
	if err != nil {
		if errors.Is(err, service.ErrMetadataOnlyBucket) {
			respondError(w, http.StatusForbidden, "bucket object listing is disabled")
			return
		}
		log.Printf("list objects in %s error: %v", name, err)
		respondError(w, http.StatusInternalServerError, "failed to list objects in bucket")
		return
	}

	respondJSON(w, http.StatusOK, schema.ObjectListResponse{
		Bucket:     name,
		Objects:    objects,
		Total:      len(objects),
		NextMarker: nextMarker,
		Truncated:  truncated,
	})
}

// objectURL responds with an access URL for an object.
// Query params: key (object key), expires (seconds).
func (h *Handler) objectURL(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	key := r.URL.Query().Get("key")
	if name == "" || key == "" {
		respondError(w, http.StatusBadRequest, "bucket name and object key are required")
		return
	}
	if h.buckets == nil {
		respondError(w, http.StatusServiceUnavailable, "OSS bucket service is not configured")
		return
	}

	// expires 参数（秒），默认 86400（1天）；公开桶忽略此参数
	expiresIn := int64(86400)
	if v := r.URL.Query().Get("expires"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			respondError(w, http.StatusBadRequest, "expires must be a positive integer")
			return
		}
		expiresIn = n
	}

	url, err := h.buckets.ObjectURL(name, key, expiresIn)
	if err != nil {
		if errors.Is(err, service.ErrMetadataOnlyBucket) {
			respondError(w, http.StatusForbidden, "bucket object URLs are disabled")
			return
		}
		log.Printf("build url for %s/%s error: %v", name, key, err)
		respondError(w, http.StatusInternalServerError, "failed to build object url")
		return
	}

	respondJSON(w, http.StatusOK, schema.ObjectURLResponse{
		Bucket:    name,
		Key:       key,
		URL:       url,
		ExpiresIn: expiresIn,
	})
}
