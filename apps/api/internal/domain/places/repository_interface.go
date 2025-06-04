package places

import (
	"context"
	"errors"
	"github.com/Oferzz/newMap/apps/api/internal/nlp"
)

var (
	ErrPlaceNotFound = errors.New("place not found")
)

// Repository defines the interface for place data access
type Repository interface {
	Create(ctx context.Context, place *Place) error
	GetByID(ctx context.Context, id string) (*Place, error)
	GetByCreator(ctx context.Context, creatorID string) ([]*Place, error)
	Update(ctx context.Context, place *Place) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, filters SearchFilters) (*SearchResult, error)
	GetNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*Place, error)
	GetInBounds(ctx context.Context, bounds Bounds) ([]*Place, error)
	GetByCategory(ctx context.Context, category string, limit, offset int) ([]*Place, error)
	GetChildren(ctx context.Context, parentID string) ([]*Place, error)
	UpdateRating(ctx context.Context, placeID string, rating float64, count int) error
	
	// Enhanced spatial search methods
	SearchWithSpatialContext(ctx context.Context, query string, spatial *nlp.SpatialSearchContext, filters SearchFilters) (*SearchResult, error)
	GetInArea(ctx context.Context, area nlp.AreaFilter) ([]*Place, error)
	GetIntersecting(ctx context.Context, area nlp.AreaFilter) ([]*Place, error)
	GetWithinDistance(ctx context.Context, area nlp.AreaFilter) ([]*Place, error)
}

// SearchFilters contains filters for place search
type SearchFilters struct {
	Category  []string
	Tags      []string
	CreatorID string
	Limit     int
	Offset    int
}

// SearchResult contains search results with metadata
type SearchResult struct {
	Places []*Place
	Total  int64
}

// Bounds represents geographical bounds
type Bounds struct {
	MinLat float64
	MaxLat float64
	MinLng float64
	MaxLng float64
}