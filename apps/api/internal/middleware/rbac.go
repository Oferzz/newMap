package middleware

import (
	"context"

	"github.com/Oferzz/newMap/apps/api/internal/domain/trips"
	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type RBACMiddleware struct {
	userRepo users.Repository
	tripRepo trips.Repository
}

func NewRBACMiddleware(userRepo users.Repository, tripRepo trips.Repository) *RBACMiddleware {
	return &RBACMiddleware{
		userRepo: userRepo,
		tripRepo: tripRepo,
	}
}

// RequireSystemPermission checks if the user has the required system-level permission
func (m *RBACMiddleware) RequireSystemPermission(permission users.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Get user from database to check their role
		user, err := m.userRepo.GetByID(context.Background(), userID)
		if err != nil {
			response.Unauthorized(c, "User not found")
			c.Abort()
			return
		}

		// For now, we'll assign roles based on user properties
		// In production, you'd have a proper role management system
		var userRole users.Role
		if user.Email == "admin@tripplatform.com" {
			userRole = users.RoleAdmin
		} else if user.IsVerified {
			userRole = users.RoleEditor
		} else {
			// Give all authenticated users at least RoleUser (which can create trips)
			userRole = users.RoleUser
		}

		// Check if the role has the required permission
		if !userRole.HasPermission(permission) {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		// Store role in context for later use
		c.Set("userRole", userRole)
		c.Next()
	}
}

// RequireTripPermission checks if the user has the required permission for a specific trip
func (m *RBACMiddleware) RequireTripPermission(permission users.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Get trip ID from URL parameter
		tripIDStr := c.Param("tripId")
		if tripIDStr == "" {
			tripIDStr = c.Param("id")
		}

		tripID := tripIDStr

		// Get trip from database
		trip, err := m.tripRepo.GetByID(context.Background(), tripID)
		if err != nil {
			if err == trips.ErrTripNotFound {
				response.NotFound(c, "Trip not found")
			} else {
				response.InternalServerError(c, "Failed to check permissions")
			}
			c.Abort()
			return
		}

		// Check if user has permission for this specific trip
		if !trip.CanUserPerform(userID, string(permission)) {
			response.Forbidden(c, "You don't have permission to perform this action on this trip")
			c.Abort()
			return
		}

		// Store trip in context for use in handlers
		c.Set("trip", trip)
		c.Next()
	}
}

// RequireTripOwnership ensures the user is the owner of the trip
func (m *RBACMiddleware) RequireTripOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Get trip ID from URL parameter
		tripIDStr := c.Param("tripId")
		if tripIDStr == "" {
			tripIDStr = c.Param("id")
		}

		tripID := tripIDStr

		// Get trip from database
		trip, err := m.tripRepo.GetByID(context.Background(), tripID)
		if err != nil {
			if err == trips.ErrTripNotFound {
				response.NotFound(c, "Trip not found")
			} else {
				response.InternalServerError(c, "Failed to check ownership")
			}
			c.Abort()
			return
		}

		// Check if user is the owner
		if trip.OwnerID != userID {
			response.Forbidden(c, "Only the trip owner can perform this action")
			c.Abort()
			return
		}

		// Store trip in context for use in handlers
		c.Set("trip", trip)
		c.Next()
	}
}

// OptionalTripPermission checks permissions if trip ID is provided, otherwise continues
func (m *RBACMiddleware) OptionalTripPermission(permission users.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.Next()
			return
		}

		// Get trip ID from URL parameter or query
		tripIDStr := c.Param("tripId")
		if tripIDStr == "" {
			tripIDStr = c.Query("tripId")
		}

		if tripIDStr == "" {
			c.Next()
			return
		}

		tripID := tripIDStr

		// Get trip from database
		trip, err := m.tripRepo.GetByID(context.Background(), tripID)
		if err != nil {
			c.Next()
			return
		}

		// Store trip permission status
		canPerform := trip.CanUserPerform(userID, string(permission))
		c.Set("canPerformTripAction", canPerform)
		c.Set("trip", trip)
		c.Next()
	}
}