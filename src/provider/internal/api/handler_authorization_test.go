package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
)

type fakeBucketBackend struct{}

func (fakeBucketBackend) ListBuckets() ([]schema.Bucket, error) {
	return []schema.Bucket{
		{Name: "qtcloud-asset-studio"},
		{Name: "qtadmin-private"},
		{Name: "quanttide-terraform-state"},
	}, nil
}

func (fakeBucketBackend) ListObjects(bucketName string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	return []schema.Object{{Key: bucketName + "/index.html"}}, "", false, nil
}

func (fakeBucketBackend) ObjectURL(bucketName, objectKey string, expiresIn int64) (string, error) {
	return fmt.Sprintf("https://signed.example.com/%s/%s?expires=%d", bucketName, objectKey, expiresIn), nil
}

func newAuthorizationTestMux(t *testing.T, user auth.User) (*http.ServeMux, *http.Cookie, *auth.MemoryAuditLogStore) {
	t.Helper()

	cfg := &config.Config{StudioOrigin: "https://asset.cloud.quanttide.com"}
	store := auth.NewMemorySessionStore()
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        store,
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	now := time.Now()
	cookie, err := sessions.CreateSession(user, now)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	users := auth.NewMemoryUserStore()
	if _, err := users.UpsertFromIdentity(user, now); err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	auditLogs := auth.NewMemoryAuditLogStore()
	buckets := service.NewBucketService(fakeBucketBackend{}, fakeBucketBackend{}, fakeBucketBackend{})
	handler := NewWithStores(cfg, buckets, sessions, fakeIdentityProvider{}, users, auditLogs)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux, cookie, auditLogs
}

func TestViewerCanListPublicBucketObjects(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtcloud-asset-studio/objects", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected viewer to list public bucket objects, got HTTP %d", res.Code)
	}
}

func TestViewerBucketListHidesMetadataOnlyBuckets(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected viewer bucket list HTTP 200, got %d", res.Code)
	}
	var body schema.BucketListResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode bucket list response: %v", err)
	}
	if body.Total != 1 || len(body.Buckets) != 1 || body.Buckets[0].Name != "qtcloud-asset-studio" {
		t.Fatalf("expected only public buckets for viewer, got %+v", body)
	}
}

func TestAdminBucketListIncludesMetadataOnlyBuckets(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "admin-1", Role: auth.RoleAdmin})
	req := httptest.NewRequest(http.MethodGet, "/buckets", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected admin bucket list HTTP 200, got %d", res.Code)
	}
	var body schema.BucketListResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode bucket list response: %v", err)
	}
	if body.Total != 3 || len(body.Buckets) != 3 {
		t.Fatalf("expected all buckets for admin, got %+v", body)
	}
}

func TestViewerCannotListPrivateBucketObjects(t *testing.T) {
	mux, cookie, auditLogs := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtadmin-private/objects", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected viewer private object listing to be forbidden, got HTTP %d", res.Code)
	}
	assertAuditEntry(t, auditLogs.List(), auth.AuditActionListObjects, auth.AuditResultDenied)
}

func TestAdminCanListPrivateBucketObjects(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "admin-1", Role: auth.RoleAdmin})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtadmin-private/objects", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected admin to list private bucket objects, got HTTP %d", res.Code)
	}
}

func TestViewerCannotGeneratePrivateObjectURL(t *testing.T) {
	mux, cookie, auditLogs := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtadmin-private/object-url?key=secret.txt", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected viewer private object URL to be forbidden, got HTTP %d", res.Code)
	}
	assertAuditEntry(t, auditLogs.List(), auth.AuditActionObjectURL, auth.AuditResultDenied)
}

func TestAdminCanGeneratePrivateObjectURL(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "admin-1", Role: auth.RoleAdmin})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtadmin-private/object-url?key=secret.txt&expires=600", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected admin to generate private object URL, got HTTP %d", res.Code)
	}
	var body schema.ObjectURLResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode object URL response: %v", err)
	}
	if body.ExpiresIn != 600 {
		t.Fatalf("expected requested private expiry 600, got %d", body.ExpiresIn)
	}
}

func TestPrivateObjectURLDefaultsToOneDay(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "admin-1", Role: auth.RoleAdmin})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtadmin-private/object-url?key=secret.txt", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected admin default private object URL HTTP 200, got %d", res.Code)
	}
	var body schema.ObjectURLResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode object URL response: %v", err)
	}
	if body.ExpiresIn != defaultPrivateObjectURLExpirySeconds {
		t.Fatalf("expected default private expiry %d, got %d", defaultPrivateObjectURLExpirySeconds, body.ExpiresIn)
	}
}

func TestPrivateObjectURLRejectsExpiryAboveMaximum(t *testing.T) {
	mux, cookie, auditLogs := newAuthorizationTestMux(t, auth.User{ID: "admin-1", Role: auth.RoleAdmin})
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/buckets/qtadmin-private/object-url?key=secret.txt&expires=%d", maxPrivateObjectURLExpirySeconds+1), nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected excessive private object URL expiry HTTP 400, got %d", res.Code)
	}
	assertAuditEntry(t, auditLogs.List(), auth.AuditActionObjectURL, auth.AuditResultFailure)
}

func TestPublicObjectURLReportsPermanentExpiry(t *testing.T) {
	mux, cookie, _ := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer})
	req := httptest.NewRequest(http.MethodGet, "/buckets/qtcloud-asset-studio/object-url?key=index.html&expires=600", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected viewer public object URL HTTP 200, got %d", res.Code)
	}
	var body schema.ObjectURLResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode object URL response: %v", err)
	}
	if body.ExpiresIn != publicObjectURLExpirySeconds {
		t.Fatalf("expected public object URL expiry %d, got %d", publicObjectURLExpirySeconds, body.ExpiresIn)
	}
}

func TestDisabledUserSessionCannotAccessBusinessRoutes(t *testing.T) {
	mux, cookie, auditLogs := newAuthorizationTestMux(t, auth.User{ID: "viewer-1", Role: auth.RoleViewer, Status: auth.UserStatusDisabled})
	req := httptest.NewRequest(http.MethodGet, "/buckets", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected disabled user HTTP 403, got %d", res.Code)
	}
	assertAuditEntry(t, auditLogs.List(), auth.AuditActionAuthFailed, auth.AuditResultDenied)
}

func assertAuditEntry(t *testing.T, entries []auth.AuditLog, action auth.AuditAction, result auth.AuditResult) {
	t.Helper()
	for _, entry := range entries {
		if entry.Action == action && entry.Result == result {
			return
		}
	}
	t.Fatalf("expected audit entry action=%s result=%s, got %+v", action, result, entries)
}
