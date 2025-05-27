package places

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Oferzz/newMap/apps/api/internal/domain/trips"
	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
)

type servicePg struct {
	repo     RepositoryInterface
	tripRepo trips.RepositoryInterface
}

func NewServicePg(repo RepositoryInterface, tripRepo trips.RepositoryInterface) Service {
	return &servicePg{
		repo:     repo,
		tripRepo: tripRepo,
	}
}

func (s *servicePg) Create(ctx context.Context, userID string, input *CreatePlaceInput) (*Place, error) {
	// For PostgreSQL, we'll create the place directly without trip association
	// The trip association will be handled separately
	
	place := &Place{
		ID:            uuid.New().String(),
		Name:          input.Name,
		Description:   input.Description,
		Type:          input.Type,
		ParentID:      input.ParentID,
		StreetAddress: input.StreetAddress,
		City:          input.City,
		State:         input.State,
		Country:       input.Country,
		PostalCode:    input.PostalCode,
		CreatedBy:     userID,
		Category:      input.Category,
		Tags:          input.Tags,
		OpeningHours:  input.OpeningHours,
		ContactInfo:   input.ContactInfo,
		Amenities:     input.Amenities,
		Privacy:       "public",
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// Handle location
	if input.Location != nil {
		place.Location = &GeoPoint{
			Type:        "Point",
			Coordinates: []float64{input.Location.Longitude, input.Location.Latitude},
		}
	}
	
	// Handle bounds
	if input.Bounds != nil {
		place.Bounds = &GeoPolygon{
			Type:        "Polygon",
			Coordinates: input.Bounds.Coordinates,
		}
	}
	
	// Set default privacy if not provided
	if input.Privacy != "" {
		place.Privacy = input.Privacy
	}
	
	if err := s.repo.Create(ctx, place); err != nil {
		return nil, fmt.Errorf("failed to create place: %w", err)
	}
	
	return place, nil
}

func (s *servicePg) GetByID(ctx context.Context, userID, placeID string) (*Place, error) {
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return nil, err
	}
	
	// Check if user has permission to view this place
	if place.Privacy == "private" && place.CreatedBy != userID && !place.HasCollaborator(userID) {
		return nil, ErrUnauthorized
	}
	
	return place, nil
}

func (s *servicePg) Update(ctx context.Context, userID, placeID string, input *UpdatePlaceInput) (*Place, error) {
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return nil, err
	}
	
	// Check if user can edit
	if !place.CanUserEdit(userID) {
		return nil, ErrUnauthorized
	}
	
	// Update fields
	if input.Name != nil {
		place.Name = *input.Name
	}
	if input.Description != nil {
		place.Description = *input.Description
	}
	if input.Type != nil {
		place.Type = *input.Type
	}
	if input.StreetAddress != nil {
		place.StreetAddress = *input.StreetAddress
	}
	if input.City != nil {
		place.City = *input.City
	}
	if input.State != nil {
		place.State = *input.State
	}
	if input.Country != nil {
		place.Country = *input.Country
	}
	if input.PostalCode != nil {
		place.PostalCode = *input.PostalCode
	}
	if len(input.Category) > 0 {
		place.Category = input.Category
	}
	if len(input.Tags) > 0 {
		place.Tags = input.Tags
	}
	if input.OpeningHours != nil {
		place.OpeningHours = input.OpeningHours
	}
	if input.ContactInfo != nil {
		place.ContactInfo = input.ContactInfo
	}
	if len(input.Amenities) > 0 {
		place.Amenities = input.Amenities
	}
	if input.Privacy != nil {
		place.Privacy = *input.Privacy
	}
	if input.Status != nil {
		place.Status = *input.Status
	}
	
	// Handle location
	if input.Location != nil {
		place.Location = &GeoPoint{
			Type:        "Point",
			Coordinates: []float64{input.Location.Longitude, input.Location.Latitude},
		}
	}
	
	// Handle bounds
	if input.Bounds != nil {
		place.Bounds = &GeoPolygon{
			Type:        "Polygon",
			Coordinates: input.Bounds.Coordinates,
		}
	}
	
	place.UpdatedAt = time.Now()
	
	if err := s.repo.Update(ctx, place); err != nil {
		return nil, fmt.Errorf("failed to update place: %w", err)
	}
	
	return place, nil
}

func (s *servicePg) Delete(ctx context.Context, userID, placeID string) error {
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return err
	}
	
	// Check if user can delete
	if !place.CanUserDelete(userID) {
		return ErrUnauthorized
	}
	
	// Check if place has children
	children, err := s.repo.GetChildPlaces(ctx, placeID)
	if err != nil {
		return fmt.Errorf("failed to check child places: %w", err)
	}
	
	if len(children) > 0 {
		return errors.New("cannot delete place with child places")
	}
	
	return s.repo.Delete(ctx, placeID)
}

func (s *servicePg) GetUserPlaces(ctx context.Context, userID string, limit, offset int) ([]*Place, int64, error) {
	return s.repo.GetByUser(ctx, userID, limit, offset)
}

func (s *servicePg) GetChildPlaces(ctx context.Context, userID, parentID string) ([]*Place, error) {
	// First check if user has access to parent place
	parent, err := s.GetByID(ctx, userID, parentID)
	if err != nil {
		return nil, err
	}
	
	children, err := s.repo.GetChildPlaces(ctx, parentID)
	if err != nil {
		return nil, err
	}
	
	// Filter children based on privacy
	filtered := make([]*Place, 0, len(children))
	for _, child := range children {
		if child.Privacy == "public" || child.CreatedBy == userID || child.HasCollaborator(userID) || parent.HasCollaborator(userID) {
			filtered = append(filtered, child)
		}
	}
	
	return filtered, nil
}

func (s *servicePg) Search(ctx context.Context, userID string, input *SearchPlacesInput) ([]*Place, int64, error) {
	// TODO: Implement search with privacy filtering
	return s.repo.Search(ctx, input)
}

func (s *servicePg) GetNearby(ctx context.Context, userID string, input *NearbyPlacesInput) ([]*Place, error) {
	// TODO: Implement nearby search with privacy filtering
	return s.repo.GetNearby(ctx, input)
}

func (s *servicePg) AddCollaborator(ctx context.Context, userID, placeID, collaboratorID, role string) error {
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return err
	}
	
	// Only owner can add collaborators
	if !place.IsOwner(userID) {
		return ErrUnauthorized
	}
	
	// Check if already a collaborator
	if place.HasCollaborator(collaboratorID) {
		return errors.New("user is already a collaborator")
	}
	
	// Validate role
	if role != "admin" && role != "editor" && role != "viewer" {
		return errors.New("invalid role")
	}
	
	return s.repo.AddCollaborator(ctx, placeID, collaboratorID, role, nil)
}

func (s *servicePg) RemoveCollaborator(ctx context.Context, userID, placeID, collaboratorID string) error {
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return err
	}
	
	// Only owner can remove collaborators
	if !place.IsOwner(userID) {
		return ErrUnauthorized
	}
	
	return s.repo.RemoveCollaborator(ctx, placeID, collaboratorID)
}

func (s *servicePg) UpdateCollaboratorRole(ctx context.Context, userID, placeID, collaboratorID, role string) error {
	place, err := s.repo.GetByID(ctx, placeID)
	if err != nil {
		return err
	}
	
	// Only owner can update collaborator roles
	if !place.IsOwner(userID) {
		return ErrUnauthorized
	}
	
	// Validate role
	if role != "admin" && role != "editor" && role != "viewer" {
		return errors.New("invalid role")
	}
	
	return s.repo.UpdateCollaboratorRole(ctx, placeID, collaboratorID, role)
}

// Implement missing interface methods with basic functionality
func (s *servicePg) List(ctx context.Context, userID string, filter *PlaceFilter, limit, offset int) ([]*Place, int64, error) {
	// TODO: Implement proper list with filter
	return s.repo.GetByUser(ctx, userID, limit, offset)
}

func (s *servicePg) GetTripPlaces(ctx context.Context, userID, tripID string) ([]*Place, error) {
	// TODO: Implement trip places retrieval
	return []*Place{}, nil
}

func (s *servicePg) AddToTrip(ctx context.Context, userID, placeID, tripID string) error {
	// TODO: Implement adding place to trip
	return nil
}

func (s *servicePg) RemoveFromTrip(ctx context.Context, userID, placeID, tripID string) error {
	// TODO: Implement removing place from trip
	return nil
}

func (s *servicePg) UpdateVisitStatus(ctx context.Context, userID, placeID string, visited bool, visitDate *time.Time) error {
	// TODO: Implement visit status update
	return nil
}

func (s *servicePg) AddImages(ctx context.Context, userID, placeID string, images []string) error {
	// TODO: Implement image management
	return nil
}

func (s *servicePg) RemoveImage(ctx context.Context, userID, placeID string, imageURL string) error {
	// TODO: Implement image removal
	return nil
}

func (s *servicePg) UpdateRating(ctx context.Context, userID, placeID string, rating float32) error {
	// TODO: Implement rating update
	return nil
}

func (s *servicePg) AddNote(ctx context.Context, userID, placeID, note string) error {
	// TODO: Implement note management
	return nil
}