package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
	LogLevel    string
	Environment string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/backend_path?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
