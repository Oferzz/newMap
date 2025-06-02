package users

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
		user := &User{
			ID:        uuid.New().String(),
			Username:  "testuser",
			Email:     "test@example.com",
			PasswordHash:  "hashedpassword",
			Profile:   Profile{Name: "Test User", Bio: "Test bio"},
			Role:      "user",
			Roles:     pq.StringArray{"user"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(user.ID, user.CreatedAt, user.UpdatedAt)

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(
				user.ID,
				user.Username,
				user.Email,
				user.PasswordHash,
				user.DisplayName,
				user.AvatarURL,
				user.Bio,
				user.Location,
				sqlmock.AnyArg(), // roles array
				user.ProfileVisibility,
				user.LocationSharing,
				user.TripDefaultPrivacy,
				user.EmailNotifications,
				user.PushNotifications,
				user.SuggestionNotifications,
				user.TripInviteNotifications,
				user.Status,
				user.CreatedAt,
				user.UpdatedAt,
				user.LastActive,
			).
			WillReturnRows(rows)

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		user := &User{
			ID:        uuid.New().String(),
			Username:  "testuser",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, user)
		assert.Error(t, err)
	})
}

func TestPostgreSQLRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("user found", func(t *testing.T) {
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "display_name", "avatar_url",
			"bio", "location", "roles", "profile_visibility", "location_sharing",
			"trip_default_privacy", "email_notifications", "push_notifications",
			"suggestion_notifications", "trip_invite_notifications", "status",
			"created_at", "updated_at", "last_active",
		}).AddRow(
			userID, "testuser", "test@example.com", "hashedpassword", "Test User", "avatar.jpg",
			"Test bio", "New York", "{user}", "public", false,
			"private", true, true, true, true, "active",
			now, now, now,
		)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		userID := uuid.New().String()

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestPostgreSQLRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("user found", func(t *testing.T) {
		email := "test@example.com"
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password", "profile_name",
			"profile_bio", "profile_avatar", "profile_location",
			"profile_website", "role", "friends", "created_at", "updated_at",
		}).AddRow(
			userID, "testuser", email, "hashedpassword",
			"Test User", "Test bio", "avatar.jpg", "New York",
			"https://example.com", "user", "{}", now, now,
		)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetByEmail(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("user found", func(t *testing.T) {
		username := "testuser"
		userID := uuid.New().String()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password", "profile_name",
			"profile_bio", "profile_avatar", "profile_location",
			"profile_website", "role", "friends", "created_at", "updated_at",
		}).AddRow(
			userID, username, "test@example.com", "hashedpassword",
			"Test User", "Test bio", "avatar.jpg", "New York",
			"https://example.com", "user", "{}", now, now,
		)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE username = \$1`).
			WithArgs(username).
			WillReturnRows(rows)

		user, err := repo.GetByUsername(ctx, username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)
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
		user := &User{
			ID:        uuid.New().String(),
			Username:  "updateduser",
			Email:     "updated@example.com",
			PasswordHash:  "newhashedpassword",
			Profile:   Profile{Name: "Updated User", Bio: "Updated bio"},
			Role:      "user",
			Roles:     pq.StringArray{"user"},
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec(`UPDATE users SET`).
			WithArgs(
				user.Username,
				user.Email,
				user.Password,
				user.Profile.Name,
				user.Profile.Bio,
				user.Profile.Avatar,
				user.Profile.Location,
				user.Profile.Website,
				user.Role,
				sqlmock.AnyArg(), // friends array
				sqlmock.AnyArg(), // updated_at
				user.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, user)
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
		userID := uuid.New().String()

		mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, userID)
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
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password", "profile_name",
			"profile_bio", "profile_avatar", "profile_location",
			"profile_website", "role", "friends", "created_at", "updated_at",
		}).
			AddRow(
				"1", "testuser1", "test1@example.com", "hash1",
				"Test User 1", "Bio 1", "", "", "", "user", "{}", now, now,
			).
			AddRow(
				"2", "testuser2", "test2@example.com", "hash2",
				"Test User 2", "Bio 2", "", "", "", "user", "{}", now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE`).
			WithArgs("%"+query+"%", "%"+query+"%", "%"+query+"%").
			WillReturnRows(rows)

		users, err := repo.Search(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_AddFriend(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful add friend", func(t *testing.T) {
		userID := uuid.New().String()
		friendID := uuid.New().String()

		mock.ExpectExec(`UPDATE users SET friends = array_append`).
			WithArgs(friendID, sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.AddFriend(ctx, userID, friendID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_RemoveFriend(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("successful remove friend", func(t *testing.T) {
		userID := uuid.New().String()
		friendID := uuid.New().String()

		mock.ExpectExec(`UPDATE users SET friends = array_remove`).
			WithArgs(friendID, sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveFriend(ctx, userID, friendID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgreSQLRepository_GetFriends(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &postgresRepository{db: db}
	ctx := context.Background()

	t.Run("get friends successfully", func(t *testing.T) {
		userID := uuid.New().String()
		now := time.Now()

		// First query to get friend IDs
		friendRows := sqlmock.NewRows([]string{"friends"}).
			AddRow("{friend1,friend2}")

		mock.ExpectQuery(`SELECT friends FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnRows(friendRows)

		// Second query to get friend details
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password", "profile_name",
			"profile_bio", "profile_avatar", "profile_location",
			"profile_website", "role", "friends", "created_at", "updated_at",
		}).
			AddRow(
				"friend1", "friend1user", "friend1@example.com", "hash1",
				"Friend 1", "Bio 1", "", "", "", "user", "{}", now, now,
			).
			AddRow(
				"friend2", "friend2user", "friend2@example.com", "hash2",
				"Friend 2", "Bio 2", "", "", "", "user", "{}", now, now,
			)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id = ANY`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		friends, err := repo.GetFriends(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, friends, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}