package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

const appSecretKeySize = 32

// Config holds the provider configuration.
type Config struct {
	Port                    string
	BaseURL                 string
	StudioOrigin            string
	StudioOrigins           []string
	AppSecretKey            string
	AuthMode                string
	LocalAuthAccount        string
	LocalAuthEmail          string
	LocalAuthName           string
	LocalAuthRole           string
	LocalAuthPasswordHash   string
	OSSEndpoint             string
	OSSAccessKeyID          string
	OSSAccessKeySecret      string
	OSSSecurityToken        string
	RDSDriver               string
	RDSConnectionString     string
	UserStoreMode           string
	UserMigration           string
	ShareStoreMode          string
	ShareMigration          string
	ShareTokenEncryptionKey string
	ShareableBuckets        []string
}

// Load reads configuration from environment variables.
func Load() *Config {
	explicitAccessKeyID := getEnv("OSS_ACCESS_KEY_ID", "")
	explicitAccessKeySecret := getEnv("OSS_ACCESS_KEY_SECRET", "")

	var accessKeyID, accessKeySecret, securityToken string
	if explicitAccessKeyID != "" && explicitAccessKeySecret != "" {
		// Keep an explicitly configured credential set together.
		accessKeyID = explicitAccessKeyID
		accessKeySecret = explicitAccessKeySecret
		securityToken = getEnv("OSS_SESSION_TOKEN", "")
	} else {
		// FC injects short-lived execution-role credentials into ALIBABA_CLOUD_*.
		accessKeyID = getEnv("ALIBABA_CLOUD_ACCESS_KEY_ID", "")
		accessKeySecret = getEnv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "")
		securityToken = getEnv("ALIBABA_CLOUD_SECURITY_TOKEN", "")
		if securityToken == "" {
			securityToken = getEnv("ALIBABA_CLOUD_SESSION_TOKEN", "")
		}
	}

	studioOrigin := getEnv("STUDIO_ORIGIN", "https://asset.cloud.quanttide.com")
	studioOrigins := []string{
		"https://asset.cloud.quanttide.com",
		"https://asset.quanttide.com",
		"http://localhost:8080",
		"http://localhost:9000",
		"http://localhost:8090",
		"http://127.0.0.1:8090",
	}
	if !contains(studioOrigins, studioOrigin) {
		studioOrigins = append(studioOrigins, studioOrigin)
	}

	return &Config{
		Port:                    getEnv("PROVIDER_PORT", "9000"),
		BaseURL:                 getEnv("PROVIDER_BASE_URL", "https://api.quanttide.com/qtcloud-asset"),
		StudioOrigin:            studioOrigin,
		StudioOrigins:           studioOrigins,
		AppSecretKey:            getEnv("APP_SECRET_KEY", ""),
		AuthMode:                strings.ToLower(getEnv("AUTH_MODE", "sso")),
		LocalAuthAccount:        getEnv("LOCAL_AUTH_ACCOUNT", getEnv("LOCAL_AUTH_EMAIL", "")),
		LocalAuthEmail:          getEnv("LOCAL_AUTH_EMAIL", ""),
		LocalAuthName:           getEnv("LOCAL_AUTH_NAME", ""),
		LocalAuthRole:           getEnv("LOCAL_AUTH_ROLE", "admin"),
		LocalAuthPasswordHash:   getEnv("LOCAL_AUTH_PASSWORD_HASH", ""),
		OSSEndpoint:             getEnv("OSS_ENDPOINT", "https://oss-cn-hangzhou.aliyuncs.com"),
		OSSAccessKeyID:          accessKeyID,
		OSSAccessKeySecret:      accessKeySecret,
		OSSSecurityToken:        securityToken,
		RDSDriver:               getEnv("RDS_DRIVER", "postgres"),
		RDSConnectionString:     getEnv("RDS_CONNECTION_STRING", ""),
		UserStoreMode:           strings.ToLower(getEnv("USER_STORE", "rds")),
		UserMigration:           getEnv("USER_MIGRATION", ""),
		ShareStoreMode:          strings.ToLower(getEnv("SHARE_STORE", "rds")),
		ShareMigration:          getEnv("SHARE_MIGRATION", ""),
		ShareTokenEncryptionKey: getEnv("SHARE_TOKEN_ENCRYPTION_KEY", ""),
		ShareableBuckets:        parseCSV(getEnv("SHAREABLE_BUCKETS", "qtcloud-asset-studio")),
	}
}

// Validate checks startup configuration that is required for safe operation.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("provider config is missing")
	}
	_, err := ParseAppSecretKey(c.AppSecretKey)
	return err
}

// ParseAppSecretKey decodes and validates the application secret key.
func ParseAppSecretKey(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("APP_SECRET_KEY is missing")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("APP_SECRET_KEY must be standard base64: %w", err)
	}
	if len(key) != appSecretKeySize {
		return nil, fmt.Errorf("APP_SECRET_KEY must decode to %d bytes", appSecretKeySize)
	}
	return key, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func parseCSV(raw string) []string {
	seen := make(map[string]struct{})
	values := make([]string, 0)
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		values = append(values, item)
	}
	return values
}
