package places

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		place := &Place{
			ID:          uuid.New().String(),
			Name:        "Test Place",
			Description: "Test Description",
			Location: Location{
				Type:        "Point",
				Coordinates: []float64{-73.935242, 40.730610},
			},
			Address:    "123 Test St, New York, NY",
			Category:   "restaurant",
			CreatorID:  uuid.New().String(),
			MediaURLs:  []string{"image1.jpg", "image2.jpg"},
			Tags:       []string{"food", "italian"},
			Rating:     4.5,
			ReviewCount: 10,
			Metadata: map[string]interface{}{
				"cuisine": "Italian",
				"price":   "$$",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(place.ID, place.CreatedAt, place.UpdatedAt)

		mock.ExpectQuery(`INSERT INTO places`).
			WithArgs(
				place.ID,
				place.Name,
				place.Description,
				sqlmock.AnyArg(), // location PostGIS point
				place.Address,
				place.Category,
				place.CreatorID,
				sqlmock.AnyArg(), // media_urls array
				sqlmock.AnyArg(), // tags array
				place.Rating,
				place.ReviewCount,
				sqlmock.AnyArg(), // metadata jsonb
				place.ParentID,
				place.CreatedAt,
				place.UpdatedAt,
			).
			WillReturnRows(rows)

		err := repo.Create(ctx, place)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		place := &Place{
			ID:        uuid.New().String(),
			Name:      "Test Place",
			CreatorID: uuid.New().String(),
			Location: Location{
				Type:        "Point",
				Coordinates: []float64{-73.935242, 40.730610},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectQuery(`INSERT INTO places`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, place)
		assert.Error(t, err)
	})
}

func TestPostgreSQLRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("place found", func(t *testing.T) {
		placeID := uuid.New().String()
		creatorID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "ST_X(location)", "ST_Y(location)",
			"address", "category", "creator_id", "media_urls", "tags",
			"rating", "review_count", "metadata", "parent_id",
			"created_at", "updated_at",
		}).AddRow(
			placeID, "Test Place", "Test Description", -73.935242, 40.730610,
			"123 Test St", "restaurant", creatorID, "{image1.jpg,image2.jpg}",
			"{food,italian}", 4.5, 10, `{"cuisine": "Italian", "price": "$$"}`,
			nil, now, now,
		)

		mock.ExpectQuery(`SELECT (.+) FROM places WHERE id = \$1`).
			WithArgs(placeID).
			WillReturnRows(rows)

		place, err := repo.GetByID(ctx, placeID)
		assert.NoError(t, err)
		assert.NotNil(t, place)
		assert.Equal(t, placeID, place.ID)
		assert.Equal(t, "Test Place", place.Name)
		assert.Equal(t, -73.935242, place.Location.Coordinates[0])
		assert.Equal(t, 40.730610, place.Location.Coordinates[1])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("place not found", func(t *testing.T) {
		placeID := uuid.New().String()

		mock.ExpectQuery(`SELECT (.+) FROM places WHERE id = \$1`).
			WithArgs(placeID).
			WillReturnError(sql.ErrNoRows)

		place, err := repo.GetByID(ctx, placeID)
		assert.Error(t, err)
		assert.Nil(t, place)
	})
}

func TestPostgreSQLRepository_GetByCreator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("places found", func(t *testing.T) {
		creatorID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "ST_X(location)", "ST_Y(location)",
			"address", "category", "creator_id", "media_urls", "tags",
			"rating", "review_count", "metadata", "parent_id",
			"created_at", "updated_at",
		}).
			AddRow(
				"place1", "Place 1", "Description 1", -73.935242, 40.730610,
				"Address 1", "restaurant", creatorID, "{}", "{}",
				4.5, 10, "{}", nil, now, now,
			).
			AddRow(
				"place2", "Place 2", "Description 2", -73.945242, 40.740610,
				"Address 2", "cafe", creatorID, "{}", "{}",
				4.0, 5, "{}", nil, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM places WHERE creator_id = \$1`).
			WithArgs(creatorID).
			WillReturnRows(rows)

		places, err := repo.GetByCreator(ctx, creatorID)
		assert.NoError(t, err)
		assert.Len(t, places, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		place := &Place{
			ID:          uuid.New().String(),
			Name:        "Updated Place",
			Description: "Updated Description",
			Location: Location{
				Type:        "Point",
				Coordinates: []float64{-73.935242, 40.730610},
			},
			Address:     "456 Updated St",
			Category:    "cafe",
			MediaURLs:   []string{"new1.jpg", "new2.jpg"},
			Tags:        []string{"coffee", "wifi"},
			Rating:      4.8,
			ReviewCount: 25,
			UpdatedAt:   time.Now(),
		}

		mock.ExpectExec(`UPDATE places SET`).
			WithArgs(
				place.Name,
				place.Description,
				sqlmock.AnyArg(), // location
				place.Address,
				place.Category,
				sqlmock.AnyArg(), // media_urls
				sqlmock.AnyArg(), // tags
				place.Rating,
				place.ReviewCount,
				sqlmock.AnyArg(), // metadata
				place.ParentID,
				sqlmock.AnyArg(), // updated_at
				place.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, place)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		placeID := uuid.New().String()

		mock.ExpectExec(`DELETE FROM places WHERE id = \$1`).
			WithArgs(placeID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, placeID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_Search(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("search with results", func(t *testing.T) {
		query := "coffee"
		filters := SearchFilters{
			Limit:  10,
			Offset: 0,
		}
		now := time.Now()

		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM places WHERE`).
			WithArgs("%"+query+"%", "%"+query+"%").
			WillReturnRows(countRows)

		// Search query
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "ST_X(location)", "ST_Y(location)",
			"address", "category", "creator_id", "media_urls", "tags",
			"rating", "review_count", "metadata", "parent_id",
			"created_at", "updated_at",
		}).
			AddRow(
				"place1", "Coffee Shop 1", "Great coffee", -73.935242, 40.730610,
				"Address 1", "cafe", "creator1", "{}", "{coffee}",
				4.5, 10, "{}", nil, now, now,
			).
			AddRow(
				"place2", "Coffee House", "Best coffee", -73.945242, 40.740610,
				"Address 2", "cafe", "creator2", "{}", "{coffee,wifi}",
				4.8, 20, "{}", nil, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM places WHERE`).
			WithArgs("%"+query+"%", "%"+query+"%", filters.Limit, filters.Offset).
			WillReturnRows(rows)

		result, err := repo.Search(ctx, query, filters)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Places, 2)
		assert.Equal(t, int64(2), result.Total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetNearby(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("get nearby places", func(t *testing.T) {
		lat := 40.730610
		lng := -73.935242
		radiusKm := 1.0
		limit := 10
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "ST_X(location)", "ST_Y(location)",
			"address", "category", "creator_id", "media_urls", "tags",
			"rating", "review_count", "metadata", "parent_id",
			"created_at", "updated_at", "distance",
		}).
			AddRow(
				"place1", "Nearby Place 1", "Description 1", -73.936242, 40.731610,
				"Address 1", "restaurant", "creator1", "{}", "{}",
				4.5, 10, "{}", nil, now, now, 0.15,
			).
			AddRow(
				"place2", "Nearby Place 2", "Description 2", -73.934242, 40.729610,
				"Address 2", "cafe", "creator2", "{}", "{}",
				4.0, 5, "{}", nil, now, now, 0.18,
			)

		mock.ExpectQuery(`SELECT (.+), ST_Distance`).
			WithArgs(lng, lat, radiusKm*1000, limit).
			WillReturnRows(rows)

		places, err := repo.GetNearby(ctx, lat, lng, radiusKm, limit)
		assert.NoError(t, err)
		assert.Len(t, places, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetInBounds(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("get places in bounds", func(t *testing.T) {
		bounds := Bounds{
			MinLat: 40.720000,
			MaxLat: 40.740000,
			MinLng: -73.940000,
			MaxLng: -73.930000,
		}
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "ST_X(location)", "ST_Y(location)",
			"address", "category", "creator_id", "media_urls", "tags",
			"rating", "review_count", "metadata", "parent_id",
			"created_at", "updated_at",
		}).
			AddRow(
				"place1", "Place in Bounds 1", "Description 1", -73.935242, 40.730610,
				"Address 1", "restaurant", "creator1", "{}", "{}",
				4.5, 10, "{}", nil, now, now,
			).
			AddRow(
				"place2", "Place in Bounds 2", "Description 2", -73.934242, 40.729610,
				"Address 2", "cafe", "creator2", "{}", "{}",
				4.0, 5, "{}", nil, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM places WHERE ST_Within`).
			WithArgs(
				bounds.MinLng, bounds.MinLat,
				bounds.MaxLng, bounds.MinLat,
				bounds.MaxLng, bounds.MaxLat,
				bounds.MinLng, bounds.MaxLat,
				bounds.MinLng, bounds.MinLat,
			).
			WillReturnRows(rows)

		places, err := repo.GetInBounds(ctx, bounds)
		assert.NoError(t, err)
		assert.Len(t, places, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetByCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("get places by category", func(t *testing.T) {
		category := "restaurant"
		limit := 10
		offset := 0
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "ST_X(location)", "ST_Y(location)",
			"address", "category", "creator_id", "media_urls", "tags",
			"rating", "review_count", "metadata", "parent_id",
			"created_at", "updated_at",
		}).
			AddRow(
				"place1", "Restaurant 1", "Description 1", -73.935242, 40.730610,
				"Address 1", category, "creator1", "{}", "{italian}",
				4.5, 10, "{}", nil, now, now,
			).
			AddRow(
				"place2", "Restaurant 2", "Description 2", -73.945242, 40.740610,
				"Address 2", category, "creator2", "{}", "{french}",
				4.8, 20, "{}", nil, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM places WHERE category = \$1`).
			WithArgs(category, limit, offset).
			WillReturnRows(rows)

		places, err := repo.GetByCategory(ctx, category, limit, offset)
		assert.NoError(t, err)
		assert.Len(t, places, 2)
		assert.Equal(t, category, places[0].Category)
		assert.Equal(t, category, places[1].Category)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetChildren(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("get child places", func(t *testing.T) {
		parentID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "ST_X(location)", "ST_Y(location)",
			"address", "category", "creator_id", "media_urls", "tags",
			"rating", "review_count", "metadata", "parent_id",
			"created_at", "updated_at",
		}).
			AddRow(
				"child1", "Child Place 1", "Description 1", -73.935242, 40.730610,
				"Address 1", "room", "creator1", "{}", "{}",
				4.5, 10, "{}", parentID, now, now,
			).
			AddRow(
				"child2", "Child Place 2", "Description 2", -73.935242, 40.730610,
				"Address 2", "floor", "creator2", "{}", "{}",
				4.0, 5, "{}", parentID, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM places WHERE parent_id = \$1`).
			WithArgs(parentID).
			WillReturnRows(rows)

		children, err := repo.GetChildren(ctx, parentID)
		assert.NoError(t, err)
		assert.Len(t, children, 2)
		assert.Equal(t, parentID, *children[0].ParentID)
		assert.Equal(t, parentID, *children[1].ParentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_UpdateRating(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful rating update", func(t *testing.T) {
		placeID := uuid.New().String()
		newRating := 4.7
		reviewCount := 25

		mock.ExpectExec(`UPDATE places SET rating = \$1, review_count = \$2`).
			WithArgs(newRating, reviewCount, sqlmock.AnyArg(), placeID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateRating(ctx, placeID, newRating, reviewCount)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}