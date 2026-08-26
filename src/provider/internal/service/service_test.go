package service

import (
	"testing"

	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

type fakeObjectLister struct {
	called bool
}

func (f *fakeObjectLister) ListObjects(bucketName string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	f.called = true
	return []schema.Object{{Key: "index.html"}}, "", false, nil
}

func TestListObjectsRejectsMetadataOnlyBuckets(t *testing.T) {
	for _, bucketName := range []string{"qtadmin-private", "quanttide-terraform-state"} {
		lister := &fakeObjectLister{}
		service := NewBucketService(nil, lister, nil)

		_, _, _, err := service.ListObjects(bucketName, schema.ListObjectsParams{})
		if err != ErrMetadataOnlyBucket {
			t.Fatalf("expected ErrMetadataOnlyBucket for %s, got %v", bucketName, err)
		}
		if lister.called {
			t.Fatalf("metadata-only bucket %s should not call object lister", bucketName)
		}
	}
}

func TestListObjectsAllowsNonPrivateBuckets(t *testing.T) {
	lister := &fakeObjectLister{}
	service := NewBucketService(nil, lister, nil)

	objects, _, _, err := service.ListObjects("qtcloud-studio", schema.ListObjectsParams{})
	if err != nil {
		t.Fatalf("expected non-private bucket to list objects, got %v", err)
	}
	if !lister.called {
		t.Fatal("expected object lister to be called")
	}
	if len(objects) != 1 || objects[0].Key != "index.html" {
		t.Fatalf("unexpected objects: %+v", objects)
	}
}
