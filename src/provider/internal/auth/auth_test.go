package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManagerCreatesAuthenticatesAndRevokesSession(t *testing.T) {
	manager := NewManager(ManagerOptions{
		Store:      NewMemorySessionStore(),
		SessionTTL: time.Hour,
	})
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	user := User{ID: "user-1", Email: "viewer@example.com", Role: RoleViewer}

	cookie, err := manager.CreateSession(user, now)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if cookie.Name != SessionCookieName || !cookie.HttpOnly {
		t.Fatalf("unexpected session cookie: %+v", cookie)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("session cookie should default to SameSite=Lax, got %+v", cookie.SameSite)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(cookie)
	authenticatedUser, sessionID, ok := manager.Authenticate(req, now.Add(time.Minute))
	if !ok {
		t.Fatal("expected session to authenticate")
	}
	if sessionID == "" || authenticatedUser.Email != user.Email {
		t.Fatalf("unexpected authenticated session: id=%q user=%+v", sessionID, authenticatedUser)
	}

	if !manager.RevokeFromRequest(req, now.Add(2*time.Minute)) {
		t.Fatal("expected revoke to succeed")
	}
	if _, _, ok := manager.Authenticate(req, now.Add(3*time.Minute)); ok {
		t.Fatal("revoked session should not authenticate")
	}
}

func TestManagerRejectsExpiredSession(t *testing.T) {
	manager := NewManager(ManagerOptions{
		Store:      NewMemorySessionStore(),
		SessionTTL: time.Minute,
	})
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	cookie, err := manager.CreateSession(User{ID: "user-1", Role: RoleViewer}, now)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(cookie)
	if _, _, ok := manager.Authenticate(req, now.Add(2*time.Minute)); ok {
		t.Fatal("expired session should not authenticate")
	}
}

func TestSignedSessionAuthenticatesAcrossManagerInstances(t *testing.T) {
	signingKey := []byte("stable-local-auth-signing-key")
	creator := NewManager(ManagerOptions{
		Store:             NewMemorySessionStore(),
		SessionTTL:        time.Hour,
		CookieSecure:      true,
		SameSite:          http.SameSiteNoneMode,
		SessionSigningKey: signingKey,
	})
	verifier := NewManager(ManagerOptions{
		Store:             NewMemorySessionStore(),
		SessionTTL:        time.Hour,
		CookieSecure:      true,
		SameSite:          http.SameSiteNoneMode,
		SessionSigningKey: signingKey,
	})
	now := time.Date(2026, 8, 26, 16, 0, 0, 0, time.UTC)
	user := User{
		ID:     "admin-1",
		Email:  "haoziteng@quanttide.com",
		Name:   "Hao Ziteng",
		Role:   RoleAdmin,
		Status: UserStatusActive,
	}

	cookie, err := creator.CreateSession(user, now)
	if err != nil {
		t.Fatalf("create signed session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(cookie)
	authenticated, sessionID, ok := verifier.Authenticate(req, now.Add(time.Minute))
	if !ok {
		t.Fatal("expected signed session to authenticate across manager instances")
	}
	if sessionID == "" || authenticated.Email != user.Email || authenticated.Role != user.Role {
		t.Fatalf("unexpected cross-instance authenticated session: id=%q user=%+v", sessionID, authenticated)
	}
}

func TestSignedSessionRevocationRejectsOldCookie(t *testing.T) {
	signingKey := []byte("stable-local-auth-signing-key")
	manager := NewManager(ManagerOptions{
		Store:             NewMemorySessionStore(),
		SessionTTL:        time.Hour,
		SessionSigningKey: signingKey,
	})
	now := time.Date(2026, 8, 26, 16, 30, 0, 0, time.UTC)
	cookie, err := manager.CreateSession(User{
		ID:     "admin-1",
		Email:  "haoziteng@quanttide.com",
		Role:   RoleAdmin,
		Status: UserStatusActive,
	}, now)
	if err != nil {
		t.Fatalf("create signed session: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(cookie)
	if !manager.RevokeFromRequest(req, now.Add(time.Minute)) {
		t.Fatal("expected signed session revoke to succeed")
	}

	if _, _, ok := manager.Authenticate(req, now.Add(2*time.Minute)); ok {
		t.Fatal("revoked signed session should not authenticate")
	}
}

func TestSignedSessionRejectsTampering(t *testing.T) {
	manager := NewManager(ManagerOptions{
		SessionSigningKey: []byte("stable-local-auth-signing-key"),
	})
	now := time.Date(2026, 8, 26, 16, 0, 0, 0, time.UTC)
	cookie, err := manager.CreateSession(User{
		ID:     "user-1",
		Email:  "user@example.com",
		Role:   RoleViewer,
		Status: UserStatusActive,
	}, now)
	if err != nil {
		t.Fatalf("create signed session: %v", err)
	}
	cookie.Value += "tampered"

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(cookie)
	if _, _, ok := manager.Authenticate(req, now.Add(time.Minute)); ok {
		t.Fatal("tampered signed session should not authenticate")
	}
}

func TestValidateStateRequiresMatchingCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=expected", nil)
	req.AddCookie(&http.Cookie{Name: LoginStateCookieName, Value: "expected"})
	if !ValidateState(req, "expected") {
		t.Fatal("expected matching login state to validate")
	}
	if ValidateState(req, "different") {
		t.Fatal("mismatched login state should not validate")
	}
}

func TestLoginStateCookieUsesHttpOnlySameSiteLaxByDefault(t *testing.T) {
	manager := NewManager(ManagerOptions{})
	cookie := manager.LoginStateCookie("state-1")

	if cookie.Name != LoginStateCookieName || !cookie.HttpOnly {
		t.Fatalf("unexpected login state cookie: %+v", cookie)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("login state cookie should default to SameSite=Lax, got %+v", cookie.SameSite)
	}
}

func TestClearCookiesExpireBrowserState(t *testing.T) {
	manager := NewManager(ManagerOptions{CookieSecure: true})

	for _, cookie := range []*http.Cookie{
		manager.ClearLoginStateCookie(),
		manager.ClearSessionCookie(),
	} {
		if cookie.MaxAge != -1 {
			t.Fatalf("expected %s to expire immediately, got MaxAge=%d", cookie.Name, cookie.MaxAge)
		}
		if !cookie.HttpOnly || !cookie.Secure {
			t.Fatalf("expected %s to keep secure HttpOnly flags: %+v", cookie.Name, cookie)
		}
	}
}

func TestLocalPasswordAuthenticatorVerifiesPBKDF2Hash(t *testing.T) {
	encoded, err := HashPasswordPBKDF2("correct-password", 1000)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	authenticator := NewLocalPasswordAuthenticator(LocalPasswordConfig{
		Account:      "admin",
		Email:        "admin@example.com",
		Name:         "Admin User",
		Role:         RoleAdmin,
		PasswordHash: encoded,
	})

	user, err := authenticator.Authenticate(t.Context(), " admin ", "correct-password")
	if err != nil {
		t.Fatalf("authenticate local password: %v", err)
	}
	if user.Account != "admin" || user.Email != "admin@example.com" || user.Role != RoleAdmin || user.ExternalID != "local:admin" {
		t.Fatalf("unexpected authenticated local user: %+v", user)
	}

	if _, err := authenticator.Authenticate(t.Context(), "admin", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials for wrong password, got %v", err)
	}
}

func TestLocalPasswordAuthenticatorUsesLegacyEmailAsAccountFallback(t *testing.T) {
	encoded, err := HashPasswordPBKDF2("correct-password", 1000)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	authenticator := NewLocalPasswordAuthenticator(LocalPasswordConfig{
		Email:        "admin@example.com",
		PasswordHash: encoded,
	})

	user, err := authenticator.Authenticate(t.Context(), " admin@example.com ", "correct-password")
	if err != nil {
		t.Fatalf("authenticate legacy local password: %v", err)
	}
	if user.Account != "admin@example.com" || user.Email != "admin@example.com" {
		t.Fatalf("unexpected legacy local user: %+v", user)
	}
}

func TestLocalPasswordAuthenticatorRequiresConfiguredHash(t *testing.T) {
	authenticator := NewLocalPasswordAuthenticator(LocalPasswordConfig{
		Email: "admin@example.com",
	})

	if _, err := authenticator.Authenticate(t.Context(), "admin@example.com", "password"); !errors.Is(err, ErrLocalPasswordNotConfigured) {
		t.Fatalf("expected local password config error, got %v", err)
	}
}
