package trips

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
)

type servicePg struct {
	repo     RepositoryInterface
	userRepo users.Repository
}

// NewService creates a new trip service
func NewService(repo RepositoryInterface, userRepo users.Repository) Service {
	return &servicePg{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *servicePg) Create(ctx context.Context, userID string, input *CreateTripInput) (*Trip, error) {
	trip := &Trip{
		ID:          uuid.New().String(),
		Title:       input.Title,
		Description: input.Description,
		OwnerID:     userID,
		CoverImage:  input.CoverImage,
		Privacy:     "private",
		Status:      "planning",
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		Timezone:    input.Timezone,
		Tags:        input.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Set default privacy if provided
	if input.Privacy != "" {
		trip.Privacy = input.Privacy
	}
	
	// Set default timezone if not provided
	if trip.Timezone == "" {
		trip.Timezone = "UTC"
	}
	
	if err := s.repo.Create(ctx, trip); err != nil {
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}
	
	return trip, nil
}

func (s *servicePg) GetByID(ctx context.Context, userID, tripID string) (*Trip, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	// Check if user has permission to view this trip
	if !s.canUserAccessTrip(trip, userID) {
		return nil, ErrUnauthorized
	}
	
	return trip, nil
}

func (s *servicePg) Update(ctx context.Context, userID, tripID string, input *UpdateTripInput) (*Trip, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	// Check if user can edit
	if !s.canUserEditTrip(trip, userID) {
		return nil, ErrUnauthorized
	}
	
	// Update fields
	if input.Title != nil {
		trip.Title = *input.Title
	}
	if input.Description != nil {
		trip.Description = *input.Description
	}
	if input.StartDate != nil {
		trip.StartDate = input.StartDate
	}
	if input.EndDate != nil {
		trip.EndDate = input.EndDate
	}
	if input.Privacy != nil {
		trip.Privacy = *input.Privacy
	}
	if input.Status != nil {
		trip.Status = *input.Status
	}
	if len(input.Tags) > 0 {
		trip.Tags = input.Tags
	}
	if input.CoverImage != nil {
		trip.CoverImage = *input.CoverImage
	}
	if input.Timezone != nil {
		trip.Timezone = *input.Timezone
	}
	
	trip.UpdatedAt = time.Now()
	
	if err := s.repo.Update(ctx, trip); err != nil {
		return nil, fmt.Errorf("failed to update trip: %w", err)
	}
	
	return trip, nil
}

func (s *servicePg) Delete(ctx context.Context, userID, tripID string) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}
	
	// Only owner can delete
	if trip.OwnerID != userID {
		return ErrUnauthorized
	}
	
	return s.repo.Delete(ctx, tripID)
}

func (s *servicePg) List(ctx context.Context, userID string, filter *TripFilter, limit, offset int) ([]*Trip, int64, error) {
	// TODO: Implement proper filtering with privacy checks
	return s.repo.GetByUser(ctx, userID, limit, offset)
}

func (s *servicePg) GetUserTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	return s.repo.GetByUser(ctx, userID, limit, offset)
}

func (s *servicePg) GetSharedTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	return s.repo.GetSharedWithUser(ctx, userID, limit, offset)
}

func (s *servicePg) Search(ctx context.Context, userID string, query string, limit, offset int) ([]*Trip, int64, error) {
	return s.repo.Search(ctx, query, userID, limit, offset)
}

func (s *servicePg) AddCollaborator(ctx context.Context, userID, tripID, collaboratorID, role string) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}
	
	// Check if user can invite
	if !s.canUserInviteToTrip(trip, userID) {
		return ErrUnauthorized
	}
	
	// Check if collaborator exists
	if _, err := s.userRepo.GetByID(ctx, collaboratorID); err != nil {
		return errors.New("user not found")
	}
	
	// Check if already a collaborator
	for _, collab := range trip.Collaborators {
		if collab.UserID == collaboratorID {
			return errors.New("user is already a collaborator")
		}
	}
	
	// Set default permissions based on role
	canEdit := role == "editor" || role == "admin"
	canDelete := role == "admin"
	canInvite := role == "admin"
	canModerate := role == "admin" || role == "editor"
	
	return s.repo.AddCollaborator(ctx, tripID, collaboratorID, role, canEdit, canDelete, canInvite, canModerate)
}

func (s *servicePg) RemoveCollaborator(ctx context.Context, userID, tripID, collaboratorID string) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}
	
	// Only owner or admin can remove collaborators
	if trip.OwnerID != userID && !s.isUserAdminOfTrip(trip, userID) {
		return ErrUnauthorized
	}
	
	return s.repo.RemoveCollaborator(ctx, tripID, collaboratorID)
}

func (s *servicePg) UpdateCollaboratorRole(ctx context.Context, userID, tripID, collaboratorID, role string) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}
	
	// Only owner can update roles
	if trip.OwnerID != userID {
		return ErrUnauthorized
	}
	
	// Update permissions based on new role
	canEdit := role == "editor" || role == "admin"
	canDelete := role == "admin"
	canInvite := role == "admin"
	canModerate := role == "admin" || role == "editor"
	
	return s.repo.UpdateCollaboratorRole(ctx, tripID, collaboratorID, role, canEdit, canDelete, canInvite, canModerate)
}

func (s *servicePg) InviteCollaborator(ctx context.Context, userID, tripID string, input *InviteCollaboratorInput) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}
	
	// Check if user can invite
	if !s.canUserInviteToTrip(trip, userID) {
		return ErrUnauthorized
	}
	
	// Check if collaborator exists
	if _, err := s.userRepo.GetByID(ctx, input.UserID); err != nil {
		return errors.New("user not found")
	}
	
	// Check if already a collaborator
	for _, collab := range trip.Collaborators {
		if collab.UserID == input.UserID {
			return errors.New("user is already a collaborator")
		}
	}
	
	return s.repo.AddCollaborator(ctx, tripID, input.UserID, input.Role, input.CanEdit, input.CanDelete, input.CanInvite, input.CanModerate)
}

func (s *servicePg) AddWaypoint(ctx context.Context, userID, tripID string, input *AddWaypointInput) (*Waypoint, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	// Check if user can edit
	if !s.canUserEditTrip(trip, userID) {
		return nil, ErrUnauthorized
	}
	
	waypoint := &Waypoint{
		ID:            uuid.New().String(),
		TripID:        tripID,
		PlaceID:       input.PlaceID,
		OrderPosition: input.OrderPosition,
		ArrivalTime:   input.ArrivalTime,
		DepartureTime: input.DepartureTime,
		Notes:         input.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	if err := s.repo.AddWaypoint(ctx, waypoint); err != nil {
		return nil, fmt.Errorf("failed to add waypoint: %w", err)
	}
	
	return waypoint, nil
}

func (s *servicePg) UpdateWaypoint(ctx context.Context, userID, tripID, waypointID string, input *UpdateWaypointInput) (*Waypoint, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	// Check if user can edit
	if !s.canUserEditTrip(trip, userID) {
		return nil, ErrUnauthorized
	}
	
	waypoint, err := s.repo.GetWaypoint(ctx, waypointID)
	if err != nil {
		return nil, err
	}
	
	if waypoint.TripID != tripID {
		return nil, errors.New("waypoint does not belong to this trip")
	}
	
	// Update fields
	if input.OrderPosition != nil {
		waypoint.OrderPosition = *input.OrderPosition
	}
	if input.ArrivalTime != nil {
		waypoint.ArrivalTime = input.ArrivalTime
	}
	if input.DepartureTime != nil {
		waypoint.DepartureTime = input.DepartureTime
	}
	if input.Notes != nil {
		waypoint.Notes = *input.Notes
	}
	
	waypoint.UpdatedAt = time.Now()
	
	if err := s.repo.UpdateWaypoint(ctx, waypoint); err != nil {
		return nil, fmt.Errorf("failed to update waypoint: %w", err)
	}
	
	return waypoint, nil
}

func (s *servicePg) RemoveWaypoint(ctx context.Context, userID, tripID, waypointID string) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}
	
	// Check if user can edit
	if !s.canUserEditTrip(trip, userID) {
		return ErrUnauthorized
	}
	
	return s.repo.RemoveWaypoint(ctx, tripID, waypointID)
}

func (s *servicePg) ReorderWaypoints(ctx context.Context, userID, tripID string, waypointIDs []string) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return err
	}
	
	// Check if user can edit
	if !s.canUserEditTrip(trip, userID) {
		return ErrUnauthorized
	}
	
	return s.repo.ReorderWaypoints(ctx, tripID, waypointIDs)
}

func (s *servicePg) GetTripStats(ctx context.Context, userID, tripID string) (*TripStats, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	// Check if user can view
	if !s.canUserAccessTrip(trip, userID) {
		return nil, ErrUnauthorized
	}
	
	// TODO: Implement actual stats gathering
	return &TripStats{
		TotalPlaces:        0,
		TotalWaypoints:     len(trip.Waypoints),
		TotalCollaborators: len(trip.Collaborators),
		TotalSuggestions:   trip.SuggestionCount,
		TotalViews:         trip.ViewCount,
		TotalShares:        trip.ShareCount,
	}, nil
}

func (s *servicePg) ExportTrip(ctx context.Context, userID, tripID, format string) ([]byte, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	// Check if user can view
	if !s.canUserAccessTrip(trip, userID) {
		return nil, ErrUnauthorized
	}
	
	// TODO: Implement export functionality
	return []byte{}, errors.New("export not implemented")
}

func (s *servicePg) CloneTrip(ctx context.Context, userID, tripID string) (*Trip, error) {
	sourceTrip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	
	// Check if user can view source trip
	if !s.canUserAccessTrip(sourceTrip, userID) {
		return nil, ErrUnauthorized
	}
	
	// Create new trip
	newTrip := &Trip{
		ID:          uuid.New().String(),
		Title:       sourceTrip.Title + " (Copy)",
		Description: sourceTrip.Description,
		OwnerID:     userID,
		CoverImage:  sourceTrip.CoverImage,
		Privacy:     "private",
		Status:      "planning",
		StartDate:   sourceTrip.StartDate,
		EndDate:     sourceTrip.EndDate,
		Timezone:    sourceTrip.Timezone,
		Tags:        sourceTrip.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := s.repo.Create(ctx, newTrip); err != nil {
		return nil, fmt.Errorf("failed to clone trip: %w", err)
	}
	
	// TODO: Clone waypoints and other related data
	
	return newTrip, nil
}

// Helper methods
func (s *servicePg) canUserAccessTrip(trip *Trip, userID string) bool {
	// Owner can always access
	if trip.OwnerID == userID {
		return true
	}
	
	// Public trips can be accessed by anyone
	if trip.Privacy == "public" {
		return true
	}
	
	// Check if user is collaborator
	for _, collab := range trip.Collaborators {
		if collab.UserID == userID {
			return true
		}
	}
	
	return false
}

func (s *servicePg) canUserEditTrip(trip *Trip, userID string) bool {
	// Owner can always edit
	if trip.OwnerID == userID {
		return true
	}
	
	// Check if user is collaborator with edit permission
	for _, collab := range trip.Collaborators {
		if collab.UserID == userID && collab.CanEdit {
			return true
		}
	}
	
	return false
}

func (s *servicePg) canUserInviteToTrip(trip *Trip, userID string) bool {
	// Owner can always invite
	if trip.OwnerID == userID {
		return true
	}
	
	// Check if user is collaborator with invite permission
	for _, collab := range trip.Collaborators {
		if collab.UserID == userID && collab.CanInvite {
			return true
		}
	}
	
	return false
}

func (s *servicePg) isUserAdminOfTrip(trip *Trip, userID string) bool {
	for _, collab := range trip.Collaborators {
		if collab.UserID == userID && collab.Role == "admin" {
			return true
		}
	}
	return false
}