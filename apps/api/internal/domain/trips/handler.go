package trips

import (
	"strconv"

	"github.com/Oferzz/newMap/apps/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// getUserID extracts the user ID from the gin context
func getUserID(c *gin.Context) (string, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	
	userID, ok := userIDValue.(string)
	if !ok {
		return "", false
	}
	
	return userID, true
}

func (h *Handler) Create(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var input CreateTripInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	trip, err := h.service.Create(c.Request.Context(), userID, &input)
	if err != nil {
		response.InternalServerError(c, "Failed to create trip")
		return
	}

	response.Created(c, trip)
}

func (h *Handler) GetByID(c *gin.Context) {
	tripIDStr := c.Param("id")
	tripID := tripIDStr

	// Get user ID if authenticated (optional for public trips)
	userID := ""
	if id, exists := getUserID(c); exists {
		userID = id
	}

	trip, err := h.service.GetByID(c.Request.Context(), tripID, userID)
	if err != nil {
		switch err {
		case ErrTripNotFound:
			response.NotFound(c, "Trip not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to view this trip")
		default:
			response.InternalServerError(c, "Failed to get trip")
		}
		return
	}

	response.Success(c, trip)
}

func (h *Handler) Update(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tripIDStr := c.Param("id")
	tripID := tripIDStr

	var input UpdateTripInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	trip, err := h.service.Update(c.Request.Context(), tripID, userID, &input)
	if err != nil {
		switch err {
		case ErrTripNotFound:
			response.NotFound(c, "Trip not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to update this trip")
		default:
			response.InternalServerError(c, "Failed to update trip")
		}
		return
	}

	response.Success(c, trip)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tripIDStr := c.Param("id")
	tripID := tripIDStr

	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	err := h.service.Delete(c.Request.Context(), userID, tripID)
	if err != nil {
		switch err {
		case ErrTripNotFound:
			response.NotFound(c, "Trip not found")
		case ErrUnauthorized:
			response.Forbidden(c, "Only the trip owner can delete the trip")
		default:
			response.InternalServerError(c, "Failed to delete trip")
		}
		return
	}

	response.NoContent(c)
}

func (h *Handler) List(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Calculate offset from page
	offset := (page - 1) * limit

	// Build filter
	filter := &TripFilter{}

	// Filter by status
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	// Filter by privacy
	if privacy := c.Query("privacy"); privacy != "" {
		filter.Privacy = privacy
	}

	// Filter by tags
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	// Get current user ID if authenticated
	userID, exists := getUserID(c)
	if !exists {
		// For unauthenticated users, only show public trips
		userID = ""
		filter.Privacy = "public"
	}

	trips, total, err := h.service.List(c.Request.Context(), userID, filter, limit, offset)
	if err != nil {
		response.InternalServerError(c, "Failed to list trips")
		return
	}

	response.SuccessWithMeta(c, trips, response.NewMeta(page, limit, total))
}

func (h *Handler) InviteCollaborator(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tripIDStr := c.Param("id")
	tripID := tripIDStr

	var input InviteCollaboratorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err := h.service.InviteCollaborator(c.Request.Context(), userID, tripID, &input)
	if err != nil {
		switch err {
		case ErrTripNotFound:
			response.NotFound(c, "Trip not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to invite collaborators")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, map[string]string{
		"message": "Collaborator invited successfully",
	})
}

func (h *Handler) RemoveCollaborator(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tripIDStr := c.Param("id")
	tripID := tripIDStr

	collaboratorIDStr := c.Param("userId")
	collaboratorID := collaboratorIDStr

	err := h.service.RemoveCollaborator(c.Request.Context(), userID, tripID, collaboratorID)
	if err != nil {
		switch err {
		case ErrTripNotFound:
			response.NotFound(c, "Trip not found")
		case ErrUnauthorized:
			response.Forbidden(c, "Only the trip owner can remove collaborators")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, map[string]string{
		"message": "Collaborator removed successfully",
	})
}

func (h *Handler) UpdateCollaboratorRole(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tripIDStr := c.Param("id")
	tripID := tripIDStr

	collaboratorID := c.Param("userId")
	
	var input struct {
		Role string `json:"role" binding:"required,oneof=viewer editor admin"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err := h.service.UpdateCollaboratorRole(c.Request.Context(), userID, tripID, collaboratorID, input.Role)
	if err != nil {
		switch err {
		case ErrTripNotFound:
			response.NotFound(c, "Trip not found")
		case ErrUnauthorized:
			response.Forbidden(c, "Only the trip owner can update collaborator roles")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, map[string]string{
		"message": "Collaborator role updated successfully",
	})
}

func (h *Handler) LeaveTrip(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tripIDStr := c.Param("id")
	tripID := tripIDStr

	err = h.service.LeaveTrip(c.Request.Context(), tripID, userID)
	if err != nil {
		switch err {
		case ErrTripNotFound:
			response.NotFound(c, "Trip not found")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Success(c, map[string]string{
		"message": "You have left the trip successfully",
	})
}