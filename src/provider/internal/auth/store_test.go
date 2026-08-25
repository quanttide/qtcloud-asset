package auth

import (
	"testing"
	"time"
)

func TestMemoryUserStoreUpsertsAndDisablesUsers(t *testing.T) {
	store := NewMemoryUserStore()
	now := time.Date(2026, 8, 25, 11, 0, 0, 0, time.UTC)
	user := User{
		ExternalID: "lark-user-1",
		Email:      "viewer@example.com",
		Name:       "Viewer User",
		Role:       RoleViewer,
	}

	saved, err := store.UpsertFromIdentity(user, now)
	if err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	if saved.ID == "" || saved.Status != UserStatusActive || !saved.LastLoginAt.Equal(now) {
		t.Fatalf("unexpected saved user: %+v", saved)
	}

	if !store.Disable(saved.ID, now.Add(time.Minute)) {
		t.Fatal("expected disable to succeed")
	}
	disabled, ok := store.GetByID(saved.ID)
	if !ok || disabled.Status != UserStatusDisabled {
		t.Fatalf("expected disabled user, got %+v ok=%v", disabled, ok)
	}
}

func TestMemoryUserStoreManagedUsersCanBeListedAndRoleUpdated(t *testing.T) {
	store := NewMemoryUserStore()
	now := time.Date(2026, 8, 25, 11, 30, 0, 0, time.UTC)

	viewer, err := store.UpsertManaged(User{
		ExternalID: "managed:viewer@example.com",
		Email:      "viewer@example.com",
		Name:       "Viewer User",
		Role:       RoleViewer,
		Status:     UserStatusActive,
	}, now)
	if err != nil {
		t.Fatalf("upsert managed viewer: %v", err)
	}
	if !viewer.LastLoginAt.IsZero() {
		t.Fatalf("managed invite should not set last login: %+v", viewer)
	}
	if _, err := store.UpsertManaged(User{
		ExternalID: "managed:admin@example.com",
		Email:      "admin@example.com",
		Name:       "Admin User",
		Role:       RoleAdmin,
		Status:     UserStatusActive,
	}, now); err != nil {
		t.Fatalf("upsert managed admin: %v", err)
	}

	users := store.List()
	if len(users) != 2 || users[0].Email != "admin@example.com" || users[1].Email != "viewer@example.com" {
		t.Fatalf("expected users sorted by email, got %+v", users)
	}
	updated, ok := store.UpdateRole(viewer.ID, RoleAdmin)
	if !ok || updated.Role != RoleAdmin {
		t.Fatalf("expected managed viewer role update, got %+v ok=%v", updated, ok)
	}
}

func TestMemorySessionStorePersistsMetadataAndRevokesUserSessions(t *testing.T) {
	store := NewMemorySessionStore()
	now := time.Date(2026, 8, 25, 11, 0, 0, 0, time.UTC)
	session := Session{
		ID:        "session-1",
		UserID:    "user-1",
		User:      User{ID: "user-1", Email: "viewer@example.com"},
		ExpiresAt: now.Add(time.Hour),
		IP:        "192.0.2.10",
		UserAgent: "test-agent",
	}

	if err := store.Create(session); err != nil {
		t.Fatalf("create session: %v", err)
	}
	saved, ok := store.Get("session-1")
	if !ok {
		t.Fatal("expected session to be persisted")
	}
	if saved.UserID != "user-1" || saved.IP != "192.0.2.10" || saved.UserAgent != "test-agent" {
		t.Fatalf("session metadata not persisted: %+v", saved)
	}

	revoked := store.RevokeUserSessions("user-1", now.Add(time.Minute))
	if revoked != 1 {
		t.Fatalf("expected one revoked session, got %d", revoked)
	}
	revokedSession, _ := store.Get("session-1")
	if revokedSession.RevokedAt == nil {
		t.Fatal("expected session to be revoked")
	}
}

func TestMemoryAuditLogStoreRecordsAndListsEntries(t *testing.T) {
	store := NewMemoryAuditLogStore()
	now := time.Date(2026, 8, 25, 11, 0, 0, 0, time.UTC)
	entry := AuditLog{
		UserID:    "user-1",
		Action:    AuditActionLogin,
		Target:    "/auth/callback",
		Result:    AuditResultSuccess,
		IP:        "192.0.2.10",
		UserAgent: "test-agent",
		CreatedAt: now,
	}

	if err := store.Record(entry); err != nil {
		t.Fatalf("record audit log: %v", err)
	}
	entries := store.List()
	if len(entries) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(entries))
	}
	if entries[0].ID == "" || entries[0].Action != AuditActionLogin || entries[0].CreatedAt != now {
		t.Fatalf("unexpected audit entry: %+v", entries[0])
	}
}
