package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
)

func TestRateLimiterRejectsExcessiveLoginStarts(t *testing.T) {
	cfg := &config.Config{
		BaseURL:      "https://api.quanttide.com/qtcloud-asset",
		StudioOrigin: "https://asset.cloud.quanttide.com",
	}
	sessions := auth.NewManager(auth.ManagerOptions{SessionTTL: time.Hour})
	identity := fakeIdentityProvider{user: auth.User{ID: "viewer-1", Role: auth.RoleViewer}}
	handler := NewWithStores(cfg, nil, sessions, identity, auth.NewMemoryUserStore(), auth.NewMemoryAuditLogStore())
	handler.rateLimiter = NewRateLimiter(2, time.Minute)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		req.RemoteAddr = "192.0.2.10:12345"
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		if res.Code != http.StatusFound {
			t.Fatalf("expected warmup login request %d HTTP 302, got %d", i+1, res.Code)
		}
	}

	limitedReq := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	limitedReq.RemoteAddr = "192.0.2.10:12345"
	limitedRes := httptest.NewRecorder()
	mux.ServeHTTP(limitedRes, limitedReq)

	if limitedRes.Code != http.StatusTooManyRequests {
		t.Fatalf("expected third login request HTTP 429, got %d", limitedRes.Code)
	}
	if got := limitedRes.Header().Get("Retry-After"); got == "" {
		t.Fatal("rate limited response should include Retry-After")
	}

	otherIPReq := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	otherIPReq.RemoteAddr = "198.51.100.20:12345"
	otherIPRes := httptest.NewRecorder()
	mux.ServeHTTP(otherIPRes, otherIPReq)

	if otherIPRes.Code != http.StatusFound {
		t.Fatalf("expected another client IP to remain allowed, got %d", otherIPRes.Code)
	}
}
