package models

import (
	"time"

	"github.com/google/uuid"
)

// InventoryEventType represents the type of inventory event
type InventoryEventType string

const (
	InventoryEventTypeCheck   InventoryEventType = "INVENTORY_CHECK"
	InventoryEventTypeReserve InventoryEventType = "INVENTORY_RESERVE"
	InventoryEventTypeConfirm InventoryEventType = "INVENTORY_CONFIRM"
	InventoryEventTypeRelease InventoryEventType = "INVENTORY_RELEASE"
	InventoryEventTypeRestock InventoryEventType = "INVENTORY_RESTOCK"
)

// InventoryEvent represents an inventory event message
type InventoryEvent struct {
	EventID       uuid.UUID              `json:"eventId"`
	EventType     InventoryEventType     `json:"eventType"`
	ProductID     string                 `json:"productId"`
	UserID        string                 `json:"userId,omitempty"`
	Quantity      int                    `json:"quantity"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID uuid.UUID              `json:"correlationId"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// InventoryState represents the current state of a product's inventory
type InventoryState struct {
	ProductID      string    `json:"productId"`
	AvailableStock int       `json:"availableStock"`
	ReservedStock  int       `json:"reservedStock"`
	TotalStock     int       `json:"totalStock"`
	LastUpdated    time.Time `json:"lastUpdated"`
	Version        int       `json:"version"`
}

// Product represents a product in the system
type Product struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	Price          float64   `json:"price" db:"price"`
	TotalStock     int       `json:"totalStock" db:"total_stock"`
	AvailableStock int       `json:"availableStock" db:"available_stock"`
	ReservedStock  int       `json:"reservedStock" db:"reserved_stock"`
	Version        int       `json:"version" db:"version"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

// UserReservation represents a user's reservation of inventory
type UserReservation struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"userId" db:"user_id"`
	ProductID     string    `json:"productId" db:"product_id"`
	Quantity      int       `json:"quantity" db:"quantity"`
	ReservedAt    time.Time `json:"reservedAt" db:"reserved_at"`
	ExpiresAt     time.Time `json:"expiresAt" db:"expires_at"`
	Status        string    `json:"status" db:"status"`
	CorrelationID uuid.UUID `json:"correlationId" db:"correlation_id"`
}

// InventoryEventRequest represents a request to process an inventory event
type InventoryEventRequest struct {
	EventType InventoryEventType     `json:"eventType" binding:"required"`
	ProductID string                 `json:"productId" binding:"required"`
	Quantity  int                    `json:"quantity" binding:"required,min=1"`
	UserID    string                 `json:"userId,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// InventoryEventResponse represents the response to an inventory event
type InventoryEventResponse struct {
	Success        bool      `json:"success"`
	EventID        uuid.UUID `json:"eventId"`
	CorrelationID  uuid.UUID `json:"correlationId"`
	ProductID      string    `json:"productId"`
	AvailableStock int       `json:"availableStock"`
	ReservedStock  int       `json:"reservedStock"`
	Message        string    `json:"message,omitempty"`
	Error          string    `json:"error,omitempty"`
}

// ProductAvailabilityRequest represents a request to check product availability
type ProductAvailabilityRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// ProductAvailabilityResponse represents the response to availability check
type ProductAvailabilityResponse struct {
	ProductID      string `json:"productId"`
	Available      bool   `json:"available"`
	AvailableStock int    `json:"availableStock"`
	TotalStock     int    `json:"totalStock"`
	ReservedStock  int    `json:"reservedStock"`
}

// PurchaseRequest represents a purchase request
type PurchaseRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	UserID    string `json:"userId" binding:"required"`
}

// PurchaseResponse represents the response to a purchase request
type PurchaseResponse struct {
	Success       bool      `json:"success"`
	OrderID       uuid.UUID `json:"orderId"`
	ProductID     string    `json:"productId"`
	Quantity      int       `json:"quantity"`
	ReservedUntil time.Time `json:"reservedUntil"`
	Message       string    `json:"message,omitempty"`
	Error         string    `json:"error,omitempty"`
}
