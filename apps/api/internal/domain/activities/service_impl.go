package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// serviceImpl implements the Service interface
type serviceImpl struct {
	repo Repository
}

// NewService creates a new activity service
func NewService(repo Repository) Service {
	return &serviceImpl{
		repo: repo,
	}
}

// Create creates a new activity
func (s *serviceImpl) Create(ctx context.Context, activity *Activity) (*Activity, error) {
	// Set defaults
	activity.ID = uuid.New().String()
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()
	activity.LikeCount = 0
	activity.CommentCount = 0
	activity.ViewCount = 0
	
	// Validate required fields
	if activity.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if activity.ActivityType == "" {
		return nil, fmt.Errorf("activity type is required")
	}
	if activity.Privacy == "" {
		activity.Privacy = "private"
	}
	
	return s.repo.Create(ctx, activity)
}

// GetByID retrieves an activity by ID
func (s *serviceImpl) GetByID(ctx context.Context, id string) (*Activity, error) {
	return s.repo.GetByID(ctx, id)
}

// Update updates an existing activity
func (s *serviceImpl) Update(ctx context.Context, id string, activity *Activity) (*Activity, error) {
	activity.UpdatedAt = time.Now()
	return s.repo.Update(ctx, id, activity)
}

// Delete deletes an activity
func (s *serviceImpl) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// List lists activities with filters and pagination
func (s *serviceImpl) List(ctx context.Context, filters ListFilters, page, limit int) ([]*Activity, int64, error) {
	offset := (page - 1) * limit
	return s.repo.List(ctx, filters, limit, offset)
}

// Like adds a like to an activity
func (s *serviceImpl) Like(ctx context.Context, activityID, userID string) error {
	return s.repo.AddLike(ctx, activityID, userID)
}

// Unlike removes a like from an activity
func (s *serviceImpl) Unlike(ctx context.Context, activityID, userID string) error {
	return s.repo.RemoveLike(ctx, activityID, userID)
}

// GetUserActivities gets activities created by a specific user
func (s *serviceImpl) GetUserActivities(ctx context.Context, userID string, page, limit int) ([]*Activity, int64, error) {
	filters := ListFilters{
		UserID: userID,
	}
	return s.List(ctx, filters, page, limit)
}

// IncrementViewCount increments the view count for an activity
func (s *serviceImpl) IncrementViewCount(ctx context.Context, activityID string) error {
	return s.repo.IncrementViewCount(ctx, activityID)
}