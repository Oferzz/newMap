package cache

import (
	"context"
	"time"
)

// noOpCache is a cache implementation that does nothing
// Used when Redis is not available
type noOpCache struct{}

func NewNoOpCache() Cache {
	return &noOpCache{}
}

func (n *noOpCache) GetTrip(ctx context.Context, tripID string) ([]byte, error) {
	return nil, nil
}

func (n *noOpCache) SetTrip(ctx context.Context, tripID string, data []byte, ttl time.Duration) error {
	return nil
}

func (n *noOpCache) DeleteTrip(ctx context.Context, tripID string) error {
	return nil
}

func (n *noOpCache) InvalidateTripRelated(ctx context.Context, tripID string) error {
	return nil
}

func (n *noOpCache) GetPlace(ctx context.Context, placeID string) ([]byte, error) {
	return nil, nil
}

func (n *noOpCache) SetPlace(ctx context.Context, placeID string, data []byte, ttl time.Duration) error {
	return nil
}

func (n *noOpCache) DeletePlace(ctx context.Context, placeID string) error {
	return nil
}

func (n *noOpCache) GetTripPlaces(ctx context.Context, tripID string) ([]byte, error) {
	return nil, nil
}

func (n *noOpCache) SetTripPlaces(ctx context.Context, tripID string, data []byte, ttl time.Duration) error {
	return nil
}

func (n *noOpCache) InvalidateTripPlaces(ctx context.Context, tripID string) error {
	return nil
}

func (n *noOpCache) GetUser(ctx context.Context, userID string) ([]byte, error) {
	return nil, nil
}

func (n *noOpCache) SetUser(ctx context.Context, userID string, data []byte, ttl time.Duration) error {
	return nil
}

func (n *noOpCache) DeleteUser(ctx context.Context, userID string) error {
	return nil
}

func (n *noOpCache) GetUserPermissions(ctx context.Context, userID, tripID string) ([]byte, error) {
	return nil, nil
}

func (n *noOpCache) SetUserPermissions(ctx context.Context, userID, tripID string, data []byte, ttl time.Duration) error {
	return nil
}

func (n *noOpCache) InvalidateUserPermissions(ctx context.Context, userID, tripID string) error {
	return nil
}