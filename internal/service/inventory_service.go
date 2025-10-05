package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"url-shortener/internal/kafka"
	"url-shortener/internal/models"
	"url-shortener/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// InventoryService handles inventory operations using Kafka
type InventoryService interface {
	CheckAvailability(ctx context.Context, req *models.ProductAvailabilityRequest) (*models.ProductAvailabilityResponse, error)
	ReserveInventory(ctx context.Context, req *models.PurchaseRequest) (*models.PurchaseResponse, error)
	ConfirmPurchase(ctx context.Context, orderID uuid.UUID, userID string) error
	ReleaseReservation(ctx context.Context, orderID uuid.UUID, userID string) error
	GetProductInventory(ctx context.Context, productID string) (*models.InventoryState, error)
	StartInventoryProcessor(ctx context.Context) error
	StopInventoryProcessor()
}

type inventoryService struct {
	db          *sql.DB
	redisClient *redis.Client
	producer    *kafka.Producer
	consumer    *kafka.Consumer
	repository  repository.InventoryRepository

	// State management
	stateCache map[string]*models.InventoryState
	stateMutex sync.RWMutex

	// Configuration
	reservationTimeout time.Duration
	cleanupInterval    time.Duration

	// Control
	stopChan chan bool
	running  bool
}

// NewInventoryService creates a new inventory service
func NewInventoryService(
	db *sql.DB,
	redisClient *redis.Client,
	producer *kafka.Producer,
	consumer *kafka.Consumer,
	repository repository.InventoryRepository,
) InventoryService {
	return &inventoryService{
		db:                 db,
		redisClient:        redisClient,
		producer:           producer,
		consumer:           consumer,
		repository:         repository,
		stateCache:         make(map[string]*models.InventoryState),
		reservationTimeout: 15 * time.Minute,
		cleanupInterval:    5 * time.Minute,
		stopChan:           make(chan bool),
	}
}

// CheckAvailability checks if a product has available inventory
func (s *inventoryService) CheckAvailability(ctx context.Context, req *models.ProductAvailabilityRequest) (*models.ProductAvailabilityResponse, error) {
	// Get current inventory state
	state, err := s.getInventoryState(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory state: %w", err)
	}

	// Check availability
	available := state.AvailableStock >= req.Quantity

	// Publish check event for analytics
	event := s.producer.CreateInventoryCheckEvent(
		req.ProductID,
		req.Quantity,
		"", // No user ID for availability checks
		map[string]interface{}{
			"available": available,
			"requested": req.Quantity,
		},
	)

	if err := s.producer.PublishInventoryEventAsync(ctx, event); err != nil {
		log.Printf("Failed to publish inventory check event: %v", err)
	}

	return &models.ProductAvailabilityResponse{
		ProductID:      req.ProductID,
		Available:      available,
		AvailableStock: state.AvailableStock,
		TotalStock:     state.TotalStock,
		ReservedStock:  state.ReservedStock,
	}, nil
}

// ReserveInventory reserves inventory for a user
func (s *inventoryService) ReserveInventory(ctx context.Context, req *models.PurchaseRequest) (*models.PurchaseResponse, error) {
	// Generate order ID
	orderID := uuid.New()

	// Create reservation event
	event := s.producer.CreateInventoryReserveEvent(
		req.ProductID,
		req.Quantity,
		req.UserID,
		map[string]interface{}{
			"orderId":       orderID.String(),
			"reservedUntil": time.Now().Add(s.reservationTimeout).Format(time.RFC3339),
		},
	)

	// Publish reservation event
	if err := s.producer.PublishInventoryEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to publish reservation event: %w", err)
	}

	// Wait for processing result (in real implementation, this would be async)
	// For now, we'll simulate a successful reservation
	reservedUntil := time.Now().Add(s.reservationTimeout)

	return &models.PurchaseResponse{
		Success:       true,
		OrderID:       orderID,
		ProductID:     req.ProductID,
		Quantity:      req.Quantity,
		ReservedUntil: reservedUntil,
		Message:       "Inventory reserved successfully",
	}, nil
}

// ConfirmPurchase confirms a purchase and reduces inventory
func (s *inventoryService) ConfirmPurchase(ctx context.Context, orderID uuid.UUID, userID string) error {
	// Get reservation details
	reservation, err := s.repository.GetReservationByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation.UserID != userID {
		return fmt.Errorf("unauthorized: reservation belongs to different user")
	}

	if reservation.Status != "ACTIVE" {
		return fmt.Errorf("reservation is not active")
	}

	// Create confirmation event
	event := s.producer.CreateInventoryConfirmEvent(
		reservation.ProductID,
		reservation.Quantity,
		reservation.UserID,
		reservation.CorrelationID,
		map[string]interface{}{
			"orderId": orderID.String(),
		},
	)

	// Publish confirmation event
	if err := s.producer.PublishInventoryEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to publish confirmation event: %w", err)
	}

	return nil
}

// ReleaseReservation releases a reservation
func (s *inventoryService) ReleaseReservation(ctx context.Context, orderID uuid.UUID, userID string) error {
	// Get reservation details
	reservation, err := s.repository.GetReservationByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation.UserID != userID {
		return fmt.Errorf("unauthorized: reservation belongs to different user")
	}

	// Create release event
	event := s.producer.CreateInventoryReleaseEvent(
		reservation.ProductID,
		reservation.Quantity,
		reservation.UserID,
		reservation.CorrelationID,
		map[string]interface{}{
			"orderId": orderID.String(),
		},
	)

	// Publish release event
	if err := s.producer.PublishInventoryEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to publish release event: %w", err)
	}

	return nil
}

// GetProductInventory gets the current inventory state for a product
func (s *inventoryService) GetProductInventory(ctx context.Context, productID string) (*models.InventoryState, error) {
	return s.getInventoryState(ctx, productID)
}

// StartInventoryProcessor starts the inventory event processor
func (s *inventoryService) StartInventoryProcessor(ctx context.Context) error {
	if s.running {
		return fmt.Errorf("inventory processor is already running")
	}

	s.running = true

	// Register event handlers
	s.registerEventHandlers()

	// Start consumer
	topics := []string{"inventory-events"}
	if err := s.consumer.Start(ctx, topics); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	// Start cleanup routine
	go s.startCleanupRoutine(ctx)

	log.Println("Inventory processor started successfully")
	return nil
}

// StopInventoryProcessor stops the inventory event processor
func (s *inventoryService) StopInventoryProcessor() {
	if !s.running {
		return
	}

	s.running = false
	s.stopChan <- true

	if err := s.consumer.Stop(); err != nil {
		log.Printf("Error stopping consumer: %v", err)
	}

	log.Println("Inventory processor stopped")
}

// registerEventHandlers registers all event handlers
func (s *inventoryService) registerEventHandlers() {
	// Reserve event handler
	reserveHandler := kafka.NewInventoryEventHandler(
		models.InventoryEventTypeReserve,
		s.handleReserveEvent,
	)
	s.consumer.RegisterHandler(reserveHandler)

	// Confirm event handler
	confirmHandler := kafka.NewInventoryEventHandler(
		models.InventoryEventTypeConfirm,
		s.handleConfirmEvent,
	)
	s.consumer.RegisterHandler(confirmHandler)

	// Release event handler
	releaseHandler := kafka.NewInventoryEventHandler(
		models.InventoryEventTypeRelease,
		s.handleReleaseEvent,
	)
	s.consumer.RegisterHandler(releaseHandler)
}

// handleReserveEvent handles inventory reserve events
func (s *inventoryService) handleReserveEvent(ctx context.Context, event *models.InventoryEvent) error {
	// Get current product state
	product, err := s.repository.GetProduct(ctx, event.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Check if enough inventory is available
	if product.AvailableStock < event.Quantity {
		log.Printf("Insufficient inventory for product %s: requested %d, available %d",
			event.ProductID, event.Quantity, product.AvailableStock)
		return nil // Don't fail, just log
	}

	// Create reservation
	reservation := &models.UserReservation{
		ID:            uuid.New().String(),
		UserID:        event.UserID,
		ProductID:     event.ProductID,
		Quantity:      event.Quantity,
		ReservedAt:    time.Now(),
		ExpiresAt:     time.Now().Add(s.reservationTimeout),
		Status:        "ACTIVE",
		CorrelationID: event.CorrelationID,
	}

	// Save reservation
	if err := s.repository.CreateReservation(ctx, reservation); err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	// Update product inventory
	if err := s.repository.UpdateProductInventory(ctx, event.ProductID,
		product.AvailableStock-event.Quantity,
		product.ReservedStock+event.Quantity,
		product.Version+1); err != nil {
		return fmt.Errorf("failed to update product inventory: %w", err)
	}

	// Update cache
	s.updateStateCache(event.ProductID, &models.InventoryState{
		ProductID:      event.ProductID,
		AvailableStock: product.AvailableStock - event.Quantity,
		ReservedStock:  product.ReservedStock + event.Quantity,
		TotalStock:     product.TotalStock,
		LastUpdated:    time.Now(),
		Version:        product.Version + 1,
	})

	// Publish state update
	state := s.getStateFromCache(event.ProductID)
	if state != nil {
		s.producer.PublishInventoryState(ctx, state)
	}

	log.Printf("Successfully reserved %d units of product %s for user %s",
		event.Quantity, event.ProductID, event.UserID)
	return nil
}

// handleConfirmEvent handles inventory confirm events
func (s *inventoryService) handleConfirmEvent(ctx context.Context, event *models.InventoryEvent) error {
	// Get reservation
	reservation, err := s.repository.GetReservationByCorrelationID(ctx, event.CorrelationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	// Update reservation status
	if err := s.repository.UpdateReservationStatus(ctx, reservation.ID, "CONFIRMED"); err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Update product inventory (move from reserved to sold)
	product, err := s.repository.GetProduct(ctx, event.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if err := s.repository.UpdateProductInventory(ctx, event.ProductID,
		product.AvailableStock,
		product.ReservedStock-event.Quantity,
		product.Version+1); err != nil {
		return fmt.Errorf("failed to update product inventory: %w", err)
	}

	// Update cache
	s.updateStateCache(event.ProductID, &models.InventoryState{
		ProductID:      event.ProductID,
		AvailableStock: product.AvailableStock,
		ReservedStock:  product.ReservedStock - event.Quantity,
		TotalStock:     product.TotalStock,
		LastUpdated:    time.Now(),
		Version:        product.Version + 1,
	})

	// Publish state update
	state := s.getStateFromCache(event.ProductID)
	if state != nil {
		s.producer.PublishInventoryState(ctx, state)
	}

	log.Printf("Successfully confirmed purchase of %d units of product %s for user %s",
		event.Quantity, event.ProductID, event.UserID)
	return nil
}

// handleReleaseEvent handles inventory release events
func (s *inventoryService) handleReleaseEvent(ctx context.Context, event *models.InventoryEvent) error {
	// Get reservation
	reservation, err := s.repository.GetReservationByCorrelationID(ctx, event.CorrelationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	// Update reservation status
	if err := s.repository.UpdateReservationStatus(ctx, reservation.ID, "RELEASED"); err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Update product inventory (move from reserved back to available)
	product, err := s.repository.GetProduct(ctx, event.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if err := s.repository.UpdateProductInventory(ctx, event.ProductID,
		product.AvailableStock+event.Quantity,
		product.ReservedStock-event.Quantity,
		product.Version+1); err != nil {
		return fmt.Errorf("failed to update product inventory: %w", err)
	}

	// Update cache
	s.updateStateCache(event.ProductID, &models.InventoryState{
		ProductID:      event.ProductID,
		AvailableStock: product.AvailableStock + event.Quantity,
		ReservedStock:  product.ReservedStock - event.Quantity,
		TotalStock:     product.TotalStock,
		LastUpdated:    time.Now(),
		Version:        product.Version + 1,
	})

	// Publish state update
	state := s.getStateFromCache(event.ProductID)
	if state != nil {
		s.producer.PublishInventoryState(ctx, state)
	}

	log.Printf("Successfully released %d units of product %s for user %s",
		event.Quantity, event.ProductID, event.UserID)
	return nil
}

// Helper methods

func (s *inventoryService) getInventoryState(ctx context.Context, productID string) (*models.InventoryState, error) {
	// Try cache first
	if state := s.getStateFromCache(productID); state != nil {
		return state, nil
	}

	// Get from database
	product, err := s.repository.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	state := &models.InventoryState{
		ProductID:      productID,
		AvailableStock: product.AvailableStock,
		ReservedStock:  product.ReservedStock,
		TotalStock:     product.TotalStock,
		LastUpdated:    product.UpdatedAt,
		Version:        product.Version,
	}

	// Update cache
	s.updateStateCache(productID, state)
	return state, nil
}

func (s *inventoryService) getStateFromCache(productID string) *models.InventoryState {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.stateCache[productID]
}

func (s *inventoryService) updateStateCache(productID string, state *models.InventoryState) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	s.stateCache[productID] = state
}

func (s *inventoryService) startCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredReservations(ctx)
		case <-s.stopChan:
			return
		}
	}
}

func (s *inventoryService) cleanupExpiredReservations(ctx context.Context) {
	expiredReservations, err := s.repository.GetExpiredReservations(ctx)
	if err != nil {
		log.Printf("Failed to get expired reservations: %v", err)
		return
	}

	for _, reservation := range expiredReservations {
		// Create release event for expired reservation
		event := s.producer.CreateInventoryReleaseEvent(
			reservation.ProductID,
			reservation.Quantity,
			reservation.UserID,
			reservation.CorrelationID,
			map[string]interface{}{
				"reason": "expired",
			},
		)

		if err := s.producer.PublishInventoryEvent(ctx, event); err != nil {
			log.Printf("Failed to publish release event for expired reservation: %v", err)
		}
	}

	if len(expiredReservations) > 0 {
		log.Printf("Cleaned up %d expired reservations", len(expiredReservations))
	}
}
