package trips

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
		trip := &Trip{
			ID:          uuid.New().String(),
			Title:       "Test Trip",
			Description: "Test Description",
			CreatorID:   uuid.New().String(),
			Collaborators: []Collaborator{
				{UserID: "user1", Role: "editor", AddedAt: time.Now()},
			},
			StartDate: time.Now(),
			EndDate:   time.Now().Add(7 * 24 * time.Hour),
			Places:    []string{"place1", "place2"},
			Waypoints: []Waypoint{
				{PlaceID: "place1", Order: 1, VisitDate: time.Now()},
			},
			IsPublic:  true,
			ViewCount: 0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(trip.ID, trip.CreatedAt, trip.UpdatedAt)

		mock.ExpectQuery(`INSERT INTO trips`).
			WithArgs(
				trip.ID,
				trip.Title,
				trip.Description,
				trip.CreatorID,
				sqlmock.AnyArg(), // collaborators jsonb
				trip.StartDate,
				trip.EndDate,
				sqlmock.AnyArg(), // places array
				sqlmock.AnyArg(), // waypoints jsonb
				trip.IsPublic,
				trip.ViewCount,
				trip.CreatedAt,
				trip.UpdatedAt,
			).
			WillReturnRows(rows)

		err := repo.Create(ctx, trip)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		trip := &Trip{
			ID:        uuid.New().String(),
			Title:     "Test Trip",
			CreatorID: uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectQuery(`INSERT INTO trips`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, trip)
		assert.Error(t, err)
	})
}

func TestPostgreSQLRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("trip found", func(t *testing.T) {
		tripID := uuid.New().String()
		creatorID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "collaborators",
			"start_date", "end_date", "places", "waypoints", "is_public",
			"view_count", "created_at", "updated_at",
		}).AddRow(
			tripID, "Test Trip", "Test Description", creatorID,
			`[{"user_id": "user1", "role": "editor", "added_at": "2024-01-01T00:00:00Z"}]`,
			now, now.Add(7*24*time.Hour), "{place1,place2}",
			`[{"place_id": "place1", "order": 1, "visit_date": "2024-01-01T00:00:00Z"}]`,
			true, 10, now, now,
		)

		mock.ExpectQuery(`SELECT (.+) FROM trips WHERE id = \$1`).
			WithArgs(tripID).
			WillReturnRows(rows)

		trip, err := repo.GetByID(ctx, tripID)
		assert.NoError(t, err)
		assert.NotNil(t, trip)
		assert.Equal(t, tripID, trip.ID)
		assert.Equal(t, "Test Trip", trip.Title)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("trip not found", func(t *testing.T) {
		tripID := uuid.New().String()

		mock.ExpectQuery(`SELECT (.+) FROM trips WHERE id = \$1`).
			WithArgs(tripID).
			WillReturnError(sql.ErrNoRows)

		trip, err := repo.GetByID(ctx, tripID)
		assert.Error(t, err)
		assert.Nil(t, trip)
	})
}

func TestPostgreSQLRepository_GetByCreator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("trips found", func(t *testing.T) {
		creatorID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "collaborators",
			"start_date", "end_date", "places", "waypoints", "is_public",
			"view_count", "created_at", "updated_at",
		}).
			AddRow(
				"trip1", "Trip 1", "Description 1", creatorID,
				"[]", now, now.Add(7*24*time.Hour), "{}", "[]",
				true, 5, now, now,
			).
			AddRow(
				"trip2", "Trip 2", "Description 2", creatorID,
				"[]", now, now.Add(14*24*time.Hour), "{}", "[]",
				false, 0, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM trips WHERE creator_id = \$1`).
			WithArgs(creatorID).
			WillReturnRows(rows)

		trips, err := repo.GetByCreator(ctx, creatorID)
		assert.NoError(t, err)
		assert.Len(t, trips, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetByCollaborator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("trips found", func(t *testing.T) {
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "collaborators",
			"start_date", "end_date", "places", "waypoints", "is_public",
			"view_count", "created_at", "updated_at",
		}).
			AddRow(
				"trip1", "Trip 1", "Description 1", "creator1",
				`[{"user_id": "`+userID+`", "role": "editor", "added_at": "2024-01-01T00:00:00Z"}]`,
				now, now.Add(7*24*time.Hour), "{}", "[]",
				true, 5, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM trips WHERE collaborators @> \$1`).
			WithArgs(sqlmock.AnyArg()). // JSONB contains query
			WillReturnRows(rows)

		trips, err := repo.GetByCollaborator(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, trips, 1)
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
		trip := &Trip{
			ID:          uuid.New().String(),
			Title:       "Updated Trip",
			Description: "Updated Description",
			CreatorID:   uuid.New().String(),
			Collaborators: []Collaborator{
				{UserID: "user1", Role: "editor", AddedAt: time.Now()},
			},
			StartDate: time.Now(),
			EndDate:   time.Now().Add(7 * 24 * time.Hour),
			Places:    []string{"place1", "place2", "place3"},
			IsPublic:  true,
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec(`UPDATE trips SET`).
			WithArgs(
				trip.Title,
				trip.Description,
				sqlmock.AnyArg(), // collaborators jsonb
				trip.StartDate,
				trip.EndDate,
				sqlmock.AnyArg(), // places array
				sqlmock.AnyArg(), // waypoints jsonb
				trip.IsPublic,
				sqlmock.AnyArg(), // updated_at
				trip.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, trip)
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
		tripID := uuid.New().String()

		mock.ExpectExec(`DELETE FROM trips WHERE id = \$1`).
			WithArgs(tripID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, tripID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_AddCollaborator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful add collaborator", func(t *testing.T) {
		tripID := uuid.New().String()
		collaborator := Collaborator{
			UserID:  "user123",
			Role:    "editor",
			AddedAt: time.Now(),
		}

		mock.ExpectExec(`UPDATE trips SET collaborators = collaborators \|\| \$1`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), tripID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.AddCollaborator(ctx, tripID, collaborator)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_RemoveCollaborator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful remove collaborator", func(t *testing.T) {
		tripID := uuid.New().String()
		userID := "user123"

		// First query to get current collaborators
		rows := sqlmock.NewRows([]string{"collaborators"}).
			AddRow(`[{"user_id": "user123", "role": "editor"}, {"user_id": "user456", "role": "viewer"}]`)

		mock.ExpectQuery(`SELECT collaborators FROM trips WHERE id = \$1`).
			WithArgs(tripID).
			WillReturnRows(rows)

		// Update query to remove collaborator
		mock.ExpectExec(`UPDATE trips SET collaborators = \$1`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), tripID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveCollaborator(ctx, tripID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_IncrementViewCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful increment", func(t *testing.T) {
		tripID := uuid.New().String()

		mock.ExpectExec(`UPDATE trips SET view_count = view_count \+ 1`).
			WithArgs(sqlmock.AnyArg(), tripID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.IncrementViewCount(ctx, tripID)
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
		query := "test"
		filters := SearchFilters{
			IsPublic: true,
			Limit:    10,
			Offset:   0,
		}
		now := time.Now()

		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM trips WHERE`).
			WithArgs("%"+query+"%", "%"+query+"%", filters.IsPublic).
			WillReturnRows(countRows)

		// Search query
		rows := sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "collaborators",
			"start_date", "end_date", "places", "waypoints", "is_public",
			"view_count", "created_at", "updated_at",
		}).
			AddRow(
				"trip1", "Test Trip 1", "Description 1", "creator1",
				"[]", now, now.Add(7*24*time.Hour), "{}", "[]",
				true, 5, now, now,
			).
			AddRow(
				"trip2", "Test Trip 2", "Description 2", "creator2",
				"[]", now, now.Add(14*24*time.Hour), "{}", "[]",
				true, 10, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM trips WHERE`).
			WithArgs("%"+query+"%", "%"+query+"%", filters.IsPublic, filters.Limit, filters.Offset).
			WillReturnRows(rows)

		result, err := repo.Search(ctx, query, filters)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Trips, 2)
		assert.Equal(t, int64(2), result.Total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetPopular(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("get popular trips", func(t *testing.T) {
		limit := 10
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "title", "description", "creator_id", "collaborators",
			"start_date", "end_date", "places", "waypoints", "is_public",
			"view_count", "created_at", "updated_at",
		}).
			AddRow(
				"trip1", "Popular Trip 1", "Description 1", "creator1",
				"[]", now, now.Add(7*24*time.Hour), "{}", "[]",
				true, 100, now, now,
			).
			AddRow(
				"trip2", "Popular Trip 2", "Description 2", "creator2",
				"[]", now, now.Add(14*24*time.Hour), "{}", "[]",
				true, 50, now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM trips WHERE is_public = true ORDER BY view_count DESC`).
			WithArgs(limit).
			WillReturnRows(rows)

		trips, err := repo.GetPopular(ctx, limit)
		assert.NoError(t, err)
		assert.Len(t, trips, 2)
		assert.Equal(t, 100, trips[0].ViewCount)
		assert.Equal(t, 50, trips[1].ViewCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}