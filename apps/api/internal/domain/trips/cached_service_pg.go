package trips

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/cache"
)

type cachedServicePg struct {
	service Service
	cache   cache.Cache
}

// NewCachedServicePg creates a new cached trip service for PostgreSQL
func NewCachedServicePg(service Service, cache cache.Cache) Service {
	return &cachedServicePg{
		service: service,
		cache:   cache,
	}
}

func (c *cachedServicePg) Create(ctx context.Context, userID string, input *CreateTripInput) (*Trip, error) {
	trip, err := c.service.Create(ctx, userID, input)
	if err != nil {
		return nil, err
	}

	// Cache the new trip
	if err := c.cacheTrip(ctx, trip); err != nil {
		// Log cache error but don't fail the operation
		fmt.Printf("Failed to cache trip: %v\n", err)
	}

	return trip, nil
}

func (c *cachedServicePg) GetByID(ctx context.Context, userID, tripID string) (*Trip, error) {
	// Try to get from cache
	data, err := c.cache.GetTrip(ctx, tripID)
	if err == nil && data != nil {
		var trip Trip
		if err := json.Unmarshal(data, &trip); err == nil {
			// Check permissions
			if c.canUserAccessTrip(&trip, userID) {
				return &trip, nil
			}
			return nil, errors.New("unauthorized")
		}
	}

	// Get from service
	trip, err := c.service.GetByID(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}

	// Cache the trip
	if err := c.cacheTrip(ctx, trip); err != nil {
		fmt.Printf("Failed to cache trip: %v\n", err)
	}

	return trip, nil
}

func (c *cachedServicePg) Update(ctx context.Context, userID, tripID string, input *UpdateTripInput) (*Trip, error) {
	trip, err := c.service.Update(ctx, userID, tripID, input)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	// Cache the updated trip
	if err := c.cacheTrip(ctx, trip); err != nil {
		fmt.Printf("Failed to cache trip: %v\n", err)
	}

	return trip, nil
}

func (c *cachedServicePg) Delete(ctx context.Context, userID, tripID string) error {
	if err := c.service.Delete(ctx, userID, tripID); err != nil {
		return err
	}

	// Invalidate cache
	if err := c.cache.InvalidateTripRelated(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return nil
}

func (c *cachedServicePg) List(ctx context.Context, userID string, filter *TripFilter, limit, offset int) ([]*Trip, int64, error) {
	// List operations are not cached due to complexity
	return c.service.List(ctx, userID, filter, limit, offset)
}

func (c *cachedServicePg) GetUserTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	// User trips are not cached due to frequent updates
	return c.service.GetUserTrips(ctx, userID, limit, offset)
}

func (c *cachedServicePg) GetSharedTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	// Shared trips are not cached due to frequent updates
	return c.service.GetSharedTrips(ctx, userID, limit, offset)
}

func (c *cachedServicePg) Search(ctx context.Context, userID string, query string, limit, offset int) ([]*Trip, int64, error) {
	// Search results are not cached
	return c.service.Search(ctx, userID, query, limit, offset)
}

func (c *cachedServicePg) AddCollaborator(ctx context.Context, userID, tripID, collaboratorID, role string) error {
	if err := c.service.AddCollaborator(ctx, userID, tripID, collaboratorID, role); err != nil {
		return err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return nil
}

func (c *cachedServicePg) RemoveCollaborator(ctx context.Context, userID, tripID, collaboratorID string) error {
	if err := c.service.RemoveCollaborator(ctx, userID, tripID, collaboratorID); err != nil {
		return err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return nil
}

func (c *cachedServicePg) UpdateCollaboratorRole(ctx context.Context, userID, tripID, collaboratorID, role string) error {
	if err := c.service.UpdateCollaboratorRole(ctx, userID, tripID, collaboratorID, role); err != nil {
		return err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return nil
}

func (c *cachedServicePg) AddWaypoint(ctx context.Context, userID, tripID string, input *AddWaypointInput) (*Waypoint, error) {
	waypoint, err := c.service.AddWaypoint(ctx, userID, tripID, input)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return waypoint, nil
}

func (c *cachedServicePg) UpdateWaypoint(ctx context.Context, userID, tripID, waypointID string, input *UpdateWaypointInput) (*Waypoint, error) {
	waypoint, err := c.service.UpdateWaypoint(ctx, userID, tripID, waypointID, input)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return waypoint, nil
}

func (c *cachedServicePg) RemoveWaypoint(ctx context.Context, userID, tripID, waypointID string) error {
	if err := c.service.RemoveWaypoint(ctx, userID, tripID, waypointID); err != nil {
		return err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return nil
}

func (c *cachedServicePg) ReorderWaypoints(ctx context.Context, userID, tripID string, waypointIDs []string) error {
	if err := c.service.ReorderWaypoints(ctx, userID, tripID, waypointIDs); err != nil {
		return err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return nil
}

// Helper methods
func (c *cachedServicePg) cacheTrip(ctx context.Context, trip *Trip) error {
	data, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	return c.cache.SetTrip(ctx, trip.ID, data, 1*time.Hour)
}

func (c *cachedServicePg) canUserAccessTrip(trip *Trip, userID string) bool {
	// Check if user is owner
	if trip.OwnerID == userID {
		return true
	}

	// Check if trip is public
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

// Implement missing interface methods
func (c *cachedServicePg) InviteCollaborator(ctx context.Context, userID, tripID string, input *InviteCollaboratorInput) error {
	if err := c.service.InviteCollaborator(ctx, userID, tripID, input); err != nil {
		return err
	}

	// Invalidate cache
	if err := c.cache.DeleteTrip(ctx, tripID); err != nil {
		fmt.Printf("Failed to invalidate trip cache: %v\n", err)
	}

	return nil
}

func (c *cachedServicePg) GetTripStats(ctx context.Context, userID, tripID string) (*TripStats, error) {
	// Stats are not cached as they may change frequently
	return c.service.GetTripStats(ctx, userID, tripID)
}

func (c *cachedServicePg) ExportTrip(ctx context.Context, userID, tripID, format string) ([]byte, error) {
	// Export operations are not cached
	return c.service.ExportTrip(ctx, userID, tripID, format)
}

func (c *cachedServicePg) CloneTrip(ctx context.Context, userID, tripID string) (*Trip, error) {
	trip, err := c.service.CloneTrip(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}

	// Cache the new trip
	if err := c.cacheTrip(ctx, trip); err != nil {
		fmt.Printf("Failed to cache trip: %v\n", err)
	}

	return trip, nil
}