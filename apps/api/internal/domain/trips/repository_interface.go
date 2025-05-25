package trips

import (
	"context"
)

// Repository defines the interface for trip data operations
type Repository interface {
	// Create creates a new trip
	Create(ctx context.Context, trip *Trip) error
	
	// GetByID retrieves a trip by ID with collaborators and waypoints
	GetByID(ctx context.Context, id string) (*Trip, error)
	
	// Update updates a trip
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	
	// Delete soft deletes a trip
	Delete(ctx context.Context, id string) error
	
	// List retrieves trips with filters
	List(ctx context.Context, filters TripFilters) ([]*Trip, error)
	
	// AddCollaborator adds a collaborator to a trip
	AddCollaborator(ctx context.Context, tripID string, collaborator Collaborator) error
	
	// UpdateCollaborator updates a collaborator's role and permissions
	UpdateCollaborator(ctx context.Context, tripID, userID string, updates map[string]interface{}) error
	
	// RemoveCollaborator removes a collaborator from a trip
	RemoveCollaborator(ctx context.Context, tripID, userID string) error
	
	// GetCollaborator retrieves a specific collaborator
	GetCollaborator(ctx context.Context, tripID, userID string) (*Collaborator, error)
	
	// IncrementViewCount increments the view count for a trip
	IncrementViewCount(ctx context.Context, tripID string) error
	
	// IncrementShareCount increments the share count for a trip
	IncrementShareCount(ctx context.Context, tripID string) error
}

// WaypointRepository defines the interface for waypoint operations
type WaypointRepository interface {
	// AddWaypoint adds a waypoint to a trip
	AddWaypoint(ctx context.Context, tripID string, waypoint *Waypoint) error
	
	// UpdateWaypoint updates a waypoint
	UpdateWaypoint(ctx context.Context, waypointID string, updates map[string]interface{}) error
	
	// RemoveWaypoint removes a waypoint from a trip
	RemoveWaypoint(ctx context.Context, waypointID string) error
	
	// ReorderWaypoints updates the order of waypoints
	ReorderWaypoints(ctx context.Context, tripID string, waypointIDs []string) error
	
	// GetWaypoints retrieves all waypoints for a trip
	GetWaypoints(ctx context.Context, tripID string) ([]Waypoint, error)
}