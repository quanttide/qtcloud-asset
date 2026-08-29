// Package service implements business logic orchestration.
//
// Service Layer: coordinates Contract, Discovery, and Snapshot operations.
// Currently a placeholder for future implementation.
package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/quanttide/qtcloud-asset/provider/internal/repository"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

// ErrMetadataOnlyBucket is returned when a bucket is intentionally limited to metadata.
var ErrMetadataOnlyBucket = errors.New("bucket is metadata-only")

// ErrBucketACLUnavailable is returned when a public-access decision cannot be verified.
var ErrBucketACLUnavailable = errors.New("bucket ACL is unavailable")

// ErrBucketNotPublic is returned when a bucket is not configured for anonymous reads.
var ErrBucketNotPublic = errors.New("bucket is not public")

// ErrObjectReaderUnavailable is returned when object streaming is not configured.
var ErrObjectReaderUnavailable = errors.New("object reader is unavailable")

// BucketService exposes OSS bucket discovery.
type BucketService struct {
	bucketLister repository.BucketLister
	objectLister repository.ObjectLister
	objectReader repository.ObjectReader
	urlBuilder   repository.ObjectURLBuilder
	aclReader    repository.BucketACLReader
}

// NewBucketService creates a BucketService backed by the given listers.
func NewBucketService(bucketLister repository.BucketLister, objectLister repository.ObjectLister, urlBuilder repository.ObjectURLBuilder) *BucketService {
	var aclReader repository.BucketACLReader
	if reader, ok := bucketLister.(repository.BucketACLReader); ok {
		aclReader = reader
	} else if reader, ok := urlBuilder.(repository.BucketACLReader); ok {
		aclReader = reader
	}
	var objectReader repository.ObjectReader
	if reader, ok := objectLister.(repository.ObjectReader); ok {
		objectReader = reader
	} else if reader, ok := urlBuilder.(repository.ObjectReader); ok {
		objectReader = reader
	}
	return &BucketService{
		bucketLister: bucketLister,
		objectLister: objectLister,
		objectReader: objectReader,
		urlBuilder:   urlBuilder,
		aclReader:    aclReader,
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
	return s.ListObjectsAuthorized(bucketName, params)
}

// ListObjectsAuthorized returns objects after the API layer has authorized access.
func (s *BucketService) ListObjectsAuthorized(bucketName string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	if s.objectLister == nil {
		return nil, "", false, fmt.Errorf("object lister not configured")
	}
	return s.objectLister.ListObjects(bucketName, params)
}

// GetObjectAuthorized opens an object after the API layer has authorized access.
func (s *BucketService) GetObjectAuthorized(bucketName, objectKey string) (io.ReadCloser, error) {
	if IsMetadataOnlyBucket(bucketName) {
		return nil, ErrMetadataOnlyBucket
	}
	if s == nil || s.objectReader == nil {
		return nil, ErrObjectReaderUnavailable
	}
	return s.objectReader.GetObject(bucketName, objectKey)
}

// ObjectURL builds an access URL for an object.
func (s *BucketService) ObjectURL(bucketName, objectKey string, expiresIn int64) (string, error) {
	if IsMetadataOnlyBucket(bucketName) {
		return "", ErrMetadataOnlyBucket
	}
	return s.ObjectURLAuthorized(bucketName, objectKey, expiresIn)
}

// ObjectURLAuthorized builds an object URL after the API layer has authorized access.
func (s *BucketService) ObjectURLAuthorized(bucketName, objectKey string, expiresIn int64) (string, error) {
	if IsMetadataOnlyBucket(bucketName) {
		return "", ErrMetadataOnlyBucket
	}
	public, err := s.IsPublicBucket(bucketName)
	if err != nil {
		return "", err
	}
	if !public {
		return "", ErrBucketNotPublic
	}
	if s.urlBuilder == nil {
		return "", fmt.Errorf("url builder not configured")
	}
	return s.urlBuilder.ObjectURL(bucketName, objectKey, expiresIn)
}

// IsPublicBucket verifies that OSS currently reports the bucket as public-read.
func (s *BucketService) IsPublicBucket(bucketName string) (bool, error) {
	if s == nil || s.aclReader == nil {
		return false, ErrBucketACLUnavailable
	}
	acl, err := s.aclReader.GetBucketACL(bucketName)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrBucketACLUnavailable, err)
	}
	return acl == "public-read", nil
}

// IsMetadataOnlyBucket reports whether object-level access should remain disabled.
func IsMetadataOnlyBucket(bucketName string) bool {
	return len(bucketName) >= len("-private") && bucketName[len(bucketName)-len("-private"):] == "-private" ||
		bucketName == "quanttide-terraform-state"
}
