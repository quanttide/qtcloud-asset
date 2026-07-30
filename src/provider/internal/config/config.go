package config

import (
	"os"
)

// Config holds the provider configuration.
type Config struct {
	Port        string
	BaseURL     string
	StudioOrigins []string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Port:    getEnv("PROVIDER_PORT", "9000"),
		BaseURL: getEnv("PROVIDER_BASE_URL", "https://api.asset.quanttide.com"),
		StudioOrigins: []string{
			"https://asset.quanttide.com",
			"http://localhost:8080",
			"http://localhost:9000",
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
