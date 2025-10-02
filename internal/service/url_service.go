package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"url-shortener/internal/models"
	"url-shortener/internal/repository"

	"github.com/go-redis/redis/v8"
)

type URLService interface {
	ShortenURL(originalURL string, expiresAt *time.Time) (*models.ShortenResponse, error)
	RedirectURL(shortCode string, ipAddress, userAgent, referer string) (string, error)
	GetAnalytics(shortCode string) (*models.AnalyticsResponse, error)
	StartPreGeneration() error
	StopPreGeneration()
}

type urlService struct {
	urlRepo       repository.URLRepository
	analyticsRepo repository.AnalyticsRepository
	redisClient   *redis.Client
	baseURL       string

	// Pre-generation management
	preGenMutex     sync.RWMutex
	stopPreGen      chan bool
	isPreGenRunning bool

	// Configuration
	minPoolSize     int
	maxPoolSize     int
	preGenBatchSize int
}

func NewURLService(
	urlRepo repository.URLRepository,
	analyticsRepo repository.AnalyticsRepository,
	redisClient *redis.Client,
	baseURL string,
) URLService {
	return &urlService{
		urlRepo:         urlRepo,
		analyticsRepo:   analyticsRepo,
		redisClient:     redisClient,
		baseURL:         baseURL,
		stopPreGen:      make(chan bool),
		minPoolSize:     100,  // Minimum pool size
		maxPoolSize:     1000, // Maximum pool size
		preGenBatchSize: 50,   // Batch size for pre-generation
	}
}

func (s *urlService) ShortenURL(originalURL string, expiresAt *time.Time) (*models.ShortenResponse, error) {
	// Validate URL
	if !isValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	// Check if URL already exists using Redis cache for fast lookup
	existingShortCode, err := s.getExistingShortCode(originalURL)
	if err == nil && existingShortCode != "" {
		// URL already exists, return existing short code
		response := &models.ShortenResponse{
			ShortCode:   existingShortCode,
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, existingShortCode),
			OriginalURL: originalURL,
			CreatedAt:   time.Now(), // We don't have the exact creation time from cache
			ExpiresAt:   expiresAt,
		}
		return response, nil
	}

	// Get pre-generated short code
	preGenURL, err := s.urlRepo.GetUnusedPreGeneratedURL()
	if err != nil {
		// Fallback to generating new short code if no pre-generated ones available
		shortCode, genErr := s.generateShortCode()
		if genErr != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", genErr)
		}
		preGenURL = &models.PreGeneratedURL{ShortCode: shortCode}
	}

	// Mark pre-generated URL as used
	err = s.urlRepo.MarkPreGeneratedURLAsUsed(preGenURL.ShortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to mark pre-generated URL as used: %w", err)
	}

	// Create URL record
	urlModel := &models.URL{
		ShortCode:   preGenURL.ShortCode,
		OriginalURL: originalURL,
		ExpiresAt:   expiresAt,
		IsActive:    true,
		IsUsed:      true,
	}

	err = s.urlRepo.Create(urlModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Cache the URL in Redis with both directions
	cacheKey := fmt.Sprintf("url:%s", preGenURL.ShortCode)
	cacheValue := originalURL
	cacheExpiration := 24 * time.Hour

	if expiresAt != nil && expiresAt.Before(time.Now().Add(cacheExpiration)) {
		cacheExpiration = time.Until(*expiresAt)
	}

	err = s.redisClient.Set(context.Background(), cacheKey, cacheValue, cacheExpiration).Err()
	if err != nil {
		fmt.Printf("Failed to cache URL: %v\n", err)
	}

	// Also cache reverse mapping for duplicate detection
	reverseCacheKey := fmt.Sprintf("reverse:%s", originalURL)
	err = s.redisClient.Set(context.Background(), reverseCacheKey, preGenURL.ShortCode, cacheExpiration).Err()
	if err != nil {
		fmt.Printf("Failed to cache reverse URL mapping: %v\n", err)
	}

	// Trigger pre-generation if pool is low
	go s.checkAndRefillPool()

	// Build response
	response := &models.ShortenResponse{
		ShortCode:   preGenURL.ShortCode,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, preGenURL.ShortCode),
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
		ShortCode:    shortCode,
		OriginalURL:  urlModel.OriginalURL,
		TotalClicks:  totalClicks,
		UniqueIPs:    uniqueIPs,
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

// Helper methods for pre-generation and optimization
func (s *urlService) getExistingShortCode(originalURL string) (string, error) {
	reverseCacheKey := fmt.Sprintf("reverse:%s", originalURL)
	shortCode, err := s.redisClient.Get(context.Background(), reverseCacheKey).Result()
	if err != nil {
		return "", err
	}
	return shortCode, nil
}

func (s *urlService) checkAndRefillPool() {
	count, err := s.urlRepo.GetPreGeneratedURLCount()
	if err != nil {
		fmt.Printf("Failed to get pre-generated URL count: %v\n", err)
		return
	}

	if count < s.minPoolSize {
		go s.preGenerateURLs(s.preGenBatchSize)
	}
}

func (s *urlService) preGenerateURLs(count int) {
	for i := 0; i < count; i++ {
		shortCode, err := s.generateShortCode()
		if err != nil {
			fmt.Printf("Failed to generate short code for pre-generation: %v\n", err)
			continue
		}

		err = s.urlRepo.CreatePreGeneratedURL(shortCode)
		if err != nil {
			fmt.Printf("Failed to create pre-generated URL: %v\n", err)
		}
	}
}

func (s *urlService) StartPreGeneration() error {
	s.preGenMutex.Lock()
	defer s.preGenMutex.Unlock()

	if s.isPreGenRunning {
		return fmt.Errorf("pre-generation is already running")
	}

	s.isPreGenRunning = true

	// Initial pre-generation
	go s.preGenerateURLs(s.maxPoolSize)

	// Start background pre-generation routine
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkAndRefillPool()
			case <-s.stopPreGen:
				return
			}
		}
	}()

	return nil
}

func (s *urlService) StopPreGeneration() {
	s.preGenMutex.Lock()
	defer s.preGenMutex.Unlock()

	if !s.isPreGenRunning {
		return
	}

	s.isPreGenRunning = false
	s.stopPreGen <- true
}
