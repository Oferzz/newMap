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
			id, username, email, password_hash, profile_name, profile_bio,
			profile_avatar, profile_location, profile_website, role, friends,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Profile.Name,
		user.Profile.Bio,
		user.Profile.Avatar,
		user.Profile.Location,
		user.Profile.Website,
		user.Role,
		pq.Array(user.Friends),
		user.CreatedAt,
		user.UpdatedAt,
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
	var friends pq.StringArray
	
	query := `
		SELECT 
			id, username, email, password_hash, profile_name, profile_bio,
			profile_avatar, profile_location, profile_website, role, friends,
			created_at, updated_at
		FROM users
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Profile.Name,
		&user.Profile.Bio,
		&user.Profile.Avatar,
		&user.Profile.Location,
		&user.Profile.Website,
		&user.Role,
		&friends,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Friends = []string(friends)
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	var friends pq.StringArray
	
	query := `
		SELECT 
			id, username, email, password_hash, profile_name, profile_bio,
			profile_avatar, profile_location, profile_website, role, friends,
			created_at, updated_at
		FROM users
		WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Profile.Name,
		&user.Profile.Bio,
		&user.Profile.Avatar,
		&user.Profile.Location,
		&user.Profile.Website,
		&user.Role,
		&friends,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Friends = []string(friends)
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *postgresRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	var friends pq.StringArray
	
	query := `
		SELECT 
			id, username, email, password_hash, profile_name, profile_bio,
			profile_avatar, profile_location, profile_website, role, friends,
			created_at, updated_at
		FROM users
		WHERE username = $1`

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Profile.Name,
		&user.Profile.Bio,
		&user.Profile.Avatar,
		&user.Profile.Location,
		&user.Profile.Website,
		&user.Role,
		&friends,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	user.Friends = []string(friends)
	return &user, nil
}

// Update updates a user
func (r *postgresRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, profile_name = $5,
			profile_bio = $6, profile_avatar = $7, profile_location = $8,
			profile_website = $9, role = $10, friends = $11, updated_at = $12
		WHERE id = $1`

	user.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.Profile.Name,
		user.Profile.Bio,
		user.Profile.Avatar,
		user.Profile.Location,
		user.Profile.Website,
		user.Role,
		pq.Array(user.Friends),
		user.UpdatedAt,
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
			id, username, email, password_hash, profile_name, profile_bio,
			profile_avatar, profile_location, profile_website, role, friends,
			created_at, updated_at
		FROM users
		WHERE username ILIKE $1 OR email ILIKE $1 OR profile_name ILIKE $1`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		var friends pq.StringArray
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Profile.Name,
			&user.Profile.Bio,
			&user.Profile.Avatar,
			&user.Profile.Location,
			&user.Profile.Website,
			&user.Role,
			&friends,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		
		user.Friends = []string(friends)
		users = append(users, &user)
	}

	return users, nil
}

// AddFriend adds a friend relationship
func (r *postgresRepository) AddFriend(ctx context.Context, userID, friendID string) error {
	query := `
		UPDATE users
		SET friends = array_append(friends, $1), updated_at = $2
		WHERE id = $3 AND NOT ($1 = ANY(friends))`

	_, err := r.db.ExecContext(ctx, query, friendID, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to add friend: %w", err)
	}

	return nil
}

// RemoveFriend removes a friend relationship
func (r *postgresRepository) RemoveFriend(ctx context.Context, userID, friendID string) error {
	query := `
		UPDATE users
		SET friends = array_remove(friends, $1), updated_at = $2
		WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, friendID, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	return nil
}

// GetFriends retrieves a user's friends
func (r *postgresRepository) GetFriends(ctx context.Context, userID string) ([]*User, error) {
	// First get the friend IDs
	var friendIDs pq.StringArray
	query := `SELECT friends FROM users WHERE id = $1`
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&friendIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend IDs: %w", err)
	}

	if len(friendIDs) == 0 {
		return []*User{}, nil
	}

	// Then get the friend details
	var users []*User
	friendQuery := `
		SELECT 
			id, username, email, password_hash, profile_name, profile_bio,
			profile_avatar, profile_location, profile_website, role, friends,
			created_at, updated_at
		FROM users
		WHERE id = ANY($1)`

	rows, err := r.db.QueryContext(ctx, friendQuery, pq.Array(friendIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		var friends pq.StringArray
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Profile.Name,
			&user.Profile.Bio,
			&user.Profile.Avatar,
			&user.Profile.Location,
			&user.Profile.Website,
			&user.Role,
			&friends,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan friend: %w", err)
		}
		
		user.Friends = []string(friends)
		users = append(users, &user)
	}

	return users, nil
}