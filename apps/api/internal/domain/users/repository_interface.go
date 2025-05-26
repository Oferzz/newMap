package users

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// Repository defines the interface for user data access
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]*User, error)
	AddFriend(ctx context.Context, userID, friendID string) error
	RemoveFriend(ctx context.Context, userID, friendID string) error
	GetFriends(ctx context.Context, userID string) ([]*User, error)
}