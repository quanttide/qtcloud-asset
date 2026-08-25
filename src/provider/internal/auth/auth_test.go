package auth

import (
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
