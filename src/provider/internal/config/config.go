package config

import (
	"os"
)

// Config holds the provider configuration.
type Config struct {
	Port               string
	BaseURL            string
	StudioOrigin       string
	StudioOrigins      []string
	OSSEndpoint        string
	OSSAccessKeyID     string
	OSSAccessKeySecret string
	OSSSecurityToken   string
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
		Port:               getEnv("PROVIDER_PORT", "9000"),
		BaseURL:            getEnv("PROVIDER_BASE_URL", "https://api.quanttide.com/qtcloud-asset"),
		StudioOrigin:       studioOrigin,
		StudioOrigins:      studioOrigins,
		OSSEndpoint:        getEnv("OSS_ENDPOINT", "https://oss-cn-hangzhou.aliyuncs.com"),
		OSSAccessKeyID:     accessKeyID,
		OSSAccessKeySecret: accessKeySecret,
		OSSSecurityToken:   securityToken,
	}
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
