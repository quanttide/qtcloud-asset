// Package repository provides data access abstractions.
//
// Repository Layer: abstracts storage backends (filesystem, GitHub, Feishu).
// Currently a placeholder for future implementation.
package repository

import (
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

// SourceAdapter defines the interface for multi-source asset discovery.
type SourceAdapter interface {
	// Discover scans the source and returns discovered assets.
	// TODO: implement for filesystem, GitHub, Feishu adapters.
	Discover() ([]byte, error)
}

// BucketLister lists OSS buckets (read-only discovery).
type BucketLister interface {
	ListBuckets() ([]schema.Bucket, error)
}

// ObjectLister lists objects inside an OSS bucket (read-only).
type ObjectLister interface {
	ListObjects(bucketName string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error)
}

// ObjectURLBuilder builds access URLs for objects.
type ObjectURLBuilder interface {
	ObjectURL(bucketName, objectKey string, expiresIn int64) (string, error)
}
