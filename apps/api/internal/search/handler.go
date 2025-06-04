package search

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
)

// Handler handles search-related HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new search handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Search handles natural language search requests
// @Summary Natural Language Search
// @Description Search for activities and places using natural language queries
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query in natural language"
// @Param limit query int false "Number of results to return (max 100)" default(20)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {object} response.Response{data=SearchResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search [get]
func (h *Handler) Search(c *gin.Context) {
	// Get query parameters
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c, "Query parameter 'q' is required")
		return
	}

	// Parse pagination parameters
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get user ID from context (if authenticated)
	userID := ""
	if user, exists := c.Get("user_id"); exists {
		if uid, ok := user.(string); ok {
			userID = uid
		}
	}

	// Get session ID for analytics
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.ClientIP() // Fallback to IP
	}

	// Create search request
	req := &SearchRequest{
		Query:     query,
		Limit:     limit,
		Offset:    offset,
		UserID:    userID,
		SessionID: sessionID,
	}

	// Perform search
	result, err := h.service.Search(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Search failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

// GetSuggestions handles autocomplete/suggestion requests
// @Summary Get Search Suggestions
// @Description Get autocomplete suggestions for search queries
// @Tags search
// @Accept json
// @Produce json
// @Param prefix query string false "Search prefix for autocomplete"
// @Param limit query int false "Number of suggestions to return" default(10)
// @Success 200 {object} response.Response{data=[]string}
// @Failure 500 {object} response.Response
// @Router /api/v1/search/suggestions [get]
func (h *Handler) GetSuggestions(c *gin.Context) {
	prefix := c.Query("prefix")
	
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	suggestions, err := h.service.GetSearchSuggestions(c.Request.Context(), prefix, limit)
	if err != nil {
		response.InternalServerError(c, "Failed to get suggestions: "+err.Error())
		return
	}

	response.Success(c, suggestions)
}

// ParseQuery handles query parsing requests (for debugging/testing)
// @Summary Parse Natural Language Query
// @Description Parse a natural language query to show how it will be interpreted
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Query to parse"
// @Success 200 {object} response.Response{data=nlp.ParsedQuery}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/search/parse [get]
func (h *Handler) ParseQuery(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c, "Query parameter 'q' is required")
		return
	}

	// This would require access to the NLP parser
	// For now, we'll return a simple response
	response.Success(c, map[string]interface{}{
		"query": query,
		"note":  "Query parsing endpoint - implementation depends on NLP service integration",
	})
}

// RegisterRoutes registers search routes with the gin router
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware ...gin.HandlerFunc) {
	search := router.Group("/search")
	{
		// Apply optional auth middleware if provided
		if len(authMiddleware) > 0 {
			search.Use(authMiddleware[0])
		}
		
		search.GET("", h.Search)
		search.GET("/suggestions", h.GetSuggestions)
		search.GET("/parse", h.ParseQuery)
	}
}