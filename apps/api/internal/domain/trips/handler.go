package trips

import (
	"strconv"

	"github.com/Oferzz/newMap/apps/api/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func getUserID(c *gin.Context) (primitive.ObjectID, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, false
	}
	
	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, false
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
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

	// Get user ID if authenticated (optional for public trips)
	userID := primitive.NilObjectID
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
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

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
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

	err = h.service.Delete(c.Request.Context(), tripID, userID)
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
	sort := c.DefaultQuery("sort", "-created_at")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build filter
	filter := TripFilter{}

	// Filter by owner
	if ownerID := c.Query("owner_id"); ownerID != "" {
		if id, err := primitive.ObjectIDFromHex(ownerID); err == nil {
			filter.OwnerID = &id
		}
	}

	// Filter by collaborator (includes owner)
	if collaboratorID := c.Query("collaborator_id"); collaboratorID != "" {
		if id, err := primitive.ObjectIDFromHex(collaboratorID); err == nil {
			filter.CollaboratorID = &id
		}
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		tripStatus := TripStatus(status)
		if tripStatus.IsValid() {
			filter.Status = &tripStatus
		}
	}

	// Filter by public/private
	if isPublic := c.Query("is_public"); isPublic != "" {
		public := isPublic == "true"
		filter.IsPublic = &public
	}

	// Filter by tags
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	// Search query
	filter.SearchQuery = c.Query("q")

	// Get current user ID if authenticated
	var userID *primitive.ObjectID
	if id, exists := getUserID(c); exists {
		userID = &id
	}

	// List options
	opts := TripListOptions{
		Filter: filter,
		Page:   page,
		Limit:  limit,
		Sort:   sort,
	}

	trips, total, err := h.service.List(c.Request.Context(), opts, userID)
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
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

	var input InviteCollaboratorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err = h.service.InviteCollaborator(c.Request.Context(), tripID, userID, &input)
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
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

	collaboratorIDStr := c.Param("userId")
	collaboratorID, err := primitive.ObjectIDFromHex(collaboratorIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	err = h.service.RemoveCollaborator(c.Request.Context(), tripID, userID, collaboratorID)
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
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

	var input UpdateCollaboratorRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err = h.service.UpdateCollaboratorRole(c.Request.Context(), tripID, userID, &input)
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
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

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