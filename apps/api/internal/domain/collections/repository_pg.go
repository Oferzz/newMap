package collections

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, collection *Collection) error {
	query := `
		INSERT INTO collections (id, name, description, user_id, privacy, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	collection.ID = uuid.New()
	collection.CreatedAt = time.Now()
	collection.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		collection.ID,
		collection.Name,
		collection.Description,
		collection.UserID,
		collection.Privacy,
		collection.CreatedAt,
		collection.UpdatedAt,
	)

	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Collection, error) {
	collection := &Collection{}
	query := `
		SELECT id, name, description, user_id, privacy, created_at, updated_at
		FROM collections
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, collection, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Load locations
	locations, err := r.GetLocations(ctx, id)
	if err != nil {
		return nil, err
	}
	collection.Locations = locations

	return collection, nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID, params GetCollectionsParams) ([]Collection, int, error) {
	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM collections
		WHERE user_id = $1
	`
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, err
	}

	// Get collections
	query := `
		SELECT id, name, description, user_id, privacy, created_at, updated_at
		FROM collections
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	var collections []Collection
	err = r.db.SelectContext(ctx, &collections, query, userID, params.Limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Load locations for each collection
	for i := range collections {
		locations, err := r.GetLocations(ctx, collections[i].ID)
		if err != nil {
			return nil, 0, err
		}
		collections[i].Locations = locations
	}

	return collections, total, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, updates UpdateCollectionRequest) (*Collection, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *updates.Name)
		argIndex++
	}

	if updates.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *updates.Description)
		argIndex++
	}

	if updates.Privacy != nil {
		setParts = append(setParts, fmt.Sprintf("privacy = $%d", argIndex))
		args = append(args, *updates.Privacy)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE collections
		SET %s
		WHERE id = $%d
	`, fmt.Sprintf("%s", setParts), argIndex)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete locations first (cascade)
	_, err := r.db.ExecContext(ctx, "DELETE FROM collection_locations WHERE collection_id = $1", id)
	if err != nil {
		return err
	}

	// Delete collection
	_, err = r.db.ExecContext(ctx, "DELETE FROM collections WHERE id = $1", id)
	return err
}

func (r *PostgresRepository) AddLocation(ctx context.Context, collectionID uuid.UUID, location *CollectionLocation) error {
	location.ID = uuid.New()
	location.CollectionID = collectionID
	location.AddedAt = time.Now()

	query := `
		INSERT INTO collection_locations (id, collection_id, name, latitude, longitude, added_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		location.ID,
		location.CollectionID,
		location.Name,
		location.Latitude,
		location.Longitude,
		location.AddedAt,
	)

	if err == nil {
		// Update collection's updated_at
		_, _ = r.db.ExecContext(ctx, "UPDATE collections SET updated_at = $1 WHERE id = $2", time.Now(), collectionID)
	}

	return err
}

func (r *PostgresRepository) RemoveLocation(ctx context.Context, collectionID uuid.UUID, locationID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, 
		"DELETE FROM collection_locations WHERE id = $1 AND collection_id = $2", 
		locationID, collectionID)

	if err == nil {
		// Update collection's updated_at
		_, _ = r.db.ExecContext(ctx, "UPDATE collections SET updated_at = $1 WHERE id = $2", time.Now(), collectionID)
	}

	return err
}

func (r *PostgresRepository) GetLocations(ctx context.Context, collectionID uuid.UUID) ([]CollectionLocation, error) {
	var locations []CollectionLocation
	query := `
		SELECT id, collection_id, name, latitude, longitude, added_at
		FROM collection_locations
		WHERE collection_id = $1
		ORDER BY added_at DESC
	`

	err := r.db.SelectContext(ctx, &locations, query, collectionID)
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *PostgresRepository) AddCollaborator(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO collection_collaborators (collection_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (collection_id, user_id) DO UPDATE SET role = $3
	`

	_, err := r.db.ExecContext(ctx, query, collectionID, userID, role, time.Now())
	return err
}

func (r *PostgresRepository) RemoveCollaborator(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, 
		"DELETE FROM collection_collaborators WHERE collection_id = $1 AND user_id = $2", 
		collectionID, userID)
	return err
}

func (r *PostgresRepository) GetCollaborators(ctx context.Context, collectionID uuid.UUID) ([]uuid.UUID, error) {
	var collaborators []uuid.UUID
	query := `
		SELECT user_id
		FROM collection_collaborators
		WHERE collection_id = $1
	`

	err := r.db.SelectContext(ctx, &collaborators, query, collectionID)
	return collaborators, err
}