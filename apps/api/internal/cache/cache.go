package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/database"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	// Trip cache operations
	GetTrip(ctx context.Context, tripID string) ([]byte, error)
	SetTrip(ctx context.Context, tripID string, data []byte, ttl time.Duration) error
	DeleteTrip(ctx context.Context, tripID string) error
	InvalidateTripRelated(ctx context.Context, tripID string) error

	// Place cache operations
	GetPlace(ctx context.Context, placeID string) ([]byte, error)
	SetPlace(ctx context.Context, placeID string, data []byte, ttl time.Duration) error
	DeletePlace(ctx context.Context, placeID string) error
	GetTripPlaces(ctx context.Context, tripID string) ([]byte, error)
	SetTripPlaces(ctx context.Context, tripID string, data []byte, ttl time.Duration) error
	InvalidateTripPlaces(ctx context.Context, tripID string) error

	// User cache operations
	GetUser(ctx context.Context, userID string) ([]byte, error)
	SetUser(ctx context.Context, userID string, data []byte, ttl time.Duration) error
	DeleteUser(ctx context.Context, userID string) error

	// Permission cache operations
	GetUserPermissions(ctx context.Context, userID, tripID string) ([]byte, error)
	SetUserPermissions(ctx context.Context, userID, tripID string, data []byte, ttl time.Duration) error
	InvalidateUserPermissions(ctx context.Context, userID, tripID string) error
}

type redisCache struct {
	client *database.RedisClient
}

func NewRedisCache(client *database.RedisClient) Cache {
	return &redisCache{
		client: client,
	}
}

// Trip cache operations

func (c *redisCache) GetTrip(ctx context.Context, tripID string) ([]byte, error) {
	key := database.BuildTripCacheKey(tripID)
	val, err := c.client.Get(ctx, key)
	if err == redis.Nil {
		return nil, nil
	}
	return []byte(val), err
}

func (c *redisCache) SetTrip(ctx context.Context, tripID string, data []byte, ttl time.Duration) error {
	key := database.BuildTripCacheKey(tripID)
	return c.client.Set(ctx, key, data, ttl)
}

func (c *redisCache) DeleteTrip(ctx context.Context, tripID string) error {
	key := database.BuildTripCacheKey(tripID)
	return c.client.Delete(ctx, key)
}

func (c *redisCache) InvalidateTripRelated(ctx context.Context, tripID string) error {
	// Delete trip cache
	if err := c.DeleteTrip(ctx, tripID); err != nil {
		return err
	}

	// Delete trip places cache
	if err := c.InvalidateTripPlaces(ctx, tripID); err != nil {
		return err
	}

	// Note: In production, you might also want to invalidate user permission caches
	// for all collaborators, but that requires fetching the trip first

	return nil
}

// Place cache operations

func (c *redisCache) GetPlace(ctx context.Context, placeID string) ([]byte, error) {
	key := database.BuildPlaceCacheKey(placeID)
	val, err := c.client.Get(ctx, key)
	if err == redis.Nil {
		return nil, nil
	}
	return []byte(val), err
}

func (c *redisCache) SetPlace(ctx context.Context, placeID string, data []byte, ttl time.Duration) error {
	key := database.BuildPlaceCacheKey(placeID)
	return c.client.Set(ctx, key, data, ttl)
}

func (c *redisCache) DeletePlace(ctx context.Context, placeID string) error {
	key := database.BuildPlaceCacheKey(placeID)
	return c.client.Delete(ctx, key)
}

func (c *redisCache) GetTripPlaces(ctx context.Context, tripID string) ([]byte, error) {
	key := database.BuildTripPlacesCacheKey(tripID)
	val, err := c.client.Get(ctx, key)
	if err == redis.Nil {
		return nil, nil
	}
	return []byte(val), err
}

func (c *redisCache) SetTripPlaces(ctx context.Context, tripID string, data []byte, ttl time.Duration) error {
	key := database.BuildTripPlacesCacheKey(tripID)
	return c.client.Set(ctx, key, data, ttl)
}

func (c *redisCache) InvalidateTripPlaces(ctx context.Context, tripID string) error {
	key := database.BuildTripPlacesCacheKey(tripID)
	return c.client.Delete(ctx, key)
}

// User cache operations

func (c *redisCache) GetUser(ctx context.Context, userID string) ([]byte, error) {
	key := database.BuildUserCacheKey(userID)
	val, err := c.client.Get(ctx, key)
	if err == redis.Nil {
		return nil, nil
	}
	return []byte(val), err
}

func (c *redisCache) SetUser(ctx context.Context, userID string, data []byte, ttl time.Duration) error {
	key := database.BuildUserCacheKey(userID)
	return c.client.Set(ctx, key, data, ttl)
}

func (c *redisCache) DeleteUser(ctx context.Context, userID string) error {
	key := database.BuildUserCacheKey(userID)
	return c.client.Delete(ctx, key)
}

// Permission cache operations

func (c *redisCache) GetUserPermissions(ctx context.Context, userID, tripID string) ([]byte, error) {
	key := database.BuildUserPermissionsCacheKey(userID, tripID)
	val, err := c.client.Get(ctx, key)
	if err == redis.Nil {
		return nil, nil
	}
	return []byte(val), err
}

func (c *redisCache) SetUserPermissions(ctx context.Context, userID, tripID string, data []byte, ttl time.Duration) error {
	key := database.BuildUserPermissionsCacheKey(userID, tripID)
	return c.client.Set(ctx, key, data, ttl)
}

func (c *redisCache) InvalidateUserPermissions(ctx context.Context, userID, tripID string) error {
	key := database.BuildUserPermissionsCacheKey(userID, tripID)
	return c.client.Delete(ctx, key)
}

// Helper function to marshal data for caching
func MarshalForCache(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Helper function to unmarshal cached data
func UnmarshalFromCache(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}