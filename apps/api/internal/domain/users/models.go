package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	FullName     string             `bson:"full_name" json:"full_name"`
	Avatar       string             `bson:"avatar" json:"avatar"`
	Bio          string             `bson:"bio" json:"bio"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	IsVerified   bool               `bson:"is_verified" json:"is_verified"`
	LastLoginAt  *time.Time         `bson:"last_login_at" json:"last_login_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
}

type UpdateUserInput struct {
	FullName *string `json:"full_name,omitempty" binding:"omitempty,min=2,max=100"`
	Bio      *string `json:"bio,omitempty" binding:"omitempty,max=500"`
	Avatar   *string `json:"avatar,omitempty" binding:"omitempty,url"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type Permission string

const (
	// Trip permissions
	PermissionTripCreate Permission = "trip.create"
	PermissionTripRead   Permission = "trip.read"
	PermissionTripUpdate Permission = "trip.update"
	PermissionTripDelete Permission = "trip.delete"
	
	// Place permissions
	PermissionPlaceCreate Permission = "place.create"
	PermissionPlaceRead   Permission = "place.read"
	PermissionPlaceUpdate Permission = "place.update"
	PermissionPlaceDelete Permission = "place.delete"
	
	// Suggestion permissions
	PermissionSuggestionCreate Permission = "suggestion.create"
	PermissionSuggestionRead   Permission = "suggestion.read"
	PermissionSuggestionUpdate Permission = "suggestion.update"
	PermissionSuggestionDelete Permission = "suggestion.delete"
	
	// User permissions
	PermissionUserRead   Permission = "user.read"
	PermissionUserUpdate Permission = "user.update"
	PermissionUserDelete Permission = "user.delete"
)

var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionTripCreate, PermissionTripRead, PermissionTripUpdate, PermissionTripDelete,
		PermissionPlaceCreate, PermissionPlaceRead, PermissionPlaceUpdate, PermissionPlaceDelete,
		PermissionSuggestionCreate, PermissionSuggestionRead, PermissionSuggestionUpdate, PermissionSuggestionDelete,
		PermissionUserRead, PermissionUserUpdate, PermissionUserDelete,
	},
	RoleEditor: {
		PermissionTripCreate, PermissionTripRead, PermissionTripUpdate,
		PermissionPlaceCreate, PermissionPlaceRead, PermissionPlaceUpdate,
		PermissionSuggestionCreate, PermissionSuggestionRead, PermissionSuggestionUpdate,
		PermissionUserRead,
	},
	RoleViewer: {
		PermissionTripRead,
		PermissionPlaceRead,
		PermissionSuggestionRead,
		PermissionUserRead,
	},
}

func (r Role) HasPermission(permission Permission) bool {
	permissions, exists := RolePermissions[r]
	if !exists {
		return false
	}
	
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}