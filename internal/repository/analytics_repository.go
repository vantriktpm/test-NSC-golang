package repository

import (
	"database/sql"
	"fmt"

	"url-shortener/internal/models"
)

type AnalyticsRepository interface {
	Create(analytics *models.Analytics) error
	GetByURLID(urlID int, limit int) ([]models.Analytics, error)
	GetTotalClicks(urlID int) (int, error)
	GetUniqueIPs(urlID int) (int, error)
	GetTopReferers(urlID int, limit int) ([]struct {
		Referer string
		Count   int
	}, error)
}

type analyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) Create(analytics *models.Analytics) error {
	query := `
		INSERT INTO analytics (url_id, ip_address, user_agent, referer)
		VALUES ($1, $2, $3, $4)
		RETURNING id, clicked_at
	`
	
	err := r.db.QueryRow(
		query,
		analytics.URLID,
		analytics.IPAddress,
		analytics.UserAgent,
		analytics.Referer,
	).Scan(&analytics.ID, &analytics.ClickedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create analytics: %w", err)
	}
	
	return nil
}

func (r *analyticsRepository) GetByURLID(urlID int, limit int) ([]models.Analytics, error) {
	query := `
		SELECT id, url_id, ip_address, user_agent, referer, clicked_at
		FROM analytics
		WHERE url_id = $1
		ORDER BY clicked_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.Query(query, urlID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}
	defer rows.Close()
	
	var analytics []models.Analytics
	for rows.Next() {
		var a models.Analytics
		err := rows.Scan(
			&a.ID,
			&a.URLID,
			&a.IPAddress,
			&a.UserAgent,
			&a.Referer,
			&a.ClickedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analytics: %w", err)
		}
		analytics = append(analytics, a)
	}
	
	return analytics, nil
}

func (r *analyticsRepository) GetTotalClicks(urlID int) (int, error) {
	query := `SELECT COUNT(*) FROM analytics WHERE url_id = $1`
	
	var count int
	err := r.db.QueryRow(query, urlID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total clicks: %w", err)
	}
	
	return count, nil
}

func (r *analyticsRepository) GetUniqueIPs(urlID int) (int, error) {
	query := `SELECT COUNT(DISTINCT ip_address) FROM analytics WHERE url_id = $1`
	
	var count int
	err := r.db.QueryRow(query, urlID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unique IPs: %w", err)
	}
	
	return count, nil
}

func (r *analyticsRepository) GetTopReferers(urlID int, limit int) ([]struct {
	Referer string
	Count   int
}, error) {
	query := `
		SELECT referer, COUNT(*) as count
		FROM analytics
		WHERE url_id = $1 AND referer IS NOT NULL AND referer != ''
		GROUP BY referer
		ORDER BY count DESC
		LIMIT $2
	`
	
	rows, err := r.db.Query(query, urlID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top referers: %w", err)
	}
	defer rows.Close()
	
	var referers []struct {
		Referer string
		Count   int
	}
	
	for rows.Next() {
		var r struct {
			Referer string
			Count   int
		}
		err := rows.Scan(&r.Referer, &r.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan referer: %w", err)
		}
		referers = append(referers, r)
	}
	
	return referers, nil
}
