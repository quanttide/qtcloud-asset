package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/storage"
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

	for _, testCase := range []struct {
		method string
		target string
		body   string
	}{
		{method: http.MethodPost, target: "/shares", body: `{}`},
		{method: http.MethodDelete, target: "/shares/token"},
	} {
		req := httptest.NewRequest(testCase.method, testCase.target, strings.NewReader(testCase.body))
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthenticated %s %s request to return HTTP 401, got %d", testCase.method, testCase.target, res.Code)
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

func TestLocalPasswordLoginCreatesHttpOnlySessionCookieAndMeReturnsUser(t *testing.T) {
	mux := newLocalPasswordAuthTestMux(t)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"correct-password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	loginRes := httptest.NewRecorder()
	mux.ServeHTTP(loginRes, loginReq)

	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected local password login HTTP 200, got %d: %s", loginRes.Code, loginRes.Body.String())
	}
	sessionCookie := findCookie(t, loginRes.Result().Cookies(), auth.SessionCookieName)
	if !sessionCookie.HttpOnly || !sessionCookie.Secure {
		t.Fatalf("local login session cookie must be secure HttpOnly, got %+v", sessionCookie)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meRes := httptest.NewRecorder()
	mux.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusOK {
		t.Fatalf("expected authenticated /auth/me after local login, got HTTP %d", meRes.Code)
	}

	var body struct {
		User auth.User `json:"user"`
	}
	if err := json.NewDecoder(meRes.Body).Decode(&body); err != nil {
		t.Fatalf("decode /auth/me response: %v", err)
	}
	if body.User.Email != "admin@example.com" || body.User.Role != auth.RoleAdmin {
		t.Fatalf("unexpected local login user: %+v", body.User)
	}
}

func TestLocalPasswordLoginRejectsBadPassword(t *testing.T) {
	mux := newLocalPasswordAuthTestMux(t)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"wrong-password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	loginRes := httptest.NewRecorder()
	mux.ServeHTTP(loginRes, loginReq)

	if loginRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid local login HTTP 401, got %d", loginRes.Code)
	}
	for _, cookie := range loginRes.Result().Cookies() {
		if cookie.Name == auth.SessionCookieName {
			t.Fatalf("invalid local login must not set session cookie: %+v", cookie)
		}
	}
}

func TestBuiltInLocalAccountCanLoginWhenRDSIsNotConfigured(t *testing.T) {
	cfg := &config.Config{
		StudioOrigin:  "https://asset.cloud.quanttide.com",
		StudioOrigins: []string{"https://asset.cloud.quanttide.com"},
		UserStoreMode: "rds",
	}
	users, closeStore, err := storage.OpenUserStore(cfg)
	if closeStore != nil {
		t.Cleanup(func() {
			if err := closeStore(); err != nil {
				t.Fatalf("close user store: %v", err)
			}
		})
	}
	if err == nil {
		t.Fatal("expected missing RDS configuration to be reported")
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        auth.NewMemorySessionStore(),
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	handler := NewWithStores(cfg, nil, sessions, fakeIdentityProvider{}, users, auth.NewMemoryAuditLogStore())
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	for _, account := range []string{"lixiang", "zhangguo", "liujingyi", "zhaoziyi", "tuyafang"} {
		loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"account":"`+account+`","password":"123456"}`))
		loginReq.Header.Set("Content-Type", "application/json")
		loginReq.Header.Set("Origin", cfg.StudioOrigin)
		loginRes := httptest.NewRecorder()
		mux.ServeHTTP(loginRes, loginReq)

		if loginRes.Code != http.StatusOK {
			t.Fatalf("expected built-in account %q login HTTP 200, got %d: %s", account, loginRes.Code, loginRes.Body.String())
		}
		responseBody := loginRes.Body.String()
		var body struct {
			User auth.User `json:"user"`
		}
		if err := json.Unmarshal([]byte(responseBody), &body); err != nil {
			t.Fatalf("decode %q login response: %v", account, err)
		}
		if body.User.Account != account || body.User.Name != account || body.User.Role != auth.RoleViewer {
			t.Fatalf("unexpected built-in account user %q: %+v", account, body.User)
		}
		if strings.Contains(responseBody, "123456") || strings.Contains(responseBody, "password_hash") {
			t.Fatalf("login response for %q must not expose password material: %s", account, responseBody)
		}
	}
}

func TestLocalPasswordLoginHasTightAttemptRateLimit(t *testing.T) {
	mux, handler := newLocalPasswordAuthTestMuxWithHandler(t)
	handler.localLoginRateLimiter = NewRateLimiter(2, time.Minute)

	for i := 0; i < 2; i++ {
		loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"wrong-password"}`))
		loginReq.Header.Set("Content-Type", "application/json")
		loginReq.RemoteAddr = "192.0.2.10:12345"
		loginRes := httptest.NewRecorder()
		mux.ServeHTTP(loginRes, loginReq)

		if loginRes.Code != http.StatusUnauthorized {
			t.Fatalf("expected bad password attempt %d HTTP 401, got %d", i+1, loginRes.Code)
		}
	}

	limitedReq := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"wrong-password"}`))
	limitedReq.Header.Set("Content-Type", "application/json")
	limitedReq.RemoteAddr = "192.0.2.10:12345"
	limitedRes := httptest.NewRecorder()
	mux.ServeHTTP(limitedRes, limitedReq)

	if limitedRes.Code != http.StatusTooManyRequests {
		t.Fatalf("expected third bad password attempt HTTP 429, got %d", limitedRes.Code)
	}
	if got := limitedRes.Header().Get("Retry-After"); got == "" {
		t.Fatal("local login rate limit response should include Retry-After")
	}
}

func TestProductionSessionCookiesAllowCrossSiteStudioRequests(t *testing.T) {
	cfg := &config.Config{
		BaseURL:      "https://api.quanttide.com/qtcloud-asset",
		StudioOrigin: "https://asset.cloud.quanttide.com",
	}
	handler := New(cfg, nil)

	cookie, err := handler.sessions.CreateSession(auth.User{
		ID:    "admin-1",
		Email: "haoziteng@quanttide.com",
		Role:  auth.RoleAdmin,
	}, time.Now())
	if err != nil {
		t.Fatalf("create production session: %v", err)
	}
	if cookie.SameSite != http.SameSiteNoneMode || !cookie.Secure || !cookie.HttpOnly {
		t.Fatalf("production session cookie must be SameSite=None; Secure; HttpOnly, got %+v", cookie)
	}
}

func TestStorelessHandlerUsesStableSignedSessionsAcrossInstances(t *testing.T) {
	passwordHash, err := auth.HashPasswordPBKDF2("correct-password", 1000)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	cfg := &config.Config{
		BaseURL:               "https://api.quanttide.com/qtcloud-asset",
		StudioOrigin:          "https://asset.cloud.quanttide.com",
		StudioOrigins:         []string{"https://asset.cloud.quanttide.com"},
		AuthMode:              "local",
		LocalAuthAccount:      "admin",
		LocalAuthEmail:        "admin@example.com",
		LocalAuthName:         "Admin User",
		LocalAuthRole:         "admin",
		LocalAuthPasswordHash: passwordHash,
	}

	loginHandler := NewWithStoresAndShares(cfg, nil, nil, nil, nil, nil, nil)
	loginMux := http.NewServeMux()
	loginHandler.RegisterRoutes(loginMux)

	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"account":"admin","password":"correct-password"}`),
	)
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("Origin", cfg.StudioOrigin)
	loginRes := httptest.NewRecorder()
	loginMux.ServeHTTP(loginRes, loginReq)
	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected local login HTTP 200, got %d: %s", loginRes.Code, loginRes.Body.String())
	}
	sessionCookie := findCookie(t, loginRes.Result().Cookies(), auth.SessionCookieName)
	if !sessionCookie.Secure || sessionCookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("production session cookie must be Secure and SameSite=None, got %+v", sessionCookie)
	}

	otherInstance := NewWithStoresAndShares(cfg, nil, nil, nil, nil, nil, nil)
	otherMux := http.NewServeMux()
	otherInstance.RegisterRoutes(otherMux)
	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meRes := httptest.NewRecorder()
	otherMux.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusOK {
		t.Fatalf("signed session should authenticate on another handler instance, got %d: %s", meRes.Code, meRes.Body.String())
	}
}

type emptyDurableUserStore struct {
	*auth.MemoryUserStore
}

func (s emptyDurableUserStore) ListWithError() ([]auth.User, error) {
	return s.List(), nil
}

func (emptyDurableUserStore) GetByIDWithError(string) (auth.User, bool, error) {
	return auth.User{}, false, nil
}

func (s emptyDurableUserStore) GetByAccountWithError(account string) (auth.User, bool, error) {
	user, ok := s.GetByAccount(account)
	return user, ok, nil
}

func (s emptyDurableUserStore) UpdateRoleWithError(id string, role auth.Role) (auth.User, bool, error) {
	user, ok := s.UpdateRole(id, role)
	return user, ok, nil
}

func (s emptyDurableUserStore) DisableWithError(id string, disabledAt time.Time) (bool, error) {
	return s.Disable(id, disabledAt), nil
}

func TestDurableUserStoreRejectsSignedSessionWithoutPersistedUser(t *testing.T) {
	cfg := &config.Config{
		BaseURL:      "https://api.quanttide.com/qtcloud-asset",
		StudioOrigin: "https://asset.cloud.quanttide.com",
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        auth.NewMemorySessionStore(),
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	users := emptyDurableUserStore{MemoryUserStore: auth.NewMemoryUserStore()}
	handler := NewWithStores(cfg, nil, sessions, fakeIdentityProvider{}, users, auth.NewMemoryAuditLogStore())
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	cookie, err := sessions.CreateSession(auth.User{
		ID:   "missing-user",
		Role: auth.RoleAdmin,
	}, time.Now())
	if err != nil {
		t.Fatalf("create signed session: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing persisted user HTTP 401, got %d: %s", res.Code, res.Body.String())
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

func newLocalPasswordAuthTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	mux, _ := newLocalPasswordAuthTestMuxWithHandler(t)
	return mux
}

func newLocalPasswordAuthTestMuxWithHandler(t *testing.T) (*http.ServeMux, *Handler) {
	t.Helper()

	passwordHash, err := auth.HashPasswordPBKDF2("correct-password", 1000)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	cfg := &config.Config{
		StudioOrigin:  "https://asset.cloud.quanttide.com",
		StudioOrigins: []string{"https://asset.cloud.quanttide.com"},
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        auth.NewMemorySessionStore(),
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	handler := NewWithStores(cfg, nil, sessions, fakeIdentityProvider{}, auth.NewMemoryUserStore(), auth.NewMemoryAuditLogStore())
	handler.localAuthenticator = auth.NewLocalPasswordAuthenticator(auth.LocalPasswordConfig{
		Email:        "admin@example.com",
		Name:         "Admin User",
		Role:         auth.RoleAdmin,
		PasswordHash: passwordHash,
	})
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux, handler
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
