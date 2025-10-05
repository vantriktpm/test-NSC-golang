package handlers

import (
	"net/http"
	"strconv"

	"url-shortener/internal/models"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InventoryHandler handles inventory-related HTTP requests
type InventoryHandler struct {
	inventoryService service.InventoryService
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// CheckAvailability handles GET /api/v1/inventory/:productId/availability
func (h *InventoryHandler) CheckAvailability(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Product ID is required",
		})
		return
	}

	quantityStr := c.Query("quantity")
	quantity := 1 // Default quantity
	if quantityStr != "" {
		var err error
		quantity, err = strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid quantity parameter",
			})
			return
		}
	}

	req := &models.ProductAvailabilityRequest{
		ProductID: productID,
		Quantity:  quantity,
	}

	response, err := h.inventoryService.CheckAvailability(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to check availability",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReserveInventory handles POST /api/v1/inventory/reserve
func (h *InventoryHandler) ReserveInventory(c *gin.Context) {
	var req models.PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.inventoryService.ReserveInventory(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to reserve inventory",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ConfirmPurchase handles POST /api/v1/inventory/confirm/:orderId
func (h *InventoryHandler) ConfirmPurchase(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid order ID",
		})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	err = h.inventoryService.ConfirmPurchase(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to confirm purchase",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase confirmed successfully",
	})
}

// ReleaseReservation handles POST /api/v1/inventory/release/:orderId
func (h *InventoryHandler) ReleaseReservation(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid order ID",
		})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	err = h.inventoryService.ReleaseReservation(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to release reservation",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reservation released successfully",
	})
}

// GetProductInventory handles GET /api/v1/inventory/:productId
func (h *InventoryHandler) GetProductInventory(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Product ID is required",
		})
		return
	}

	state, err := h.inventoryService.GetProductInventory(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get product inventory",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, state)
}

// BulkCheckAvailability handles POST /api/v1/inventory/bulk-check
func (h *InventoryHandler) BulkCheckAvailability(c *gin.Context) {
	var req struct {
		Products []struct {
			ProductID string `json:"productId" binding:"required"`
			Quantity  int    `json:"quantity" binding:"required,min=1"`
		} `json:"products" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	var responses []models.ProductAvailabilityResponse
	for _, product := range req.Products {
		availabilityReq := &models.ProductAvailabilityRequest{
			ProductID: product.ProductID,
			Quantity:  product.Quantity,
		}

		response, err := h.inventoryService.CheckAvailability(c.Request.Context(), availabilityReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Failed to check availability",
				Message: err.Error(),
			})
			return
		}

		responses = append(responses, *response)
	}

	c.JSON(http.StatusOK, gin.H{
		"results": responses,
	})
}

// GetInventoryMetrics handles GET /api/v1/inventory/metrics
func (h *InventoryHandler) GetInventoryMetrics(c *gin.Context) {
	// This would typically aggregate metrics from multiple products
	// For now, return a simple response
	c.JSON(http.StatusOK, gin.H{
		"totalProducts":    0,
		"totalStock":       0,
		"availableStock":   0,
		"reservedStock":    0,
		"lowStockProducts": 0,
		"lastUpdated":      "2024-01-01T00:00:00Z",
	})
}
