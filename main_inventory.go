package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/handlers"
	"url-shortener/internal/kafka"
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

	// Create inventory tables
	if err := database.CreateInventoryTables(db); err != nil {
		log.Fatal("Failed to create inventory tables:", err)
	}

	// Initialize Redis
	redisClient, err := database.InitializeRedis(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize Kafka Producer
	producerConfig := &kafka.ProducerConfig{
		Brokers:       cfg.Kafka.Brokers,
		TopicPrefix:   cfg.Kafka.TopicPrefix,
		RetryAttempts: cfg.Kafka.RetryAttempts,
		RetryDelay:    cfg.Kafka.RetryDelay,
	}

	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		log.Fatal("Failed to create Kafka producer:", err)
	}
	defer producer.Close()

	// Initialize Kafka Consumer
	consumerConfig := &kafka.ConsumerConfig{
		Brokers:           cfg.Kafka.Brokers,
		GroupID:           cfg.Kafka.ConsumerGroupID,
		TopicPrefix:       cfg.Kafka.TopicPrefix,
		SessionTimeout:    cfg.Kafka.SessionTimeout,
		HeartbeatInterval: cfg.Kafka.HeartbeatInterval,
		RetryAttempts:     cfg.Kafka.RetryAttempts,
		RetryDelay:        cfg.Kafka.RetryDelay,
	}

	consumer, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}
	defer consumer.Stop()

	// Initialize repositories
	urlRepo := repository.NewURLRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)

	// Initialize services
	urlService := service.NewURLService(urlRepo, analyticsRepo, redisClient, cfg.BaseURL)
	inventoryService := service.NewInventoryService(db, redisClient, producer, consumer, inventoryRepo)

	// Start URL pre-generation service
	if err := urlService.StartPreGeneration(); err != nil {
		log.Printf("Failed to start pre-generation service: %v", err)
	} else {
		log.Println("Pre-generation service started successfully")
	}

	// Start inventory processor
	ctx := context.Background()
	if err := inventoryService.StartInventoryProcessor(ctx); err != nil {
		log.Fatal("Failed to start inventory processor:", err)
	}
	defer inventoryService.StopInventoryProcessor()

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(urlService)
	analyticsHandler := handlers.NewAnalyticsHandler(urlService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)

	// Setup Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())

	// Health check
	router.GET("/api/v1/health", handlers.HealthCheck)

	// URL Shortener API routes
	urlAPI := router.Group("/api/v1")
	{
		urlAPI.POST("/shorten", urlHandler.ShortenURL)
		urlAPI.GET("/analytics/:shortCode", analyticsHandler.GetAnalytics)
	}

	// Inventory API routes
	inventoryAPI := router.Group("/api/v1/inventory")
	{
		// Product availability
		inventoryAPI.GET("/:productId/availability", inventoryHandler.CheckAvailability)
		inventoryAPI.GET("/:productId", inventoryHandler.GetProductInventory)

		// Inventory operations
		inventoryAPI.POST("/reserve", inventoryHandler.ReserveInventory)
		inventoryAPI.POST("/confirm/:orderId", inventoryHandler.ConfirmPurchase)
		inventoryAPI.POST("/release/:orderId", inventoryHandler.ReleaseReservation)

		// Bulk operations
		inventoryAPI.POST("/bulk-check", inventoryHandler.BulkCheckAvailability)

		// Metrics
		inventoryAPI.GET("/metrics", inventoryHandler.GetInventoryMetrics)
	}

	// Redirect route (must be last to avoid conflicts)
	router.GET("/:shortCode", urlHandler.RedirectURL)

	// Start server with graceful shutdown
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	// Create a channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop services
	inventoryService.StopInventoryProcessor()
	urlService.StopPreGeneration()

	log.Println("Server stopped")
}
