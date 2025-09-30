package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck handles GET /api/v1/health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": gin.H{},
		"service":   "url-shortener",
		"version":   "1.0.0",
	})
}
