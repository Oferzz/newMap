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
	repo     Repository
	userRepo users.Repository
}

// NewService creates a new trip service
func NewService(repo Repository, userRepo users.Repository) Service {
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
		
		// Activity-specific fields
		ActivityType:       input.ActivityType,
		DifficultyLevel:    input.DifficultyLevel,
		DurationHours:      input.DurationHours,
		DistanceKm:         input.DistanceKm,
		ElevationGainM:     input.ElevationGainM,
		MaxElevationM:      input.MaxElevationM,
		RouteType:          input.RouteType,
		RouteGeoJSON:       input.RouteGeoJSON,
		WaterFeatures:      input.WaterFeatures,
		TerrainTypes:       input.TerrainTypes,
		EssentialGear:      input.EssentialGear,
		BestSeasons:        input.BestSeasons,
		TrailConditions:    input.TrailConditions,
		AccessibilityNotes: input.AccessibilityNotes,
		ParkingInfo:        input.ParkingInfo,
		PermitsRequired:    input.PermitsRequired,
		Hazards:            input.Hazards,
		EmergencyContacts:  input.EmergencyContacts,
		Visibility:         "private",
		SharedWith:         input.SharedWith,
		CompletionCount:    0,
		RatingCount:        0,
		Featured:           false,
		Verified:           false,
	}
	
	// Set default privacy if provided
	if input.Privacy != "" {
		trip.Privacy = input.Privacy
	}
	
	// Set visibility (for activity features)
	if input.Visibility != "" {
		trip.Visibility = input.Visibility
	}
	
	// Set default timezone if not provided
	if trip.Timezone == "" {
		trip.Timezone = "UTC"
	}
	
	// Set default activity type if not provided
	if trip.ActivityType == "" {
		trip.ActivityType = "general"
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
	
	// Build updates map for dynamic update
	updates := make(map[string]interface{})
	
	// Basic fields
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.StartDate != nil {
		updates["start_date"] = input.StartDate
	}
	if input.EndDate != nil {
		updates["end_date"] = input.EndDate
	}
	if input.Privacy != nil {
		updates["privacy"] = *input.Privacy
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if len(input.Tags) > 0 {
		updates["tags"] = input.Tags
	}
	if input.CoverImage != nil {
		updates["cover_image"] = *input.CoverImage
	}
	if input.Timezone != nil {
		updates["timezone"] = *input.Timezone
	}
	
	// Activity-specific fields
	if input.ActivityType != nil {
		updates["activity_type"] = *input.ActivityType
	}
	if input.DifficultyLevel != nil {
		updates["difficulty_level"] = *input.DifficultyLevel
	}
	if input.DurationHours != nil {
		updates["duration_hours"] = input.DurationHours
	}
	if input.DistanceKm != nil {
		updates["distance_km"] = input.DistanceKm
	}
	if input.ElevationGainM != nil {
		updates["elevation_gain_m"] = input.ElevationGainM
	}
	if input.MaxElevationM != nil {
		updates["max_elevation_m"] = input.MaxElevationM
	}
	if input.RouteType != nil {
		updates["route_type"] = *input.RouteType
	}
	if input.RouteGeoJSON != nil {
		updates["route_geojson"] = input.RouteGeoJSON
	}
	if len(input.WaterFeatures) > 0 {
		updates["water_features"] = input.WaterFeatures
	}
	if len(input.TerrainTypes) > 0 {
		updates["terrain_types"] = input.TerrainTypes
	}
	if len(input.EssentialGear) > 0 {
		updates["essential_gear"] = input.EssentialGear
	}
	if len(input.BestSeasons) > 0 {
		updates["best_seasons"] = input.BestSeasons
	}
	if input.TrailConditions != nil {
		updates["trail_conditions"] = *input.TrailConditions
	}
	if input.AccessibilityNotes != nil {
		updates["accessibility_notes"] = *input.AccessibilityNotes
	}
	if input.ParkingInfo != nil {
		updates["parking_info"] = input.ParkingInfo
	}
	if len(input.PermitsRequired) > 0 {
		updates["permits_required"] = input.PermitsRequired
	}
	if len(input.Hazards) > 0 {
		updates["hazards"] = input.Hazards
	}
	if input.EmergencyContacts != nil {
		updates["emergency_contacts"] = input.EmergencyContacts
	}
	if input.Visibility != nil {
		updates["visibility"] = *input.Visibility
	}
	if len(input.SharedWith) > 0 {
		updates["shared_with"] = input.SharedWith
	}
	
	if err := s.repo.Update(ctx, tripID, updates); err != nil {
		return nil, fmt.Errorf("failed to update trip: %w", err)
	}
	
	// Get updated trip
	updatedTrip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated trip: %w", err)
	}
	
	return updatedTrip, nil
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
	// Build filters for repository
	filters := TripFilters{
		CollaboratorID: userID,
		Status:         filter.Status,
		Privacy:        filter.Privacy,
		Tags:           filter.Tags,
		Limit:          limit,
		Offset:         offset,
	}
	
	trips, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	
	// TODO: Get total count properly
	total := int64(len(trips))
	
	return trips, total, nil
}

func (s *servicePg) GetUserTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	// Get trips where user is owner
	filters := TripFilters{
		OwnerID: userID,
		Limit:   limit,
		Offset:  offset,
	}
	
	trips, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	
	total := int64(len(trips))
	return trips, total, nil
}

func (s *servicePg) GetSharedTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	// Get trips where user is collaborator
	filters := TripFilters{
		CollaboratorID: userID,
		Limit:          limit,
		Offset:         offset,
	}
	
	trips, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	
	total := int64(len(trips))
	return trips, total, nil
}

func (s *servicePg) Search(ctx context.Context, userID string, query string, limit, offset int) ([]*Trip, int64, error) {
	filters := TripFilters{
		CollaboratorID: userID,
		Search:         query,
		Limit:          limit,
		Offset:         offset,
	}
	
	trips, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	
	total := int64(len(trips))
	return trips, total, nil
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
	
	collaborator := Collaborator{
		TripID:                 tripID,
		UserID:                 collaboratorID,
		Role:                   role,
		CanEdit:                canEdit,
		CanDelete:              canDelete,
		CanInvite:              canInvite,
		CanModerateSuggestions: canModerate,
		InvitedAt:              time.Now(),
	}
	
	return s.repo.AddCollaborator(ctx, tripID, collaborator)
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
	
	updates := map[string]interface{}{
		"role":                     role,
		"can_edit":                 canEdit,
		"can_delete":               canDelete,
		"can_invite":               canInvite,
		"can_moderate_suggestions": canModerate,
	}
	
	return s.repo.UpdateCollaborator(ctx, tripID, collaboratorID, updates)
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
	
	collaborator := Collaborator{
		TripID:                 tripID,
		UserID:                 input.UserID,
		Role:                   input.Role,
		CanEdit:                input.CanEdit,
		CanDelete:              input.CanDelete,
		CanInvite:              input.CanInvite,
		CanModerateSuggestions: input.CanModerate,
		InvitedAt:              time.Now(),
	}
	
	return s.repo.AddCollaborator(ctx, tripID, collaborator)
}

func (s *servicePg) AddWaypoint(ctx context.Context, userID, tripID string, input *AddWaypointInput) (*Waypoint, error) {
	// TODO: Implement waypoint functionality with WaypointRepository
	return nil, errors.New("waypoint functionality not yet implemented")
}

func (s *servicePg) UpdateWaypoint(ctx context.Context, userID, tripID, waypointID string, input *UpdateWaypointInput) (*Waypoint, error) {
	// TODO: Implement waypoint functionality with WaypointRepository
	return nil, errors.New("waypoint functionality not yet implemented")
}

func (s *servicePg) RemoveWaypoint(ctx context.Context, userID, tripID, waypointID string) error {
	// TODO: Implement waypoint functionality with WaypointRepository
	return errors.New("waypoint functionality not yet implemented")
}

func (s *servicePg) ReorderWaypoints(ctx context.Context, userID, tripID string, waypointIDs []string) error {
	// TODO: Implement waypoint functionality with WaypointRepository
	return errors.New("waypoint functionality not yet implemented")
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