package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	BaseURL     string
	Port        string
	Environment string
	RateLimit   time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/urlshortener?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		RateLimit:   time.Duration(100) * time.Millisecond, // 10 requests per second
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
