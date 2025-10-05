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

// CreateInventoryTables creates the inventory management tables
func CreateInventoryTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
			total_stock INTEGER NOT NULL DEFAULT 0 CHECK (total_stock >= 0),
			available_stock INTEGER NOT NULL DEFAULT 0 CHECK (available_stock >= 0),
			reserved_stock INTEGER NOT NULL DEFAULT 0 CHECK (reserved_stock >= 0),
			version INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT check_stock_consistency CHECK (total_stock = available_stock + reserved_stock)
		)`,
		`CREATE TABLE IF NOT EXISTS user_reservations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			reserved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CONFIRMED', 'RELEASED', 'EXPIRED')),
			correlation_id UUID NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS inventory_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event_id UUID NOT NULL,
			event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('INVENTORY_CHECK', 'INVENTORY_RESERVE', 'INVENTORY_CONFIRM', 'INVENTORY_RELEASE', 'INVENTORY_RESTOCK')),
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			user_id UUID,
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			correlation_id UUID,
			processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			metadata JSONB
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			product_id UUID NOT NULL REFERENCES products(id),
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
			total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
			status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'COMPLETED')),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			correlation_id UUID NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_available_stock ON products(available_stock)`,
		`CREATE INDEX IF NOT EXISTS idx_user_reservations_user_id ON user_reservations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_reservations_product_id ON user_reservations(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_reservations_expires_at ON user_reservations(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_user_reservations_correlation_id ON user_reservations(correlation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_reservations_status ON user_reservations(status)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_events_product_id ON inventory_events(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_events_event_type ON inventory_events(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_events_processed_at ON inventory_events(processed_at)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_events_correlation_id ON inventory_events(correlation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_product_id ON orders(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_correlation_id ON orders(correlation_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create inventory table: %w", err)
		}
	}

	log.Println("Inventory tables created successfully")
	return nil
}
