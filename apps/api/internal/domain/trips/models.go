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