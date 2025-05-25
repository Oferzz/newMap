package users

import (
	"database/sql/driver"
	"strings"
	"time"

	"github.com/lib/pq"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
	RoleUser   Role = "user"
)

type User struct {
	ID                      string         `db:"id" json:"id"`
	Email                   string         `db:"email" json:"email"`
	Username                string         `db:"username" json:"username"`
	PasswordHash            string         `db:"password_hash" json:"-"`
	DisplayName             string         `db:"display_name" json:"display_name"`
	AvatarURL               string         `db:"avatar_url" json:"avatar_url"`
	Bio                     string         `db:"bio" json:"bio"`
	Location                string         `db:"location" json:"location"`
	Roles                   pq.StringArray `db:"roles" json:"roles"`
	ProfileVisibility       string         `db:"profile_visibility" json:"profile_visibility"`
	LocationSharing         bool           `db:"location_sharing" json:"location_sharing"`
	TripDefaultPrivacy      string         `db:"trip_default_privacy" json:"trip_default_privacy"`
	EmailNotifications      bool           `db:"email_notifications" json:"email_notifications"`
	PushNotifications       bool           `db:"push_notifications" json:"push_notifications"`
	SuggestionNotifications bool           `db:"suggestion_notifications" json:"suggestion_notifications"`
	TripInviteNotifications bool           `db:"trip_invite_notifications" json:"trip_invite_notifications"`
	CreatedAt               time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time      `db:"updated_at" json:"updated_at"`
	LastActive              time.Time      `db:"last_active" json:"last_active"`
	Status                  string         `db:"status" json:"status"`
}

type CreateUserInput struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=30"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
}

type UpdateUserInput struct {
	DisplayName             *string `json:"display_name,omitempty" binding:"omitempty,min=2,max=100"`
	Bio                     *string `json:"bio,omitempty" binding:"omitempty,max=500"`
	AvatarURL               *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	Location                *string `json:"location,omitempty"`
	ProfileVisibility       *string `json:"profile_visibility,omitempty" binding:"omitempty,oneof=public friends private"`
	LocationSharing         *bool   `json:"location_sharing,omitempty"`
	TripDefaultPrivacy      *string `json:"trip_default_privacy,omitempty" binding:"omitempty,oneof=public friends private invite_only"`
	EmailNotifications      *bool   `json:"email_notifications,omitempty"`
	PushNotifications       *bool   `json:"push_notifications,omitempty"`
	SuggestionNotifications *bool   `json:"suggestion_notifications,omitempty"`
	TripInviteNotifications *bool   `json:"trip_invite_notifications,omitempty"`
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

// UserFriend represents a friendship relationship
type UserFriend struct {
	ID          string     `db:"id" json:"id"`
	UserID      string     `db:"user_id" json:"user_id"`
	FriendID    string     `db:"friend_id" json:"friend_id"`
	Status      string     `db:"status" json:"status"`
	RequestedAt time.Time  `db:"requested_at" json:"requested_at"`
	RespondedAt *time.Time `db:"responded_at" json:"responded_at"`
}

// Permission types for RBAC
type Permission string

const (
	// Trip permissions
	PermissionTripCreate Permission = "trip.create"
	PermissionTripRead   Permission = "trip.read"
	PermissionTripUpdate Permission = "trip.update"
	PermissionTripDelete Permission = "trip.delete"
	PermissionTripShare  Permission = "trip.share"
	PermissionTripInvite Permission = "trip.invite"
	
	// Place permissions
	PermissionPlaceCreate Permission = "place.create"
	PermissionPlaceRead   Permission = "place.read"
	PermissionPlaceUpdate Permission = "place.update"
	PermissionPlaceDelete Permission = "place.delete"
	PermissionPlaceMedia  Permission = "place.media"
	
	// Suggestion permissions
	PermissionSuggestionCreate   Permission = "suggestion.create"
	PermissionSuggestionRead     Permission = "suggestion.read"
	PermissionSuggestionModerate Permission = "suggestion.moderate"
	
	// User permissions
	PermissionUserRead   Permission = "user.read"
	PermissionUserUpdate Permission = "user.update"
	PermissionUserDelete Permission = "user.delete"
)

var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionTripCreate, PermissionTripRead, PermissionTripUpdate, PermissionTripDelete, PermissionTripShare, PermissionTripInvite,
		PermissionPlaceCreate, PermissionPlaceRead, PermissionPlaceUpdate, PermissionPlaceDelete, PermissionPlaceMedia,
		PermissionSuggestionCreate, PermissionSuggestionRead, PermissionSuggestionModerate,
		PermissionUserRead, PermissionUserUpdate, PermissionUserDelete,
	},
	RoleEditor: {
		PermissionTripCreate, PermissionTripRead, PermissionTripUpdate, PermissionTripShare,
		PermissionPlaceCreate, PermissionPlaceRead, PermissionPlaceUpdate, PermissionPlaceMedia,
		PermissionSuggestionCreate, PermissionSuggestionRead,
		PermissionUserRead,
	},
	RoleViewer: {
		PermissionTripRead,
		PermissionPlaceRead,
		PermissionSuggestionCreate, PermissionSuggestionRead,
		PermissionUserRead,
	},
	RoleUser: {
		PermissionTripCreate, PermissionTripRead,
		PermissionPlaceRead,
		PermissionSuggestionCreate,
		PermissionUserRead, PermissionUserUpdate,
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
	case RoleAdmin, RoleEditor, RoleViewer, RoleUser:
		return true
	default:
		return false
	}
}

// Value implements the driver.Valuer interface
func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}

// Scan implements the sql.Scanner interface
func (r *Role) Scan(value interface{}) error {
	if value == nil {
		*r = RoleUser
		return nil
	}
	switch s := value.(type) {
	case string:
		*r = Role(s)
	case []byte:
		*r = Role(s)
	default:
		*r = RoleUser
	}
	return nil
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role Role) bool {
	for _, r := range u.Roles {
		if strings.EqualFold(r, string(role)) {
			return true
		}
	}
	return false
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission Permission) bool {
	for _, roleStr := range u.Roles {
		role := Role(roleStr)
		if role.HasPermission(permission) {
			return true
		}
	}
	return false
}

// GetHighestRole returns the user's highest role
func (u *User) GetHighestRole() Role {
	if u.HasRole(RoleAdmin) {
		return RoleAdmin
	}
	if u.HasRole(RoleEditor) {
		return RoleEditor
	}
	if u.HasRole(RoleViewer) {
		return RoleViewer
	}
	return RoleUser
}