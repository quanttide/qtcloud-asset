package repository

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/quanttide/qtcloud-asset/provider/internal/schema"
)

// OssAdapter implements BucketLister backed by Alibaba Cloud OSS.
// It is the first SourceAdapter-style data source for asset discovery.
type OssAdapter struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
}

// NewOssAdapter builds an OssAdapter from explicit or temporary credentials.
func NewOssAdapter(endpoint, accessKeyID, accessKeySecret, securityToken string) *OssAdapter {
	return &OssAdapter{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		SecurityToken:   securityToken,
	}
}

func (a *OssAdapter) client() (*oss.Client, error) {
	options := make([]oss.ClientOption, 0, 1)
	if a.SecurityToken != "" {
		options = append(options, oss.SecurityToken(a.SecurityToken))
	}
	return oss.New(a.Endpoint, a.AccessKeyID, a.AccessKeySecret, options...)
}

// ListBuckets lists all OSS buckets (read-only).
func (a *OssAdapter) ListBuckets() ([]schema.Bucket, error) {
	client, err := a.client()
	if err != nil {
		return nil, fmt.Errorf("create oss client: %w", err)
	}

	result, err := client.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	buckets := make([]schema.Bucket, 0, len(result.Buckets))
	for _, b := range result.Buckets {
		buckets = append(buckets, schema.Bucket{
			Name:         b.Name,
			Region:       b.Location,
			StorageClass: b.StorageClass,
			CreatedAt:    b.CreationDate.Format("2006-01-02"),
		})
	}
	return buckets, nil
}

// ListObjects lists objects (files) inside a bucket (read-only).
//
// prefix / marker / limit 走 OSS 原生能力；sort / order 在内存中排序（OSS 不支持服务端排序）。
// 返回：对象列表、下一页 marker、是否被截断。
func (a *OssAdapter) ListObjects(bucketName string, params schema.ListObjectsParams) ([]schema.Object, string, bool, error) {
	client, err := a.client()
	if err != nil {
		return nil, "", false, fmt.Errorf("create oss client: %w", err)
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, "", false, fmt.Errorf("open bucket %s: %w", bucketName, err)
	}

	// 每页数量：默认 1000（OSS 单次上限），显式传 limit 则用较小值
	maxKeys := 1000
	if params.Limit > 0 && params.Limit < maxKeys {
		maxKeys = params.Limit
	}

	options := []oss.Option{oss.MaxKeys(maxKeys)}
	if params.Prefix != "" {
		options = append(options, oss.Prefix(params.Prefix))
	}
	if params.Marker != "" {
		options = append(options, oss.Marker(params.Marker))
	}

	result, err := bucket.ListObjects(options...)
	if err != nil {
		return nil, "", false, fmt.Errorf("list objects in %s: %w", bucketName, err)
	}

	objects := make([]schema.Object, 0, len(result.Objects))
	for _, o := range result.Objects {
		objects = append(objects, schema.Object{
			Key:          o.Key,
			Size:         o.Size,
			Type:         o.Type,
			StorageClass: o.StorageClass,
			LastModified: o.LastModified.Format("2006-01-02 15:04:05"),
		})
	}

	sortObjects(objects, params.Sort, params.Order)

	return objects, result.NextMarker, result.IsTruncated, nil
}

// sortObjects sorts objects in memory by the given field and order.
func sortObjects(objects []schema.Object, sortKey, order string) {
	if sortKey == "" {
		return
	}
	desc := order == "desc"
	switch sortKey {
	case "key":
		sort.Slice(objects, func(i, j int) bool {
			if desc {
				return objects[i].Key > objects[j].Key
			}
			return objects[i].Key < objects[j].Key
		})
	case "size":
		sort.Slice(objects, func(i, j int) bool {
			if desc {
				return objects[i].Size > objects[j].Size
			}
			return objects[i].Size < objects[j].Size
		})
	case "date":
		sort.Slice(objects, func(i, j int) bool {
			if desc {
				return objects[i].LastModified > objects[j].LastModified
			}
			return objects[i].LastModified < objects[j].LastModified
		})
	}
}

// ObjectURL builds an access URL for a given object.
//
// Public buckets get a plain permanent URL.
func (a *OssAdapter) ObjectURL(bucketName, objectKey string, expiresIn int64) (string, error) {
	publicURL, err := a.validateObjectURLRequest(bucketName, objectKey, expiresIn)
	if err != nil {
		return "", err
	}
	if publicURL != "" {
		return publicURL, nil
	}

	client, err := a.client()
	if err != nil {
		return "", fmt.Errorf("create oss client: %w", err)
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", fmt.Errorf("open bucket %s: %w", bucketName, err)
	}

	signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, expiresIn)
	if err != nil {
		return "", fmt.Errorf("sign object url for %s/%s: %w", bucketName, objectKey, err)
	}
	return beautifyPath(signedURL), nil
}

func (a *OssAdapter) validateObjectURLRequest(bucketName, objectKey string, expiresIn int64) (string, error) {
	if !isPrivateBucket(bucketName) {
		return a.publicObjectURL(bucketName, objectKey), nil
	}
	if expiresIn <= 0 {
		return "", fmt.Errorf("private object url expiry must be positive")
	}
	return "", nil
}

// host extracts the host (without scheme) from the endpoint.
func (a *OssAdapter) host() string {
	h := a.Endpoint
	for _, prefix := range []string{"https://", "http://"} {
		if len(h) > len(prefix) && h[:len(prefix)] == prefix {
			h = h[len(prefix):]
			break
		}
	}
	return h
}

func (a *OssAdapter) publicObjectURL(bucketName, objectKey string) string {
	u := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s.%s", bucketName, a.host()),
		Path:   "/" + objectKey,
	}
	return u.String()
}

func isPrivateBucket(bucketName string) bool {
	return strings.HasSuffix(bucketName, "-private") || bucketName == "quanttide-terraform-state"
}

// beautifyPath restores "/" in the path part of a signed URL.
//
// The OSS SDK encodes the object key's "/" as "%2F" (QueryEscape behavior).
// Functionally equivalent, but ugly. We revert it in the path portion only —
// the query string (which carries the signature) must stay untouched,
// otherwise the signature becomes invalid.
func beautifyPath(rawURL string) string {
	idx := strings.Index(rawURL, "?")
	if idx == -1 {
		// no query string, whole URL is path
		return strings.ReplaceAll(rawURL, "%2F", "/")
	}
	path := rawURL[:idx]
	query := rawURL[idx:]
	return strings.ReplaceAll(path, "%2F", "/") + query
}
