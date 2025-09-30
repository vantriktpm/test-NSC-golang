package service

import (
	"testing"
	"time"

	"url-shortener/internal/models"
	"url-shortener/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Create(url *models.URL) error {
	args := m.Called(url)
	url.ID = 1
	url.CreatedAt = time.Now()
	return args.Error(0)
}

func (m *MockURLRepository) GetByShortCode(shortCode string) (*models.URL, error) {
	args := m.Called(shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.URL), args.Error(1)
}

func (m *MockURLRepository) GetByID(id int) (*models.URL, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.URL), args.Error(1)
}

func (m *MockURLRepository) Update(url *models.URL) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockURLRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockURLRepository) IsShortCodeExists(shortCode string) (bool, error) {
	args := m.Called(shortCode)
	return args.Bool(0), args.Error(1)
}

type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) Create(analytics *models.Analytics) error {
	args := m.Called(analytics)
	analytics.ID = 1
	analytics.ClickedAt = time.Now()
	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetByURLID(urlID int, limit int) ([]models.Analytics, error) {
	args := m.Called(urlID, limit)
	return args.Get(0).([]models.Analytics), args.Error(1)
}

func (m *MockAnalyticsRepository) GetTotalClicks(urlID int) (int, error) {
	args := m.Called(urlID)
	return args.Int(0), args.Error(1)
}

func (m *MockAnalyticsRepository) GetUniqueIPs(urlID int) (int, error) {
	args := m.Called(urlID)
	return args.Int(0), args.Error(1)
}

func (m *MockAnalyticsRepository) GetTopReferers(urlID int, limit int) ([]struct {
	Referer string
	Count   int
}, error) {
	args := m.Called(urlID, limit)
	return args.Get(0).([]struct {
		Referer string
		Count   int
	}), args.Error(1)
}

func TestURLService_ShortenURL(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	
	service := NewURLService(mockURLRepo, mockAnalyticsRepo, mockRedis, "http://localhost:8080")

	tests := []struct {
		name        string
		url         string
		expiresAt   *time.Time
		setupMocks  func()
		expectError bool
	}{
		{
			name: "Valid URL",
			url:  "https://example.com",
			setupMocks: func() {
				mockURLRepo.On("IsShortCodeExists", mock.AnythingOfType("string")).Return(false, nil)
				mockURLRepo.On("Create", mock.AnythingOfType("*models.URL")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Invalid URL",
			url:  "not-a-url",
			setupMocks: func() {
				// No mocks needed for invalid URL
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			response, err := service.ShortenURL(tt.url, tt.expiresAt)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.url, response.OriginalURL)
				assert.NotEmpty(t, response.ShortCode)
			}
			
			mockURLRepo.AssertExpectations(t)
		})
	}
}

func TestURLService_RedirectURL(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	
	service := NewURLService(mockURLRepo, mockAnalyticsRepo, mockRedis, "http://localhost:8080")

	tests := []struct {
		name        string
		shortCode   string
		setupMocks  func()
		expectError bool
		expectedURL string
	}{
		{
			name:      "Valid short code",
			shortCode: "abc123",
			setupMocks: func() {
				url := &models.URL{
					ID:          1,
					ShortCode:   "abc123",
					OriginalURL: "https://example.com",
					IsActive:    true,
				}
				mockURLRepo.On("GetByShortCode", "abc123").Return(url, nil)
			},
			expectError: false,
			expectedURL: "https://example.com",
		},
		{
			name:      "Invalid short code",
			shortCode: "invalid",
			setupMocks: func() {
				mockURLRepo.On("GetByShortCode", "invalid").Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			originalURL, err := service.RedirectURL(tt.shortCode, "127.0.0.1", "test-agent", "test-referer")
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, originalURL)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, originalURL)
			}
			
			mockURLRepo.AssertExpectations(t)
		})
	}
}
