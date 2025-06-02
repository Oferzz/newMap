package users

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// postgresRepository implements the repository interface for PostgreSQL
type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{
		db: db,
	}
}

// Create creates a new user
func (r *postgresRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, status,
			created_at, updated_at, last_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
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
		user.CreatedAt,
		user.UpdatedAt,
		user.LastActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

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
func (r *postgresRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	
	query := `
		SELECT 
			id, username, email, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, status,
			created_at, updated_at, last_active
		FROM users
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Bio,
		&user.Location,
		pq.Array(&user.Roles),
		&user.ProfileVisibility,
		&user.LocationSharing,
		&user.TripDefaultPrivacy,
		&user.EmailNotifications,
		&user.PushNotifications,
		&user.SuggestionNotifications,
		&user.TripInviteNotifications,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	
	query := `
		SELECT 
			id, username, email, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, status,
			created_at, updated_at, last_active
		FROM users
		WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Bio,
		&user.Location,
		pq.Array(&user.Roles),
		&user.ProfileVisibility,
		&user.LocationSharing,
		&user.TripDefaultPrivacy,
		&user.EmailNotifications,
		&user.PushNotifications,
		&user.SuggestionNotifications,
		&user.TripInviteNotifications,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *postgresRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	
	query := `
		SELECT 
			id, username, email, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, status,
			created_at, updated_at, last_active
		FROM users
		WHERE username = $1`

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Bio,
		&user.Location,
		pq.Array(&user.Roles),
		&user.ProfileVisibility,
		&user.LocationSharing,
		&user.TripDefaultPrivacy,
		&user.EmailNotifications,
		&user.PushNotifications,
		&user.SuggestionNotifications,
		&user.TripInviteNotifications,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *postgresRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, display_name = $5,
			avatar_url = $6, bio = $7, location = $8, roles = $9,
			profile_visibility = $10, location_sharing = $11, trip_default_privacy = $12,
			email_notifications = $13, push_notifications = $14, suggestion_notifications = $15,
			trip_invite_notifications = $16, status = $17, updated_at = $18, last_active = $19
		WHERE id = $1`

	user.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
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
		user.UpdatedAt,
		user.LastActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Search searches for users by query
func (r *postgresRepository) Search(ctx context.Context, query string) ([]*User, error) {
	var users []*User
	
	searchQuery := `
		SELECT 
			id, username, email, password_hash, display_name, avatar_url,
			bio, location, roles, profile_visibility, location_sharing,
			trip_default_privacy, email_notifications, push_notifications,
			suggestion_notifications, trip_invite_notifications, status,
			created_at, updated_at, last_active
		FROM users
		WHERE username ILIKE $1 OR email ILIKE $1 OR display_name ILIKE $1`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Bio,
			&user.Location,
			pq.Array(&user.Roles),
			&user.ProfileVisibility,
			&user.LocationSharing,
			&user.TripDefaultPrivacy,
			&user.EmailNotifications,
			&user.PushNotifications,
			&user.SuggestionNotifications,
			&user.TripInviteNotifications,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		
		users = append(users, &user)
	}

	return users, nil
}

// AddFriend adds a friend relationship using the user_friends table
func (r *postgresRepository) AddFriend(ctx context.Context, userID, friendID string) error {
	query := `
		INSERT INTO user_friends (user_id, friend_id, status)
		VALUES ($1, $2, 'accepted')
		ON CONFLICT (user_id, friend_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to add friend: %w", err)
	}

	return nil
}

// RemoveFriend removes a friend relationship
func (r *postgresRepository) RemoveFriend(ctx context.Context, userID, friendID string) error {
	query := `DELETE FROM user_friends WHERE user_id = $1 AND friend_id = $2`

	_, err := r.db.ExecContext(ctx, query, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	return nil
}

// GetFriends retrieves a user's friends using the user_friends table
func (r *postgresRepository) GetFriends(ctx context.Context, userID string) ([]*User, error) {
	var users []*User
	
	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.display_name, u.avatar_url,
			u.bio, u.location, u.roles, u.profile_visibility, u.location_sharing,
			u.trip_default_privacy, u.email_notifications, u.push_notifications,
			u.suggestion_notifications, u.trip_invite_notifications, u.status,
			u.created_at, u.updated_at, u.last_active
		FROM users u
		INNER JOIN user_friends uf ON u.id = uf.friend_id
		WHERE uf.user_id = $1 AND uf.status = 'accepted'`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Bio,
			&user.Location,
			pq.Array(&user.Roles),
			&user.ProfileVisibility,
			&user.LocationSharing,
			&user.TripDefaultPrivacy,
			&user.EmailNotifications,
			&user.PushNotifications,
			&user.SuggestionNotifications,
			&user.TripInviteNotifications,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan friend: %w", err)
		}
		
		users = append(users, &user)
	}

	return users, nil
}