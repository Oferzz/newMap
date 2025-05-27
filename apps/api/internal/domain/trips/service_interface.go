package trips

import (
	"context"
	"errors"
	"time"
)

// Service defines the interface for trip operations
type Service interface {
	// Basic CRUD operations
	Create(ctx context.Context, userID string, input *CreateTripInput) (*Trip, error)
	GetByID(ctx context.Context, userID, tripID string) (*Trip, error)
	Update(ctx context.Context, userID, tripID string, input *UpdateTripInput) (*Trip, error)
	Delete(ctx context.Context, userID, tripID string) error
	
	// Query operations
	List(ctx context.Context, userID string, filter *TripFilter, limit, offset int) ([]*Trip, int64, error)
	GetUserTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error)
	GetSharedTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error)
	Search(ctx context.Context, userID string, query string, limit, offset int) ([]*Trip, int64, error)
	
	// Collaborator management
	AddCollaborator(ctx context.Context, userID, tripID, collaboratorID, role string) error
	RemoveCollaborator(ctx context.Context, userID, tripID, collaboratorID string) error
	UpdateCollaboratorRole(ctx context.Context, userID, tripID, collaboratorID, role string) error
	InviteCollaborator(ctx context.Context, userID, tripID string, input *InviteCollaboratorInput) error
	
	// Waypoint management
	AddWaypoint(ctx context.Context, userID, tripID string, input *AddWaypointInput) (*Waypoint, error)
	UpdateWaypoint(ctx context.Context, userID, tripID, waypointID string, input *UpdateWaypointInput) (*Waypoint, error)
	RemoveWaypoint(ctx context.Context, userID, tripID, waypointID string) error
	ReorderWaypoints(ctx context.Context, userID, tripID string, waypointIDs []string) error
	
	// Additional operations
	GetTripStats(ctx context.Context, userID, tripID string) (*TripStats, error)
	ExportTrip(ctx context.Context, userID, tripID, format string) ([]byte, error)
	CloneTrip(ctx context.Context, userID, tripID string) (*Trip, error)
}

// Common errors
var (
	ErrTripNotFound = errors.New("trip not found")
	ErrUnauthorized = errors.New("unauthorized")
)

// TripFilter contains filter criteria for trips
type TripFilter struct {
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	Privacy   string
	Tags      []string
}

// TripStats contains trip statistics
type TripStats struct {
	TotalPlaces      int `json:"total_places"`
	TotalWaypoints   int `json:"total_waypoints"`
	TotalCollaborators int `json:"total_collaborators"`
	TotalSuggestions int `json:"total_suggestions"`
	TotalViews       int `json:"total_views"`
	TotalShares      int `json:"total_shares"`
}

// InviteCollaboratorInput for service compatibility
type InviteCollaboratorInput struct {
	UserID      string `json:"user_id" binding:"required,uuid"`
	Role        string `json:"role" binding:"required,oneof=viewer editor admin"`
	CanEdit     bool   `json:"can_edit"`
	CanDelete   bool   `json:"can_delete"`
	CanInvite   bool   `json:"can_invite"`
	CanModerate bool   `json:"can_moderate_suggestions"`
}