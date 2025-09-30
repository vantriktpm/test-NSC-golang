package handlers

import (
	"net/http"

	"url-shortener/internal/models"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	urlService service.URLService
}

func NewURLHandler(urlService service.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// ShortenURL handles POST /api/v1/shorten
func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.urlService.ShortenURL(req.URL, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to shorten URL",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// RedirectURL handles GET /{shortCode}
func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Short code is required",
		})
		return
	}

	// Get client information
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	originalURL, err := h.urlService.RedirectURL(shortCode, ipAddress, userAgent, referer)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "URL not found",
			Message: err.Error(),
		})
		return
	}

	// Redirect to original URL
	c.Redirect(http.StatusMovedPermanently, originalURL)
}
