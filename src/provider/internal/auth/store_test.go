package auth

import (
	"bytes"
	"encoding/json"
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
	if len(users) != 2 || users[0].Account != "admin@example.com" || users[1].Account != "viewer@example.com" {
		t.Fatalf("expected users sorted by account, got %+v", users)
	}
	updated, ok := store.UpdateRole(viewer.ID, RoleAdmin)
	if !ok || updated.Role != RoleAdmin {
		t.Fatalf("expected managed viewer role update, got %+v ok=%v", updated, ok)
	}
}

func TestMemoryUserStoreFindsManagedPasswordByNormalizedAccountWithoutJSONLeak(t *testing.T) {
	store := NewMemoryUserStore()
	now := time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC)
	passwordHash, err := HashPasswordPBKDF2("123456", 1000)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	saved, err := store.UpsertManaged(User{
		ExternalID:   "managed:lixiang",
		Account:      " LiXiang ",
		Name:         "LiXiang",
		Role:         RoleViewer,
		Status:       UserStatusActive,
		PasswordHash: passwordHash,
	}, now)
	if err != nil {
		t.Fatalf("upsert managed user: %v", err)
	}
	if saved.Account != "lixiang" {
		t.Fatalf("expected normalized account, got %q", saved.Account)
	}

	found, ok := store.GetByAccount(" lixiang ")
	if !ok || found.PasswordHash != passwordHash {
		t.Fatalf("expected stored password hash by account, got %+v ok=%v", found, ok)
	}
	encoded, err := json.Marshal(found)
	if err != nil {
		t.Fatalf("marshal managed user: %v", err)
	}
	if bytes.Contains(encoded, []byte("password")) || bytes.Contains(encoded, []byte(passwordHash)) {
		t.Fatalf("user JSON must not expose password material: %s", string(encoded))
	}
}

func TestMemoryIdentityLoginDoesNotOverwriteManagedRoleOrDisabledStatus(t *testing.T) {
	store := NewMemoryUserStore()
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	managed, err := store.UpsertManaged(User{
		ExternalID:   "lark-user-1",
		Account:      "lixiang",
		Name:         "LiXiang",
		Role:         RoleAdmin,
		Status:       UserStatusDisabled,
		PasswordHash: "existing-hash",
	}, now)
	if err != nil {
		t.Fatalf("upsert managed user: %v", err)
	}

	saved, err := store.UpsertFromIdentity(User{
		ExternalID: "lark-user-1",
		Account:    "lixiang",
		Name:       "Li Xiang",
		Role:       RoleViewer,
		Status:     UserStatusActive,
	}, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("upsert identity user: %v", err)
	}
	if saved.ID != managed.ID || saved.Role != RoleAdmin || saved.Status != UserStatusDisabled {
		t.Fatalf("identity login must preserve managed role and status, got %+v", saved)
	}
	if saved.PasswordHash != "existing-hash" {
		t.Fatalf("identity login must preserve managed password hash, got %q", saved.PasswordHash)
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

func TestJSONAuditLogStoreWritesStructuredAuditEntry(t *testing.T) {
	var out bytes.Buffer
	store := NewJSONAuditLogStore(&out)
	now := time.Date(2026, 8, 26, 19, 30, 0, 0, time.UTC)

	if err := store.Record(AuditLog{
		UserID:    "user-1",
		Action:    AuditActionObjectURL,
		Target:    "/buckets/qtcloud-asset-studio/object-url?key=index.html",
		Result:    AuditResultSuccess,
		IP:        "192.0.2.10",
		UserAgent: "test-agent",
		CreatedAt: now,
	}); err != nil {
		t.Fatalf("record structured audit log: %v", err)
	}

	var logged struct {
		Event     string      `json:"event"`
		UserID    string      `json:"user_id"`
		Action    AuditAction `json:"action"`
		Target    string      `json:"target"`
		Result    AuditResult `json:"result"`
		IP        string      `json:"ip"`
		UserAgent string      `json:"user_agent"`
		CreatedAt time.Time   `json:"created_at"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &logged); err != nil {
		t.Fatalf("decode structured audit log: %v output=%q", err, out.String())
	}
	if logged.Event != "qtcloud_asset_audit" || logged.Action != AuditActionObjectURL || logged.Result != AuditResultSuccess {
		t.Fatalf("unexpected structured audit log: %+v", logged)
	}
	if logged.UserID != "user-1" || logged.IP != "192.0.2.10" || !logged.CreatedAt.Equal(now) {
		t.Fatalf("audit metadata not preserved: %+v", logged)
	}
}

func TestMultiAuditLogStoreRecordsEverySink(t *testing.T) {
	memory := NewMemoryAuditLogStore()
	var out bytes.Buffer
	store := NewMultiAuditLogStore(memory, NewJSONAuditLogStore(&out))

	if err := store.Record(AuditLog{Action: AuditActionLogin, Result: AuditResultDenied, Target: "/auth/login"}); err != nil {
		t.Fatalf("record multi audit log: %v", err)
	}
	if got := len(memory.List()); got != 1 {
		t.Fatalf("expected memory sink to receive audit entry, got %d", got)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"event":"qtcloud_asset_audit"`)) {
		t.Fatalf("expected JSON sink to receive audit event, got %q", out.String())
	}
}
