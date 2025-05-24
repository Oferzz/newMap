package trips

import (
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/domain/users"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripStatus string

const (
	StatusPlanning  TripStatus = "planning"
	StatusUpcoming  TripStatus = "upcoming"
	StatusOngoing   TripStatus = "ongoing"
	StatusCompleted TripStatus = "completed"
	StatusCancelled TripStatus = "cancelled"
)

type Collaborator struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Role      users.Role         `bson:"role" json:"role"`
	JoinedAt  time.Time          `bson:"joined_at" json:"joined_at"`
	InvitedBy primitive.ObjectID `bson:"invited_by" json:"invited_by"`
}

type Trip struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description" json:"description"`
	CoverImage    string             `bson:"cover_image" json:"cover_image"`
	OwnerID       primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	Collaborators []Collaborator     `bson:"collaborators" json:"collaborators"`
	StartDate     time.Time          `bson:"start_date" json:"start_date"`
	EndDate       time.Time          `bson:"end_date" json:"end_date"`
	Status        TripStatus         `bson:"status" json:"status"`
	IsPublic      bool               `bson:"is_public" json:"is_public"`
	Tags          []string           `bson:"tags" json:"tags"`
	Budget        *Budget            `bson:"budget,omitempty" json:"budget,omitempty"`
	PlaceCount    int                `bson:"place_count" json:"place_count"`
	ViewCount     int                `bson:"view_count" json:"view_count"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type Budget struct {
	Currency string  `bson:"currency" json:"currency"`
	Total    float64 `bson:"total" json:"total"`
	Spent    float64 `bson:"spent" json:"spent"`
}

type CreateTripInput struct {
	Name        string     `json:"name" binding:"required,min=3,max=100"`
	Description string     `json:"description" binding:"max=1000"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     time.Time  `json:"end_date" binding:"required,gtfield=StartDate"`
	IsPublic    bool       `json:"is_public"`
	Tags        []string   `json:"tags" binding:"max=10,dive,min=1,max=30"`
	Budget      *Budget    `json:"budget,omitempty"`
}

type UpdateTripInput struct {
	Name        *string     `json:"name,omitempty" binding:"omitempty,min=3,max=100"`
	Description *string     `json:"description,omitempty" binding:"omitempty,max=1000"`
	CoverImage  *string     `json:"cover_image,omitempty" binding:"omitempty,url"`
	StartDate   *time.Time  `json:"start_date,omitempty"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	Status      *TripStatus `json:"status,omitempty"`
	IsPublic    *bool       `json:"is_public,omitempty"`
	Tags        []string    `json:"tags,omitempty" binding:"omitempty,max=10,dive,min=1,max=30"`
	Budget      *Budget     `json:"budget,omitempty"`
}

type InviteCollaboratorInput struct {
	Email string     `json:"email" binding:"required,email"`
	Role  users.Role `json:"role" binding:"required"`
}

type UpdateCollaboratorRoleInput struct {
	UserID string     `json:"user_id" binding:"required"`
	Role   users.Role `json:"role" binding:"required"`
}

type TripFilter struct {
	OwnerID       *primitive.ObjectID
	CollaboratorID *primitive.ObjectID
	Status        *TripStatus
	IsPublic      *bool
	Tags          []string
	StartDateFrom *time.Time
	StartDateTo   *time.Time
	SearchQuery   string
}

type TripListOptions struct {
	Filter TripFilter
	Page   int
	Limit  int
	Sort   string
}

func (s TripStatus) IsValid() bool {
	switch s {
	case StatusPlanning, StatusUpcoming, StatusOngoing, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

func (t *Trip) HasCollaborator(userID primitive.ObjectID) bool {
	if t.OwnerID == userID {
		return true
	}
	
	for _, collaborator := range t.Collaborators {
		if collaborator.UserID == userID {
			return true
		}
	}
	return false
}

func (t *Trip) GetUserRole(userID primitive.ObjectID) (users.Role, bool) {
	if t.OwnerID == userID {
		return users.RoleAdmin, true
	}
	
	for _, collaborator := range t.Collaborators {
		if collaborator.UserID == userID {
			return collaborator.Role, true
		}
	}
	return "", false
}

func (t *Trip) CanUserPerform(userID primitive.ObjectID, permission users.Permission) bool {
	role, hasRole := t.GetUserRole(userID)
	if !hasRole {
		return false
	}
	return role.HasPermission(permission)
}