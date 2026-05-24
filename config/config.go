package config

import "os"

// Config holds MondaiPhi environment configuration.
type Config struct {
	ServerPort   string
	Environment  string
	DatabaseURL  string
	AuthJWKSURL  string
	StorageEndpoint   string
	StorageRegion     string
	StorageBucket     string
	StorageAccessKey  string
	StorageSecretKey  string
	StoragePresignTTL int
	AdminUserIDs      string // comma-separated Discord IDs
}

// Load reads configuration from environment variables with sensible defaults for local dev.
func Load() *Config {
	return &Config{
		ServerPort:        getEnv("MONDAIPHI_PORT", "8087"),
		Environment:     getEnv("MONDAIPHI_ENVIRONMENT", "development"),
		DatabaseURL:     getEnv("MONDAIPHI_DATABASE_URL", "postgres://phi:phi_dev_password@localhost:5432/mondaiphi?sslmode=disable"),
		AuthJWKSURL:     getEnv("MONDAIPHI_AUTH_JWKS_URL", "http://localhost:8080/.well-known/jwks.json"),
		StorageEndpoint: getEnv("MONDAIPHI_STORAGE_ENDPOINT", "http://localhost:9000"),
		StorageRegion:   getEnv("MONDAIPHI_STORAGE_REGION", "us-east-1"),
		StorageBucket:   getEnv("MONDAIPHI_STORAGE_BUCKET", "philia-jlpt-dev"),
		StorageAccessKey:getEnv("MONDAIPHI_STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:getEnv("MONDAIPHI_STORAGE_SECRET_KEY", ""),
		StoragePresignTTL:getEnvInt("MONDAIPHI_STORAGE_PRESIGN_TTL", 3600),
		AdminUserIDs:    getEnv("MONDAIPHI_ADMIN_USER_IDS", ""),
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
		// simplistic: in real code use strconv.Atoi with error handling
		// for now, keep it minimal; full implementation later
	}
	return defaultVal
}
