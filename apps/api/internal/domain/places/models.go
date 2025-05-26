package places

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Place struct {
	ID            string         `db:"id" json:"id"`
	Name          string         `db:"name" json:"name"`
	Description   string         `db:"description" json:"description"`
	Type          string         `db:"type" json:"type"` // 'poi', 'area', 'region'
	ParentID      *string        `db:"parent_id" json:"parent_id,omitempty"`
	Location      *GeoPoint      `db:"location" json:"location,omitempty"`
	Bounds        *GeoPolygon    `db:"bounds" json:"bounds,omitempty"`
	StreetAddress string         `db:"street_address" json:"street_address"`
	City          string         `db:"city" json:"city"`
	State         string         `db:"state" json:"state"`
	Country       string         `db:"country" json:"country"`
	PostalCode    string         `db:"postal_code" json:"postal_code"`
	CreatedBy     string         `db:"created_by" json:"created_by"`
	Category      pq.StringArray `db:"category" json:"category"`
	Tags          pq.StringArray `db:"tags" json:"tags"`
	OpeningHours  *OpeningHours  `db:"opening_hours" json:"opening_hours,omitempty"`
	ContactInfo   *ContactInfo   `db:"contact_info" json:"contact_info,omitempty"`
	Amenities     pq.StringArray `db:"amenities" json:"amenities"`
	AverageRating *float32       `db:"average_rating" json:"average_rating,omitempty"`
	RatingCount   int            `db:"rating_count" json:"rating_count"`
	Privacy       string         `db:"privacy" json:"privacy"`
	Status        string         `db:"status" json:"status"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`

	// Joined fields
	Media         []Media        `json:"media,omitempty"`
	Collaborators []Collaborator `json:"collaborators,omitempty"`
}

// GeoPoint represents a PostGIS geography point
type GeoPoint struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude]
}

// GeoPolygon represents a PostGIS geography polygon
type GeoPolygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// OpeningHours stores business hours in JSONB
type OpeningHours struct {
	Monday    []TimeRange `json:"monday,omitempty"`
	Tuesday   []TimeRange `json:"tuesday,omitempty"`
	Wednesday []TimeRange `json:"wednesday,omitempty"`
	Thursday  []TimeRange `json:"thursday,omitempty"`
	Friday    []TimeRange `json:"friday,omitempty"`
	Saturday  []TimeRange `json:"saturday,omitempty"`
	Sunday    []TimeRange `json:"sunday,omitempty"`
}

type TimeRange struct {
	Open  string `json:"open"`  // "09:00"
	Close string `json:"close"` // "17:00"
}

// ContactInfo stores contact information in JSONB
type ContactInfo struct {
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
	Social  Social `json:"social,omitempty"`
}

type Social struct {
	Facebook  string `json:"facebook,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
}

type Media struct {
	ID              string     `db:"id" json:"id"`
	MediaID         string     `db:"media_id" json:"media_id"`
	PlaceID         string     `db:"place_id" json:"place_id"`
	Caption         string     `db:"caption" json:"caption"`
	OrderPosition   int        `db:"order_position" json:"order_position"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	
	// Joined media details
	URL           string     `json:"url,omitempty"`
	ThumbnailURL  string     `json:"thumbnail_url,omitempty"`
	MimeType      string     `json:"mime_type,omitempty"`
	UploadedBy    string     `json:"uploaded_by,omitempty"`
}

type Collaborator struct {
	ID          string          `db:"id" json:"id"`
	PlaceID     string          `db:"place_id" json:"place_id"`
	UserID      string          `db:"user_id" json:"user_id"`
	Role        string          `db:"role" json:"role"`
	Permissions json.RawMessage `db:"permissions" json:"permissions"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`

	// Joined user info
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// Value implementations for custom types
func (g GeoPoint) Value() (driver.Value, error) {
	if len(g.Coordinates) == 0 {
		return nil, nil
	}
	// Convert to PostGIS format: POINT(longitude latitude)
	return json.Marshal(g)
}

func (g *GeoPoint) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(data, g)
}

func (g GeoPolygon) Value() (driver.Value, error) {
	if len(g.Coordinates) == 0 {
		return nil, nil
	}
	return json.Marshal(g)
}

func (g *GeoPolygon) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(data, g)
}

func (o OpeningHours) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *OpeningHours) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(data, o)
}

func (c ContactInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *ContactInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	
	return json.Unmarshal(data, c)
}

// Input types
type CreatePlaceInput struct {
	Name          string        `json:"name" binding:"required,min=1,max=255"`
	Description   string        `json:"description" binding:"max=1000"`
	Type          string        `json:"type" binding:"required,oneof=poi area region"`
	ParentID      *string       `json:"parent_id,omitempty" binding:"omitempty,uuid"`
	Location      *LocationInput `json:"location,omitempty"`
	Bounds        *BoundsInput   `json:"bounds,omitempty"`
	StreetAddress string        `json:"street_address" binding:"max=255"`
	City          string        `json:"city" binding:"max=100"`
	State         string        `json:"state" binding:"max=100"`
	Country       string        `json:"country" binding:"max=100"`
	PostalCode    string        `json:"postal_code" binding:"max=20"`
	Category      []string      `json:"category"`
	Tags          []string      `json:"tags"`
	OpeningHours  *OpeningHours `json:"opening_hours,omitempty"`
	ContactInfo   *ContactInfo  `json:"contact_info,omitempty"`
	Amenities     []string      `json:"amenities"`
	Privacy       string        `json:"privacy" binding:"omitempty,oneof=public friends private"`
}

type LocationInput struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

type BoundsInput struct {
	Coordinates [][][]float64 `json:"coordinates" binding:"required"`
}

type UpdatePlaceInput struct {
	Name          *string        `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description   *string        `json:"description,omitempty" binding:"omitempty,max=1000"`
	Type          *string        `json:"type,omitempty" binding:"omitempty,oneof=poi area region"`
	Location      *LocationInput `json:"location,omitempty"`
	Bounds        *BoundsInput   `json:"bounds,omitempty"`
	StreetAddress *string        `json:"street_address,omitempty" binding:"omitempty,max=255"`
	City          *string        `json:"city,omitempty" binding:"omitempty,max=100"`
	State         *string        `json:"state,omitempty" binding:"omitempty,max=100"`
	Country       *string        `json:"country,omitempty" binding:"omitempty,max=100"`
	PostalCode    *string        `json:"postal_code,omitempty" binding:"omitempty,max=20"`
	Category      []string       `json:"category,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	OpeningHours  *OpeningHours  `json:"opening_hours,omitempty"`
	ContactInfo   *ContactInfo   `json:"contact_info,omitempty"`
	Amenities     []string       `json:"amenities,omitempty"`
	Privacy       *string        `json:"privacy,omitempty" binding:"omitempty,oneof=public friends private"`
	Status        *string        `json:"status,omitempty" binding:"omitempty,oneof=active pending archived"`
}

type SearchPlacesInput struct {
	Query     string   `form:"q" binding:"max=100"`
	Type      string   `form:"type" binding:"omitempty,oneof=poi area region"`
	Category  []string `form:"category"`
	Tags      []string `form:"tags"`
	City      string   `form:"city"`
	Country   string   `form:"country"`
	Latitude  *float64 `form:"lat" binding:"omitempty,min=-90,max=90"`
	Longitude *float64 `form:"lng" binding:"omitempty,min=-180,max=180"`
	Radius    *int     `form:"radius" binding:"omitempty,min=1,max=50000"` // meters
	Limit     int      `form:"limit" binding:"min=1,max=100"`
	Offset    int      `form:"offset" binding:"min=0"`
}

type NearbyPlacesInput struct {
	Latitude  float64  `form:"lat" binding:"required,min=-90,max=90"`
	Longitude float64  `form:"lng" binding:"required,min=-180,max=180"`
	Radius    int      `form:"radius" binding:"required,min=1,max=50000"` // meters
	Type      string   `form:"type" binding:"omitempty,oneof=poi area region"`
	Category  []string `form:"category"`
	Tags      []string `form:"tags"`
	Limit     int      `form:"limit" binding:"min=1,max=100"`
	Offset    int      `form:"offset" binding:"min=0"`
}

// Helper methods
func (p *Place) IsOwner(userID string) bool {
	return p.CreatedBy == userID
}

func (p *Place) HasCollaborator(userID string) bool {
	for _, c := range p.Collaborators {
		if c.UserID == userID {
			return true
		}
	}
	return false
}

func (p *Place) GetCollaborator(userID string) *Collaborator {
	for _, c := range p.Collaborators {
		if c.UserID == userID {
			return &c
		}
	}
	return nil
}

func (p *Place) CanUserEdit(userID string) bool {
	if p.IsOwner(userID) {
		return true
	}
	
	collaborator := p.GetCollaborator(userID)
	if collaborator == nil {
		return false
	}
	
	return collaborator.Role == "admin" || collaborator.Role == "editor"
}

func (p *Place) CanUserDelete(userID string) bool {
	if p.IsOwner(userID) {
		return true
	}
	
	collaborator := p.GetCollaborator(userID)
	if collaborator == nil {
		return false
	}
	
	return collaborator.Role == "admin"
}

// PlaceFilter contains filter criteria for places
type PlaceFilter struct {
	TripID      *primitive.ObjectID
	ParentID    *primitive.ObjectID
	Category    *PlaceCategory
	IsVisited   *bool
	Tags        []string
	MinRating   *float32
	MaxCost     *float64
	SearchQuery string
}

// PlaceCategory represents place categories
type PlaceCategory string

const (
	PlaceCategoryRestaurant PlaceCategory = "restaurant"
	PlaceCategoryHotel      PlaceCategory = "hotel"
	PlaceCategoryAttraction PlaceCategory = "attraction"
	PlaceCategoryShopping   PlaceCategory = "shopping"
	PlaceCategoryTransport  PlaceCategory = "transport"
	PlaceCategoryOther      PlaceCategory = "other"
)

// IsValid checks if the category is valid
func (c PlaceCategory) IsValid() bool {
	switch c {
	case PlaceCategoryRestaurant, PlaceCategoryHotel, PlaceCategoryAttraction,
		PlaceCategoryShopping, PlaceCategoryTransport, PlaceCategoryOther:
		return true
	}
	return false
}