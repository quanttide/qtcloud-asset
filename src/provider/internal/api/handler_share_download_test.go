package api

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
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

type downloadShareBackend struct{}

func (downloadShareBackend) ListBuckets() ([]schema.Bucket, error) {
	return []schema.Bucket{{Name: "qtcloud-asset-studio"}}, nil
}

func (downloadShareBackend) GetBucketACL(string) (string, error) {
	return "public-read", nil
}

func (downloadShareBackend) ListObjects(_ string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	if params.Prefix != "docs/" {
		return nil, "", false, nil
	}
	if params.Marker == "" {
		return []schema.Object{
			{Key: "docs/", Type: "Directory"},
			{Key: "docs/readme.txt", Size: 5, Type: "Normal"},
		}, "page-2", true, nil
	}
	return []schema.Object{
		{Key: "docs/nested/guide.txt", Size: 5, Type: "Normal"},
		{Key: "other/secret.txt", Size: 6, Type: "Normal"},
	}, "", false, nil
}

func (downloadShareBackend) ObjectURL(bucketName, objectKey string, _ int64) (string, error) {
	return fmt.Sprintf("https://%s.example.com/%s", bucketName, objectKey), nil
}

func (downloadShareBackend) GetObject(_ string, objectKey string) (io.ReadCloser, error) {
	contents := map[string]string{
		"docs/readme.txt":       "hello",
		"docs/nested/guide.txt": "guide",
		"exact/.last_build_id":  "a550f78fe5bbe0f41643d324e05d05c8",
		"other/secret.txt":      "secret",
	}
	content, ok := contents[objectKey]
	if !ok {
		return nil, fmt.Errorf("missing object %q", objectKey)
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func newDownloadShareMux(t *testing.T, backend any, record share.Record) (*http.ServeMux, string) {
	t.Helper()

	store := share.NewMemoryStore()
	created, err := store.Create(record)
	if err != nil {
		t.Fatalf("create test share: %v", err)
	}
	cfg := &config.Config{
		StudioOrigin:     "https://asset.cloud.quanttide.com",
		StudioOrigins:    []string{"https://asset.cloud.quanttide.com"},
		ShareableBuckets: []string{"qtcloud-asset-studio"},
	}
	buckets := service.NewBucketService(backend.(interface {
		ListBuckets() ([]schema.Bucket, error)
	}), backend.(interface {
		ListObjects(string, schema.ListObjectsParams) ([]schema.Object, string, bool, error)
	}), backend.(interface {
		ObjectURL(string, string, int64) (string, error)
	}))
	handler := NewWithStoresAndShares(
		cfg,
		buckets,
		nil,
		nil,
		nil,
		auth.NewMemoryAuditLogStore(),
		store,
	)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux, created.Token
}

func TestDownloadShareReturnsAuthorizedZip(t *testing.T) {
	mux, token := newDownloadShareMux(t, downloadShareBackend{}, share.Record{
		Title:    "设计稿",
		Bucket:   "qtcloud-asset-studio",
		Prefixes: []string{"docs/"},
		Keys:     []string{"exact/.last_build_id"},
	})

	req := httptest.NewRequest(http.MethodGet, "/shares/"+token+"/download", nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("download share HTTP %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.Contains(got, "application/zip") {
		t.Fatalf("expected zip content type, got %q", got)
	}
	if got := res.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected no-store cache policy, got %q", got)
	}
	if got := res.Header().Get("Content-Disposition"); !strings.Contains(got, "filename*=UTF-8''%E8%AE%BE%E8%AE%A1%E7%A8%BF.zip") {
		t.Fatalf("expected title-based download name, got %q", got)
	}

	reader, err := zip.NewReader(bytes.NewReader(res.Body.Bytes()), int64(res.Body.Len()))
	if err != nil {
		t.Fatalf("open response as zip: %v", err)
	}
	files := make(map[string]string, len(reader.File))
	for _, file := range reader.File {
		if file.Method != zip.Store {
			t.Fatalf("expected stored zip entry %q, got method %d", file.Name, file.Method)
		}
		content, err := file.Open()
		if err != nil {
			t.Fatalf("open zip entry %q: %v", file.Name, err)
		}
		data, err := io.ReadAll(content)
		_ = content.Close()
		if err != nil {
			t.Fatalf("read zip entry %q: %v", file.Name, err)
		}
		files[file.Name] = string(data)
	}
	if len(files) != 3 {
		t.Fatalf("expected three authorized files, got %v", files)
	}
	if files["docs/readme.txt"] != "hello" ||
		files["docs/nested/guide.txt"] != "guide" ||
		files["exact/.last_build_id"] != "a550f78fe5bbe0f41643d324e05d05c8" {
		t.Fatalf("unexpected zip contents: %v", files)
	}
	if _, ok := files["other/secret.txt"]; ok {
		t.Fatalf("zip contains an unauthorized object: %v", files)
	}
}

func TestDownloadShareRejectsRevokedShare(t *testing.T) {
	revokedAt := time.Now()
	mux, token := newDownloadShareMux(t, downloadShareBackend{}, share.Record{
		Title:     "已撤销",
		Bucket:    "qtcloud-asset-studio",
		Prefixes:  []string{"docs/"},
		RevokedAt: &revokedAt,
	})

	req := httptest.NewRequest(http.MethodGet, "/shares/"+token+"/download", nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("revoked share should return 404, got %d: %s", res.Code, res.Body.String())
	}
}

func TestDownloadShareFailsClosedWithoutObjectReader(t *testing.T) {
	mux, token := newDownloadShareMux(t, shareTestBackend{}, share.Record{
		Title:    "无读取器",
		Bucket:   "qtcloud-asset-studio",
		Prefixes: []string{"design/"},
	})

	req := httptest.NewRequest(http.MethodGet, "/shares/"+token+"/download", nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing object reader should return 503, got %d: %s", res.Code, res.Body.String())
	}
}
