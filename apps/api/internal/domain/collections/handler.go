package collections

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Oferzz/newMap/apps/api/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// POST /collections
func (h *Handler) CreateCollection(c *gin.Context) {
	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	collection, err := h.service.CreateCollection(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		response.InternalServerError(c, "Failed to create collection")
		return
	}

	response.Created(c, collection)
}

// GET /collections/:id
func (h *Handler) GetCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	collection, err := h.service.GetCollection(c.Request.Context(), id, userID.(uuid.UUID))
	if err != nil {
		if err == ErrCollectionNotFound {
			response.NotFound(c, "Collection not found")
			return
		}
		if err == ErrUnauthorized {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to get collection")
		return
	}

	response.Success(c, collection)
}

// GET /collections
func (h *Handler) GetUserCollections(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var params GetCollectionsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	collections, total, err := h.service.GetUserCollections(c.Request.Context(), userID.(uuid.UUID), params)
	if err != nil {
		response.InternalServerError(c, "Failed to get collections")
		return
	}

	meta := response.NewMeta(params.Page, params.Limit, int64(total))
	response.SuccessWithMeta(c, collections, meta)
}

// PUT /collections/:id
func (h *Handler) UpdateCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	var req UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	collection, err := h.service.UpdateCollection(c.Request.Context(), id, userID.(uuid.UUID), req)
	if err != nil {
		if err == ErrCollectionNotFound {
			response.NotFound(c, "Collection not found")
			return
		}
		if err == ErrUnauthorized {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to update collection")
		return
	}

	response.Success(c, collection)
}

// DELETE /collections/:id
func (h *Handler) DeleteCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.service.DeleteCollection(c.Request.Context(), id, userID.(uuid.UUID))
	if err != nil {
		if err == ErrCollectionNotFound {
			response.NotFound(c, "Collection not found")
			return
		}
		if err == ErrUnauthorized {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to delete collection")
		return
	}

	response.Success(c, gin.H{"message": "Collection deleted successfully"})
}

// POST /collections/:id/locations
func (h *Handler) AddLocationToCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	var req AddLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	location, err := h.service.AddLocationToCollection(c.Request.Context(), id, userID.(uuid.UUID), req)
	if err != nil {
		if err == ErrCollectionNotFound {
			response.NotFound(c, "Collection not found")
			return
		}
		if err == ErrUnauthorized {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to add location")
		return
	}

	response.Created(c, location)
}

// DELETE /collections/:id/locations/:locationId
func (h *Handler) RemoveLocationFromCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	locationIdStr := c.Param("locationId")
	locationId, err := uuid.Parse(locationIdStr)
	if err != nil {
		response.BadRequest(c, "Invalid location ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.service.RemoveLocationFromCollection(c.Request.Context(), id, locationId, userID.(uuid.UUID))
	if err != nil {
		if err == ErrCollectionNotFound {
			response.NotFound(c, "Collection not found")
			return
		}
		if err == ErrUnauthorized {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to remove location")
		return
	}

	response.Success(c, gin.H{"message": "Location removed successfully"})
}

// POST /collections/:id/collaborators
func (h *Handler) AddCollaborator(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.service.AddCollaborator(c.Request.Context(), id, targetUserID, req.Role, userID.(uuid.UUID))
	if err != nil {
		if err == ErrCollectionNotFound {
			response.NotFound(c, "Collection not found")
			return
		}
		if err == ErrUnauthorized {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to add collaborator")
		return
	}

	response.Success(c, gin.H{"message": "Collaborator added successfully"})
}

// DELETE /collections/:id/collaborators/:userId
func (h *Handler) RemoveCollaborator(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	targetUserIdStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIdStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.service.RemoveCollaborator(c.Request.Context(), id, targetUserID, userID.(uuid.UUID))
	if err != nil {
		if err == ErrCollectionNotFound {
			response.NotFound(c, "Collection not found")
			return
		}
		if err == ErrUnauthorized {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to remove collaborator")
		return
	}

	response.Success(c, gin.H{"message": "Collaborator removed successfully"})
}