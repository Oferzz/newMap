package users

import (
	"context"
)

// Repository defines the interface for user data operations
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)
	
	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*User, error)
	
	// Update updates a user
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	
	// UpdateLastActive updates the user's last active timestamp
	UpdateLastActive(ctx context.Context, id string) error
	
	// Delete soft deletes a user
	Delete(ctx context.Context, id string) error
	
	// List retrieves users with pagination
	List(ctx context.Context, limit, offset int) ([]*User, error)
	
	// Search searches for users by username or display name
	Search(ctx context.Context, query string, limit, offset int) ([]*User, error)
	
	// GetFriends retrieves a user's friends
	GetFriends(ctx context.Context, userID string) ([]*User, error)
	
	// AddFriend sends a friend request
	AddFriend(ctx context.Context, userID, friendID string) error
	
	// UpdateFriendship updates the status of a friendship
	UpdateFriendship(ctx context.Context, userID, friendID, status string) error
	
	// RemoveFriend removes a friendship
	RemoveFriend(ctx context.Context, userID, friendID string) error
	
	// GetPendingFriendRequests retrieves pending friend requests for a user
	GetPendingFriendRequests(ctx context.Context, userID string) ([]*UserFriend, error)
	
	// CountByStatus counts users by status
	CountByStatus(ctx context.Context, status string) (int64, error)
}