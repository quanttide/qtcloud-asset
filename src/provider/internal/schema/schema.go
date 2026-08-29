package schema

// HealthResponse is the health check response.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// RootResponse is the root endpoint response.
type RootResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// ConfigResponse is the configuration endpoint response.
type ConfigResponse struct {
	ProviderBaseURL string `json:"provider_base_url"`
	StudioOrigin    string `json:"studio_origin"`
	CORS            string `json:"cors"`
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Bucket represents a single OSS bucket.
type Bucket struct {
	Name         string `json:"name"`
	Region       string `json:"region"`
	StorageClass string `json:"storage_class"`
	CreatedAt    string `json:"created_at"`
}

// BucketListResponse is the response for the /buckets endpoint.
type BucketListResponse struct {
	Buckets []Bucket `json:"buckets"`
	Total   int      `json:"total"`
}

// Object represents a single object (file) inside a bucket.
type Object struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	Type         string `json:"type"`
	StorageClass string `json:"storage_class"`
	LastModified string `json:"last_modified"`
}

// ObjectListResponse is the response for the /buckets/{name}/objects endpoint.
type ObjectListResponse struct {
	Bucket     string   `json:"bucket"`
	Objects    []Object `json:"objects"`
	Total      int      `json:"total"`
	NextMarker string   `json:"next_marker,omitempty"` // 下一页 marker，空表示已到末尾
	Truncated  bool     `json:"truncated"`             // 是否被 limit 截断
}

// ListObjectsParams carries query parameters for listing objects.
type ListObjectsParams struct {
	Prefix string // 对象 key 前缀过滤（OSS 原生）
	Sort   string // 排序字段：key / size / date（空 = 不排序）
	Order  string // asc / desc（默认 asc）
	Limit  int    // 每页数量，0 = 不限制（最大 1000）
	Marker string // 分页游标（OSS 原生）
}

// ObjectURLResponse is the response for the /buckets/{name}/objects/{key}/url endpoint.
type ObjectURLResponse struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expires_in"` // 有效期秒数（0 = 永久/公开）
}

// FolderShareResponse describes a public, read-only content share.
type FolderShareResponse struct {
	Token     string   `json:"token"`
	Title     string   `json:"title"`
	Bucket    string   `json:"bucket"`
	Prefixes  []string `json:"prefixes,omitempty"`
	Keys      []string `json:"keys,omitempty"`
	URL       string   `json:"url"`
	CreatedAt string   `json:"created_at"`
}

// FolderShareEnvelope is the response wrapper used by share metadata routes.
type FolderShareEnvelope struct {
	Share FolderShareResponse `json:"share"`
}

// FolderShareListResponse is the response for the authenticated share list.
type FolderShareListResponse struct {
	Shares []FolderShareResponse `json:"shares"`
	Total  int                   `json:"total"`
}
