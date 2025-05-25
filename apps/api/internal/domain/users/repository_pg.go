package users

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

// Create creates a new user
func (r *PostgresRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (
			email, username, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id, created_at, updated_at, last_active`

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.Bio,
		user.Location,
		pq.Array(user.Roles),
		user.ProfileVisibility,
		user.LocationSharing,
		user.TripDefaultPrivacy,
		user.EmailNotifications,
		user.PushNotifications,
		user.SuggestionNotifications,
		user.TripInviteNotifications,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.LastActive)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return fmt.Errorf("email already exists")
				}
				if pqErr.Constraint == "users_username_key" {
					return fmt.Errorf("username already exists")
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := `
		SELECT 
			id, email, username, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, 
			created_at, updated_at, last_active, status
		FROM users
		WHERE id = $1 AND status = 'active'`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
		SELECT 
			id, email, username, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, 
			created_at, updated_at, last_active, status
		FROM users
		WHERE email = $1 AND status = 'active'`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	query := `
		SELECT 
			id, email, username, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, 
			created_at, updated_at, last_active, status
		FROM users
		WHERE username = $1 AND status = 'active'`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// Update updates a user
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
		if field == "roles" {
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
		UPDATE users
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'
	`, setClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastActive updates the user's last active timestamp
func (r *PostgresRepository) UpdateLastActive(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET last_active = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}

	return nil
}

// Delete soft deletes a user
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET status = 'deleted', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'active'`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves users with pagination
func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	var users []*User
	query := `
		SELECT 
			id, email, username, display_name, avatar_url,
			bio, location, roles, profile_visibility,
			created_at, updated_at, last_active, status
		FROM users
		WHERE status = 'active'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Search searches for users by username or display name
func (r *PostgresRepository) Search(ctx context.Context, query string, limit, offset int) ([]*User, error) {
	var users []*User
	searchQuery := `
		SELECT 
			id, email, username, display_name, avatar_url,
			bio, location, roles, profile_visibility,
			created_at, updated_at, last_active, status
		FROM users
		WHERE status = 'active' 
		AND (
			username ILIKE $1 
			OR display_name ILIKE $1
			OR email ILIKE $1
		)
		ORDER BY 
			CASE 
				WHEN username ILIKE $1 THEN 1
				WHEN display_name ILIKE $1 THEN 2
				ELSE 3
			END,
			created_at DESC
		LIMIT $2 OFFSET $3`

	searchPattern := "%" + query + "%"
	err := r.db.SelectContext(ctx, &users, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// GetFriends retrieves a user's friends
func (r *PostgresRepository) GetFriends(ctx context.Context, userID string) ([]*User, error) {
	var users []*User
	query := `
		SELECT DISTINCT
			u.id, u.email, u.username, u.display_name, u.avatar_url,
			u.bio, u.location, u.roles, u.profile_visibility,
			u.created_at, u.updated_at, u.last_active, u.status
		FROM users u
		JOIN user_friends uf ON (
			(uf.user_id = $1 AND uf.friend_id = u.id) OR
			(uf.friend_id = $1 AND uf.user_id = u.id)
		)
		WHERE uf.status = 'accepted' AND u.status = 'active'
		ORDER BY u.display_name`

	err := r.db.SelectContext(ctx, &users, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	return users, nil
}

// AddFriend sends a friend request
func (r *PostgresRepository) AddFriend(ctx context.Context, userID, friendID string) error {
	// Check if friendship already exists
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM user_friends 
			WHERE (user_id = $1 AND friend_id = $2) 
			   OR (user_id = $2 AND friend_id = $1)
		)`
	
	err := r.db.GetContext(ctx, &exists, checkQuery, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to check friendship: %w", err)
	}
	
	if exists {
		return fmt.Errorf("friendship already exists")
	}

	// Create friend request
	query := `
		INSERT INTO user_friends (user_id, friend_id, status)
		VALUES ($1, $2, 'pending')`

	_, err = r.db.ExecContext(ctx, query, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to add friend: %w", err)
	}

	return nil
}

// UpdateFriendship updates the status of a friendship
func (r *PostgresRepository) UpdateFriendship(ctx context.Context, userID, friendID, status string) error {
	query := `
		UPDATE user_friends
		SET status = $3, responded_at = CURRENT_TIMESTAMP
		WHERE friend_id = $1 AND user_id = $2 AND status = 'pending'`

	result, err := r.db.ExecContext(ctx, query, userID, friendID, status)
	if err != nil {
		return fmt.Errorf("failed to update friendship: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("friend request not found")
	}

	return nil
}

// RemoveFriend removes a friendship
func (r *PostgresRepository) RemoveFriend(ctx context.Context, userID, friendID string) error {
	query := `
		DELETE FROM user_friends
		WHERE (user_id = $1 AND friend_id = $2) 
		   OR (user_id = $2 AND friend_id = $1)`

	_, err := r.db.ExecContext(ctx, query, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	return nil
}

// GetPendingFriendRequests retrieves pending friend requests for a user
func (r *PostgresRepository) GetPendingFriendRequests(ctx context.Context, userID string) ([]*UserFriend, error) {
	var requests []*UserFriend
	query := `
		SELECT 
			uf.id, uf.user_id, uf.friend_id, uf.status, 
			uf.requested_at, uf.responded_at
		FROM user_friends uf
		WHERE uf.friend_id = $1 AND uf.status = 'pending'
		ORDER BY uf.requested_at DESC`

	err := r.db.SelectContext(ctx, &requests, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending friend requests: %w", err)
	}

	return requests, nil
}

// CountByStatus counts users by status
func (r *PostgresRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE status = $1`

	err := r.db.GetContext(ctx, &count, query, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}