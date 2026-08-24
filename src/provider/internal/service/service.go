// Package service implements business logic orchestration.
//
// Service Layer: coordinates Contract, Discovery, and Snapshot operations.
// Currently a placeholder for future implementation.
package service

import (
	"errors"
	"fmt"

	"github.com/quanttide/qtcloud-asset/provider/internal/repository"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

// ErrMetadataOnlyBucket is returned when a bucket is intentionally limited to metadata.
var ErrMetadataOnlyBucket = errors.New("bucket is metadata-only")

// BucketService exposes OSS bucket discovery.
type BucketService struct {
	bucketLister repository.BucketLister
	objectLister repository.ObjectLister
	urlBuilder   repository.ObjectURLBuilder
}

// NewBucketService creates a BucketService backed by the given listers.
func NewBucketService(bucketLister repository.BucketLister, objectLister repository.ObjectLister, urlBuilder repository.ObjectURLBuilder) *BucketService {
	return &BucketService{
		bucketLister: bucketLister,
		objectLister: objectLister,
		urlBuilder:   urlBuilder,
	}
}

// ListBuckets returns all discovered OSS buckets.
func (s *BucketService) ListBuckets() ([]schema.Bucket, error) {
	return s.bucketLister.ListBuckets()
}

// ListObjects returns objects inside a given bucket.
// 返回：对象列表、下一页 marker、是否被截断。
func (s *BucketService) ListObjects(bucketName string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	if IsMetadataOnlyBucket(bucketName) {
		return nil, "", false, ErrMetadataOnlyBucket
	}
	if s.objectLister == nil {
		return nil, "", false, fmt.Errorf("object lister not configured")
	}
	return s.objectLister.ListObjects(bucketName, params)
}

// ObjectURL builds an access URL for an object.
func (s *BucketService) ObjectURL(bucketName, objectKey string, expiresIn int64) (string, error) {
	if IsMetadataOnlyBucket(bucketName) {
		return "", ErrMetadataOnlyBucket
	}
	if s.urlBuilder == nil {
		return "", fmt.Errorf("url builder not configured")
	}
	return s.urlBuilder.ObjectURL(bucketName, objectKey, expiresIn)
}

// IsMetadataOnlyBucket reports whether object-level access should remain disabled.
func IsMetadataOnlyBucket(bucketName string) bool {
	return len(bucketName) >= len("-private") && bucketName[len(bucketName)-len("-private"):] == "-private" ||
		bucketName == "quanttide-terraform-state"
}
