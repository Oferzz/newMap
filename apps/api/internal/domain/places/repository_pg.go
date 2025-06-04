package places

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/Oferzz/newMap/apps/api/internal/nlp"
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

// Create creates a new place
func (r *PostgresRepository) Create(ctx context.Context, place *Place) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare location as GeoJSON
	var locationGeoJSON interface{}
	if place.Location != nil {
		locationGeoJSON = fmt.Sprintf(
			"ST_GeomFromGeoJSON('{\"type\":\"Point\",\"coordinates\":[%f,%f]}')",
			place.Location.Coordinates[0],
			place.Location.Coordinates[1],
		)
	}

	// Prepare bounds as GeoJSON
	var boundsGeoJSON interface{}
	if place.Bounds != nil {
		boundsJSON, _ := json.Marshal(place.Bounds)
		boundsGeoJSON = fmt.Sprintf("ST_GeomFromGeoJSON('%s')", string(boundsJSON))
	}

	// Insert place
	query := `
		INSERT INTO places (
			name, description, type, parent_id, location, bounds,
			street_address, city, state, country, postal_code,
			created_by, category, tags, opening_hours, contact_info,
			amenities, privacy, status
		) VALUES (
			$1, $2, $3, $4, %s, %s, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id, created_at, updated_at`

	// Build query with spatial functions
	if locationGeoJSON != nil && boundsGeoJSON != nil {
		query = fmt.Sprintf(query, locationGeoJSON, boundsGeoJSON)
	} else if locationGeoJSON != nil {
		query = fmt.Sprintf(query, locationGeoJSON, "NULL")
	} else if boundsGeoJSON != nil {
		query = fmt.Sprintf(query, "NULL", boundsGeoJSON)
	} else {
		query = fmt.Sprintf(query, "NULL", "NULL")
	}

	// Execute query
	args := []interface{}{
		place.Name,
		place.Description,
		place.Type,
		place.ParentID,
		place.StreetAddress,
		place.City,
		place.State,
		place.Country,
		place.PostalCode,
		place.CreatedBy,
		pq.Array(place.Category),
		pq.Array(place.Tags),
		place.OpeningHours,
		place.ContactInfo,
		pq.Array(place.Amenities),
		place.Privacy,
		place.Status,
	}

	// Remove location and bounds from args if they're included in the query
	if locationGeoJSON != nil || boundsGeoJSON != nil {
		// Adjust args to skip the spatial placeholders
		filteredArgs := []interface{}{
			args[0], args[1], args[2], args[3], // name, desc, type, parent_id
		}
		filteredArgs = append(filteredArgs, args[4:]...) // skip location/bounds positions
		args = filteredArgs
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&place.ID, &place.CreatedAt, &place.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create place: %w", err)
	}

	// Add creator as admin collaborator if needed
	if place.Type != "poi" { // Only for areas and regions
		collaboratorQuery := `
			INSERT INTO place_collaborators (
				place_id, user_id, role, permissions
			) VALUES (
				$1, $2, 'admin', '{"all": true}'::jsonb
			)`

		_, err = tx.ExecContext(ctx, collaboratorQuery, place.ID, place.CreatedBy)
		if err != nil {
			return fmt.Errorf("failed to add creator as collaborator: %w", err)
		}
	}

	return tx.Commit()
}

// GetByID retrieves a place by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Place, error) {
	var place Place
	query := `
		SELECT 
			id, name, description, type, parent_id,
			ST_AsGeoJSON(location) as location,
			ST_AsGeoJSON(bounds) as bounds,
			street_address, city, state, country, postal_code,
			created_by, category, tags, opening_hours, contact_info,
			amenities, average_rating, rating_count, privacy, status,
			created_at, updated_at
		FROM places
		WHERE id = $1 AND status = 'active'`

	var locationJSON, boundsJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&place.ID,
		&place.Name,
		&place.Description,
		&place.Type,
		&place.ParentID,
		&locationJSON,
		&boundsJSON,
		&place.StreetAddress,
		&place.City,
		&place.State,
		&place.Country,
		&place.PostalCode,
		&place.CreatedBy,
		pq.Array(&place.Category),
		pq.Array(&place.Tags),
		&place.OpeningHours,
		&place.ContactInfo,
		pq.Array(&place.Amenities),
		&place.AverageRating,
		&place.RatingCount,
		&place.Privacy,
		&place.Status,
		&place.CreatedAt,
		&place.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("place not found")
		}
		return nil, fmt.Errorf("failed to get place: %w", err)
	}

	// Parse spatial data
	if locationJSON.Valid {
		var geoPoint GeoPoint
		if err := json.Unmarshal([]byte(locationJSON.String), &geoPoint); err == nil {
			place.Location = &geoPoint
		}
	}

	if boundsJSON.Valid {
		var geoPolygon GeoPolygon
		if err := json.Unmarshal([]byte(boundsJSON.String), &geoPolygon); err == nil {
			place.Bounds = &geoPolygon
		}
	}

	// Get media
	media, err := r.getPlaceMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	place.Media = media

	// Get collaborators
	collaborators, err := r.getCollaborators(ctx, id)
	if err != nil {
		return nil, err
	}
	place.Collaborators = collaborators

	return &place, nil
}

// Update updates a place
func (r *PostgresRepository) UpdateByID(ctx context.Context, id string, updates map[string]interface{}) error {
	// Build dynamic update query
	setClause := ""
	args := []interface{}{id}
	argCount := 2

	for field, value := range updates {
		if setClause != "" {
			setClause += ", "
		}
		
		// Handle special fields
		switch field {
		case "category", "tags", "amenities":
			setClause += fmt.Sprintf("%s = $%d", field, argCount)
			args = append(args, pq.Array(value))
		case "location":
			if loc, ok := value.(*LocationInput); ok && loc != nil {
				setClause += fmt.Sprintf("location = ST_GeomFromGeoJSON('{\"type\":\"Point\",\"coordinates\":[%f,%f]}')", 
					loc.Longitude, loc.Latitude)
				argCount-- // Don't increment as we're not using a placeholder
			}
		case "opening_hours", "contact_info":
			jsonData, _ := json.Marshal(value)
			setClause += fmt.Sprintf("%s = $%d::jsonb", field, argCount)
			args = append(args, string(jsonData))
		default:
			setClause += fmt.Sprintf("%s = $%d", field, argCount)
			args = append(args, value)
		}
		argCount++
	}

	if setClause == "" {
		return nil // No updates
	}

	query := fmt.Sprintf(`
		UPDATE places
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'
	`, setClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update place: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("place not found")
	}

	return nil
}

// Delete soft deletes a place
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE places
		SET status = 'archived', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete place: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("place not found")
	}

	return nil
}

// Search searches for places
func (r *PostgresRepository) SearchPlaces(ctx context.Context, input SearchPlacesInput) ([]*Place, error) {
	var places []*Place
	query := `
		SELECT 
			id, name, description, type, parent_id,
			ST_AsGeoJSON(location) as location,
			street_address, city, state, country, postal_code,
			created_by, category, tags, average_rating, rating_count,
			privacy, status, created_at, updated_at
		FROM places
		WHERE status = 'active'`

	args := []interface{}{}
	argCount := 1

	// Text search
	if input.Query != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		searchPattern := "%" + input.Query + "%"
		args = append(args, searchPattern)
		argCount++
	}

	// Type filter
	if input.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, input.Type)
		argCount++
	}

	// Category filter
	if len(input.Category) > 0 {
		query += fmt.Sprintf(" AND category && $%d", argCount)
		args = append(args, pq.Array(input.Category))
		argCount++
	}

	// Tags filter
	if len(input.Tags) > 0 {
		query += fmt.Sprintf(" AND tags && $%d", argCount)
		args = append(args, pq.Array(input.Tags))
		argCount++
	}

	// Location filter
	if input.City != "" {
		query += fmt.Sprintf(" AND city ILIKE $%d", argCount)
		args = append(args, "%"+input.City+"%")
		argCount++
	}

	if input.Country != "" {
		query += fmt.Sprintf(" AND country ILIKE $%d", argCount)
		args = append(args, "%"+input.Country+"%")
		argCount++
	}

	// Spatial query
	if input.Latitude != nil && input.Longitude != nil && input.Radius != nil {
		query += fmt.Sprintf(` AND ST_DWithin(
			location::geography,
			ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography,
			$%d
		)`, argCount, argCount+1, argCount+2)
		args = append(args, *input.Longitude, *input.Latitude, *input.Radius)
		argCount += 3
	}

	// Ordering
	if input.Latitude != nil && input.Longitude != nil {
		query += fmt.Sprintf(` ORDER BY location <-> ST_SetSRID(ST_MakePoint($%d, $%d), 4326)`, argCount, argCount+1)
		args = append(args, *input.Longitude, *input.Latitude)
		argCount += 2
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, input.Limit, input.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search places: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var place Place
		var locationJSON sql.NullString

		err := rows.Scan(
			&place.ID,
			&place.Name,
			&place.Description,
			&place.Type,
			&place.ParentID,
			&locationJSON,
			&place.StreetAddress,
			&place.City,
			&place.State,
			&place.Country,
			&place.PostalCode,
			&place.CreatedBy,
			pq.Array(&place.Category),
			pq.Array(&place.Tags),
			&place.AverageRating,
			&place.RatingCount,
			&place.Privacy,
			&place.Status,
			&place.CreatedAt,
			&place.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}

		// Parse location
		if locationJSON.Valid {
			var geoPoint GeoPoint
			if err := json.Unmarshal([]byte(locationJSON.String), &geoPoint); err == nil {
				place.Location = &geoPoint
			}
		}

		places = append(places, &place)
	}

	return places, nil
}

// GetNearby finds nearby places
func (r *PostgresRepository) GetNearbyPlaces(ctx context.Context, input NearbyPlacesInput) ([]*Place, error) {
	// Convert to SearchPlacesInput and use Search method
	searchInput := SearchPlacesInput{
		Type:      input.Type,
		Category:  input.Category,
		Tags:      input.Tags,
		Latitude:  &input.Latitude,
		Longitude: &input.Longitude,
		Radius:    &input.Radius,
		Limit:     input.Limit,
		Offset:    input.Offset,
	}

	result, err := r.Search(ctx, searchInput.Query, SearchFilters{
		Category: searchInput.Category,
		Tags:     searchInput.Tags,
		Limit:    searchInput.Limit,
		Offset:   searchInput.Offset,
	})
	if err != nil {
		return nil, err
	}
	return result.Places, nil
}

// GetByTripID retrieves all places for a trip
func (r *PostgresRepository) GetByTripID(ctx context.Context, tripID string) ([]*Place, error) {
	var places []*Place
	query := `
		SELECT DISTINCT
			p.id, p.name, p.description, p.type, p.parent_id,
			ST_AsGeoJSON(p.location) as location,
			p.street_address, p.city, p.state, p.country, p.postal_code,
			p.created_by, p.category, p.tags, p.average_rating, p.rating_count,
			p.privacy, p.status, p.created_at, p.updated_at
		FROM places p
		JOIN trip_waypoints tw ON p.id = tw.place_id
		WHERE tw.trip_id = $1 AND p.status = 'active'
		ORDER BY tw.order_position`

	rows, err := r.db.QueryContext(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get places by trip: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var place Place
		var locationJSON sql.NullString

		err := rows.Scan(
			&place.ID,
			&place.Name,
			&place.Description,
			&place.Type,
			&place.ParentID,
			&locationJSON,
			&place.StreetAddress,
			&place.City,
			&place.State,
			&place.Country,
			&place.PostalCode,
			&place.CreatedBy,
			pq.Array(&place.Category),
			pq.Array(&place.Tags),
			&place.AverageRating,
			&place.RatingCount,
			&place.Privacy,
			&place.Status,
			&place.CreatedAt,
			&place.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}

		// Parse location
		if locationJSON.Valid {
			var geoPoint GeoPoint
			if err := json.Unmarshal([]byte(locationJSON.String), &geoPoint); err == nil {
				place.Location = &geoPoint
			}
		}

		places = append(places, &place)
	}

	return places, nil
}

// GetChildren retrieves child places
func (r *PostgresRepository) GetChildren(ctx context.Context, parentID string) ([]*Place, error) {
	var places []*Place
	query := `
		SELECT 
			id, name, description, type, parent_id,
			ST_AsGeoJSON(location) as location,
			street_address, city, state, country, postal_code,
			created_by, category, tags, average_rating, rating_count,
			privacy, status, created_at, updated_at
		FROM places
		WHERE parent_id = $1 AND status = 'active'
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child places: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var place Place
		var locationJSON sql.NullString

		err := rows.Scan(
			&place.ID,
			&place.Name,
			&place.Description,
			&place.Type,
			&place.ParentID,
			&locationJSON,
			&place.StreetAddress,
			&place.City,
			&place.State,
			&place.Country,
			&place.PostalCode,
			&place.CreatedBy,
			pq.Array(&place.Category),
			pq.Array(&place.Tags),
			&place.AverageRating,
			&place.RatingCount,
			&place.Privacy,
			&place.Status,
			&place.CreatedAt,
			&place.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}

		// Parse location
		if locationJSON.Valid {
			var geoPoint GeoPoint
			if err := json.Unmarshal([]byte(locationJSON.String), &geoPoint); err == nil {
				place.Location = &geoPoint
			}
		}

		places = append(places, &place)
	}

	return places, nil
}

// Helper functions

func (r *PostgresRepository) getPlaceMedia(ctx context.Context, placeID string) ([]Media, error) {
	var media []Media
	query := `
		SELECT 
			pm.id, pm.media_id, pm.place_id, pm.caption, pm.order_position,
			pm.created_at, m.cdn_url, m.thumbnail_medium, m.mime_type, m.uploaded_by
		FROM place_media pm
		JOIN media m ON pm.media_id = m.id
		WHERE pm.place_id = $1
		ORDER BY pm.order_position, pm.created_at`

	err := r.db.SelectContext(ctx, &media, query, placeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get place media: %w", err)
	}

	return media, nil
}

func (r *PostgresRepository) getCollaborators(ctx context.Context, placeID string) ([]Collaborator, error) {
	var collaborators []Collaborator
	query := `
		SELECT 
			pc.id, pc.place_id, pc.user_id, pc.role, pc.permissions, pc.created_at,
			u.username, u.display_name, u.avatar_url
		FROM place_collaborators pc
		JOIN users u ON pc.user_id = u.id
		WHERE pc.place_id = $1
		ORDER BY pc.created_at`

	err := r.db.SelectContext(ctx, &collaborators, query, placeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	return collaborators, nil
}

// UpdateRating updates the average rating for a place
func (r *PostgresRepository) UpdateRating(ctx context.Context, placeID string, rating float64, count int) error {
	query := `
		UPDATE places
		SET average_rating = $2,
			rating_count = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, placeID, rating, count)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	return nil
}

// IncrementRatingCount increments the rating count
func (r *PostgresRepository) IncrementRatingCount(ctx context.Context, placeID string) error {
	query := `
		UPDATE places
		SET rating_count = rating_count + 1,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'`

	_, err := r.db.ExecContext(ctx, query, placeID)
	if err != nil {
		return fmt.Errorf("failed to increment rating count: %w", err)
	}

	return nil
}

// Add missing string import
// import "strings" - should be added at the top

// GetByCreator retrieves all places created by a specific user
func (r *PostgresRepository) GetByCreator(ctx context.Context, creatorID string) ([]*Place, error) {
	query := `
		SELECT 
			id, name, description, type, parent_id,
			ST_AsGeoJSON(location) as location,
			street_address, city, state, country, postal_code,
			created_by, category, tags, average_rating, rating_count,
			privacy, status, created_at, updated_at
		FROM places
		WHERE created_by = $1 AND status = 'active'
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get places by creator: %w", err)
	}
	defer rows.Close()
	
	var places []*Place
	for rows.Next() {
		place := &Place{}
		var locationJSON sql.NullString
		var tagsArray pq.StringArray
		
		err := rows.Scan(
			&place.ID, &place.Name, &place.Description, &place.Type, &place.ParentID,
			&locationJSON, &place.StreetAddress, &place.City, &place.State,
			&place.Country, &place.PostalCode, &place.CreatedBy, &place.Category,
			&tagsArray, &place.AverageRating, &place.RatingCount,
			&place.Privacy, &place.Status, &place.CreatedAt, &place.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse location
		if locationJSON.Valid {
			var geoJSON struct {
				Type        string    `json:"type"`
				Coordinates []float64 `json:"coordinates"`
			}
			if err := json.Unmarshal([]byte(locationJSON.String), &geoJSON); err == nil && len(geoJSON.Coordinates) >= 2 {
				place.Location = &GeoPoint{
					Type:        geoJSON.Type,
					Coordinates: geoJSON.Coordinates,
				}
			}
		}
		
		place.Tags = tagsArray
		places = append(places, place)
	}
	
	return places, nil
}

// Update updates a place - implementation matches interface
func (r *PostgresRepository) Update(ctx context.Context, place *Place) error {
	// Convert to map and use existing Update method
	updates := map[string]interface{}{
		"name":           place.Name,
		"description":    place.Description,
		"type":           place.Type,
		"street_address": place.StreetAddress,
		"city":           place.City,
		"state":          place.State,
		"country":        place.Country,
		"postal_code":    place.PostalCode,
		"category":       place.Category,
		"tags":           place.Tags,
		"privacy":        place.Privacy,
		"status":         place.Status,
		"updated_at":     time.Now(),
	}
	
	return r.UpdateByID(ctx, place.ID, updates)
}

// GetByCategory retrieves places by category with pagination
func (r *PostgresRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*Place, error) {
	query := `
		SELECT 
			id, name, description, type, parent_id,
			ST_AsGeoJSON(location) as location,
			street_address, city, state, country, postal_code,
			created_by, category, tags, average_rating, rating_count,
			privacy, status, created_at, updated_at
		FROM places
		WHERE category = $1 AND status = 'active'
		ORDER BY average_rating DESC, created_at DESC
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.QueryContext(ctx, query, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get places by category: %w", err)
	}
	defer rows.Close()
	
	var places []*Place
	for rows.Next() {
		place := &Place{}
		var locationJSON sql.NullString
		var tagsArray pq.StringArray
		
		err := rows.Scan(
			&place.ID, &place.Name, &place.Description, &place.Type, &place.ParentID,
			&locationJSON, &place.StreetAddress, &place.City, &place.State,
			&place.Country, &place.PostalCode, &place.CreatedBy, &place.Category,
			&tagsArray, &place.AverageRating, &place.RatingCount,
			&place.Privacy, &place.Status, &place.CreatedAt, &place.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse location
		if locationJSON.Valid {
			var geoJSON struct {
				Type        string    `json:"type"`
				Coordinates []float64 `json:"coordinates"`
			}
			if err := json.Unmarshal([]byte(locationJSON.String), &geoJSON); err == nil && len(geoJSON.Coordinates) >= 2 {
				place.Location = &GeoPoint{
					Type:        geoJSON.Type,
					Coordinates: geoJSON.Coordinates,
				}
			}
		}
		
		place.Tags = tagsArray
		places = append(places, place)
	}
	
	return places, nil
}


// Search implements the Repository interface Search method
func (r *PostgresRepository) Search(ctx context.Context, query string, filters SearchFilters) (*SearchResult, error) {
	// Use existing SearchPlaces method
	input := SearchPlacesInput{
		Query:    query,
		Category: filters.Category,
		Tags:     filters.Tags,
		Limit:    filters.Limit,
		Offset:   filters.Offset,
	}
	
	places, err := r.SearchPlaces(ctx, input)
	if err != nil {
		return nil, err
	}
	
	// Count total results - simplified for now
	total := int64(len(places))
	
	return &SearchResult{
		Places: places,
		Total:  total,
	}, nil
}

// GetNearby implements the Repository interface GetNearby method
func (r *PostgresRepository) GetNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*Place, error) {
	// Use existing GetNearbyPlaces method
	input := NearbyPlacesInput{
		Latitude:  lat,
		Longitude: lng,
		Radius:    int(radiusKm * 1000), // Convert km to meters
		Limit:     limit,
	}
	
	return r.GetNearbyPlaces(ctx, input)
}

// GetInBounds retrieves places within geographical bounds
func (r *PostgresRepository) GetInBounds(ctx context.Context, bounds Bounds) ([]*Place, error) {
	query := `
		SELECT 
			id, name, description, type, parent_id,
			ST_AsGeoJSON(location) as location,
			street_address, city, state, country, postal_code,
			created_by, category, tags, average_rating, rating_count,
			privacy, status, created_at, updated_at
		FROM places
		WHERE status = 'active'
			AND ST_Within(
				location,
				ST_MakeEnvelope($1, $2, $3, $4, 4326)
			)
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, bounds.MinLng, bounds.MinLat, bounds.MaxLng, bounds.MaxLat)
	if err != nil {
		return nil, fmt.Errorf("failed to get places in bounds: %w", err)
	}
	defer rows.Close()
	
	var places []*Place
	for rows.Next() {
		place := &Place{}
		var locationJSON sql.NullString
		var tagsArray pq.StringArray
		
		err := rows.Scan(
			&place.ID, &place.Name, &place.Description, &place.Type, &place.ParentID,
			&locationJSON, &place.StreetAddress, &place.City, &place.State,
			&place.Country, &place.PostalCode, &place.CreatedBy, &place.Category,
			&tagsArray, &place.AverageRating, &place.RatingCount,
			&place.Privacy, &place.Status, &place.CreatedAt, &place.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse location
		if locationJSON.Valid {
			var geoJSON struct {
				Type        string    `json:"type"`
				Coordinates []float64 `json:"coordinates"`
			}
			if err := json.Unmarshal([]byte(locationJSON.String), &geoJSON); err == nil && len(geoJSON.Coordinates) >= 2 {
				place.Location = &GeoPoint{
					Type:        geoJSON.Type,
					Coordinates: geoJSON.Coordinates,
				}
			}
		}
		
		place.Tags = tagsArray
		places = append(places, place)
	}
	
	return places, nil
}
// SearchWithSpatialContext performs spatial search with enhanced area filtering
func (r *PostgresRepository) SearchWithSpatialContext(ctx context.Context, query string, spatial *nlp.SpatialSearchContext, filters SearchFilters) (*SearchResult, error) {
	baseQuery := `
		SELECT 
			id, name, description, type, parent_id,
			ST_AsGeoJSON(location) as location,
			ST_AsGeoJSON(bounds) as bounds,
			street_address, city, state, country, postal_code,
			created_by, category, tags, average_rating, rating_count,
			privacy, status, created_at, updated_at
		FROM places
		WHERE status = 'active'`
	
	var conditions []string
	var args []interface{}
	argCount := 0
	
	// Text search
	if query != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+query+"%")
	}
	
	// Category filter
	if len(filters.Category) > 0 {
		argCount++
		conditions = append(conditions, fmt.Sprintf("category = ANY($%d)", argCount))
		args = append(args, pq.Array(filters.Category))
	}
	
	// Spatial filters
	if spatial != nil {
		spatialConditions, spatialArgs := r.buildSpatialConditions(spatial, argCount)
		conditions = append(conditions, spatialConditions...)
		args = append(args, spatialArgs...)
	}
	
	// Build final query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}
	
	// Add ordering
	baseQuery += " ORDER BY created_at DESC"
	
	// Add limit/offset
	if filters.Limit > 0 {
		argCount = len(args) + 1
		baseQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}
	if filters.Offset > 0 {
		argCount = len(args) + 1
		baseQuery += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}
	
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search places with spatial context: %w", err)
	}
	defer rows.Close()
	
	var places []*Place
	for rows.Next() {
		place := &Place{}
		var locationJSON, boundsJSON sql.NullString
		var tagsArray pq.StringArray
		
		err := rows.Scan(
			&place.ID, &place.Name, &place.Description, &place.Type, &place.ParentID,
			&locationJSON, &boundsJSON, &place.StreetAddress, &place.City, &place.State,
			&place.Country, &place.PostalCode, &place.CreatedBy, &place.Category,
			&tagsArray, &place.AverageRating, &place.RatingCount,
			&place.Privacy, &place.Status, &place.CreatedAt, &place.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Parse location
		if locationJSON.Valid {
			if err := r.parseLocationJSON(locationJSON.String, place); err != nil {
				continue // Skip invalid locations
			}
		}
		
		// Parse bounds
		if boundsJSON.Valid {
			if err := r.parseBoundsJSON(boundsJSON.String, place); err != nil {
				// Bounds are optional, don't skip
			}
		}
		
		place.Tags = tagsArray
		places = append(places, place)
	}
	
	return &SearchResult{
		Places: places,
		Total:  int64(len(places)), // Simplified total count
	}, nil
}

// buildSpatialConditions creates PostGIS spatial query conditions
func (r *PostgresRepository) buildSpatialConditions(spatial *nlp.SpatialSearchContext, startArgCount int) ([]string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := startArgCount
	
	// Within area - place must be completely within the specified area
	if spatial.Within != nil {
		condition, spatialArgs := r.buildAreaCondition("ST_Within", spatial.Within, argCount)
		if condition != "" {
			conditions = append(conditions, condition)
			args = append(args, spatialArgs...)
			argCount += len(spatialArgs)
		}
	}
	
	// Near area - place must be within distance of the specified area
	if spatial.Near != nil {
		condition, spatialArgs := r.buildDistanceCondition(spatial.Near, argCount)
		if condition != "" {
			conditions = append(conditions, condition)
			args = append(args, spatialArgs...)
			argCount += len(spatialArgs)
		}
	}
	
	// Intersects area - place must intersect with the specified area
	if spatial.Intersects != nil {
		condition, spatialArgs := r.buildAreaCondition("ST_Intersects", spatial.Intersects, argCount)
		if condition != "" {
			conditions = append(conditions, condition)
			args = append(args, spatialArgs...)
			argCount += len(spatialArgs)
		}
	}
	
	// Multiple areas - place must be in one of the specified areas
	for _, area := range spatial.Areas {
		condition, spatialArgs := r.buildAreaCondition("ST_Within", &area, argCount)
		if condition != "" {
			conditions = append(conditions, condition)
			args = append(args, spatialArgs...)
			argCount += len(spatialArgs)
		}
	}
	
	return conditions, args
}

// buildAreaCondition creates spatial conditions for geometric areas
func (r *PostgresRepository) buildAreaCondition(operation string, area *nlp.AreaFilter, argCount int) (string, []interface{}) {
	if area == nil {
		return "", nil
	}
	
	switch area.Type {
	case "circle":
		// For circles, we need coordinates and radius
		if coords, ok := area.Coordinates.([]interface{}); ok && len(coords) >= 2 && area.Radius != nil {
			lat, latOk := coords[1].(float64)
			lng, lngOk := coords[0].(float64)
			if latOk && lngOk {
				condition := fmt.Sprintf(`%s(
					location::geography,
					ST_Buffer(
						ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography,
						$%d * 1000
					)
				)`, operation, argCount+1, argCount+2, argCount+3)
				return condition, []interface{}{lng, lat, *area.Radius}
			}
		}
		
	case "polygon":
		// For polygons, coordinates should be a GeoJSON-style coordinate array
		if coords, ok := area.Coordinates.([]interface{}); ok {
			// Build GeoJSON polygon
			coordsJSON, err := json.Marshal(map[string]interface{}{
				"type":        "Polygon",
				"coordinates": coords,
			})
			if err == nil {
				condition := fmt.Sprintf(`%s(
					location,
					ST_GeomFromGeoJSON($%d)
				)`, operation, argCount+1)
				return condition, []interface{}{string(coordsJSON)}
			}
		}
		
	case "region":
		// For named regions, we'd typically look up in a regions table
		// For now, we'll do a simple text match on location fields
		condition := fmt.Sprintf(`(
			city ILIKE $%d OR 
			state ILIKE $%d OR 
			country ILIKE $%d
		)`, argCount+1, argCount+1, argCount+1)
		return condition, []interface{}{"%" + area.Name + "%"}
		
	case "bounds":
		// For rectangular bounds
		if coords, ok := area.Coordinates.([]interface{}); ok && len(coords) >= 4 {
			minLng, _ := coords[0].(float64)
			minLat, _ := coords[1].(float64)
			maxLng, _ := coords[2].(float64)
			maxLat, _ := coords[3].(float64)
			
			condition := fmt.Sprintf(`%s(
				location,
				ST_MakeEnvelope($%d, $%d, $%d, $%d, 4326)
			)`, operation, argCount+1, argCount+2, argCount+3, argCount+4)
			return condition, []interface{}{minLng, minLat, maxLng, maxLat}
		}
	}
	
	return "", nil
}

// buildDistanceCondition creates distance-based spatial conditions
func (r *PostgresRepository) buildDistanceCondition(area *nlp.AreaFilter, argCount int) (string, []interface{}) {
	if area == nil || area.Radius == nil {
		return "", nil
	}
	
	switch area.Type {
	case "circle":
		if coords, ok := area.Coordinates.([]interface{}); ok && len(coords) >= 2 {
			lat, latOk := coords[1].(float64)
			lng, lngOk := coords[0].(float64)
			if latOk && lngOk {
				condition := fmt.Sprintf(`ST_DWithin(
					location::geography,
					ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography,
					$%d * 1000
				)`, argCount+1, argCount+2, argCount+3)
				return condition, []interface{}{lng, lat, *area.Radius}
			}
		}
	}
	
	return "", nil
}

// GetInArea retrieves places within a specific geometric area
func (r *PostgresRepository) GetInArea(ctx context.Context, area nlp.AreaFilter) ([]*Place, error) {
	spatial := &nlp.SpatialSearchContext{
		Within: &area,
	}
	
	result, err := r.SearchWithSpatialContext(ctx, "", spatial, SearchFilters{Limit: 100})
	if err != nil {
		return nil, err
	}
	
	return result.Places, nil
}

// GetIntersecting retrieves places that intersect with a specific geometric area
func (r *PostgresRepository) GetIntersecting(ctx context.Context, area nlp.AreaFilter) ([]*Place, error) {
	spatial := &nlp.SpatialSearchContext{
		Intersects: &area,
	}
	
	result, err := r.SearchWithSpatialContext(ctx, "", spatial, SearchFilters{Limit: 100})
	if err != nil {
		return nil, err
	}
	
	return result.Places, nil
}

// GetWithinDistance retrieves places within a specific distance of an area
func (r *PostgresRepository) GetWithinDistance(ctx context.Context, area nlp.AreaFilter) ([]*Place, error) {
	spatial := &nlp.SpatialSearchContext{
		Near: &area,
	}
	
	result, err := r.SearchWithSpatialContext(ctx, "", spatial, SearchFilters{Limit: 100})
	if err != nil {
		return nil, err
	}
	
	return result.Places, nil
}

// Helper methods for JSON parsing
func (r *PostgresRepository) parseLocationJSON(locationStr string, place *Place) error {
	var geoJSON struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	}
	if err := json.Unmarshal([]byte(locationStr), &geoJSON); err != nil {
		return err
	}
	if len(geoJSON.Coordinates) >= 2 {
		place.Location = &GeoPoint{
			Type:        geoJSON.Type,
			Coordinates: geoJSON.Coordinates,
		}
	}
	return nil
}

func (r *PostgresRepository) parseBoundsJSON(boundsStr string, place *Place) error {
	var geoJSON struct {
		Type        string          `json:"type"`
		Coordinates [][][]float64   `json:"coordinates"`
	}
	if err := json.Unmarshal([]byte(boundsStr), &geoJSON); err != nil {
		return err
	}
	if len(geoJSON.Coordinates) > 0 {
		place.Bounds = &GeoPolygon{
			Type:        geoJSON.Type,
			Coordinates: geoJSON.Coordinates,
		}
	}
	return nil
}
