package collections

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Collection CRUD
	Create(ctx context.Context, collection *Collection) error
	GetByID(ctx context.Context, id uuid.UUID) (*Collection, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, params GetCollectionsParams) ([]Collection, int, error)
	Update(ctx context.Context, id uuid.UUID, updates UpdateCollectionRequest) (*Collection, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Collection locations
	AddLocation(ctx context.Context, collectionID uuid.UUID, location *CollectionLocation) error
	RemoveLocation(ctx context.Context, collectionID uuid.UUID, locationID uuid.UUID) error
	GetLocations(ctx context.Context, collectionID uuid.UUID) ([]CollectionLocation, error)

	// Collaboration
	AddCollaborator(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID, role string) error
	RemoveCollaborator(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID) error
	GetCollaborators(ctx context.Context, collectionID uuid.UUID) ([]uuid.UUID, error)
}