package activities

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
)

// Handler handles activity-related HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new activities handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateActivity creates a new activity
// @Summary Create activity
// @Description Create a new outdoor activity
// @Tags activities
// @Accept json
// @Produce json
// @Param activity body CreateActivityInput true "Activity data"
// @Success 201 {object} response.Response{data=Activity}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/activities [post]
func (h *Handler) CreateActivity(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var input CreateActivityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid activity data")
		return
	}

	activity := &Activity{
		Title:        input.Title,
		Description:  input.Description,
		ActivityType: input.ActivityType,
		CreatedBy:    userID.(string),
		Privacy:      input.Privacy,
		Route:        input.Route,
		Metadata:     input.Metadata,
	}

	created, err := h.service.Create(c.Request.Context(), activity)
	if err != nil {
		response.InternalServerError(c, "Failed to create activity")
		return
	}

	response.Created(c, created)
}

// GetActivity retrieves an activity by ID
// @Summary Get activity
// @Description Get an activity by ID
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} response.Response{data=Activity}
// @Failure 404 {object} response.Response
// @Router /api/v1/activities/{id} [get]
func (h *Handler) GetActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, hasUser := c.Get("user_id")

	activity, err := h.service.GetByID(c.Request.Context(), activityID)
	if err != nil {
		response.NotFound(c, "Activity not found")
		return
	}

	// Check visibility permissions
	if activity.Privacy == "private" && (!hasUser || activity.CreatedBy != userID.(string)) {
		response.Forbidden(c, "You don't have permission to view this activity")
		return
	}

	response.Success(c, activity)
}

// UpdateActivity updates an activity
// @Summary Update activity
// @Description Update an existing activity
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Param activity body UpdateActivityInput true "Updated activity data"
// @Success 200 {object} response.Response{data=Activity}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/activities/{id} [put]
func (h *Handler) UpdateActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")

	var input UpdateActivityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid activity data")
		return
	}

	// Check ownership
	activity, err := h.service.GetByID(c.Request.Context(), activityID)
	if err != nil {
		response.NotFound(c, "Activity not found")
		return
	}

	if activity.CreatedBy != userID.(string) {
		response.Forbidden(c, "You don't have permission to update this activity")
		return
	}

	// Update fields
	if input.Title != nil {
		activity.Title = *input.Title
	}
	if input.Description != nil {
		activity.Description = *input.Description
	}
	if input.Privacy != nil {
		activity.Privacy = *input.Privacy
	}
	if input.Route != nil {
		activity.Route = input.Route
	}
	if input.Metadata != nil {
		activity.Metadata = input.Metadata
	}

	updated, err := h.service.Update(c.Request.Context(), activityID, activity)
	if err != nil {
		response.InternalServerError(c, "Failed to update activity")
		return
	}

	response.Success(c, updated)
}

// DeleteActivity deletes an activity
// @Summary Delete activity
// @Description Delete an activity
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Success 204
// @Failure 404 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/activities/{id} [delete]
func (h *Handler) DeleteActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Check ownership
	activity, err := h.service.GetByID(c.Request.Context(), activityID)
	if err != nil {
		response.NotFound(c, "Activity not found")
		return
	}

	if activity.CreatedBy != userID.(string) {
		response.Forbidden(c, "You don't have permission to delete this activity")
		return
	}

	if err := h.service.Delete(c.Request.Context(), activityID); err != nil {
		response.InternalServerError(c, "Failed to delete activity")
		return
	}

	response.NoContent(c)
}

// ListActivities lists activities with filters
// @Summary List activities
// @Description Get a list of activities with optional filters
// @Tags activities
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param activity_type query []string false "Filter by activity types"
// @Param difficulty query []string false "Filter by difficulty levels"
// @Param privacy query string false "Filter by privacy setting"
// @Success 200 {object} response.Response{data=[]Activity,meta=response.Meta}
// @Router /api/v1/activities [get]
func (h *Handler) ListActivities(c *gin.Context) {
	userID, hasUser := c.Get("user_id")
	
	// Parse pagination
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Parse filters
	filters := ListFilters{
		ActivityTypes: c.QueryArray("activity_type"),
		Difficulty:    c.QueryArray("difficulty"),
		Privacy:       c.Query("privacy"),
	}

	// If user is not authenticated, only show public activities
	if !hasUser {
		filters.Privacy = "public"
	} else {
		filters.UserID = userID.(string)
	}

	activities, total, err := h.service.List(c.Request.Context(), filters, page, limit)
	if err != nil {
		response.InternalServerError(c, "Failed to list activities")
		return
	}

	meta := response.NewMeta(page, limit, total)
	response.SuccessWithMeta(c, activities, meta)
}

// LikeActivity likes an activity
// @Summary Like activity
// @Description Add a like to an activity
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/activities/{id}/like [post]
func (h *Handler) LikeActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")

	if err := h.service.Like(c.Request.Context(), activityID, userID.(string)); err != nil {
		response.InternalServerError(c, "Failed to like activity")
		return
	}

	response.Success(c, map[string]interface{}{
		"message": "Activity liked successfully",
	})
}

// UnlikeActivity removes a like from an activity
// @Summary Unlike activity
// @Description Remove a like from an activity
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/activities/{id}/like [delete]
func (h *Handler) UnlikeActivity(c *gin.Context) {
	activityID := c.Param("id")
	userID, _ := c.Get("user_id")

	if err := h.service.Unlike(c.Request.Context(), activityID, userID.(string)); err != nil {
		response.InternalServerError(c, "Failed to unlike activity")
		return
	}

	response.Success(c, map[string]interface{}{
		"message": "Activity unliked successfully",
	})
}

// RegisterRoutes registers activity routes with the router
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	activities := router.Group("/activities")
	
	// Public routes (no auth required)
	activities.GET("", h.ListActivities)
	activities.GET("/shared/:token", h.GetSharedActivity)
	
	// Protected routes
	protected := activities.Group("")
	protected.Use(authMiddleware)
	{
		protected.POST("", h.CreateActivity)
		protected.GET("/:id", h.GetActivity)
		protected.PUT("/:id", h.UpdateActivity)
		protected.DELETE("/:id", h.DeleteActivity)
		protected.POST("/:id/like", h.LikeActivity)
		protected.DELETE("/:id/like", h.UnlikeActivity)
		
		// Share endpoints
		protected.POST("/:id/share", h.GenerateShareLink)
		protected.GET("/:id/share", h.ListShareLinks)
		protected.DELETE("/:id/share/:linkId", h.RevokeShareLink)
	}
}