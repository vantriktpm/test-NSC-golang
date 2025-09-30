package handlers

import (
	"net/http"

	"url-shortener/internal/models"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	urlService service.URLService
}

func NewAnalyticsHandler(urlService service.URLService) *AnalyticsHandler {
	return &AnalyticsHandler{
		urlService: urlService,
	}
}

// GetAnalytics handles GET /api/v1/analytics/{shortCode}
func (h *AnalyticsHandler) GetAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Short code is required",
		})
		return
	}

	analytics, err := h.urlService.GetAnalytics(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Analytics not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
