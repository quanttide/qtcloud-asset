package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
	"github.com/quanttide/qtcloud-asset/provider/internal/share"
)

type shareTestBackend struct{}

func (shareTestBackend) ListBuckets() ([]schema.Bucket, error) {
	return []schema.Bucket{
		{Name: "qtcloud-asset-studio"},
		{Name: "qtadmin-private"},
		{Name: "team-data"},
	}, nil
}

func (shareTestBackend) GetBucketACL(bucketName string) (string, error) {
	if bucketName == "qtadmin-private" || bucketName == "team-data" {
		return "private", nil
	}
	return "public-read", nil
}

func (shareTestBackend) ListObjects(bucketName string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	if params.Prefix == "design/" {
		return []schema.Object{
			{Key: "design/logo.svg", Size: 20, Type: "Normal"},
			{Key: "design/nested/readme.txt", Size: 30, Type: "Normal"},
		}, "", false, nil
	}
	return []schema.Object{{Key: "other/secret.txt", Size: 40, Type: "Normal"}}, "", false, nil
}

func (shareTestBackend) ObjectURL(bucketName, objectKey string, expiresIn int64) (string, error) {
	return fmt.Sprintf("https://%s.example.com/%s", bucketName, objectKey), nil
}

type shareTestHarness struct {
	mux      *http.ServeMux
	sessions *auth.Manager
	users    *auth.MemoryUserStore
	audit    *auth.MemoryAuditLogStore
}

func newShareTestHarness(t *testing.T, user auth.User) (shareTestHarness, *http.Cookie) {
	t.Helper()

	cfg := &config.Config{
		StudioOrigin:     "https://asset.cloud.quanttide.com",
		StudioOrigins:    []string{"https://asset.cloud.quanttide.com"},
		ShareableBuckets: []string{"qtcloud-asset-studio", "team-data"},
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        auth.NewMemorySessionStore(),
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	users := auth.NewMemoryUserStore()
	now := time.Now()
	saved, err := users.UpsertManaged(user, now)
	if err != nil {
		t.Fatalf("save test user: %v", err)
	}
	cookie, err := sessions.CreateSession(saved, now)
	if err != nil {
		t.Fatalf("create test session: %v", err)
	}
	buckets := service.NewBucketService(
		shareTestBackend{},
		shareTestBackend{},
		shareTestBackend{},
	)
	audit := auth.NewMemoryAuditLogStore()
	handler := NewWithStores(
		cfg,
		buckets,
		sessions,
		fakeIdentityProvider{},
		users,
		audit,
	)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return shareTestHarness{
		mux:      mux,
		sessions: sessions,
		users:    users,
		audit:    audit,
	}, cookie
}

func (h shareTestHarness) createShare(t *testing.T, cookie *http.Cookie) schema.FolderShareResponse {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost,
		"/shares",
		bytes.NewBufferString(`{"title":"设计稿","bucket":"qtcloud-asset-studio","prefixes":["design/"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	h.mux.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("create share HTTP %d: %s", res.Code, res.Body.String())
	}
	var body schema.FolderShareEnvelope
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode share response: %v", err)
	}
	return body.Share
}

func TestViewerCanCreateAndAnonymouslyBrowsePublicFolderShare(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})
	share := harness.createShare(t, cookie)

	if share.Token == "" || share.URL != "https://asset.cloud.quanttide.com/#/share/"+share.Token {
		t.Fatalf("expected opaque share URL, got %+v", share)
	}
	if share.Bucket != "qtcloud-asset-studio" || len(share.Prefixes) != 1 {
		t.Fatalf("unexpected share metadata: %+v", share)
	}

	metadataReq := httptest.NewRequest(http.MethodGet, "/shares/"+share.Token, nil)
	metadataRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(metadataRes, metadataReq)
	if metadataRes.Code != http.StatusOK {
		t.Fatalf("anonymous share metadata HTTP %d: %s", metadataRes.Code, metadataRes.Body.String())
	}

	objectsReq := httptest.NewRequest(
		http.MethodGet,
		"/shares/"+share.Token+"/objects?prefix=design/",
		nil,
	)
	objectsRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(objectsRes, objectsReq)
	if objectsRes.Code != http.StatusOK {
		t.Fatalf("anonymous share object list HTTP %d: %s", objectsRes.Code, objectsRes.Body.String())
	}
	var objects schema.ObjectListResponse
	if err := json.NewDecoder(objectsRes.Body).Decode(&objects); err != nil {
		t.Fatalf("decode shared objects: %v", err)
	}
	if len(objects.Objects) != 2 || objects.Objects[0].Key != "design/logo.svg" {
		t.Fatalf("unexpected shared objects: %+v", objects)
	}

	urlReq := httptest.NewRequest(
		http.MethodGet,
		"/shares/"+share.Token+"/object-url?key=design/logo.svg",
		nil,
	)
	urlRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(urlRes, urlReq)
	if urlRes.Code != http.StatusOK {
		t.Fatalf("anonymous shared object URL HTTP %d: %s", urlRes.Code, urlRes.Body.String())
	}
}

func TestViewerCanCreateAndAnonymouslyBrowseIndividualFileShare(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/shares",
		bytes.NewBufferString(`{"title":"单文件","bucket":"qtcloud-asset-studio","keys":["design/logo.svg"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("create file share HTTP %d: %s", res.Code, res.Body.String())
	}

	var envelope schema.FolderShareEnvelope
	if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode file share response: %v", err)
	}
	if len(envelope.Share.Prefixes) != 0 || len(envelope.Share.Keys) != 1 ||
		envelope.Share.Keys[0] != "design/logo.svg" {
		t.Fatalf("unexpected file share response: %+v", envelope.Share)
	}

	objectsReq := httptest.NewRequest(
		http.MethodGet,
		"/shares/"+envelope.Share.Token+"/objects?prefix=design/",
		nil,
	)
	objectsRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(objectsRes, objectsReq)
	if objectsRes.Code != http.StatusOK {
		t.Fatalf("list file share objects HTTP %d: %s", objectsRes.Code, objectsRes.Body.String())
	}
	var objects schema.ObjectListResponse
	if err := json.NewDecoder(objectsRes.Body).Decode(&objects); err != nil {
		t.Fatalf("decode file share objects: %v", err)
	}
	if len(objects.Objects) != 1 || objects.Objects[0].Key != "design/logo.svg" {
		t.Fatalf("unexpected file share objects: %+v", objects.Objects)
	}
}

func TestViewerCanListOnlyTheirOwnShares(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "owner-1",
		Role: auth.RoleViewer,
	})
	harness.createShare(t, cookie)
	otherCookie := sessionCookieForExistingMuxUser(t, harness, auth.User{
		ID:   "other-1",
		Role: auth.RoleViewer,
	})
	_ = otherCookie

	req := httptest.NewRequest(http.MethodGet, "/shares", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("list shares HTTP %d: %s", res.Code, res.Body.String())
	}
	var body schema.FolderShareListResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode shares response: %v", err)
	}
	if body.Total != 1 || len(body.Shares) != 1 {
		t.Fatalf("expected one owner share, got %+v", body)
	}
}

func TestFolderShareRejectsPrivateBucketsAndInvalidPrefixes(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})
	for _, body := range []string{
		`{"title":"私密","bucket":"qtadmin-private","prefixes":["design/"]}`,
		`{"title":"不是目录","bucket":"qtcloud-asset-studio","prefixes":["design"]}`,
		`{"title":"根目录","bucket":"qtcloud-asset-studio","prefixes":[""]}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/shares", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
		req.AddCookie(cookie)
		res := httptest.NewRecorder()
		harness.mux.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest && res.Code != http.StatusForbidden {
			t.Fatalf("invalid share request HTTP %d for %s", res.Code, body)
		}
	}
}

func TestFolderShareRejectsUnknownBuckets(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})
	req := httptest.NewRequest(
		http.MethodPost,
		"/shares",
		bytes.NewBufferString(`{"title":"未知","bucket":"unknown-public","prefixes":["design/"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected unknown bucket HTTP 404, got %d: %s", res.Code, res.Body.String())
	}
}

func TestFolderShareReturnsServiceUnavailableWhenDurableStoreIsMissing(t *testing.T) {
	cfg := &config.Config{
		StudioOrigin:     "https://asset.cloud.quanttide.com",
		StudioOrigins:    []string{"https://asset.cloud.quanttide.com"},
		ShareableBuckets: []string{"qtcloud-asset-studio"},
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        auth.NewMemorySessionStore(),
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	users := auth.NewMemoryUserStore()
	now := time.Now()
	user, err := users.UpsertManaged(auth.User{ID: "viewer-1", Role: auth.RoleViewer}, now)
	if err != nil {
		t.Fatalf("save test user: %v", err)
	}
	cookie, err := sessions.CreateSession(user, now)
	if err != nil {
		t.Fatalf("create test session: %v", err)
	}
	buckets := service.NewBucketService(shareTestBackend{}, shareTestBackend{}, shareTestBackend{})
	handler := NewWithStoresAndShares(
		cfg,
		buckets,
		sessions,
		fakeIdentityProvider{},
		users,
		auth.NewMemoryAuditLogStore(),
		share.NewUnavailableStore(errors.New("RDS is not configured")),
	)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(
		http.MethodPost,
		"/shares",
		bytes.NewBufferString(`{"title":"设计稿","bucket":"qtcloud-asset-studio","prefixes":["design/"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected missing durable store HTTP 503, got %d: %s", res.Code, res.Body.String())
	}
}

func TestFolderShareRejectsBucketThatIsNotPublicReadEvenWithoutPrivateSuffix(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})
	req := httptest.NewRequest(
		http.MethodPost,
		"/shares",
		bytes.NewBufferString(`{"title":"内部","bucket":"team-data","prefixes":["design/"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected private ACL share HTTP 403, got %d: %s", res.Code, res.Body.String())
	}
}

func TestFolderShareCannotBrowseOrSignOutsideSharedPrefixes(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})
	share := harness.createShare(t, cookie)

	for _, path := range []string{
		"/shares/" + share.Token + "/objects?prefix=other/",
		"/shares/" + share.Token + "/object-url?key=other/secret.txt",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		harness.mux.ServeHTTP(res, req)
		if res.Code != http.StatusForbidden {
			t.Fatalf("expected outside-prefix request HTTP 403, got %d for %s", res.Code, path)
		}
	}
}

func TestFolderShareOwnerCanRevokeAndAnonymousAccessStops(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})
	share := harness.createShare(t, cookie)

	revokeReq := httptest.NewRequest(http.MethodDelete, "/shares/"+share.Token, nil)
	revokeReq.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	revokeReq.AddCookie(cookie)
	revokeRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(revokeRes, revokeReq)
	if revokeRes.Code != http.StatusNoContent {
		t.Fatalf("revoke share HTTP %d: %s", revokeRes.Code, revokeRes.Body.String())
	}

	metadataReq := httptest.NewRequest(http.MethodGet, "/shares/"+share.Token, nil)
	metadataRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(metadataRes, metadataReq)
	if metadataRes.Code != http.StatusNotFound {
		t.Fatalf("revoked share should return 404, got %d", metadataRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/shares", nil)
	listReq.AddCookie(cookie)
	listRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("list shares after revoke HTTP %d: %s", listRes.Code, listRes.Body.String())
	}
	var listBody schema.FolderShareListResponse
	if err := json.NewDecoder(listRes.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode shares after revoke: %v", err)
	}
	if listBody.Total != 0 || len(listBody.Shares) != 0 {
		t.Fatalf("revoked shares should be hidden from active owner list: %+v", listBody)
	}
}

func TestFolderShareMutationsRejectUnregisteredOriginsAndRedactTokensInAudit(t *testing.T) {
	harness, cookie := newShareTestHarness(t, auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/shares",
		bytes.NewBufferString(`{"title":"设计稿","bucket":"qtcloud-asset-studio","prefixes":["design/"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example.com")
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected cross-origin share creation HTTP 403, got %d", res.Code)
	}

	share := harness.createShare(t, cookie)
	viewReq := httptest.NewRequest(http.MethodGet, "/shares/"+share.Token, nil)
	viewRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(viewRes, viewReq)
	if viewRes.Code != http.StatusOK {
		t.Fatalf("view share HTTP %d: %s", viewRes.Code, viewRes.Body.String())
	}
	urlReq := httptest.NewRequest(
		http.MethodGet,
		"/shares/"+share.Token+"/object-url?key=design/logo.svg",
		nil,
	)
	urlRes := httptest.NewRecorder()
	harness.mux.ServeHTTP(urlRes, urlReq)
	if urlRes.Code != http.StatusOK {
		t.Fatalf("shared object URL HTTP %d: %s", urlRes.Code, urlRes.Body.String())
	}
	for _, entry := range harness.audit.List() {
		if strings.Contains(entry.Target, share.Token) {
			t.Fatalf("audit target must not contain share token: %+v", entry)
		}
		if strings.Contains(entry.Target, "design/logo.svg") {
			t.Fatalf("audit target must not contain object key: %+v", entry)
		}
	}
}

func TestViewerCannotRevokeAnotherUsersShare(t *testing.T) {
	harness, ownerCookie := newShareTestHarness(t, auth.User{
		ID:   "owner-1",
		Role: auth.RoleViewer,
	})
	share := harness.createShare(t, ownerCookie)

	otherCookie := sessionCookieForExistingMuxUser(t, harness, auth.User{
		ID:   "other-1",
		Role: auth.RoleViewer,
	})
	req := httptest.NewRequest(http.MethodDelete, "/shares/"+share.Token, nil)
	req.Header.Set("Origin", "https://asset.cloud.quanttide.com")
	req.AddCookie(otherCookie)
	res := httptest.NewRecorder()
	harness.mux.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected another viewer revoke HTTP 403, got %d", res.Code)
	}
}

func sessionCookieForExistingMuxUser(t *testing.T, harness shareTestHarness, user auth.User) *http.Cookie {
	t.Helper()
	now := time.Now()
	saved, err := harness.users.UpsertManaged(user, now)
	if err != nil {
		t.Fatalf("save other user: %v", err)
	}
	cookie, err := harness.sessions.CreateSession(saved, now)
	if err != nil {
		t.Fatalf("create other session: %v", err)
	}
	return cookie
}
