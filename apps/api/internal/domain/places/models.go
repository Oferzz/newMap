package places

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlaceCategory string

const (
	CategoryAccommodation PlaceCategory = "accommodation"
	CategoryRestaurant    PlaceCategory = "restaurant"
	CategoryAttraction    PlaceCategory = "attraction"
	CategoryTransport     PlaceCategory = "transport"
	CategoryShopping      PlaceCategory = "shopping"
	CategoryNightlife     PlaceCategory = "nightlife"
	CategoryOutdoor       PlaceCategory = "outdoor"
	CategoryCultural      PlaceCategory = "cultural"
	CategoryOther         PlaceCategory = "other"
)

type Location struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}

type Place struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	TripID      primitive.ObjectID   `bson:"trip_id" json:"trip_id"`
	ParentID    *primitive.ObjectID  `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description" json:"description"`
	Category    PlaceCategory        `bson:"category" json:"category"`
	Location    Location             `bson:"location" json:"location"`
	Address     string               `bson:"address" json:"address"`
	GooglePlaceID string             `bson:"google_place_id,omitempty" json:"google_place_id,omitempty"`
	Images      []string             `bson:"images" json:"images"`
	Tags        []string             `bson:"tags" json:"tags"`
	Notes       string               `bson:"notes" json:"notes"`
	Rating      float32              `bson:"rating" json:"rating"`
	Cost        *Cost                `bson:"cost,omitempty" json:"cost,omitempty"`
	VisitDate   *time.Time           `bson:"visit_date,omitempty" json:"visit_date,omitempty"`
	Duration    int                  `bson:"duration" json:"duration"` // in minutes
	IsVisited   bool                 `bson:"is_visited" json:"is_visited"`
	CreatedBy   primitive.ObjectID   `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}

type Cost struct {
	Currency string  `bson:"currency" json:"currency"`
	Amount   float64 `bson:"amount" json:"amount"`
	PerPerson bool   `bson:"per_person" json:"per_person"`
}

type CreatePlaceInput struct {
	TripID      string         `json:"trip_id" binding:"required"`
	ParentID    *string        `json:"parent_id,omitempty"`
	Name        string         `json:"name" binding:"required,min=1,max=200"`
	Description string         `json:"description" binding:"max=1000"`
	Category    PlaceCategory  `json:"category" binding:"required"`
	Latitude    float64        `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude   float64        `json:"longitude" binding:"required,min=-180,max=180"`
	Address     string         `json:"address" binding:"max=500"`
	GooglePlaceID *string      `json:"google_place_id,omitempty"`
	Images      []string       `json:"images,omitempty" binding:"omitempty,max=10,dive,url"`
	Tags        []string       `json:"tags,omitempty" binding:"omitempty,max=10,dive,min=1,max=30"`
	Notes       string         `json:"notes,omitempty" binding:"omitempty,max=2000"`
	Rating      float32        `json:"rating,omitempty" binding:"omitempty,min=0,max=5"`
	Cost        *Cost          `json:"cost,omitempty"`
	VisitDate   *time.Time     `json:"visit_date,omitempty"`
	Duration    int            `json:"duration,omitempty" binding:"omitempty,min=0,max=1440"`
}

type UpdatePlaceInput struct {
	Name        *string        `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Description *string        `json:"description,omitempty" binding:"omitempty,max=1000"`
	Category    *PlaceCategory `json:"category,omitempty"`
	Latitude    *float64       `json:"latitude,omitempty" binding:"omitempty,min=-90,max=90"`
	Longitude   *float64       `json:"longitude,omitempty" binding:"omitempty,min=-180,max=180"`
	Address     *string        `json:"address,omitempty" binding:"omitempty,max=500"`
	Images      []string       `json:"images,omitempty" binding:"omitempty,max=10,dive,url"`
	Tags        []string       `json:"tags,omitempty" binding:"omitempty,max=10,dive,min=1,max=30"`
	Notes       *string        `json:"notes,omitempty" binding:"omitempty,max=2000"`
	Rating      *float32       `json:"rating,omitempty" binding:"omitempty,min=0,max=5"`
	Cost        *Cost          `json:"cost,omitempty"`
	VisitDate   *time.Time     `json:"visit_date,omitempty"`
	Duration    *int           `json:"duration,omitempty" binding:"omitempty,min=0,max=1440"`
	IsVisited   *bool          `json:"is_visited,omitempty"`
}

type PlaceFilter struct {
	TripID     *primitive.ObjectID
	ParentID   *primitive.ObjectID
	Category   *PlaceCategory
	IsVisited  *bool
	Tags       []string
	MinRating  *float32
	MaxCost    *float64
	DateFrom   *time.Time
	DateTo     *time.Time
	Bounds     *GeoBounds
	SearchQuery string
}

type GeoBounds struct {
	MinLat float64 `json:"min_lat"`
	MaxLat float64 `json:"max_lat"`
	MinLng float64 `json:"min_lng"`
	MaxLng float64 `json:"max_lng"`
}

type PlaceListOptions struct {
	Filter PlaceFilter
	Page   int
	Limit  int
	Sort   string
}

func (c PlaceCategory) IsValid() bool {
	switch c {
	case CategoryAccommodation, CategoryRestaurant, CategoryAttraction,
		CategoryTransport, CategoryShopping, CategoryNightlife,
		CategoryOutdoor, CategoryCultural, CategoryOther:
		return true
	default:
		return false
	}
}

func NewLocation(longitude, latitude float64) Location {
	return Location{
		Type:        "Point",
		Coordinates: []float64{longitude, latitude},
	}
}