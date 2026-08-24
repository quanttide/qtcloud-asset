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
