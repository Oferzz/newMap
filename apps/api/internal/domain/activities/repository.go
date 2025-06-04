package activities

import (
	"context"
)

// Repository defines the interface for activity data persistence
type Repository interface {
	// Create creates a new activity
	Create(ctx context.Context, activity *Activity) (*Activity, error)
	
	// GetByID retrieves an activity by ID
	GetByID(ctx context.Context, id string) (*Activity, error)
	
	// Update updates an existing activity
	Update(ctx context.Context, id string, activity *Activity) (*Activity, error)
	
	// Delete deletes an activity
	Delete(ctx context.Context, id string) error
	
	// List lists activities with filters and pagination
	List(ctx context.Context, filters ListFilters, limit, offset int) ([]*Activity, int64, error)
	
	// AddLike adds a like to an activity
	AddLike(ctx context.Context, activityID, userID string) error
	
	// RemoveLike removes a like from an activity
	RemoveLike(ctx context.Context, activityID, userID string) error
	
	// IncrementViewCount increments the view count for an activity
	IncrementViewCount(ctx context.Context, activityID string) error
	
	// GetLikedByUser checks if a user has liked an activity
	GetLikedByUser(ctx context.Context, activityID, userID string) (bool, error)
}