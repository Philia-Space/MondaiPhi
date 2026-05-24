package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds MondaiPhi environment configuration.
type Config struct {
	ServerPort        string
	Environment       string
	DatabaseURL       string
	AuthJWKSURL       string
	StorageEndpoint   string
	StorageRegion     string
	StorageBucket     string
	StorageAccessKey  string
	StorageSecretKey  string
	StoragePresignTTL time.Duration
	UseSSL            bool
	AdminUserIDs      string // comma-separated Discord IDs
}

// Load reads configuration from environment variables with sensible defaults for local dev.
func Load() *Config {
	return &Config{
		ServerPort:        getEnv("MONDAIPHI_PORT", "8087"),
		Environment:       getEnv("MONDAIPHI_ENVIRONMENT", "development"),
		DatabaseURL:       getEnv("MONDAIPHI_DATABASE_URL", "postgres://phi:phi_dev_password@localhost:5432/mondaiphi?sslmode=disable"),
		AuthJWKSURL:       getEnv("MONDAIPHI_AUTH_JWKS_URL", "http://localhost:8080"),
		StorageEndpoint:   getEnv("MONDAIPHI_STORAGE_ENDPOINT", "nos.wjv-1.neo.id"),
		StorageRegion:     getEnv("MONDAIPHI_STORAGE_REGION", "us-east-1"),
		StorageBucket:     getEnv("MONDAIPHI_STORAGE_BUCKET", "philia-space"),
		StorageAccessKey:  getEnv("MONDAIPHI_STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:  getEnv("MONDAIPHI_STORAGE_SECRET_KEY", ""),
		StoragePresignTTL: getEnvDuration("MONDAIPHI_STORAGE_PRESIGN_TTL", 1*time.Hour),
		UseSSL:            getEnv("MONDAIPHI_STORAGE_USE_SSL", "true") == "true",
		AdminUserIDs:      getEnv("MONDAIPHI_ADMIN_USER_IDS", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
