package users

import (
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

func (h *Handler) Register(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	authResponse, err := h.service.Register(c.Request.Context(), &input)
	if err != nil {
		switch err {
		case ErrEmailExists:
			response.Conflict(c, "Email already exists")
		case ErrUsernameExists:
			response.Conflict(c, "Username already exists")
		default:
			response.InternalServerError(c, "Failed to register user")
		}
		return
	}
	
	response.Created(c, authResponse)
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	authResponse, err := h.service.Login(c.Request.Context(), &input)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			response.Unauthorized(c, "Invalid email or password")
		default:
			response.InternalServerError(c, "Failed to login")
		}
		return
	}
	
	response.Success(c, authResponse)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var input RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	authResponse, err := h.service.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Invalid refresh token")
		return
	}
	
	response.Success(c, authResponse)
}

func (h *Handler) GetProfile(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok {
		response.Unauthorized(c, "Invalid user ID")
		return
	}
	
	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == ErrUserNotFound {
			response.NotFound(c, "User not found")
		} else {
			response.InternalServerError(c, "Failed to get user profile")
		}
		return
	}
	
	response.Success(c, user)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok {
		response.Unauthorized(c, "Invalid user ID")
		return
	}
	
	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	user, err := h.service.Update(c.Request.Context(), userID, &input)
	if err != nil {
		if err == ErrUserNotFound {
			response.NotFound(c, "User not found")
		} else {
			response.InternalServerError(c, "Failed to update profile")
		}
		return
	}
	
	response.Success(c, user)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok {
		response.Unauthorized(c, "Invalid user ID")
		return
	}
	
	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	err := h.service.ChangePassword(c.Request.Context(), userID, &input)
	if err != nil {
		switch err {
		case ErrInvalidPassword:
			response.BadRequest(c, "Current password is incorrect")
		case ErrUserNotFound:
			response.NotFound(c, "User not found")
		default:
			response.InternalServerError(c, "Failed to change password")
		}
		return
	}
	
	response.Success(c, map[string]string{
		"message": "Password changed successfully",
	})
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok {
		response.Unauthorized(c, "Invalid user ID")
		return
	}
	
	err := h.service.Delete(c.Request.Context(), userID)
	if err != nil {
		if err == ErrUserNotFound {
			response.NotFound(c, "User not found")
		} else {
			response.InternalServerError(c, "Failed to delete account")
		}
		return
	}
	
	response.NoContent(c)
}