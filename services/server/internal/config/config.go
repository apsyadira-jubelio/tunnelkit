package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	APISecret    string
	BaseDomain   string // e.g., "tunnel.example.com"
	TLSEnabled   bool
	TLSCertPath  string
	TLSKeyPath   string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://tunnelkit:tunnelkit@localhost:5432/tunnelkit?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
		APISecret:    getEnv("API_SECRET", "change-me-in-production"),
		BaseDomain:   getEnv("BASE_DOMAIN", "localhost:8080"),
		TLSEnabled:   getEnv("TLS_ENABLED", "false") == "true",
		TLSCertPath:  getEnv("TLS_CERT_PATH", ""),
		TLSKeyPath:   getEnv("TLS_KEY_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
