package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
)

type adminTestHarness struct {
	mux      *http.ServeMux
	sessions *auth.Manager
	store    *auth.MemorySessionStore
	users    *auth.MemoryUserStore
	audit    *auth.MemoryAuditLogStore
}

func newAdminTestHarness(t *testing.T) adminTestHarness {
	t.Helper()

	cfg := &config.Config{
		StudioOrigin:  "https://asset.cloud.quanttide.com",
		StudioOrigins: []string{"https://asset.cloud.quanttide.com"},
	}
	store := auth.NewMemorySessionStore()
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        store,
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	users := auth.NewMemoryUserStore()
	auditLogs := auth.NewMemoryAuditLogStore()
	handler := NewWithStores(cfg, nil, sessions, fakeIdentityProvider{}, users, auditLogs)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return adminTestHarness{mux: mux, sessions: sessions, store: store, users: users, audit: auditLogs}
}

func (h adminTestHarness) sessionCookie(t *testing.T, user auth.User) *http.Cookie {
	t.Helper()

	now := time.Now()
	saved, err := h.users.UpsertManaged(user, now)
	if err != nil {
		t.Fatalf("upsert managed user: %v", err)
	}
	cookie, err := h.sessions.CreateSession(saved, now)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return cookie
}

func TestAdminCanInviteListAndChangeUserRole(t *testing.T) {
	harness := newAdminTestHarness(t)
	adminCookie := harness.sessionCookie(t, auth.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin", Role: auth.RoleAdmin})

	inviteReq := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBufferString(`{"email":"new.viewer@example.com","name":"New Viewer","role":"viewer"}`))
	inviteReq.AddCookie(adminCookie)
	inviteRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(inviteRes, inviteReq)

	if inviteRes.Code != http.StatusCreated {
		t.Fatalf("expected invite HTTP 201, got %d: %s", inviteRes.Code, inviteRes.Body.String())
	}
	var invited struct {
		User auth.User `json:"user"`
	}
	if err := json.NewDecoder(inviteRes.Body).Decode(&invited); err != nil {
		t.Fatalf("decode invite response: %v", err)
	}
	if invited.User.ID == "" || invited.User.Role != auth.RoleViewer || invited.User.Status != auth.UserStatusActive {
		t.Fatalf("unexpected invited user: %+v", invited.User)
	}
	if !invited.User.LastLoginAt.IsZero() {
		t.Fatalf("invited user should not have a last login timestamp: %+v", invited.User)
	}

	roleReq := httptest.NewRequest(http.MethodPatch, "/admin/users/"+invited.User.ID+"/role", bytes.NewBufferString(`{"role":"admin"}`))
	roleReq.AddCookie(adminCookie)
	roleRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(roleRes, roleReq)

	if roleRes.Code != http.StatusOK {
		t.Fatalf("expected role update HTTP 200, got %d: %s", roleRes.Code, roleRes.Body.String())
	}
	updated, ok := harness.users.GetByID(invited.User.ID)
	if !ok || updated.Role != auth.RoleAdmin {
		t.Fatalf("expected invited user role to be admin, got %+v ok=%v", updated, ok)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	listReq.AddCookie(adminCookie)
	listRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(listRes, listReq)

	if listRes.Code != http.StatusOK {
		t.Fatalf("expected user list HTTP 200, got %d", listRes.Code)
	}
	var listBody struct {
		Users []auth.User `json:"users"`
		Total int         `json:"total"`
	}
	if err := json.NewDecoder(listRes.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode users response: %v", err)
	}
	if listBody.Total != 2 || len(listBody.Users) != 2 {
		t.Fatalf("expected two users, got total=%d users=%+v", listBody.Total, listBody.Users)
	}

	assertAuditEntry(t, harness.audit.List(), auth.AuditActionInviteUser, auth.AuditResultSuccess)
	assertAuditEntry(t, harness.audit.List(), auth.AuditActionUpdateUserRole, auth.AuditResultSuccess)
}

func TestViewerCannotUseAdminUserRoutes(t *testing.T) {
	harness := newAdminTestHarness(t)
	viewerCookie := harness.sessionCookie(t, auth.User{ID: "viewer-1", Email: "viewer@example.com", Name: "Viewer", Role: auth.RoleViewer})

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	req.AddCookie(viewerCookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected viewer admin list HTTP 403, got %d", res.Code)
	}
	assertAuditEntry(t, harness.audit.List(), auth.AuditActionListUsers, auth.AuditResultDenied)
}

func TestAdminDisableUserRevokesSessions(t *testing.T) {
	harness := newAdminTestHarness(t)
	adminCookie := harness.sessionCookie(t, auth.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin", Role: auth.RoleAdmin})
	targetCookie := harness.sessionCookie(t, auth.User{ID: "target-1", Email: "target@example.com", Name: "Target", Role: auth.RoleViewer})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/target-1/disable", nil)
	req.AddCookie(adminCookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected disable HTTP 200, got %d: %s", res.Code, res.Body.String())
	}
	disabled, ok := harness.users.GetByID("target-1")
	if !ok || disabled.Status != auth.UserStatusDisabled {
		t.Fatalf("expected disabled target user, got %+v ok=%v", disabled, ok)
	}
	if session, ok := harness.store.Get(targetCookie.Value); !ok || session.RevokedAt == nil {
		t.Fatalf("expected target session to be revoked, got %+v ok=%v", session, ok)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.AddCookie(targetCookie)
	meRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked target session HTTP 401, got %d", meRes.Code)
	}
	assertAuditEntry(t, harness.audit.List(), auth.AuditActionDisableUser, auth.AuditResultSuccess)
}

func TestAdminCanRevokeUserSessionsWithoutDisablingAccount(t *testing.T) {
	harness := newAdminTestHarness(t)
	adminCookie := harness.sessionCookie(t, auth.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin", Role: auth.RoleAdmin})
	targetCookie := harness.sessionCookie(t, auth.User{ID: "target-1", Email: "target@example.com", Name: "Target", Role: auth.RoleViewer})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/target-1/sessions/revoke", nil)
	req.AddCookie(adminCookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected revoke sessions HTTP 200, got %d: %s", res.Code, res.Body.String())
	}
	var body struct {
		Revoked int `json:"revoked"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode revoke sessions response: %v", err)
	}
	if body.Revoked != 1 {
		t.Fatalf("expected one revoked session, got %d", body.Revoked)
	}
	user, ok := harness.users.GetByID("target-1")
	if !ok || user.Status != auth.UserStatusActive {
		t.Fatalf("expected target account to remain active, got %+v ok=%v", user, ok)
	}
	if session, ok := harness.store.Get(targetCookie.Value); !ok || session.RevokedAt == nil {
		t.Fatalf("expected target session to be revoked, got %+v ok=%v", session, ok)
	}
	assertAuditEntry(t, harness.audit.List(), auth.AuditActionRevokeSessions, auth.AuditResultSuccess)
}

func TestAdminUserRoutesValidateInput(t *testing.T) {
	harness := newAdminTestHarness(t)
	adminCookie := harness.sessionCookie(t, auth.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin", Role: auth.RoleAdmin})

	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBufferString(`{"email":"not-an-email","name":"Bad","role":"owner"}`))
	req.AddCookie(adminCookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid invite HTTP 400, got %d", res.Code)
	}
	assertAuditEntry(t, harness.audit.List(), auth.AuditActionInviteUser, auth.AuditResultFailure)
}

func TestAdminUserMutationsRejectUnregisteredOrigin(t *testing.T) {
	harness := newAdminTestHarness(t)
	adminCookie := harness.sessionCookie(t, auth.User{ID: "admin-1", Email: "admin@example.com", Name: "Admin", Role: auth.RoleAdmin})

	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBufferString(`{"email":"new.viewer@example.com","name":"New Viewer","role":"viewer"}`))
	req.Header.Set("Origin", "https://evil.example.com")
	req.AddCookie(adminCookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected unregistered origin HTTP 403, got %d", res.Code)
	}
	assertAuditEntry(t, harness.audit.List(), auth.AuditActionInviteUser, auth.AuditResultDenied)
}
