package main

import (
	"log"
	"os"

	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/handlers"
	"url-shortener/internal/middleware"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create tables
	if err := database.CreateTables(db); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Initialize Redis
	redisClient, err := database.InitializeRedis(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize repository
	urlRepo := repository.NewURLRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Initialize service
	urlService := service.NewURLService(urlRepo, analyticsRepo, redisClient, cfg.BaseURL)

	// Start pre-generation service
	if err := urlService.StartPreGeneration(); err != nil {
		log.Printf("Failed to start pre-generation service: %v", err)
	} else {
		log.Println("Pre-generation service started successfully")
	}

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(urlService)
	analyticsHandler := handlers.NewAnalyticsHandler(urlService)

	// Setup Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())

	// Health check
	router.GET("/api/v1/health", handlers.HealthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/shorten", urlHandler.ShortenURL)
		api.GET("/analytics/:shortCode", analyticsHandler.GetAnalytics)
	}

	// Redirect route (must be last to avoid conflicts)
	router.GET("/:shortCode", urlHandler.RedirectURL)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
