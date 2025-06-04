package activities

import (
	"context"
)

// Service defines the interface for activity operations
type Service interface {
	// Create creates a new activity
	Create(ctx context.Context, activity *Activity) (*Activity, error)
	
	// GetByID retrieves an activity by ID
	GetByID(ctx context.Context, id string) (*Activity, error)
	
	// Update updates an existing activity
	Update(ctx context.Context, id string, activity *Activity) (*Activity, error)
	
	// Delete deletes an activity
	Delete(ctx context.Context, id string) error
	
	// List lists activities with filters and pagination
	List(ctx context.Context, filters ListFilters, page, limit int) ([]*Activity, int64, error)
	
	// Like adds a like to an activity
	Like(ctx context.Context, activityID, userID string) error
	
	// Unlike removes a like from an activity
	Unlike(ctx context.Context, activityID, userID string) error
	
	// GetUserActivities gets activities created by a specific user
	GetUserActivities(ctx context.Context, userID string, page, limit int) ([]*Activity, int64, error)
	
	// IncrementViewCount increments the view count for an activity
	IncrementViewCount(ctx context.Context, activityID string) error
}