package places

import (
	"context"
	"time"
)

// Service defines the interface for place operations
type Service interface {
	// Basic CRUD operations
	Create(ctx context.Context, userID string, input *CreatePlaceInput) (*Place, error)
	GetByID(ctx context.Context, userID, placeID string) (*Place, error)
	Update(ctx context.Context, userID, placeID string, input *UpdatePlaceInput) (*Place, error)
	Delete(ctx context.Context, userID, placeID string) error
	
	// Query operations
	GetUserPlaces(ctx context.Context, userID string, limit, offset int) ([]*Place, int64, error)
	GetChildPlaces(ctx context.Context, userID, parentID string) ([]*Place, error)
	Search(ctx context.Context, userID string, input *SearchPlacesInput) ([]*Place, int64, error)
	GetNearby(ctx context.Context, userID string, input *NearbyPlacesInput) ([]*Place, error)
	
	// Collaborator management
	AddCollaborator(ctx context.Context, userID, placeID, collaboratorID, role string) error
	RemoveCollaborator(ctx context.Context, userID, placeID, collaboratorID string) error
	UpdateCollaboratorRole(ctx context.Context, userID, placeID, collaboratorID, role string) error
	
	// Extended operations (for compatibility)
	List(ctx context.Context, userID string, filter *PlaceFilter, limit, offset int) ([]*Place, int64, error)
	GetTripPlaces(ctx context.Context, userID, tripID string) ([]*Place, error)
	AddToTrip(ctx context.Context, userID, placeID, tripID string) error
	RemoveFromTrip(ctx context.Context, userID, placeID, tripID string) error
	UpdateVisitStatus(ctx context.Context, userID, placeID string, visited bool, visitDate *time.Time) error
	AddImages(ctx context.Context, userID, placeID string, images []string) error
	RemoveImage(ctx context.Context, userID, placeID string, imageURL string) error
	UpdateRating(ctx context.Context, userID, placeID string, rating float32) error
	AddNote(ctx context.Context, userID, placeID, note string) error
}

