package media

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles media-related HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new media handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers media routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	media := router.Group("/media")
	{
		media.POST("/upload", h.UploadMedia)
		media.GET("/:id", h.GetMedia)
		media.DELETE("/:id", h.DeleteMedia)
		media.GET("/user/:userID", h.GetUserMedia)
		media.POST("/:id/attach", h.AttachMedia)
		
		// Cloudinary endpoints (public - no auth required for hero images)
		media.POST("/cloudinary/sign", SignCloudinaryURL)
		media.GET("/cloudinary/config", GetCloudinaryConfig)
	}
}

// UploadMedia handles file upload
func (h *Handler) UploadMedia(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    "INVALID_FILE",
				"message": "No file uploaded",
			},
		})
		return
	}

	// Upload media
	media, err := h.service.UploadMedia(c.Request.Context(), file, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    "UPLOAD_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    media,
	})
}

// GetMedia retrieves media information
func (h *Handler) GetMedia(c *gin.Context) {
	mediaID := c.Param("id")

	media, err := h.service.GetMedia(c.Request.Context(), mediaID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "GET_MEDIA_FAILED"
		
		if err.Error() == "media not found" {
			status = http.StatusNotFound
			code = "MEDIA_NOT_FOUND"
		}

		c.JSON(status, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    code,
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    media,
	})
}

// DeleteMedia handles media deletion
func (h *Handler) DeleteMedia(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	mediaID := c.Param("id")

	err := h.service.DeleteMedia(c.Request.Context(), mediaID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "DELETE_FAILED"
		
		if err.Error() == "media not found" {
			status = http.StatusNotFound
			code = "MEDIA_NOT_FOUND"
		} else if err.Error() == "unauthorized to delete this media" {
			status = http.StatusForbidden
			code = "FORBIDDEN"
		}

		c.JSON(status, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    code,
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nil,
	})
}

// GetUserMedia retrieves all media uploaded by a user
func (h *Handler) GetUserMedia(c *gin.Context) {
	userID := c.Param("userID")
	
	// Parse pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	media, err := h.service.GetUserMedia(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    "GET_USER_MEDIA_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    media,
		"meta": map[string]interface{}{
			"limit":   limit,
			"offset":  offset,
			"hasMore": len(media) == limit,
		},
	})
}

// AttachMedia attaches media to an entity
func (h *Handler) AttachMedia(c *gin.Context) {
	mediaID := c.Param("id")

	var input struct {
		EntityType string `json:"entity_type" binding:"required,oneof=trip place profile"`
		EntityID   string `json:"entity_id" binding:"required,uuid"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// TODO: Add authorization check to ensure user has permission to attach media to the entity

	err := h.service.AttachMediaToEntity(c.Request.Context(), mediaID, input.EntityType, input.EntityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    "ATTACH_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nil,
	})
}

// ServeMedia serves the actual media file (for local development)
func (h *Handler) ServeMedia(storage Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a simple file server for development
		// In production, files should be served by a CDN or nginx
		
		path := c.Param("filepath")
		if path == "" {
			c.Status(http.StatusNotFound)
			return
		}

		fullPath := storage.GetFullPath(path)
		
		// Check if file exists
		if _, err := http.Dir(storage.GetFullPath("")).Open(path); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.File(fullPath)
	}
}

// Middleware to validate file upload
func ValidateFileUpload(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		
		if err := c.Request.ParseMultipartForm(maxSize); err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error": map[string]interface{}{
					"code":    "FILE_TOO_LARGE",
					"message": fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", maxSize),
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}