package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/Oferzz/newMap/apps/api/internal/database"
)

// Handler handles health check requests
type Handler struct {
	db    *sqlx.DB
	redis *database.RedisClient
}

// NewHandler creates a new health handler
func NewHandler(db *sqlx.DB, redis *database.RedisClient) *Handler {
	return &Handler{
		db:    db,
		redis: redis,
	}
}

// RegisterRoutes registers health check routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.Health)
	router.GET("/api/health", h.Health)
	router.GET("/ready", h.Ready)
}

// Health performs a basic health check
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

// Ready performs a comprehensive readiness check
func (h *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check database
	dbHealthy := true
	var dbError string
	if err := h.checkDatabase(ctx); err != nil {
		dbHealthy = false
		dbError = err.Error()
	}

	// Check Redis
	redisHealthy := true
	var redisError string
	if h.redis != nil {
		if err := h.redis.HealthCheck(ctx); err != nil {
			redisHealthy = false
			redisError = err.Error()
		}
	}

	// Overall health
	healthy := dbHealthy && redisHealthy
	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": map[string]interface{}{
			"healthy": healthy,
			"checks": map[string]interface{}{
				"database": map[string]interface{}{
					"healthy": dbHealthy,
					"error":   dbError,
				},
				"redis": map[string]interface{}{
					"healthy": redisHealthy,
					"error":   redisError,
				},
			},
		},
		"time": time.Now().UTC(),
	})
}

// checkDatabase performs a database health check
func (h *Handler) checkDatabase(ctx context.Context) error {
	var result int
	return h.db.GetContext(ctx, &result, "SELECT 1")
}