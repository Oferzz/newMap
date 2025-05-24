package places

import (
	"strconv"

	"github.com/Oferzz/newMap/apps/api/internal/middleware"
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

func (h *Handler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var input CreatePlaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	place, err := h.service.Create(c.Request.Context(), userID, &input)
	if err != nil {
		switch err {
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to create places in this trip")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Created(c, place)
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	placeIDStr := c.Param("id")
	placeID, err := primitive.ObjectIDFromHex(placeIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid place ID")
		return
	}

	place, err := h.service.GetByID(c.Request.Context(), placeID, userID)
	if err != nil {
		switch err {
		case ErrPlaceNotFound:
			response.NotFound(c, "Place not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to view this place")
		default:
			response.InternalServerError(c, "Failed to get place")
		}
		return
	}

	response.Success(c, place)
}

func (h *Handler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	placeIDStr := c.Param("id")
	placeID, err := primitive.ObjectIDFromHex(placeIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid place ID")
		return
	}

	var input UpdatePlaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	place, err := h.service.Update(c.Request.Context(), placeID, userID, &input)
	if err != nil {
		switch err {
		case ErrPlaceNotFound:
			response.NotFound(c, "Place not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to update this place")
		default:
			response.InternalServerError(c, "Failed to update place")
		}
		return
	}

	response.Success(c, place)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	placeIDStr := c.Param("id")
	placeID, err := primitive.ObjectIDFromHex(placeIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid place ID")
		return
	}

	err = h.service.Delete(c.Request.Context(), placeID, userID)
	if err != nil {
		switch err {
		case ErrPlaceNotFound:
			response.NotFound(c, "Place not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to delete this place")
		default:
			response.InternalServerError(c, "Failed to delete place")
		}
		return
	}

	response.NoContent(c)
}

func (h *Handler) List(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

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
	filter := PlaceFilter{}

	// Filter by trip ID
	if tripID := c.Query("trip_id"); tripID != "" {
		if id, err := primitive.ObjectIDFromHex(tripID); err == nil {
			filter.TripID = &id
		}
	}

	// Filter by parent ID
	if parentID := c.Query("parent_id"); parentID != "" {
		if id, err := primitive.ObjectIDFromHex(parentID); err == nil {
			filter.ParentID = &id
		}
	}

	// Filter by category
	if category := c.Query("category"); category != "" {
		cat := PlaceCategory(category)
		if cat.IsValid() {
			filter.Category = &cat
		}
	}

	// Filter by visited status
	if isVisited := c.Query("is_visited"); isVisited != "" {
		visited := isVisited == "true"
		filter.IsVisited = &visited
	}

	// Filter by tags
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	// Filter by minimum rating
	if minRating := c.Query("min_rating"); minRating != "" {
		if rating, err := strconv.ParseFloat(minRating, 32); err == nil {
			r := float32(rating)
			filter.MinRating = &r
		}
	}

	// Filter by maximum cost
	if maxCost := c.Query("max_cost"); maxCost != "" {
		if cost, err := strconv.ParseFloat(maxCost, 64); err == nil {
			filter.MaxCost = &cost
		}
	}

	// Search query
	filter.SearchQuery = c.Query("q")

	// List options
	opts := PlaceListOptions{
		Filter: filter,
		Page:   page,
		Limit:  limit,
		Sort:   sort,
	}

	places, total, err := h.service.List(c.Request.Context(), opts, userID)
	if err != nil {
		if err == ErrUnauthorized {
			response.Forbidden(c, "You don't have permission to view these places")
		} else {
			response.InternalServerError(c, "Failed to list places")
		}
		return
	}

	response.SuccessWithMeta(c, places, response.NewMeta(page, limit, total))
}

func (h *Handler) GetByTripID(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tripIDStr := c.Param("tripId")
	tripID, err := primitive.ObjectIDFromHex(tripIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid trip ID")
		return
	}

	places, err := h.service.GetByTripID(c.Request.Context(), tripID, userID)
	if err != nil {
		switch err {
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to view places in this trip")
		default:
			response.InternalServerError(c, "Failed to get places")
		}
		return
	}

	response.Success(c, places)
}

func (h *Handler) MarkAsVisited(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	placeIDStr := c.Param("id")
	placeID, err := primitive.ObjectIDFromHex(placeIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid place ID")
		return
	}

	var input struct {
		IsVisited bool `json:"is_visited"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err = h.service.MarkAsVisited(c.Request.Context(), placeID, userID, input.IsVisited)
	if err != nil {
		switch err {
		case ErrPlaceNotFound:
			response.NotFound(c, "Place not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to update this place")
		default:
			response.InternalServerError(c, "Failed to update place")
		}
		return
	}

	response.Success(c, map[string]string{
		"message": "Place visit status updated successfully",
	})
}

func (h *Handler) GetChildren(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	parentIDStr := c.Param("id")
	parentID, err := primitive.ObjectIDFromHex(parentIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid parent ID")
		return
	}

	places, err := h.service.GetChildren(c.Request.Context(), parentID, userID)
	if err != nil {
		switch err {
		case ErrPlaceNotFound:
			response.NotFound(c, "Parent place not found")
		case ErrUnauthorized:
			response.Forbidden(c, "You don't have permission to view these places")
		default:
			response.InternalServerError(c, "Failed to get child places")
		}
		return
	}

	response.Success(c, places)
}