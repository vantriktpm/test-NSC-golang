package models

import (
	"time"
)

// URL represents a shortened URL
type URL struct {
	ID          int       `json:"id" db:"id"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// Analytics represents click analytics for a URL
type Analytics struct {
	ID        int       `json:"id" db:"id"`
	URLID     int       `json:"url_id" db:"url_id"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Referer   string    `json:"referer" db:"referer"`
	ClickedAt time.Time `json:"clicked_at" db:"clicked_at"`
}

// ShortenRequest represents the request to shorten a URL
type ShortenRequest struct {
	URL      string     `json:"url" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ShortenResponse represents the response after shortening a URL
type ShortenResponse struct {
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// AnalyticsResponse represents analytics data for a URL
type AnalyticsResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	TotalClicks int    `json:"total_clicks"`
	UniqueIPs   int    `json:"unique_ips"`
	TopReferers []struct {
		Referer string `json:"referer"`
		Count   int    `json:"count"`
	} `json:"top_referers"`
	RecentClicks []Analytics `json:"recent_clicks"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
