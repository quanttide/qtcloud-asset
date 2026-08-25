package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSMiddlewareAllowsFormalAndCompatibilityStudioOrigins(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := CORSMiddleware([]string{
		"https://asset.cloud.quanttide.com",
		"https://asset.quanttide.com",
	}, next)

	for _, origin := range []string{
		"https://asset.cloud.quanttide.com",
		"https://asset.quanttide.com",
	} {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", origin)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		if got := res.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("expected CORS origin %q, got %q", origin, got)
		}
	}
}

func TestCORSMiddlewareRejectsUnregisteredOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := CORSMiddleware([]string{"https://asset.cloud.quanttide.com"}, next)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://unregistered.example.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected unregistered origin to be rejected, got %q", got)
	}
}

func TestCORSMiddlewarePreflightAllowsCredentialedAdminHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight should not call downstream handler")
	})
	handler := CORSMiddleware([]string{"https://asset.cloud.quanttide.com"}, next)
	req := httptest.NewRequest(http.MethodOptions, "/admin/users", nil)
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, X-CSRF-Token")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected preflight HTTP 204, got %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentialed CORS preflight, got %q", got)
	}
	allowedHeaders := res.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(allowedHeaders, "Content-Type") || !strings.Contains(allowedHeaders, "X-CSRF-Token") {
		t.Fatalf("expected admin headers to be allowed, got %q", allowedHeaders)
	}
}
