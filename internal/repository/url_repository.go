package repository

import (
	"database/sql"
	"fmt"
	"time"

	"url-shortener/internal/models"
)

type URLRepository interface {
	Create(url *models.URL) error
	GetByShortCode(shortCode string) (*models.URL, error)
	GetByID(id int) (*models.URL, error)
	Update(url *models.URL) error
	Delete(id int) error
	IsShortCodeExists(shortCode string) (bool, error)
}

type urlRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(url *models.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, expires_at, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	
	err := r.db.QueryRow(
		query,
		url.ShortCode,
		url.OriginalURL,
		url.ExpiresAt,
		url.IsActive,
	).Scan(&url.ID, &url.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}
	
	return nil
}

func (r *urlRepository) GetByShortCode(shortCode string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, is_active
		FROM urls
		WHERE short_code = $1 AND is_active = TRUE
	`
	
	url := &models.URL{}
	err := r.db.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	
	// Check if URL has expired
	if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("URL has expired")
	}
	
	return url, nil
}

func (r *urlRepository) GetByID(id int) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, is_active
		FROM urls
		WHERE id = $1
	`
	
	url := &models.URL{}
	err := r.db.QueryRow(query, id).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	
	return url, nil
}

func (r *urlRepository) Update(url *models.URL) error {
	query := `
		UPDATE urls
		SET original_url = $2, expires_at = $3, is_active = $4
		WHERE id = $1
	`
	
	_, err := r.db.Exec(query, url.ID, url.OriginalURL, url.ExpiresAt, url.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update URL: %w", err)
	}
	
	return nil
}

func (r *urlRepository) Delete(id int) error {
	query := `UPDATE urls SET is_active = FALSE WHERE id = $1`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}
	
	return nil
}

func (r *urlRepository) IsShortCodeExists(shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`
	
	var exists bool
	err := r.db.QueryRow(query, shortCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check short code existence: %w", err)
	}
	
	return exists, nil
}
