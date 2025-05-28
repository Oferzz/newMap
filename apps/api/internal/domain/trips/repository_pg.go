package trips

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// PostgresRepository implements the repository interface for PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Create creates a new trip
func (r *PostgresRepository) Create(ctx context.Context, trip *Trip) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert trip
	query := `
		INSERT INTO trips (
			title, description, owner_id, cover_image, privacy, status,
			start_date, end_date, timezone, tags
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at, updated_at`

	err = tx.QueryRowContext(ctx, query,
		trip.Title,
		trip.Description,
		trip.OwnerID,
		trip.CoverImage,
		trip.Privacy,
		trip.Status,
		trip.StartDate,
		trip.EndDate,
		trip.Timezone,
		pq.Array(trip.Tags),
	).Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create trip: %w", err)
	}

	// Add owner as admin collaborator
	collaboratorQuery := `
		INSERT INTO trip_collaborators (
			trip_id, user_id, role, can_edit, can_delete, can_invite, 
			can_moderate_suggestions, joined_at
		) VALUES (
			$1, $2, 'admin', true, true, true, true, CURRENT_TIMESTAMP
		)`

	_, err = tx.ExecContext(ctx, collaboratorQuery, trip.ID, trip.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to add owner as collaborator: %w", err)
	}

	return tx.Commit()
}

// GetByID retrieves a trip by ID with collaborators and waypoints
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Trip, error) {
	var trip Trip
	
	// Get trip
	tripQuery := `
		SELECT 
			id, title, description, owner_id, cover_image, privacy, status,
			start_date, end_date, timezone, tags, view_count, share_count,
			suggestion_count, created_at, updated_at, deleted_at
		FROM trips
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &trip, tripQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("trip not found")
		}
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	// Get collaborators
	collaborators, err := r.getCollaborators(ctx, id)
	if err != nil {
		return nil, err
	}
	trip.Collaborators = collaborators

	// Get waypoints
	waypoints, err := r.getWaypoints(ctx, id)
	if err != nil {
		return nil, err
	}
	trip.Waypoints = waypoints

	return &trip, nil
}

// Update updates a trip
func (r *PostgresRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	// Build dynamic update query
	setClause := ""
	args := []interface{}{id}
	argCount := 2

	for field, value := range updates {
		if setClause != "" {
			setClause += ", "
		}
		// Handle array fields
		if field == "tags" {
			setClause += fmt.Sprintf("%s = $%d", field, argCount)
			args = append(args, pq.Array(value))
		} else {
			setClause += fmt.Sprintf("%s = $%d", field, argCount)
			args = append(args, value)
		}
		argCount++
	}

	if setClause == "" {
		return nil // No updates
	}

	query := fmt.Sprintf(`
		UPDATE trips
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`, setClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update trip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trip not found")
	}

	return nil
}

// Delete soft deletes a trip
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE trips
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete trip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trip not found")
	}

	return nil
}

// List retrieves trips with filters
func (r *PostgresRepository) List(ctx context.Context, filters TripFilters) ([]*Trip, error) {
	var trips []*Trip
	query := `
		SELECT 
			t.id, t.title, t.description, t.owner_id, t.cover_image, 
			t.privacy, t.status, t.start_date, t.end_date, t.timezone, 
			t.tags, t.view_count, t.share_count, t.suggestion_count,
			t.created_at, t.updated_at
		FROM trips t
		WHERE t.deleted_at IS NULL`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if filters.OwnerID != "" {
		query += fmt.Sprintf(" AND t.owner_id = $%d", argCount)
		args = append(args, filters.OwnerID)
		argCount++
	}

	if filters.CollaboratorID != "" {
		query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM trip_collaborators tc WHERE tc.trip_id = t.id AND tc.user_id = $%d)", argCount)
		args = append(args, filters.CollaboratorID)
		argCount++
	}

	if filters.Privacy != "" {
		query += fmt.Sprintf(" AND t.privacy = $%d", argCount)
		args = append(args, filters.Privacy)
		argCount++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND t.status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if len(filters.Tags) > 0 {
		query += fmt.Sprintf(" AND t.tags && $%d", argCount)
		args = append(args, pq.Array(filters.Tags))
		argCount++
	}

	if filters.StartDateFrom != nil {
		query += fmt.Sprintf(" AND t.start_date >= $%d", argCount)
		args = append(args, filters.StartDateFrom)
		argCount++
	}

	if filters.StartDateTo != nil {
		query += fmt.Sprintf(" AND t.start_date <= $%d", argCount)
		args = append(args, filters.StartDateTo)
		argCount++
	}

	if filters.Search != "" {
		query += fmt.Sprintf(" AND (t.title ILIKE $%d OR t.description ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern)
		argCount++
	}

	// Add sorting
	orderBy := " ORDER BY "
	switch filters.SortBy {
	case "title":
		orderBy += "t.title"
	case "start_date":
		orderBy += "t.start_date"
	case "updated_at":
		orderBy += "t.updated_at"
	default:
		orderBy += "t.created_at"
	}

	if strings.ToUpper(filters.SortOrder) == "ASC" {
		orderBy += " ASC"
	} else {
		orderBy += " DESC"
	}
	query += orderBy

	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filters.Limit, filters.Offset)

	err := r.db.SelectContext(ctx, &trips, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list trips: %w", err)
	}

	// Load collaborators and waypoints for each trip
	for _, trip := range trips {
		collaborators, err := r.getCollaborators(ctx, trip.ID)
		if err != nil {
			return nil, err
		}
		trip.Collaborators = collaborators

		waypoints, err := r.getWaypoints(ctx, trip.ID)
		if err != nil {
			return nil, err
		}
		trip.Waypoints = waypoints
	}

	return trips, nil
}

// AddCollaborator adds a collaborator to a trip
func (r *PostgresRepository) AddCollaborator(ctx context.Context, tripID string, collaborator Collaborator) error {
	query := `
		INSERT INTO trip_collaborators (
			trip_id, user_id, role, can_edit, can_delete, can_invite,
			can_moderate_suggestions
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`

	_, err := r.db.ExecContext(ctx, query,
		tripID,
		collaborator.UserID,
		collaborator.Role,
		collaborator.CanEdit,
		collaborator.CanDelete,
		collaborator.CanInvite,
		collaborator.CanModerateSuggestions,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("user is already a collaborator")
		}
		return fmt.Errorf("failed to add collaborator: %w", err)
	}

	return nil
}

// UpdateCollaborator updates a collaborator's role and permissions
func (r *PostgresRepository) UpdateCollaborator(ctx context.Context, tripID, userID string, updates map[string]interface{}) error {
	// Build dynamic update query
	setClause := ""
	args := []interface{}{tripID, userID}
	argCount := 3

	for field, value := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", field, argCount)
		args = append(args, value)
		argCount++
	}

	if setClause == "" {
		return nil // No updates
	}

	query := fmt.Sprintf(`
		UPDATE trip_collaborators
		SET %s
		WHERE trip_id = $1 AND user_id = $2
	`, setClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update collaborator: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("collaborator not found")
	}

	return nil
}

// RemoveCollaborator removes a collaborator from a trip
func (r *PostgresRepository) RemoveCollaborator(ctx context.Context, tripID, userID string) error {
	query := `
		DELETE FROM trip_collaborators
		WHERE trip_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, tripID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove collaborator: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("collaborator not found")
	}

	return nil
}

// GetCollaborator retrieves a specific collaborator
func (r *PostgresRepository) GetCollaborator(ctx context.Context, tripID, userID string) (*Collaborator, error) {
	var collaborator Collaborator
	query := `
		SELECT 
			tc.id, tc.trip_id, tc.user_id, tc.role, tc.can_edit, 
			tc.can_delete, tc.can_invite, tc.can_moderate_suggestions,
			tc.invited_at, tc.joined_at,
			u.username, u.display_name, u.avatar_url
		FROM trip_collaborators tc
		JOIN users u ON tc.user_id = u.id
		WHERE tc.trip_id = $1 AND tc.user_id = $2`

	err := r.db.GetContext(ctx, &collaborator, query, tripID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("collaborator not found")
		}
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}

	return &collaborator, nil
}

// Helper functions

func (r *PostgresRepository) getCollaborators(ctx context.Context, tripID string) ([]Collaborator, error) {
	var collaborators []Collaborator
	query := `
		SELECT 
			tc.id, tc.trip_id, tc.user_id, tc.role, tc.can_edit, 
			tc.can_delete, tc.can_invite, tc.can_moderate_suggestions,
			tc.invited_at, tc.joined_at,
			u.username, u.display_name, u.avatar_url
		FROM trip_collaborators tc
		JOIN users u ON tc.user_id = u.id
		WHERE tc.trip_id = $1
		ORDER BY tc.joined_at`

	err := r.db.SelectContext(ctx, &collaborators, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	return collaborators, nil
}

func (r *PostgresRepository) getWaypoints(ctx context.Context, tripID string) ([]Waypoint, error) {
	var waypoints []Waypoint
	query := `
		SELECT 
			tw.id, tw.trip_id, tw.place_id, tw.order_position,
			tw.arrival_time, tw.departure_time, tw.notes,
			tw.created_at, tw.updated_at,
			p.id as "place.id", p.name as "place.name", 
			p.description as "place.description", p.type as "place.type",
			ST_AsGeoJSON(p.location) as "place.location",
			p.street_address as "place.street_address", 
			p.city as "place.city", p.country as "place.country"
		FROM trip_waypoints tw
		JOIN places p ON tw.place_id = p.id
		WHERE tw.trip_id = $1
		ORDER BY tw.order_position`

	rows, err := r.db.QueryContext(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get waypoints: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var w Waypoint
		var placeLocation sql.NullString

		err := rows.Scan(
			&w.ID, &w.TripID, &w.PlaceID, &w.OrderPosition,
			&w.ArrivalTime, &w.DepartureTime, &w.Notes,
			&w.CreatedAt, &w.UpdatedAt,
			&w.Place.ID, &w.Place.Name, &w.Place.Description, &w.Place.Type,
			&placeLocation, &w.Place.Address, &w.Place.City, &w.Place.Country,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan waypoint: %w", err)
		}

		// Parse GeoJSON location
		if placeLocation.Valid {
			var geoJSON GeoJSON
			if err := geoJSON.Scan(placeLocation.String); err == nil {
				w.Place.Location = &geoJSON
			}
		}

		waypoints = append(waypoints, w)
	}

	return waypoints, nil
}

// IncrementViewCount increments the view count for a trip
func (r *PostgresRepository) IncrementViewCount(ctx context.Context, tripID string) error {
	query := `
		UPDATE trips
		SET view_count = view_count + 1
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, tripID)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

// IncrementShareCount increments the share count for a trip
func (r *PostgresRepository) IncrementShareCount(ctx context.Context, tripID string) error {
	query := `
		UPDATE trips
		SET share_count = share_count + 1
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, tripID)
	if err != nil {
		return fmt.Errorf("failed to increment share count: %w", err)
	}

	return nil
}