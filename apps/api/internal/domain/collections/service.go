package collections

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrCollectionNotFound   = errors.New("collection not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrLocationNotFound    = errors.New("location not found")
	ErrInvalidInput        = errors.New("invalid input")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Collection CRUD operations
func (s *Service) CreateCollection(ctx context.Context, userID uuid.UUID, req CreateCollectionRequest) (*Collection, error) {
	collection := &Collection{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
		Privacy:     req.Privacy,
	}

	err := s.repo.Create(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return collection, nil
}

func (s *Service) GetCollection(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Collection, error) {
	collection, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return nil, ErrCollectionNotFound
	}

	// Check permission
	if !s.canAccessCollection(ctx, collection, userID) {
		return nil, ErrUnauthorized
	}

	return collection, nil
}

func (s *Service) GetUserCollections(ctx context.Context, userID uuid.UUID, params GetCollectionsParams) ([]Collection, int, error) {
	return s.repo.GetByUserID(ctx, userID, params)
}

func (s *Service) UpdateCollection(ctx context.Context, id uuid.UUID, userID uuid.UUID, req UpdateCollectionRequest) (*Collection, error) {
	collection, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return nil, ErrCollectionNotFound
	}

	// Check permission - only owner can update
	if collection.UserID != userID {
		return nil, ErrUnauthorized
	}

	return s.repo.Update(ctx, id, req)
}

func (s *Service) DeleteCollection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	collection, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return ErrCollectionNotFound
	}

	// Check permission - only owner can delete
	if collection.UserID != userID {
		return ErrUnauthorized
	}

	return s.repo.Delete(ctx, id)
}

// Location operations
func (s *Service) AddLocationToCollection(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID, req AddLocationRequest) (*CollectionLocation, error) {
	collection, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return nil, ErrCollectionNotFound
	}

	// Check permission - owner or collaborator can add
	if !s.canModifyCollection(ctx, collection, userID) {
		return nil, ErrUnauthorized
	}

	location := &CollectionLocation{
		Name:      req.Name,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	err = s.repo.AddLocation(ctx, collectionID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to add location: %w", err)
	}

	return location, nil
}

func (s *Service) RemoveLocationFromCollection(ctx context.Context, collectionID uuid.UUID, locationID uuid.UUID, userID uuid.UUID) error {
	collection, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return ErrCollectionNotFound
	}

	// Check permission - owner or collaborator can remove
	if !s.canModifyCollection(ctx, collection, userID) {
		return ErrUnauthorized
	}

	return s.repo.RemoveLocation(ctx, collectionID, locationID)
}

// Collaboration operations
func (s *Service) AddCollaborator(ctx context.Context, collectionID uuid.UUID, targetUserID uuid.UUID, role string, userID uuid.UUID) error {
	collection, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return ErrCollectionNotFound
	}

	// Only owner can add collaborators
	if collection.UserID != userID {
		return ErrUnauthorized
	}

	return s.repo.AddCollaborator(ctx, collectionID, targetUserID, role)
}

func (s *Service) RemoveCollaborator(ctx context.Context, collectionID uuid.UUID, targetUserID uuid.UUID, userID uuid.UUID) error {
	collection, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return ErrCollectionNotFound
	}

	// Only owner can remove collaborators
	if collection.UserID != userID {
		return ErrUnauthorized
	}

	return s.repo.RemoveCollaborator(ctx, collectionID, targetUserID)
}

// Helper methods for permissions
func (s *Service) canAccessCollection(ctx context.Context, collection *Collection, userID uuid.UUID) bool {
	// Owner can always access
	if collection.UserID == userID {
		return true
	}

	// Public collections can be accessed by anyone
	if collection.Privacy == "public" {
		return true
	}

	// For private/friends collections, check if user is a collaborator
	collaborators, err := s.repo.GetCollaborators(ctx, collection.ID)
	if err != nil {
		return false
	}

	for _, collaboratorID := range collaborators {
		if collaboratorID == userID {
			return true
		}
	}

	return false
}

func (s *Service) canModifyCollection(ctx context.Context, collection *Collection, userID uuid.UUID) bool {
	// Owner can always modify
	if collection.UserID == userID {
		return true
	}

	// Check if user is a collaborator (collaborators can modify locations)
	collaborators, err := s.repo.GetCollaborators(ctx, collection.ID)
	if err != nil {
		return false
	}

	for _, collaboratorID := range collaborators {
		if collaboratorID == userID {
			return true
		}
	}

	return false
}