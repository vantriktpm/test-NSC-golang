package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

// Initialize creates a new database connection
func Initialize(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Database connection established")
	return db, nil
}

// InitializeRedis creates a new Redis connection
func InitializeRedis(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	_, err = client.Ping(client.Context()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connection established")
	return client, nil
}

// CreateTables creates the necessary database tables
func CreateTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE,
			is_used BOOLEAN DEFAULT TRUE
		)`,
		`CREATE TABLE IF NOT EXISTS analytics (
			id SERIAL PRIMARY KEY,
			url_id INTEGER REFERENCES urls(id) ON DELETE CASCADE,
			ip_address INET,
			user_agent TEXT,
			referer TEXT,
			clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pre_generated_urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(8) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_used BOOLEAN DEFAULT FALSE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_url_id ON analytics(url_id)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_clicked_at ON analytics(clicked_at)`,
		`CREATE INDEX IF NOT EXISTS idx_pre_generated_urls_unused ON pre_generated_urls(is_used) WHERE is_used = FALSE`,
		`CREATE INDEX IF NOT EXISTS idx_pre_generated_urls_short_code ON pre_generated_urls(short_code)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	log.Println("Database tables created successfully")
	return nil
}
