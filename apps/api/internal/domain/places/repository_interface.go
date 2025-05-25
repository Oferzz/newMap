package places

import (
	"context"
)

// Repository defines the interface for place data operations
type Repository interface {
	// Create creates a new place
	Create(ctx context.Context, place *Place) error
	
	// GetByID retrieves a place by ID
	GetByID(ctx context.Context, id string) (*Place, error)
	
	// Update updates a place
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	
	// Delete soft deletes a place
	Delete(ctx context.Context, id string) error
	
	// Search searches for places
	Search(ctx context.Context, input SearchPlacesInput) ([]*Place, error)
	
	// GetNearby finds nearby places
	GetNearby(ctx context.Context, input NearbyPlacesInput) ([]*Place, error)
	
	// GetByTripID retrieves all places for a trip
	GetByTripID(ctx context.Context, tripID string) ([]*Place, error)
	
	// GetChildren retrieves child places
	GetChildren(ctx context.Context, parentID string) ([]*Place, error)
	
	// UpdateRating updates the average rating for a place
	UpdateRating(ctx context.Context, placeID string, newRating float32) error
	
	// IncrementRatingCount increments the rating count
	IncrementRatingCount(ctx context.Context, placeID string) error
}

// MediaRepository defines the interface for place media operations
type MediaRepository interface {
	// AddMedia adds media to a place
	AddMedia(ctx context.Context, placeID string, media Media) error
	
	// RemoveMedia removes media from a place
	RemoveMedia(ctx context.Context, placeID, mediaID string) error
	
	// UpdateMediaOrder updates the order of media items
	UpdateMediaOrder(ctx context.Context, placeID string, mediaIDs []string) error
	
	// GetMedia retrieves all media for a place
	GetMedia(ctx context.Context, placeID string) ([]Media, error)
}

// CollaboratorRepository defines the interface for place collaborator operations
type CollaboratorRepository interface {
	// AddCollaborator adds a collaborator to a place
	AddCollaborator(ctx context.Context, placeID string, collaborator Collaborator) error
	
	// UpdateCollaborator updates a collaborator's role and permissions
	UpdateCollaborator(ctx context.Context, placeID, userID string, updates map[string]interface{}) error
	
	// RemoveCollaborator removes a collaborator from a place
	RemoveCollaborator(ctx context.Context, placeID, userID string) error
	
	// GetCollaborators retrieves all collaborators for a place
	GetCollaborators(ctx context.Context, placeID string) ([]Collaborator, error)
}