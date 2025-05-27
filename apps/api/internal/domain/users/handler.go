package users

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
)

type Handler struct {
	service Service
}

// NewHandler creates a new user handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Create(c.Request.Context(), &input)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	response.Success(c, http.StatusCreated, user)
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	loginResp, err := h.service.Login(c.Request.Context(), &input)
	if err != nil {
		if err.Error() == "invalid credentials" {
			response.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to login")
		return
	}

	response.Success(c, http.StatusOK, loginResp)
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var input RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	loginResp, err := h.service.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	response.Success(c, http.StatusOK, loginResp)
}

// GetProfile retrieves the current user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}

	response.Success(c, http.StatusOK, user)
}

// GetUser retrieves a user by ID
func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	
	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}

	response.Success(c, http.StatusOK, user)
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Update(c.Request.Context(), userID.(string), &input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	response.Success(c, http.StatusOK, user)
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.ChangePassword(c.Request.Context(), userID.(string), &input)
	if err != nil {
		if err.Error() == "invalid current password" {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to change password")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// SendPasswordReset sends a password reset email
func (h *Handler) SendPasswordReset(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.SendPasswordResetEmail(c.Request.Context(), input.Email)
	if err != nil {
		// Don't reveal if email exists or not
		response.Success(c, http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
}

// ResetPassword resets user password with token
func (h *Handler) ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.ResetPassword(c.Request.Context(), &input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid or expired reset token")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// SearchUsers searches for users
func (h *Handler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.service.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search users")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, users, gin.H{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetFriends retrieves user's friends
func (h *Handler) GetFriends(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	friends, total, err := h.service.GetFriends(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get friends")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, friends, gin.H{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// SendFriendRequest sends a friend request
func (h *Handler) SendFriendRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	toUserID := c.Param("id")
	if toUserID == userID.(string) {
		response.Error(c, http.StatusBadRequest, "Cannot send friend request to yourself")
		return
	}

	err := h.service.SendFriendRequest(c.Request.Context(), userID.(string), toUserID)
	if err != nil {
		if err.Error() == "friend request already exists" {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to send friend request")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Friend request sent successfully"})
}

// AcceptFriendRequest accepts a friend request
func (h *Handler) AcceptFriendRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	requestID := c.Param("id")

	err := h.service.AcceptFriendRequest(c.Request.Context(), userID.(string), requestID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to accept friend request")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Friend request accepted"})
}

// RejectFriendRequest rejects a friend request
func (h *Handler) RejectFriendRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	requestID := c.Param("id")

	err := h.service.RejectFriendRequest(c.Request.Context(), userID.(string), requestID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reject friend request")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Friend request rejected"})
}

// RemoveFriend removes a friend
func (h *Handler) RemoveFriend(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	friendID := c.Param("id")

	err := h.service.RemoveFriend(c.Request.Context(), userID.(string), friendID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove friend")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Friend removed successfully"})
}

// GetFriendRequests retrieves friend requests
func (h *Handler) GetFriendRequests(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	incoming := c.Query("type") != "sent"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	requests, total, err := h.service.GetFriendRequests(c.Request.Context(), userID.(string), incoming, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get friend requests")
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, requests, gin.H{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}