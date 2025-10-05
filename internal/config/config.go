package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port string

	// Database configuration
	DatabaseURL string

	// Redis configuration
	RedisURL string

	// Kafka configuration
	Kafka KafkaConfig

	// Base URL for the application
	BaseURL string

	// Inventory configuration
	Inventory InventoryConfig
}

// KafkaConfig holds Kafka-specific configuration
type KafkaConfig struct {
	Brokers           []string
	TopicPrefix       string
	ConsumerGroupID   string
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	RetryAttempts     int
	RetryDelay        time.Duration
}

// InventoryConfig holds inventory-specific configuration
type InventoryConfig struct {
	ReservationTimeout time.Duration
	CleanupInterval    time.Duration
	MinPoolSize        int
	MaxPoolSize        int
	PreGenBatchSize    int
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/url_shortener?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),

		Kafka: KafkaConfig{
			Brokers:           getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			TopicPrefix:       getEnv("KAFKA_TOPIC_PREFIX", ""),
			ConsumerGroupID:   getEnv("KAFKA_CONSUMER_GROUP_ID", "inventory-service"),
			SessionTimeout:    getEnvDuration("KAFKA_SESSION_TIMEOUT", 30*time.Second),
			HeartbeatInterval: getEnvDuration("KAFKA_HEARTBEAT_INTERVAL", 3*time.Second),
			RetryAttempts:     getEnvInt("KAFKA_RETRY_ATTEMPTS", 3),
			RetryDelay:        getEnvDuration("KAFKA_RETRY_DELAY", 1*time.Second),
		},

		Inventory: InventoryConfig{
			ReservationTimeout: getEnvDuration("INVENTORY_RESERVATION_TIMEOUT", 15*time.Minute),
			CleanupInterval:    getEnvDuration("INVENTORY_CLEANUP_INTERVAL", 5*time.Minute),
			MinPoolSize:        getEnvInt("INVENTORY_MIN_POOL_SIZE", 100),
			MaxPoolSize:        getEnvInt("INVENTORY_MAX_POOL_SIZE", 1000),
			PreGenBatchSize:    getEnvInt("INVENTORY_PRE_GEN_BATCH_SIZE", 50),
		},
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
