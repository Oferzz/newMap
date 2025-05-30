package places

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/Oferzz/newMap/apps/api/internal/domain/trips"
)

type servicePg struct {
	repo          Repository
	tripRepo      trips.Repository
	mapboxService *MapboxService
}

func NewServicePg(repo Repository, tripRepo trips.Repository, mapboxAPIKey string) Service {
	var mapboxService *MapboxService
	if mapboxAPIKey != "" {
		log.Printf("[PlaceService] Initializing with Mapbox API key (length: %d)", len(mapboxAPIKey))
		mapboxService = NewMapboxService(mapboxAPIKey)
	} else {
		log.Printf("[PlaceService] WARNING: No Mapbox API key provided. Mapbox search will not be available.")
	}
	
	return &servicePg{
		repo:          repo,
		tripRepo:      tripRepo,
		mapboxService: mapboxService,
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
	
	// TODO: Check if place has children when GetChildPlaces is implemented
	// children, err := s.repo.GetChildPlaces(ctx, placeID)
	// if err != nil {
	// 	return fmt.Errorf("failed to check child places: %w", err)
	// }
	
	// if len(children) > 0 {
	// 	return errors.New("cannot delete place with child places")
	// }
	
	return s.repo.Delete(ctx, placeID)
}

func (s *servicePg) GetUserPlaces(ctx context.Context, userID string, limit, offset int) ([]*Place, int64, error) {
	places, err := s.repo.GetByCreator(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	
	// Apply pagination manually
	start := offset
	end := offset + limit
	if start > len(places) {
		return []*Place{}, int64(len(places)), nil
	}
	if end > len(places) {
		end = len(places)
	}
	
	return places[start:end], int64(len(places)), nil
}

func (s *servicePg) GetChildPlaces(ctx context.Context, userID, parentID string) ([]*Place, error) {
	// First check if user has access to parent place
	parent, err := s.GetByID(ctx, userID, parentID)
	if err != nil {
		return nil, err
	}
	
	children, err := s.repo.GetChildren(ctx, parentID)
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
	filters := SearchFilters{
		Category: input.Category,
		Tags:     input.Tags,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}
	
	result, err := s.repo.Search(ctx, input.Query, filters)
	if err != nil {
		return nil, 0, err
	}
	
	return result.Places, result.Total, nil
}

func (s *servicePg) GetNearby(ctx context.Context, userID string, input *NearbyPlacesInput) ([]*Place, error) {
	// TODO: Implement nearby search with privacy filtering
	// Convert radius from meters to kilometers
	radiusKM := float64(input.Radius) / 1000.0
	return s.repo.GetNearby(ctx, input.Latitude, input.Longitude, radiusKM, input.Limit)
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
	
	// TODO: Implement when repository supports collaborators
	return errors.New("collaborator functionality not yet implemented")
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
	
	// TODO: Implement when repository supports collaborators
	return errors.New("collaborator functionality not yet implemented")
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
	
	// TODO: Implement when repository supports collaborators
	return errors.New("collaborator functionality not yet implemented")
}

// Implement missing interface methods with basic functionality
func (s *servicePg) List(ctx context.Context, userID string, filter *PlaceFilter, limit, offset int) ([]*Place, int64, error) {
	log.Printf("[PlaceService] List called with userID: %s, filter: %+v, limit: %d, offset: %d", userID, filter, limit, offset)
	
	// Handle public search (when userID is empty)
	if userID == "" && filter != nil && filter.SearchQuery != "" {
		log.Printf("[PlaceService] Public search detected (no userID, has search query)")
		return s.handlePublicSearch(ctx, filter, limit, offset)
	}
	
	// TODO: Implement proper list with filter
	places, err := s.repo.GetByCreator(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	
	// Apply pagination manually
	start := offset
	end := offset + limit
	if start > len(places) {
		return []*Place{}, int64(len(places)), nil
	}
	if end > len(places) {
		end = len(places)
	}
	
	return places[start:end], int64(len(places)), nil
}

func (s *servicePg) handlePublicSearch(ctx context.Context, filter *PlaceFilter, limit, offset int) ([]*Place, int64, error) {
	query := filter.SearchQuery
	log.Printf("[PlaceService] handlePublicSearch called with query: %s, limit: %d, offset: %d", query, limit, offset)
	
	// First, try to search in our database for user-created public places
	log.Printf("[PlaceService] Searching database for public places...")
	dbPlaces, total, err := s.searchDatabasePlaces(ctx, filter, limit, offset)
	if err == nil && total > 0 {
		log.Printf("[PlaceService] Found %d places in database", total)
		return dbPlaces, total, nil
	}
	
	log.Printf("[PlaceService] No database results. Mapbox service configured: %v", s.mapboxService != nil)
	
	// If no results from database and we have Mapbox service, search Mapbox
	if s.mapboxService != nil && query != "" {
		log.Printf("[PlaceService] Searching Mapbox for query: %s", query)
		// For Mapbox, we ignore offset and just use limit
		mapboxPlaces, err := s.mapboxService.SearchPlaces(ctx, query, limit)
		if err != nil {
			log.Printf("[PlaceService] ERROR: Mapbox search failed: %v", err)
			// Log error but don't fail - return empty results
			return []*Place{}, 0, nil
		}
		
		log.Printf("[PlaceService] Mapbox returned %d places", len(mapboxPlaces))
		return mapboxPlaces, int64(len(mapboxPlaces)), nil
	}
	
	// Fallback to empty results if no Mapbox service configured
	log.Printf("[PlaceService] No Mapbox service configured or empty query. Returning empty results.")
	return []*Place{}, 0, nil
}

func (s *servicePg) searchDatabasePlaces(ctx context.Context, filter *PlaceFilter, limit, offset int) ([]*Place, int64, error) {
	// For now, return empty results since we don't have database search implemented
	// TODO: Implement actual database search for public places
	return []*Place{}, 0, nil
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