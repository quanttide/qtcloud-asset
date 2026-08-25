package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
)

type fakeIdentityProvider struct {
	user auth.User
}

func (p fakeIdentityProvider) LoginURL(state string) (string, error) {
	loginURL := url.URL{Scheme: "https", Host: "sso.example.com", Path: "/login"}
	query := loginURL.Query()
	query.Set("state", state)
	loginURL.RawQuery = query.Encode()
	return loginURL.String(), nil
}

func (p fakeIdentityProvider) Exchange(context.Context, string, string) (auth.User, error) {
	return p.user, nil
}

func newAuthTestMux(t *testing.T) *http.ServeMux {
	mux, _ := newAuthTestMuxWithAudit(t)
	return mux
}

func newAuthTestMuxWithAudit(t *testing.T) (*http.ServeMux, *auth.MemoryAuditLogStore) {
	t.Helper()

	cfg := &config.Config{
		StudioOrigin:  "https://asset.cloud.quanttide.com",
		StudioOrigins: []string{"https://asset.cloud.quanttide.com"},
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        auth.NewMemorySessionStore(),
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	provider := fakeIdentityProvider{user: auth.User{
		ID:         "user-1",
		ExternalID: "lark-user-1",
		Email:      "viewer@example.com",
		Name:       "Viewer User",
		Role:       auth.RoleViewer,
	}}
	users := auth.NewMemoryUserStore()
	auditLogs := auth.NewMemoryAuditLogStore()

	handler := NewWithStores(cfg, nil, sessions, provider, users, auditLogs)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux, auditLogs
}

func TestSensitiveRoutesRequireAuthentication(t *testing.T) {
	mux := newAuthTestMux(t)

	for _, target := range []string{
		"/buckets",
		"/buckets/qtcloud-asset-studio/objects",
		"/buckets/qtcloud-asset-studio/object-url?key=index.html",
	} {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		res := httptest.NewRecorder()

		mux.ServeHTTP(res, req)

		if res.Code != http.StatusUnauthorized {
			t.Fatalf("expected %s to require authentication, got HTTP %d", target, res.Code)
		}
	}
}

func TestAuthCallbackCreatesHttpOnlySessionCookieAndMeReturnsUser(t *testing.T) {
	mux := newAuthTestMux(t)

	loginRes := httptest.NewRecorder()
	mux.ServeHTTP(loginRes, httptest.NewRequest(http.MethodGet, "/auth/login", nil))
	if loginRes.Code != http.StatusFound {
		t.Fatalf("expected login redirect, got HTTP %d", loginRes.Code)
	}
	stateCookie := findCookie(t, loginRes.Result().Cookies(), auth.LoginStateCookieName)
	if !stateCookie.HttpOnly {
		t.Fatal("login state cookie must be HttpOnly")
	}
	state := callbackState(t, loginRes.Header().Get("Location"))

	callbackReq := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+url.QueryEscape(state)+"&code=ok", nil)
	callbackReq.AddCookie(stateCookie)
	callbackRes := httptest.NewRecorder()
	mux.ServeHTTP(callbackRes, callbackReq)
	if callbackRes.Code != http.StatusSeeOther {
		t.Fatalf("expected callback to redirect to Studio, got HTTP %d", callbackRes.Code)
	}
	sessionCookie := findCookie(t, callbackRes.Result().Cookies(), auth.SessionCookieName)
	if !sessionCookie.HttpOnly {
		t.Fatal("session cookie must be HttpOnly")
	}
	if !sessionCookie.Secure {
		t.Fatal("session cookie must be Secure")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meRes := httptest.NewRecorder()
	mux.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusOK {
		t.Fatalf("expected authenticated /auth/me, got HTTP %d", meRes.Code)
	}

	var body struct {
		User auth.User `json:"user"`
	}
	if err := json.NewDecoder(meRes.Body).Decode(&body); err != nil {
		t.Fatalf("decode /auth/me response: %v", err)
	}
	if body.User.Email != "viewer@example.com" || body.User.Role != auth.RoleViewer {
		t.Fatalf("unexpected authenticated user: %+v", body.User)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	mux := newAuthTestMux(t)
	sessionCookie := loginSessionCookie(t, mux)

	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutRes := httptest.NewRecorder()
	mux.ServeHTTP(logoutRes, logoutReq)
	if logoutRes.Code != http.StatusNoContent {
		t.Fatalf("expected logout HTTP 204, got %d", logoutRes.Code)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meRes := httptest.NewRecorder()
	mux.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked session to be unauthorized, got HTTP %d", meRes.Code)
	}
}

func TestAuthenticationFailureWritesAuditLog(t *testing.T) {
	mux, auditLogs := newAuthTestMuxWithAudit(t)
	req := httptest.NewRequest(http.MethodGet, "/buckets", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	req.Header.Set("User-Agent", "test-agent")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized response, got HTTP %d", res.Code)
	}
	entries := auditLogs.List()
	if len(entries) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(entries))
	}
	if entries[0].Action != auth.AuditActionAuthFailed || entries[0].Target != "/buckets" || entries[0].IP != "192.0.2.10" {
		t.Fatalf("unexpected auth failure audit entry: %+v", entries[0])
	}
}

func TestLogoutWritesAuditLog(t *testing.T) {
	mux, auditLogs := newAuthTestMuxWithAudit(t)
	sessionCookie := loginSessionCookie(t, mux)

	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutRes := httptest.NewRecorder()
	mux.ServeHTTP(logoutRes, logoutReq)

	if logoutRes.Code != http.StatusNoContent {
		t.Fatalf("expected logout HTTP 204, got %d", logoutRes.Code)
	}
	entries := auditLogs.List()
	found := false
	for _, entry := range entries {
		if entry.Action == auth.AuditActionLogout && entry.UserID == "user-1" && entry.Result == auth.AuditResultSuccess {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected logout audit entry, got %+v", entries)
	}
}

func TestCORSMiddlewareAllowsCredentialsForRegisteredOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := CORSMiddleware([]string{"https://asset.cloud.quanttide.com"}, next)
	req := httptest.NewRequest(http.MethodOptions, "/auth/me", nil)
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials to be allowed for registered origin, got %q", got)
	}
}

func loginSessionCookie(t *testing.T, mux *http.ServeMux) *http.Cookie {
	t.Helper()

	loginRes := httptest.NewRecorder()
	mux.ServeHTTP(loginRes, httptest.NewRequest(http.MethodGet, "/auth/login", nil))
	stateCookie := findCookie(t, loginRes.Result().Cookies(), auth.LoginStateCookieName)
	state := callbackState(t, loginRes.Header().Get("Location"))

	callbackReq := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+url.QueryEscape(state)+"&code=ok", nil)
	callbackReq.AddCookie(stateCookie)
	callbackRes := httptest.NewRecorder()
	mux.ServeHTTP(callbackRes, callbackReq)
	return findCookie(t, callbackRes.Result().Cookies(), auth.SessionCookieName)
}

func callbackState(t *testing.T, location string) string {
	t.Helper()

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse login redirect: %v", err)
	}
	state := parsed.Query().Get("state")
	if state == "" {
		t.Fatal("login redirect must include state")
	}
	return state
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q was not set", name)
	return nil
}
