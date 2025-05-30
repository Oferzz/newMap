package collections

import (
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description *string           `json:"description,omitempty" db:"description"`
	UserID      uuid.UUID         `json:"user_id" db:"user_id"`
	Privacy     string            `json:"privacy" db:"privacy"`
	Locations   []CollectionLocation `json:"locations,omitempty"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

type CollectionLocation struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CollectionID uuid.UUID `json:"collection_id" db:"collection_id"`
	Name         *string   `json:"name,omitempty" db:"name"`
	Latitude     float64   `json:"latitude" db:"latitude"`
	Longitude    float64   `json:"longitude" db:"longitude"`
	AddedAt      time.Time `json:"added_at" db:"added_at"`
}

type CreateCollectionRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Privacy     string  `json:"privacy" validate:"required,oneof=public private friends"`
}

type UpdateCollectionRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Privacy     *string `json:"privacy,omitempty" validate:"omitempty,oneof=public private friends"`
}

type AddLocationRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

type GetCollectionsParams struct {
	Page   int    `query:"page" validate:"omitempty,min=1"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
	UserID string `query:"user_id" validate:"omitempty,uuid"`
}