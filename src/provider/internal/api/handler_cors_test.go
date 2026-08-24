package api

import (
	"net/http"
	"net/http/httptest"
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
