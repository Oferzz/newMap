package trips

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/cache"
	"github.com/Oferzz/newMap/apps/api/internal/database"
	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type cachedService struct {
	service Service
	cache   cache.Cache
}

// NewCachedService wraps the trip service with caching
func NewCachedService(service Service, cache cache.Cache) Service {
	return &cachedService{
		service: service,
		cache:   cache,
	}
}

func (s *cachedService) Create(ctx context.Context, userID primitive.ObjectID, input *CreateTripInput) (*Trip, error) {
	// Create trip without cache (no need to cache new trips)
	trip, err := s.service.Create(ctx, userID, input)
	if err != nil {
		return nil, err
	}

	// Cache the new trip
	if data, err := json.Marshal(trip); err == nil {
		_ = s.cache.SetTrip(ctx, trip.ID.Hex(), data, database.CacheTTLMedium)
	}

	return trip, nil
}

func (s *cachedService) GetByID(ctx context.Context, tripID, userID primitive.ObjectID) (*Trip, error) {
	// Try to get from cache first
	cacheKey := tripID.Hex()
	if cached, err := s.cache.GetTrip(ctx, cacheKey); err == nil && cached != nil {
		var trip Trip
		if err := json.Unmarshal(cached, &trip); err == nil {
			// Still need to check permissions
			if !trip.IsPublic && !trip.HasCollaborator(userID) {
				return nil, ErrUnauthorized
			}
			return &trip, nil
		}
	}

	// Get from service
	trip, err := s.service.GetByID(ctx, tripID, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(trip); err == nil {
		_ = s.cache.SetTrip(ctx, cacheKey, data, database.CacheTTLMedium)
	}

	return trip, nil
}

func (s *cachedService) Update(ctx context.Context, tripID, userID primitive.ObjectID, input *UpdateTripInput) (*Trip, error) {
	// Update trip
	trip, err := s.service.Update(ctx, tripID, userID, input)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	_ = s.cache.InvalidateTripRelated(ctx, tripID.Hex())

	// Cache the updated trip
	if data, err := json.Marshal(trip); err == nil {
		_ = s.cache.SetTrip(ctx, tripID.Hex(), data, database.CacheTTLMedium)
	}

	return trip, nil
}

func (s *cachedService) Delete(ctx context.Context, tripID, userID primitive.ObjectID) error {
	// Delete trip
	err := s.service.Delete(ctx, tripID, userID)
	if err != nil {
		return err
	}

	// Invalidate cache
	_ = s.cache.InvalidateTripRelated(ctx, tripID.Hex())

	return nil
}

func (s *cachedService) List(ctx context.Context, opts TripListOptions, userID *primitive.ObjectID) ([]*Trip, int64, error) {
	// For lists, we'll skip caching for now as it's complex with filters
	// In production, you might want to cache common queries
	return s.service.List(ctx, opts, userID)
}

func (s *cachedService) InviteCollaborator(ctx context.Context, tripID, inviterID primitive.ObjectID, input *InviteCollaboratorInput) error {
	// Invite collaborator
	err := s.service.InviteCollaborator(ctx, tripID, inviterID, input)
	if err != nil {
		return err
	}

	// Invalidate trip cache
	_ = s.cache.InvalidateTripRelated(ctx, tripID.Hex())

	return nil
}

func (s *cachedService) RemoveCollaborator(ctx context.Context, tripID, removerID, userID primitive.ObjectID) error {
	// Remove collaborator
	err := s.service.RemoveCollaborator(ctx, tripID, removerID, userID)
	if err != nil {
		return err
	}

	// Invalidate trip cache and user permissions
	_ = s.cache.InvalidateTripRelated(ctx, tripID.Hex())
	_ = s.cache.InvalidateUserPermissions(ctx, userID.Hex(), tripID.Hex())

	return nil
}

func (s *cachedService) UpdateCollaboratorRole(ctx context.Context, tripID, updaterID primitive.ObjectID, input *UpdateCollaboratorRoleInput) error {
	// Update role
	err := s.service.UpdateCollaboratorRole(ctx, tripID, updaterID, input)
	if err != nil {
		return err
	}

	// Parse user ID to invalidate their permissions
	if userID, err := primitive.ObjectIDFromHex(input.UserID); err == nil {
		_ = s.cache.InvalidateUserPermissions(ctx, userID.Hex(), tripID.Hex())
	}

	// Invalidate trip cache
	_ = s.cache.InvalidateTripRelated(ctx, tripID.Hex())

	return nil
}

func (s *cachedService) LeaveTrip(ctx context.Context, tripID, userID primitive.ObjectID) error {
	// Leave trip
	err := s.service.LeaveTrip(ctx, tripID, userID)
	if err != nil {
		return err
	}

	// Invalidate caches
	_ = s.cache.InvalidateTripRelated(ctx, tripID.Hex())
	_ = s.cache.InvalidateUserPermissions(ctx, userID.Hex(), tripID.Hex())

	return nil
}

// CachedPermissionChecker wraps permission checking with cache
type CachedPermissionChecker struct {
	tripRepo Repository
	cache    cache.Cache
}

func NewCachedPermissionChecker(tripRepo Repository, cache cache.Cache) *CachedPermissionChecker {
	return &CachedPermissionChecker{
		tripRepo: tripRepo,
		cache:    cache,
	}
}

func (c *CachedPermissionChecker) CanUserPerformOnTrip(ctx context.Context, userID, tripID primitive.ObjectID, permission users.Permission) (bool, error) {
	// Try cache first
	cacheData, err := c.cache.GetUserPermissions(ctx, userID.Hex(), tripID.Hex())
	if err == nil && cacheData != nil {
		var perms map[string]bool
		if err := json.Unmarshal(cacheData, &perms); err == nil {
			if allowed, exists := perms[string(permission)]; exists {
				return allowed, nil
			}
		}
	}

	// Get trip and check permissions
	trip, err := c.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return false, err
	}

	// Check permission
	allowed := trip.CanUserPerform(userID, permission)

	// Cache the result
	perms := map[string]bool{
		string(permission): allowed,
	}
	if data, err := json.Marshal(perms); err == nil {
		_ = c.cache.SetUserPermissions(ctx, userID.Hex(), tripID.Hex(), data, database.CacheTTLShort)
	}

	return allowed, nil
}