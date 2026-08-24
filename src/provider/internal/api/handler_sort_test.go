package api

import (
	"testing"

	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

func TestSortBucketsByNameAsc(t *testing.T) {
	buckets := []schema.Bucket{
		{Name: "zeta"},
		{Name: "alpha"},
		{Name: "mid"},
	}
	sortBuckets(buckets, "name", "asc")

	if buckets[0].Name != "alpha" || buckets[1].Name != "mid" || buckets[2].Name != "zeta" {
		t.Fatalf("name asc 排序错误: %+v", buckets)
	}
}

func TestSortBucketsByCreatedDesc(t *testing.T) {
	buckets := []schema.Bucket{
		{Name: "old", CreatedAt: "2026-01-01 00:00:00"},
		{Name: "new", CreatedAt: "2026-08-21 10:00:00"},
		{Name: "mid", CreatedAt: "2026-05-15 12:00:00"},
	}
	sortBuckets(buckets, "created", "desc")

	if buckets[0].Name != "new" || buckets[1].Name != "mid" || buckets[2].Name != "old" {
		t.Fatalf("created desc 排序错误: %+v", buckets)
	}
}

func TestSortBucketsEmptySortKey(t *testing.T) {
	buckets := []schema.Bucket{
		{Name: "a"},
		{Name: "b"},
	}
	sortBuckets(buckets, "", "desc")

	if buckets[0].Name != "a" || buckets[1].Name != "b" {
		t.Fatalf("空 sortKey 不应排序: %+v", buckets)
	}
}
