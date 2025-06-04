package trips

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type Trip struct {
	ID              string         `db:"id" json:"id"`
	Title           string         `db:"title" json:"title"`
	Description     string         `db:"description" json:"description"`
	OwnerID         string         `db:"owner_id" json:"owner_id"`
	CoverImage      string         `db:"cover_image" json:"cover_image"`
	Privacy         string         `db:"privacy" json:"privacy"`
	Status          string         `db:"status" json:"status"`
	StartDate       *time.Time     `db:"start_date" json:"start_date"`
	EndDate         *time.Time     `db:"end_date" json:"end_date"`
	Timezone        string         `db:"timezone" json:"timezone"`
	Tags            pq.StringArray `db:"tags" json:"tags"`
	ViewCount       int            `db:"view_count" json:"view_count"`
	ShareCount      int            `db:"share_count" json:"share_count"`
	SuggestionCount int            `db:"suggestion_count" json:"suggestion_count"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time     `db:"deleted_at" json:"deleted_at,omitempty"`

	// Activity-specific fields
	ActivityType       string         `db:"activity_type" json:"activity_type"`
	DifficultyLevel    string         `db:"difficulty_level" json:"difficulty_level"`
	DurationHours      *float64       `db:"duration_hours" json:"duration_hours"`
	DistanceKm         *float64       `db:"distance_km" json:"distance_km"`
	ElevationGainM     *int           `db:"elevation_gain_m" json:"elevation_gain_m"`
	MaxElevationM      *int           `db:"max_elevation_m" json:"max_elevation_m"`
	RouteType          string         `db:"route_type" json:"route_type"`
	RouteGeoJSON       *GeoJSONRoute  `db:"route_geojson" json:"route_geojson"`
	WaterFeatures      pq.StringArray `db:"water_features" json:"water_features"`
	TerrainTypes       pq.StringArray `db:"terrain_types" json:"terrain_types"`
	EssentialGear      pq.StringArray `db:"essential_gear" json:"essential_gear"`
	BestSeasons        pq.StringArray `db:"best_seasons" json:"best_seasons"`
	TrailConditions    string         `db:"trail_conditions" json:"trail_conditions"`
	AccessibilityNotes string         `db:"accessibility_notes" json:"accessibility_notes"`
	ParkingInfo        *JSONB         `db:"parking_info" json:"parking_info"`
	PermitsRequired    pq.StringArray `db:"permits_required" json:"permits_required"`
	Hazards            pq.StringArray `db:"hazards" json:"hazards"`
	EmergencyContacts  *JSONB         `db:"emergency_contacts" json:"emergency_contacts"`
	Visibility         string         `db:"visibility" json:"visibility"`
	SharedWith         pq.StringArray `db:"shared_with" json:"shared_with"`
	CompletionCount    int            `db:"completion_count" json:"completion_count"`
	AverageRating      *float64       `db:"average_rating" json:"average_rating"`
	RatingCount        int            `db:"rating_count" json:"rating_count"`
	Featured           bool           `db:"featured" json:"featured"`
	Verified           bool           `db:"verified" json:"verified"`

	// Joined fields
	Collaborators []Collaborator `json:"collaborators,omitempty"`
	Waypoints     []Waypoint     `json:"waypoints,omitempty"`
}

type Collaborator struct {
	ID                     string     `db:"id" json:"id"`
	TripID                 string     `db:"trip_id" json:"trip_id"`
	UserID                 string     `db:"user_id" json:"user_id"`
	Role                   string     `db:"role" json:"role"`
	CanEdit                bool       `db:"can_edit" json:"can_edit"`
	CanDelete              bool       `db:"can_delete" json:"can_delete"`
	CanInvite              bool       `db:"can_invite" json:"can_invite"`
	CanModerateSuggestions bool       `db:"can_moderate_suggestions" json:"can_moderate_suggestions"`
	InvitedAt              time.Time  `db:"invited_at" json:"invited_at"`
	JoinedAt               *time.Time `db:"joined_at" json:"joined_at"`

	// Joined fields
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type Waypoint struct {
	ID            string     `db:"id" json:"id"`
	TripID        string     `db:"trip_id" json:"trip_id"`
	PlaceID       string     `db:"place_id" json:"place_id"`
	OrderPosition int        `db:"order_position" json:"order_position"`
	ArrivalTime   *time.Time `db:"arrival_time" json:"arrival_time"`
	DepartureTime *time.Time `db:"departure_time" json:"departure_time"`
	Notes         string     `db:"notes" json:"notes"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`

	// Joined place info
	Place *Place `json:"place,omitempty"`
}

type Place struct {
	ID          string   `db:"id" json:"id"`
	Name        string   `db:"name" json:"name"`
	Description string   `db:"description" json:"description"`
	Type        string   `db:"type" json:"type"`
	Location    *GeoJSON `db:"location" json:"location"`
	Address     string   `db:"street_address" json:"address"`
	City        string   `db:"city" json:"city"`
	Country     string   `db:"country" json:"country"`
}

// GeoJSON represents a PostGIS geography point
type GeoJSON struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// GeoJSONRoute represents a PostGIS LineString or Polygon for routes/areas
type GeoJSONRoute struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

// JSONB represents a PostgreSQL JSONB column
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (g GeoJSON) Value() (driver.Value, error) {
	if len(g.Coordinates) == 0 {
		return nil, nil
	}
	return json.Marshal(g)
}

// Scan implements the sql.Scanner interface
func (g *GeoJSON) Scan(value interface{}) error {
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

// Value implements the driver.Valuer interface for GeoJSONRoute
func (g GeoJSONRoute) Value() (driver.Value, error) {
	if g.Coordinates == nil {
		return nil, nil
	}
	return json.Marshal(g)
}

// Scan implements the sql.Scanner interface for GeoJSONRoute
func (g *GeoJSONRoute) Scan(value interface{}) error {
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

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
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
	
	return json.Unmarshal(data, j)
}

// Input types
type CreateTripInput struct {
	Title       string     `json:"title" binding:"required,min=3,max=255"`
	Description string     `json:"description" binding:"max=1000"`
	Privacy     string     `json:"privacy" binding:"omitempty,oneof=public friends private invite_only"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Timezone    string     `json:"timezone"`
	Tags        []string   `json:"tags"`
	CoverImage  string     `json:"cover_image"`
	
	// Activity-specific fields
	ActivityType       string         `json:"activity_type" binding:"omitempty,oneof=hiking biking climbing skiing snowboarding kayaking canoeing rafting swimming surfing running walking backpacking camping fishing birdwatching photography sightseeing general"`
	DifficultyLevel    string         `json:"difficulty_level" binding:"omitempty,oneof=easy moderate hard expert"`
	DurationHours      *float64       `json:"duration_hours" binding:"omitempty,min=0,max=240"`
	DistanceKm         *float64       `json:"distance_km" binding:"omitempty,min=0,max=1000"`
	ElevationGainM     *int           `json:"elevation_gain_m" binding:"omitempty,min=-500,max=10000"`
	MaxElevationM      *int           `json:"max_elevation_m" binding:"omitempty,min=-500,max=10000"`
	RouteType          string         `json:"route_type" binding:"omitempty,oneof=out_and_back loop point_to_point area"`
	RouteGeoJSON       *GeoJSONRoute  `json:"route_geojson"`
	WaterFeatures      []string       `json:"water_features"`
	TerrainTypes       []string       `json:"terrain_types"`
	EssentialGear      []string       `json:"essential_gear"`
	BestSeasons        []string       `json:"best_seasons"`
	TrailConditions    string         `json:"trail_conditions" binding:"max=500"`
	AccessibilityNotes string         `json:"accessibility_notes" binding:"max=500"`
	ParkingInfo        *JSONB         `json:"parking_info"`
	PermitsRequired    []string       `json:"permits_required"`
	Hazards            []string       `json:"hazards"`
	EmergencyContacts  *JSONB         `json:"emergency_contacts"`
	Visibility         string         `json:"visibility" binding:"omitempty,oneof=public private"`
	SharedWith         []string       `json:"shared_with"`
}

type UpdateTripInput struct {
	Title       *string    `json:"title,omitempty" binding:"omitempty,min=3,max=255"`
	Description *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
	Privacy     *string    `json:"privacy,omitempty" binding:"omitempty,oneof=public friends private invite_only"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Timezone    *string    `json:"timezone,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	CoverImage  *string    `json:"cover_image,omitempty"`
	Status      *string    `json:"status,omitempty" binding:"omitempty,oneof=planning active completed cancelled"`
	
	// Activity-specific fields
	ActivityType       *string        `json:"activity_type,omitempty" binding:"omitempty,oneof=hiking biking climbing skiing snowboarding kayaking canoeing rafting swimming surfing running walking backpacking camping fishing birdwatching photography sightseeing general"`
	DifficultyLevel    *string        `json:"difficulty_level,omitempty" binding:"omitempty,oneof=easy moderate hard expert"`
	DurationHours      *float64       `json:"duration_hours,omitempty" binding:"omitempty,min=0,max=240"`
	DistanceKm         *float64       `json:"distance_km,omitempty" binding:"omitempty,min=0,max=1000"`
	ElevationGainM     *int           `json:"elevation_gain_m,omitempty" binding:"omitempty,min=-500,max=10000"`
	MaxElevationM      *int           `json:"max_elevation_m,omitempty" binding:"omitempty,min=-500,max=10000"`
	RouteType          *string        `json:"route_type,omitempty" binding:"omitempty,oneof=out_and_back loop point_to_point area"`
	RouteGeoJSON       *GeoJSONRoute  `json:"route_geojson,omitempty"`
	WaterFeatures      []string       `json:"water_features,omitempty"`
	TerrainTypes       []string       `json:"terrain_types,omitempty"`
	EssentialGear      []string       `json:"essential_gear,omitempty"`
	BestSeasons        []string       `json:"best_seasons,omitempty"`
	TrailConditions    *string        `json:"trail_conditions,omitempty" binding:"omitempty,max=500"`
	AccessibilityNotes *string        `json:"accessibility_notes,omitempty" binding:"omitempty,max=500"`
	ParkingInfo        *JSONB         `json:"parking_info,omitempty"`
	PermitsRequired    []string       `json:"permits_required,omitempty"`
	Hazards            []string       `json:"hazards,omitempty"`
	EmergencyContacts  *JSONB         `json:"emergency_contacts,omitempty"`
	Visibility         *string        `json:"visibility,omitempty" binding:"omitempty,oneof=public private"`
	SharedWith         []string       `json:"shared_with,omitempty"`
}

type AddCollaboratorInput struct {
	UserID                 string `json:"user_id" binding:"required,uuid"`
	Role                   string `json:"role" binding:"required,oneof=admin editor viewer"`
	CanEdit                bool   `json:"can_edit"`
	CanDelete              bool   `json:"can_delete"`
	CanInvite              bool   `json:"can_invite"`
	CanModerateSuggestions bool   `json:"can_moderate_suggestions"`
}

type UpdateCollaboratorInput struct {
	Role                   *string `json:"role,omitempty" binding:"omitempty,oneof=admin editor viewer"`
	CanEdit                *bool   `json:"can_edit,omitempty"`
	CanDelete              *bool   `json:"can_delete,omitempty"`
	CanInvite              *bool   `json:"can_invite,omitempty"`
	CanModerateSuggestions *bool   `json:"can_moderate_suggestions,omitempty"`
}

type AddWaypointInput struct {
	PlaceID       string     `json:"place_id" binding:"required,uuid"`
	OrderPosition int        `json:"order_position" binding:"min=0"`
	ArrivalTime   *time.Time `json:"arrival_time"`
	DepartureTime *time.Time `json:"departure_time"`
	Notes         string     `json:"notes" binding:"max=500"`
}

type UpdateWaypointInput struct {
	OrderPosition *int       `json:"order_position,omitempty" binding:"omitempty,min=0"`
	ArrivalTime   *time.Time `json:"arrival_time,omitempty"`
	DepartureTime *time.Time `json:"departure_time,omitempty"`
	Notes         *string    `json:"notes,omitempty" binding:"omitempty,max=500"`
}

type TripFilters struct {
	OwnerID       string    `form:"owner_id"`
	CollaboratorID string    `form:"collaborator_id"`
	Privacy       string    `form:"privacy"`
	Status        string    `form:"status"`
	Tags          []string  `form:"tags"`
	StartDateFrom *time.Time `form:"start_date_from"`
	StartDateTo   *time.Time `form:"start_date_to"`
	Search        string    `form:"search"`
	Limit         int       `form:"limit"`
	Offset        int       `form:"offset"`
	SortBy        string    `form:"sort_by"`
	SortOrder     string    `form:"sort_order"`
	
	// Activity-specific filters
	ActivityTypes   []string `form:"activity_types"`
	DifficultyLevels []string `form:"difficulty_levels"`
	MinDuration     *float64 `form:"min_duration"`
	MaxDuration     *float64 `form:"max_duration"`
	MinDistance     *float64 `form:"min_distance"`
	MaxDistance     *float64 `form:"max_distance"`
	WaterFeatures   []string `form:"water_features"`
	TerrainTypes    []string `form:"terrain_types"`
	Visibility      string   `form:"visibility"`
	Featured        *bool    `form:"featured"`
	Verified        *bool    `form:"verified"`
	
	// Geospatial filters
	NearLat         *float64 `form:"near_lat"`
	NearLng         *float64 `form:"near_lng"`
	RadiusKm        *float64 `form:"radius_km"`
	BoundsNorthEast []float64 `form:"bounds_ne"`
	BoundsSouthWest []float64 `form:"bounds_sw"`
}

// Helper methods
func (t *Trip) IsOwner(userID string) bool {
	return t.OwnerID == userID
}

func (t *Trip) HasCollaborator(userID string) bool {
	for _, c := range t.Collaborators {
		if c.UserID == userID {
			return true
		}
	}
	return false
}

func (t *Trip) GetCollaborator(userID string) *Collaborator {
	for _, c := range t.Collaborators {
		if c.UserID == userID {
			return &c
		}
	}
	return nil
}

func (t *Trip) CanUserEdit(userID string) bool {
	if t.IsOwner(userID) {
		return true
	}
	
	collaborator := t.GetCollaborator(userID)
	if collaborator == nil {
		return false
	}
	
	return collaborator.CanEdit || collaborator.Role == "admin"
}

func (t *Trip) CanUserDelete(userID string) bool {
	if t.IsOwner(userID) {
		return true
	}
	
	collaborator := t.GetCollaborator(userID)
	if collaborator == nil {
		return false
	}
	
	return collaborator.CanDelete || collaborator.Role == "admin"
}

func (t *Trip) CanUserInvite(userID string) bool {
	if t.IsOwner(userID) {
		return true
	}
	
	collaborator := t.GetCollaborator(userID)
	if collaborator == nil {
		return false
	}
	
	return collaborator.CanInvite || collaborator.Role == "admin"
}

func (t *Trip) CanUserModerateSuggestions(userID string) bool {
	if t.IsOwner(userID) {
		return true
	}
	
	collaborator := t.GetCollaborator(userID)
	if collaborator == nil {
		return false
	}
	
	return collaborator.CanModerateSuggestions || collaborator.Role == "admin"
}

func (t *Trip) CanUserPerform(userID string, permission string) bool {
	// Convert permission string to check specific capabilities
	switch permission {
	case "trip.read":
		// For read, check if user is owner or collaborator
		return t.IsOwner(userID) || t.HasCollaborator(userID)
	case "trip.update":
		return t.CanUserEdit(userID)
	case "trip.delete":
		return t.CanUserDelete(userID)
	case "trip.invite":
		return t.CanUserInvite(userID)
	case "suggestion.moderate":
		return t.CanUserModerateSuggestions(userID)
	default:
		// For any other permission, check if user is owner
		return t.IsOwner(userID)
	}
}

// Activity-specific models
type ActivityCompletion struct {
	ID                 string         `db:"id" json:"id"`
	TripID             string         `db:"trip_id" json:"trip_id"`
	UserID             string         `db:"user_id" json:"user_id"`
	CompletedAt        time.Time      `db:"completed_at" json:"completed_at"`
	DurationMinutes    *int           `db:"duration_minutes" json:"duration_minutes"`
	DifficultyRating   *int           `db:"difficulty_rating" json:"difficulty_rating"`
	OverallRating      *int           `db:"overall_rating" json:"overall_rating"`
	WeatherConditions  string         `db:"weather_conditions" json:"weather_conditions"`
	TrailConditions    string         `db:"trail_conditions" json:"trail_conditions"`
	Notes              string         `db:"notes" json:"notes"`
	Photos             pq.StringArray `db:"photos" json:"photos"`
	GPXTrack           *JSONB         `db:"gpx_track" json:"gpx_track"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at"`
}

type ActivityRating struct {
	ID                   string    `db:"id" json:"id"`
	TripID               string    `db:"trip_id" json:"trip_id"`
	UserID               string    `db:"user_id" json:"user_id"`
	OverallRating        int       `db:"overall_rating" json:"overall_rating"`
	DifficultyAccuracy   *int      `db:"difficulty_accuracy" json:"difficulty_accuracy"`
	DescriptionAccuracy  *int      `db:"description_accuracy" json:"description_accuracy"`
	SceneryRating        *int      `db:"scenery_rating" json:"scenery_rating"`
	ReviewText           string    `db:"review_text" json:"review_text"`
	HelpfulCount         int       `db:"helpful_count" json:"helpful_count"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

type ActivityCondition struct {
	ID             string         `db:"id" json:"id"`
	TripID         string         `db:"trip_id" json:"trip_id"`
	ReportedBy     string         `db:"reported_by" json:"reported_by"`
	ConditionType  string         `db:"condition_type" json:"condition_type"`
	Severity       string         `db:"severity" json:"severity"`
	Description    string         `db:"description" json:"description"`
	Location       *GeoJSON       `db:"location" json:"location"`
	Photos         pq.StringArray `db:"photos" json:"photos"`
	ValidFrom      time.Time      `db:"valid_from" json:"valid_from"`
	ValidUntil     *time.Time     `db:"valid_until" json:"valid_until"`
	Verified       bool           `db:"verified" json:"verified"`
	VerifiedBy     *string        `db:"verified_by" json:"verified_by"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
}

type ActivityShareLink struct {
	ID          string     `db:"id" json:"id"`
	TripID      string     `db:"trip_id" json:"trip_id"`
	CreatedBy   string     `db:"created_by" json:"created_by"`
	ShareToken  string     `db:"share_token" json:"share_token"`
	Permissions string     `db:"permissions" json:"permissions"`
	MaxUses     *int       `db:"max_uses" json:"max_uses"`
	UseCount    int        `db:"use_count" json:"use_count"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	LastUsedAt  *time.Time `db:"last_used_at" json:"last_used_at"`
}

// Input types for activity features
type CreateActivityCompletionInput struct {
	TripID             string   `json:"trip_id" binding:"required,uuid"`
	CompletedAt        time.Time `json:"completed_at" binding:"required"`
	DurationMinutes    *int     `json:"duration_minutes" binding:"omitempty,min=1,max=10000"`
	DifficultyRating   *int     `json:"difficulty_rating" binding:"omitempty,min=1,max=5"`
	OverallRating      *int     `json:"overall_rating" binding:"omitempty,min=1,max=5"`
	WeatherConditions  string   `json:"weather_conditions" binding:"max=500"`
	TrailConditions    string   `json:"trail_conditions" binding:"max=500"`
	Notes              string   `json:"notes" binding:"max=1000"`
	Photos             []string `json:"photos"`
	GPXTrack           *JSONB   `json:"gpx_track"`
}

type CreateActivityRatingInput struct {
	TripID               string `json:"trip_id" binding:"required,uuid"`
	OverallRating        int    `json:"overall_rating" binding:"required,min=1,max=5"`
	DifficultyAccuracy   *int   `json:"difficulty_accuracy" binding:"omitempty,min=1,max=5"`
	DescriptionAccuracy  *int   `json:"description_accuracy" binding:"omitempty,min=1,max=5"`
	SceneryRating        *int   `json:"scenery_rating" binding:"omitempty,min=1,max=5"`
	ReviewText           string `json:"review_text" binding:"max=2000"`
}

type CreateActivityConditionInput struct {
	TripID         string     `json:"trip_id" binding:"required,uuid"`
	ConditionType  string     `json:"condition_type" binding:"required,oneof=trail weather closure hazard"`
	Severity       string     `json:"severity" binding:"omitempty,oneof=info warning danger"`
	Description    string     `json:"description" binding:"required,min=10,max=1000"`
	Location       *GeoJSON   `json:"location"`
	Photos         []string   `json:"photos"`
	ValidUntil     *time.Time `json:"valid_until"`
}

type CreateShareLinkInput struct {
	TripID      string     `json:"trip_id" binding:"required,uuid"`
	Permissions string     `json:"permissions" binding:"omitempty,oneof=view edit"`
	MaxUses     *int       `json:"max_uses" binding:"omitempty,min=1,max=1000"`
	ExpiresAt   *time.Time `json:"expires_at"`
}