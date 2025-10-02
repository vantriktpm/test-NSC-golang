-- Migration: Create pre_generated_urls table
-- This table stores pre-generated short codes for faster URL shortening

CREATE TABLE IF NOT EXISTS pre_generated_urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(8) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_used BOOLEAN DEFAULT FALSE
);

-- Add index for faster lookups
CREATE INDEX IF NOT EXISTS idx_pre_generated_urls_unused ON pre_generated_urls(is_used) WHERE is_used = FALSE;
CREATE INDEX IF NOT EXISTS idx_pre_generated_urls_short_code ON pre_generated_urls(short_code);

-- Add is_used column to urls table if it doesn't exist
ALTER TABLE urls ADD COLUMN IF NOT EXISTS is_used BOOLEAN DEFAULT TRUE;

