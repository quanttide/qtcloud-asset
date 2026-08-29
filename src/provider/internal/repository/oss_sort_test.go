package repository

import (
	"testing"

	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

func TestSortObjectsBySizeDesc(t *testing.T) {
	objects := []schema.Object{
		{Key: "a", Size: 10},
		{Key: "b", Size: 30},
		{Key: "c", Size: 20},
	}
	sortObjects(objects, "size", "desc")

	if objects[0].Key != "b" || objects[1].Key != "c" || objects[2].Key != "a" {
		t.Fatalf("size desc 排序错误: %+v", objects)
	}
}

func TestSortObjectsByKeyAsc(t *testing.T) {
	objects := []schema.Object{
		{Key: "z.txt"},
		{Key: "a.txt"},
		{Key: "m.txt"},
	}
	sortObjects(objects, "key", "asc")

	if objects[0].Key != "a.txt" || objects[1].Key != "m.txt" || objects[2].Key != "z.txt" {
		t.Fatalf("key asc 排序错误: %+v", objects)
	}
}

func TestSortObjectsByDateDesc(t *testing.T) {
	objects := []schema.Object{
		{Key: "old", LastModified: "2026-01-01 00:00:00"},
		{Key: "new", LastModified: "2026-08-21 10:00:00"},
		{Key: "mid", LastModified: "2026-05-15 12:00:00"},
	}
	sortObjects(objects, "date", "desc")

	if objects[0].Key != "new" || objects[1].Key != "mid" || objects[2].Key != "old" {
		t.Fatalf("date desc 排序错误: %+v", objects)
	}
}

func TestSortObjectsEmptySortKey(t *testing.T) {
	objects := []schema.Object{
		{Key: "a"},
		{Key: "b"},
	}
	sortObjects(objects, "", "desc")

	// 不排序，保持原顺序
	if objects[0].Key != "a" || objects[1].Key != "b" {
		t.Fatalf("空 sortKey 不应排序: %+v", objects)
	}
}

func TestObjectURLEscapesReservedObjectKeyCharacters(t *testing.T) {
	adapter := &OssAdapter{Endpoint: "https://oss-cn-hangzhou.aliyuncs.com"}

	got := adapter.publicObjectURL("public-bucket", "dir/a file?#.txt")
	want := "https://public-bucket.oss-cn-hangzhou.aliyuncs.com/dir/a%20file%3F%23.txt"

	if got != want {
		t.Fatalf("expected encoded object URL %q, got %q", want, got)
	}
}

func TestMetadataOnlyBucketObjectURLIsDisabled(t *testing.T) {
	adapter := &OssAdapter{Endpoint: "https://oss-cn-hangzhou.aliyuncs.com"}

	for _, bucketName := range []string{"qtadmin-private", "quanttide-terraform-state"} {
		t.Run(bucketName, func(t *testing.T) {
			_, err := adapter.validateObjectURLRequest(bucketName, "secret.txt", 600)
			if err == nil {
				t.Fatal("expected metadata-only bucket URL to be disabled")
			}
		})
	}
}

func TestPublicBucketObjectURLAllowsDefaultExpiry(t *testing.T) {
	adapter := &OssAdapter{Endpoint: "https://oss-cn-hangzhou.aliyuncs.com"}

	got, err := adapter.validateObjectURLRequest("qtcloud-asset-studio", "index.html", 0)
	if err != nil {
		t.Fatalf("public object URL should allow default expiry: %v", err)
	}
	if got != "https://qtcloud-asset-studio.oss-cn-hangzhou.aliyuncs.com/index.html" {
		t.Fatalf("unexpected public object URL: %s", got)
	}
}

func TestNewOssAdapterStoresCredentials(t *testing.T) {
	adapter := NewOssAdapter("https://oss-cn-hangzhou.aliyuncs.com", "ak", "secret", "token")

	if adapter.Endpoint != "https://oss-cn-hangzhou.aliyuncs.com" || adapter.AccessKeyID != "ak" || adapter.AccessKeySecret != "secret" || adapter.SecurityToken != "token" {
		t.Fatalf("unexpected adapter config: %+v", adapter)
	}
}

func TestOssAdapterReusesClient(t *testing.T) {
	adapter := NewOssAdapter("https://oss-cn-hangzhou.aliyuncs.com", "ak", "secret", "")

	first, err := adapter.client()
	if err != nil {
		t.Fatalf("create first client: %v", err)
	}
	second, err := adapter.client()
	if err != nil {
		t.Fatalf("create second client: %v", err)
	}
	if first != second {
		t.Fatal("expected OSS client to be reused")
	}
}

func TestHostStripsEndpointScheme(t *testing.T) {
	adapter := &OssAdapter{Endpoint: "https://oss-cn-hangzhou.aliyuncs.com"}
	if got := adapter.host(); got != "oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("unexpected host: %s", got)
	}
}

func TestBeautifyPathOnlyRestoresPathSlashes(t *testing.T) {
	rawURL := "https://bucket.oss-cn-hangzhou.aliyuncs.com/dir%2Fa.txt?Signature=keep%2Fencoded"
	got := beautifyPath(rawURL)
	want := "https://bucket.oss-cn-hangzhou.aliyuncs.com/dir/a.txt?Signature=keep%2Fencoded"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBeautifyPathHandlesURLWithoutQuery(t *testing.T) {
	rawURL := "https://bucket.oss-cn-hangzhou.aliyuncs.com/dir%2Fa.txt"
	got := beautifyPath(rawURL)
	want := "https://bucket.oss-cn-hangzhou.aliyuncs.com/dir/a.txt"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
