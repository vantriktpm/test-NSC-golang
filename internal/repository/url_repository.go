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

	// Pre-generated URL methods
	CreatePreGeneratedURL(shortCode string) error
	GetUnusedPreGeneratedURL() (*models.PreGeneratedURL, error)
	MarkPreGeneratedURLAsUsed(shortCode string) error
	GetPreGeneratedURLCount() (int, error)
}

type urlRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(url *models.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, expires_at, is_active, is_used)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		url.ShortCode,
		url.OriginalURL,
		url.ExpiresAt,
		url.IsActive,
		url.IsUsed,
	).Scan(&url.ID, &url.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

func (r *urlRepository) GetByShortCode(shortCode string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, is_active, is_used
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
		&url.IsUsed,
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
		SELECT id, short_code, original_url, created_at, expires_at, is_active, is_used
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
		&url.IsUsed,
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

// Pre-generated URL methods
func (r *urlRepository) CreatePreGeneratedURL(shortCode string) error {
	query := `
		INSERT INTO pre_generated_urls (short_code, is_used)
		VALUES ($1, FALSE)
	`

	_, err := r.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to create pre-generated URL: %w", err)
	}

	return nil
}

func (r *urlRepository) GetUnusedPreGeneratedURL() (*models.PreGeneratedURL, error) {
	query := `
		SELECT short_code, created_at, is_used
		FROM pre_generated_urls
		WHERE is_used = FALSE
		ORDER BY created_at ASC
		LIMIT 1
	`

	url := &models.PreGeneratedURL{}
	err := r.db.QueryRow(query).Scan(
		&url.ShortCode,
		&url.CreatedAt,
		&url.IsUsed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no unused pre-generated URL available")
		}
		return nil, fmt.Errorf("failed to get unused pre-generated URL: %w", err)
	}

	return url, nil
}

func (r *urlRepository) MarkPreGeneratedURLAsUsed(shortCode string) error {
	query := `UPDATE pre_generated_urls SET is_used = TRUE WHERE short_code = $1`

	_, err := r.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to mark pre-generated URL as used: %w", err)
	}

	return nil
}

func (r *urlRepository) GetPreGeneratedURLCount() (int, error) {
	query := `SELECT COUNT(*) FROM pre_generated_urls WHERE is_used = FALSE`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get pre-generated URL count: %w", err)
	}

	return count, nil
}
