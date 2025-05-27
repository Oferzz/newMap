package middleware

import (
	"strings"

	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"github.com/Oferzz/newMap/apps/api/internal/utils"
	"github.com/Oferzz/newMap/apps/api/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "userID"
	UserEmailKey        = "userEmail"
)

type AuthMiddleware struct {
	jwtManager *utils.JWTManager
}

func NewAuthMiddleware(jwtManager *utils.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			response.Unauthorized(c, "Missing authentication token")
			c.Abort()
			return
		}
		
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			if err == utils.ErrExpiredToken {
				response.Unauthorized(c, "Token has expired")
			} else {
				response.Unauthorized(c, "Invalid authentication token")
			}
			c.Abort()
			return
		}
		
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}
		
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}
		
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Next()
	}
}

func (m *AuthMiddleware) RequirePermission(permission users.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware should be used after RequireAuth
		// In a real implementation, we would fetch the user's role from the database
		// For now, we'll implement this when we have the user service
		c.Next()
	}
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	header := c.GetHeader(AuthorizationHeader)
	if header == "" {
		return ""
	}
	
	if !strings.HasPrefix(header, BearerPrefix) {
		return ""
	}
	
	return strings.TrimPrefix(header, BearerPrefix)
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	
	id, ok := userID.(string)
	return id, ok
}

func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(UserEmailKey)
	if !exists {
		return "", false
	}
	
	str, ok := email.(string)
	return str, ok
}