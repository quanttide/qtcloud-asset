package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/quanttide/qtcloud-asset/provider/internal/auth"
	"github.com/quanttide/qtcloud-asset/provider/internal/config"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
	"github.com/quanttide/qtcloud-asset/provider/internal/service"
)

func TestClassifyObjectListError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantText   string
	}{
		{
			name:       "access denied",
			err:        fmt.Errorf("list objects: %w", oss.ServiceError{Code: "AccessDenied", StatusCode: http.StatusForbidden}),
			wantStatus: http.StatusServiceUnavailable,
			wantText:   "Provider 无权访问此 OSS 桶，请检查生产权限配置",
		},
		{
			name:       "missing bucket",
			err:        fmt.Errorf("list objects: %w", oss.ServiceError{Code: "NoSuchBucket", StatusCode: http.StatusNotFound}),
			wantStatus: http.StatusNotFound,
			wantText:   "OSS 桶不存在或当前 Provider 无权查看",
		},
		{
			name:       "invalid marker",
			err:        fmt.Errorf("list objects: %w", oss.ServiceError{Code: "InvalidArgument", StatusCode: http.StatusBadRequest}),
			wantStatus: http.StatusBadRequest,
			wantText:   "文件列表分页参数无效，请刷新后重试",
		},
		{
			name:       "wrong endpoint",
			err:        fmt.Errorf("list objects: %w", oss.ServiceError{Code: "PermanentRedirect", StatusCode: http.StatusMovedPermanently}),
			wantStatus: http.StatusServiceUnavailable,
			wantText:   "OSS 区域或访问端点配置不匹配",
		},
		{
			name:       "upstream unavailable",
			err:        fmt.Errorf("list objects: %w", oss.ServiceError{Code: "InternalError", StatusCode: http.StatusInternalServerError}),
			wantStatus: http.StatusBadGateway,
			wantText:   "OSS 暂时不可用，请稍后重试",
		},
		{
			name:       "unknown error",
			err:        errors.New("unexpected object lister failure"),
			wantStatus: http.StatusInternalServerError,
			wantText:   "文件列表加载失败，请联系管理员",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, text := classifyObjectListError(tt.err)
			if status != tt.wantStatus {
				t.Fatalf("expected HTTP %d, got %d", tt.wantStatus, status)
			}
			if text != tt.wantText {
				t.Fatalf("expected message %q, got %q", tt.wantText, text)
			}
		})
	}
}

type objectListErrorBackend struct {
	err error
}

func (b objectListErrorBackend) ListBuckets() ([]schema.Bucket, error) {
	return []schema.Bucket{{Name: "qtcloud-asset-studio"}}, nil
}

func (objectListErrorBackend) GetBucketACL(string) (string, error) {
	return "public-read", nil
}

func (b objectListErrorBackend) ListObjects(string, schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	return nil, "", false, b.err
}

func (objectListErrorBackend) ObjectURL(string, string, int64) (string, error) {
	return "https://example.com/object", nil
}

func TestBucketObjectsListReturnsSafeClassifiedError(t *testing.T) {
	cfg := &config.Config{
		StudioOrigin:  "https://asset.cloud.quanttide.com",
		StudioOrigins: []string{"https://asset.cloud.quanttide.com"},
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:        auth.NewMemorySessionStore(),
		SessionTTL:   time.Hour,
		CookieSecure: true,
	})
	user := auth.User{ID: "viewer-1", Role: auth.RoleViewer}
	cookie, err := sessions.CreateSession(user, time.Now())
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	users := auth.NewMemoryUserStore()
	if _, err := users.UpsertFromIdentity(user, time.Now()); err != nil {
		t.Fatalf("upsert user: %v", err)
	}
	backend := objectListErrorBackend{
		err: fmt.Errorf("oss details must stay server-side: %w", oss.ServiceError{
			Code:       "AccessDenied",
			Message:    "secret permission details",
			StatusCode: http.StatusForbidden,
		}),
	}
	buckets := service.NewBucketService(backend, backend, backend)
	handler := NewWithStores(
		cfg,
		buckets,
		sessions,
		fakeIdentityProvider{},
		users,
		auth.NewMemoryAuditLogStore(),
	)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/buckets/qtcloud-asset-studio/objects", nil)
	req.AddCookie(cookie)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected classified OSS permission failure HTTP 503, got %d", res.Code)
	}
	var body schema.ErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Message != objectListPermissionMessage {
		t.Fatalf("unexpected safe error message: %q", body.Message)
	}
	if strings.Contains(body.Message, "secret permission details") {
		t.Fatal("OSS error details must not be exposed")
	}
}
