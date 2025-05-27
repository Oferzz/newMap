package users

import (
	"context"
)

// Service defines the interface for user service operations
type Service interface {
	// User CRUD operations
	Create(ctx context.Context, input *CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id string, input *UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id string) error

	// Authentication operations
	Login(ctx context.Context, input *LoginInput) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	ChangePassword(ctx context.Context, userID string, input *ChangePasswordInput) error
	ResetPassword(ctx context.Context, input *ResetPasswordInput) error
	SendPasswordResetEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error

	// Search and social operations
	Search(ctx context.Context, query string, limit, offset int) ([]*User, int64, error)
	GetFriends(ctx context.Context, userID string, limit, offset int) ([]*User, int64, error)
	SendFriendRequest(ctx context.Context, fromUserID, toUserID string) error
	AcceptFriendRequest(ctx context.Context, userID, requestID string) error
	RejectFriendRequest(ctx context.Context, userID, requestID string) error
	RemoveFriend(ctx context.Context, userID, friendID string) error
	GetFriendRequests(ctx context.Context, userID string, incoming bool, limit, offset int) ([]*FriendRequest, int64, error)
}