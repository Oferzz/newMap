package activities

import (
	"time"
)

// Activity represents an outdoor activity
type Activity struct {
	ID           string                 `json:"id" db:"id"`
	Title        string                 `json:"title" db:"title"`
	Description  string                 `json:"description" db:"description"`
	ActivityType string                 `json:"activity_type" db:"activity_type"`
	CreatedBy    string                 `json:"created_by" db:"created_by"`
	Privacy      string                 `json:"privacy" db:"privacy"`
	Route        *Route                 `json:"route,omitempty" db:"route"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	LikeCount    int                    `json:"like_count" db:"like_count"`
	CommentCount int                    `json:"comment_count" db:"comment_count"`
	ViewCount    int                    `json:"view_count" db:"view_count"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// Route represents the geographic route of an activity
type Route struct {
	Type       string      `json:"type"` // out-and-back, loop, point-to-point
	Waypoints  []Waypoint  `json:"waypoints"`
	Distance   float64     `json:"distance,omitempty"`    // in kilometers
	ElevationGain float64  `json:"elevation_gain,omitempty"` // in meters
	ElevationLoss float64  `json:"elevation_loss,omitempty"` // in meters
}

// Waypoint represents a point on a route
type Waypoint struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Elevation float64 `json:"elevation,omitempty"`
}

// CreateActivityInput represents the input for creating a new activity
type CreateActivityInput struct {
	Title        string                 `json:"title" binding:"required"`
	Description  string                 `json:"description"`
	ActivityType string                 `json:"activity_type" binding:"required"`
	Privacy      string                 `json:"privacy" binding:"required,oneof=public friends private"`
	Route        *Route                 `json:"route,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateActivityInput represents the input for updating an activity
type UpdateActivityInput struct {
	Title        *string                `json:"title,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Privacy      *string                `json:"privacy,omitempty" binding:"omitempty,oneof=public friends private"`
	Route        *Route                 `json:"route,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ListFilters represents filters for listing activities
type ListFilters struct {
	ActivityTypes []string `json:"activity_types,omitempty"`
	Difficulty    []string `json:"difficulty,omitempty"`
	Privacy       string   `json:"privacy,omitempty"`
	UserID        string   `json:"user_id,omitempty"`
	MinDistance   *float64 `json:"min_distance,omitempty"`
	MaxDistance   *float64 `json:"max_distance,omitempty"`
	MinDuration   *float64 `json:"min_duration,omitempty"`
	MaxDuration   *float64 `json:"max_duration,omitempty"`
}