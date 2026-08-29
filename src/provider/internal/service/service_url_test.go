package service

import (
	"testing"

	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

type fakeBucketLister struct {
	buckets []schema.Bucket
	err     error
}

func (f fakeBucketLister) ListBuckets() ([]schema.Bucket, error) {
	return f.buckets, f.err
}

func (f fakeBucketLister) GetBucketACL(bucketName string) (string, error) {
	if bucketName == "qtadmin-private" || bucketName == "quanttide-terraform-state" || bucketName == "team-data" {
		return "private", nil
	}
	return "public-read", nil
}

type fakeObjectURLBuilder struct {
	called bool
	url    string
	err    error
}

func (f *fakeObjectURLBuilder) ObjectURL(bucketName, objectKey string, expiresIn int64) (string, error) {
	f.called = true
	return f.url, f.err
}

func TestListBucketsDelegatesToLister(t *testing.T) {
	service := NewBucketService(fakeBucketLister{buckets: []schema.Bucket{{Name: "qtcloud-asset-studio"}}}, nil, nil)

	buckets, err := service.ListBuckets()
	if err != nil {
		t.Fatalf("list buckets: %v", err)
	}
	if len(buckets) != 1 || buckets[0].Name != "qtcloud-asset-studio" {
		t.Fatalf("unexpected buckets: %+v", buckets)
	}
}

func TestObjectURLRejectsMetadataOnlyBuckets(t *testing.T) {
	builder := &fakeObjectURLBuilder{url: "https://example.com/secret.txt"}
	service := NewBucketService(nil, nil, builder)

	_, err := service.ObjectURL("qtadmin-private", "secret.txt", 600)
	if err != ErrMetadataOnlyBucket {
		t.Fatalf("expected metadata-only error, got %v", err)
	}
	if builder.called {
		t.Fatal("metadata-only bucket should not call URL builder")
	}
}

func TestObjectURLAuthorizedRejectsMetadataOnlyBuckets(t *testing.T) {
	for _, bucketName := range []string{"qtadmin-private", "quanttide-terraform-state"} {
		t.Run(bucketName, func(t *testing.T) {
			builder := &fakeObjectURLBuilder{url: "https://example.com/secret.txt"}
			service := NewBucketService(nil, nil, builder)

			_, err := service.ObjectURLAuthorized(bucketName, "secret.txt", 600)
			if err != ErrMetadataOnlyBucket {
				t.Fatalf("expected metadata-only error, got %v", err)
			}
			if builder.called {
				t.Fatal("metadata-only bucket should not call URL builder")
			}
		})
	}
}

func TestObjectURLAuthorizedDelegatesPublicBucketsToBuilder(t *testing.T) {
	builder := &fakeObjectURLBuilder{url: "https://example.com/index.html"}
	service := NewBucketService(fakeBucketLister{}, nil, builder)

	got, err := service.ObjectURLAuthorized("qtcloud-asset-studio", "index.html", 600)
	if err != nil {
		t.Fatalf("authorized object URL: %v", err)
	}
	if !builder.called || got != builder.url {
		t.Fatalf("expected URL builder result %q, got %q called=%v", builder.url, got, builder.called)
	}
}

func TestObjectURLAuthorizedRejectsNonPublicACLWithoutNameInference(t *testing.T) {
	builder := &fakeObjectURLBuilder{url: "https://example.com/internal.txt"}
	service := NewBucketService(fakeBucketLister{}, nil, builder)

	_, err := service.ObjectURLAuthorized("team-data", "internal.txt", 600)
	if err != ErrBucketNotPublic {
		t.Fatalf("expected non-public ACL to be rejected, got %v", err)
	}
	if builder.called {
		t.Fatal("non-public bucket must not call URL builder")
	}
}

func TestObjectURLAuthorizedRequiresBuilder(t *testing.T) {
	service := NewBucketService(fakeBucketLister{}, nil, nil)

	_, err := service.ObjectURLAuthorized("qtcloud-asset-studio", "index.html", 600)
	if err == nil {
		t.Fatal("expected missing URL builder error")
	}
}

func TestListObjectsAuthorizedRequiresLister(t *testing.T) {
	service := NewBucketService(nil, nil, nil)

	_, _, _, err := service.ListObjectsAuthorized("qtcloud-asset-studio", schema.ListObjectsParams{})
	if err == nil {
		t.Fatal("expected missing object lister error")
	}
}

func TestIsPublicBucketRequiresACLReader(t *testing.T) {
	service := NewBucketService(nil, nil, nil)

	_, err := service.IsPublicBucket("qtcloud-asset-studio")
	if err != ErrBucketACLUnavailable {
		t.Fatalf("expected missing ACL reader error, got %v", err)
	}
}

func TestIsPublicBucketUsesCurrentACL(t *testing.T) {
	service := NewBucketService(fakeBucketLister{}, nil, nil)

	public, err := service.IsPublicBucket("qtcloud-asset-studio")
	if err != nil || !public {
		t.Fatalf("expected public ACL, got public=%v err=%v", public, err)
	}
	private, err := service.IsPublicBucket("qtadmin-private")
	if err != nil {
		t.Fatalf("read private ACL: %v", err)
	}
	if private {
		t.Fatal("private ACL must not be treated as public")
	}
}
