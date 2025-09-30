package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"url-shortener/internal/models"
	"url-shortener/internal/repository"

	"github.com/go-redis/redis/v8"
)

type URLService interface {
	ShortenURL(originalURL string, expiresAt *time.Time) (*models.ShortenResponse, error)
	RedirectURL(shortCode string, ipAddress, userAgent, referer string) (string, error)
	GetAnalytics(shortCode string) (*models.AnalyticsResponse, error)
}

type urlService struct {
	urlRepo       repository.URLRepository
	analyticsRepo repository.AnalyticsRepository
	redisClient   *redis.Client
	baseURL       string
}

func NewURLService(
	urlRepo repository.URLRepository,
	analyticsRepo repository.AnalyticsRepository,
	redisClient *redis.Client,
	baseURL string,
) URLService {
	return &urlService{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
		redisClient:   redisClient,
		baseURL:       baseURL,
	}
}

func (s *urlService) ShortenURL(originalURL string, expiresAt *time.Time) (*models.ShortenResponse, error) {
	// Validate URL
	if !isValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	// Generate unique short code
	shortCode, err := s.generateShortCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}

	// Create URL record
	urlModel := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	err = s.urlRepo.Create(urlModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Cache the URL in Redis
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	cacheValue := originalURL
	cacheExpiration := 24 * time.Hour

	if expiresAt != nil && expiresAt.Before(time.Now().Add(cacheExpiration)) {
		cacheExpiration = time.Until(*expiresAt)
	}

	err = s.redisClient.Set(context.Background(), cacheKey, cacheValue, cacheExpiration).Err()
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to cache URL: %v\n", err)
	}

	// Build response
	response := &models.ShortenResponse{
		ShortCode:   shortCode,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		OriginalURL: originalURL,
		CreatedAt:   urlModel.CreatedAt,
		ExpiresAt:   expiresAt,
	}

	return response, nil
}

func (s *urlService) RedirectURL(shortCode string, ipAddress, userAgent, referer string) (string, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("url:%s", shortCode)
	cachedURL, err := s.redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		// URL found in cache, record analytics
		go s.recordAnalytics(shortCode, ipAddress, userAgent, referer)
		return cachedURL, nil
	}

	// Get from database
	urlModel, err := s.urlRepo.GetByShortCode(shortCode)
	if err != nil {
		return "", fmt.Errorf("URL not found: %w", err)
	}

	// Cache the URL for future requests
	cacheExpiration := 24 * time.Hour
	if urlModel.ExpiresAt != nil && urlModel.ExpiresAt.Before(time.Now().Add(cacheExpiration)) {
		cacheExpiration = time.Until(*urlModel.ExpiresAt)
	}

	s.redisClient.Set(context.Background(), cacheKey, urlModel.OriginalURL, cacheExpiration)

	// Record analytics
	go s.recordAnalytics(shortCode, ipAddress, userAgent, referer)

	return urlModel.OriginalURL, nil
}

func (s *urlService) GetAnalytics(shortCode string) (*models.AnalyticsResponse, error) {
	// Get URL by short code
	urlModel, err := s.urlRepo.GetByShortCode(shortCode)
	if err != nil {
		return nil, fmt.Errorf("URL not found: %w", err)
	}

	// Get analytics data
	totalClicks, err := s.analyticsRepo.GetTotalClicks(urlModel.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total clicks: %w", err)
	}

	uniqueIPs, err := s.analyticsRepo.GetUniqueIPs(urlModel.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique IPs: %w", err)
	}

	topReferers, err := s.analyticsRepo.GetTopReferers(urlModel.ID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top referers: %w", err)
	}

	recentClicks, err := s.analyticsRepo.GetByURLID(urlModel.ID, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent clicks: %w", err)
	}

	// Build response
	response := &models.AnalyticsResponse{
		ShortCode:   shortCode,
		OriginalURL: urlModel.OriginalURL,
		TotalClicks: totalClicks,
		UniqueIPs:   uniqueIPs,
		RecentClicks: recentClicks,
	}

	// Convert top referers
	for _, referer := range topReferers {
		response.TopReferers = append(response.TopReferers, struct {
			Referer string `json:"referer"`
			Count   int    `json:"count"`
		}{
			Referer: referer.Referer,
			Count:   referer.Count,
		})
	}

	return response, nil
}

func (s *urlService) generateShortCode() (string, error) {
	const maxAttempts = 10
	
	for i := 0; i < maxAttempts; i++ {
		// Generate 4 random bytes
		bytes := make([]byte, 4)
		_, err := rand.Read(bytes)
		if err != nil {
			return "", err
		}
		
		// Convert to hex string and take first 8 characters
		shortCode := hex.EncodeToString(bytes)[:8]
		
		// Check if short code already exists
		exists, err := s.urlRepo.IsShortCodeExists(shortCode)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return shortCode, nil
		}
	}
	
	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxAttempts)
}

func (s *urlService) recordAnalytics(shortCode, ipAddress, userAgent, referer string) {
	// Get URL by short code
	urlModel, err := s.urlRepo.GetByShortCode(shortCode)
	if err != nil {
		return // Silently fail for analytics
	}

	// Create analytics record
	analytics := &models.Analytics{
		URLID:     urlModel.ID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
	}

	// Save analytics
	err = s.analyticsRepo.Create(analytics)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to record analytics: %v\n", err)
	}
}

func isValidURL(rawURL string) bool {
	// Add protocol if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}
