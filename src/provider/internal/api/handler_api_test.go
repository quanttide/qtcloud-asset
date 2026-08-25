package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

func TestPublicEndpointsReturnExpectedResponses(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.quanttide.com/qtcloud-asset", StudioOrigin: "https://asset.cloud.quanttide.com"}
	handler := New(cfg, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	for _, tc := range []struct {
		path string
		want int
	}{
		{path: "/health", want: http.StatusOK},
		{path: "/", want: http.StatusOK},
		{path: "/config", want: http.StatusOK},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)

		if res.Code != tc.want {
			t.Fatalf("expected %s HTTP %d, got %d", tc.path, tc.want, res.Code)
		}
	}
}

func TestConfigEndpointDoesNotExposeSecrets(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.quanttide.com/qtcloud-asset", StudioOrigin: "https://asset.cloud.quanttide.com"}
	handler := New(cfg, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	var body schema.ConfigResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if body.ProviderBaseURL != cfg.BaseURL || body.StudioOrigin != cfg.StudioOrigin || body.CORS != "enabled" {
		t.Fatalf("unexpected config response: %+v", body)
	}
}

func TestAuthLoginReturnsServiceUnavailableWhenIdentityProviderMissing(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://api.quanttide.com/qtcloud-asset", StudioOrigin: "https://asset.cloud.quanttide.com"}
	handler := New(cfg, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected missing identity provider HTTP 503, got %d", res.Code)
	}
}

func TestAuthenticatedBucketsListSucceeds(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected authenticated bucket list HTTP 200, got %d", res.Code)
	}
}

func TestObjectListRejectsInvalidLimit(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtcloud-asset-studio/objects?limit=bad", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid limit HTTP 400, got %d", res.Code)
	}
}

func TestObjectURLRejectsInvalidExpiry(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtcloud-asset-studio/object-url?key=index.html&expires=bad", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid expiry HTTP 400, got %d", res.Code)
	}
}

func TestClientIPUsesFirstForwardedAddress(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/buckets", nil)
	req.RemoteAddr = "198.51.100.20:12345"
	req.Header.Set("X-Forwarded-For", "192.0.2.10, 198.51.100.20")

	if got := clientIP(req); got != "192.0.2.10" {
		t.Fatalf("expected first forwarded IP, got %q", got)
	}
}
