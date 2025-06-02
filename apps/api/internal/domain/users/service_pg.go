package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/config"
	"github.com/Oferzz/newMap/apps/api/internal/utils"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// postgresService implements the service layer for PostgreSQL
type postgresService struct {
	repo       Repository
	jwtManager *utils.JWTManager
}

// NewPostgreSQLService creates a new PostgreSQL service
func NewPostgreSQLService(repo Repository, cfg *config.Config) *postgresService {
	return &postgresService{
		repo:       repo,
		jwtManager: utils.NewJWTManager(&cfg.JWT),
	}
}


// Register creates a new user account
func (s *postgresService) Register(ctx context.Context, username, email, password string) (*User, error) {
	// Check if email already exists
	existingUser, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	existingUser, err = s.repo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:         uuid.New().String(),
		Username:   username,
		Email:      email,
		Password:   hashedPassword,
		Role:       "user",
		IsVerified: false,
		Profile:    Profile{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user
func (s *postgresService) LoginOriginal(ctx context.Context, emailOrUsername, password string) (*User, error) {
	// Try to find user by email first
	user, err := s.repo.GetByEmail(ctx, emailOrUsername)
	if err != nil {
		// Try username
		user, err = s.repo.GetByUsername(ctx, emailOrUsername)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
	}

	// Check password
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// UpdateProfile updates user profile information
func (s *postgresService) UpdateProfile(ctx context.Context, userID string, updates ProfileUpdate) error {
	// Get existing user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update profile fields
	if updates.Name != "" {
		user.Profile.Name = updates.Name
	}
	if updates.Bio != "" {
		user.Profile.Bio = updates.Bio
	}
	if updates.Avatar != "" {
		user.Profile.Avatar = updates.Avatar
	}
	if updates.Location != "" {
		user.Profile.Location = updates.Location
	}
	if updates.Website != "" {
		user.Profile.Website = updates.Website
	}

	user.UpdatedAt = time.Now()

	// Save updates
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// AddFriend adds a friend relationship
func (s *postgresService) AddFriend(ctx context.Context, userID, friendID string) error {
	// Verify both users exist
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	_, err = s.repo.GetByID(ctx, friendID)
	if err != nil {
		return fmt.Errorf("friend not found: %w", err)
	}

	// Note: Friend check is now handled by database constraint in repository layer

	// Add friend relationship (bidirectional)
	if err := s.repo.AddFriend(ctx, userID, friendID); err != nil {
		return fmt.Errorf("failed to add friend: %w", err)
	}

	if err := s.repo.AddFriend(ctx, friendID, userID); err != nil {
		return fmt.Errorf("failed to add reverse friend relationship: %w", err)
	}

	return nil
}

// RemoveFriend removes a friend relationship
func (s *postgresService) RemoveFriend(ctx context.Context, userID, friendID string) error {
	// Verify both users exist
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	_, err = s.repo.GetByID(ctx, friendID)
	if err != nil {
		return fmt.Errorf("friend not found: %w", err)
	}

	// Note: Friend relationship check is now handled by database operations

	// Remove friend relationship (bidirectional)
	if err := s.repo.RemoveFriend(ctx, userID, friendID); err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	if err := s.repo.RemoveFriend(ctx, friendID, userID); err != nil {
		return fmt.Errorf("failed to remove reverse friend relationship: %w", err)
	}

	return nil
}

// SearchUsers searches for users by query
func (s *postgresService) SearchUsers(ctx context.Context, query string) ([]*User, error) {
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	users, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// GetFriends returns a user's friends
func (s *postgresService) GetFriendsOriginal(ctx context.Context, userID string) ([]*User, error) {
	friends, err := s.repo.GetFriends(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	return friends, nil
}

// GetByID returns a user by ID
func (s *postgresService) GetByID(ctx context.Context, userID string) (*User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Missing interface methods - stub implementations for now

func (s *postgresService) Create(ctx context.Context, input *CreateUserInput) (*User, error) {
	fmt.Printf("DEBUG: Service.Create called with input: %+v\n", input)

	// Check if email already exists
	existingUser, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		fmt.Printf("DEBUG: Error checking existing email: %v\n", err)
	}
	if existingUser != nil {
		fmt.Printf("DEBUG: Email already exists for user: %+v\n", existingUser)
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	existingUser, err = s.repo.GetByUsername(ctx, input.Username)
	if err != nil {
		fmt.Printf("DEBUG: Error checking existing username: %v\n", err)
	}
	if existingUser != nil {
		fmt.Printf("DEBUG: Username already exists for user: %+v\n", existingUser)
		return nil, errors.New("username already exists")
	}

	// Hash password
	fmt.Printf("DEBUG: Hashing password...\n")
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		fmt.Printf("DEBUG: Failed to hash password: %v\n", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	fmt.Printf("DEBUG: Password hashed successfully\n")

	// Create user
	user := &User{
		ID:                      uuid.New().String(),
		Username:                input.Username,
		Email:                   input.Email,
		PasswordHash:            hashedPassword,
		DisplayName:             input.DisplayName,
		Roles:                   pq.StringArray{"user"},
		ProfileVisibility:       "public",
		LocationSharing:         false,
		TripDefaultPrivacy:      "private",
		EmailNotifications:      true,
		PushNotifications:       true,
		SuggestionNotifications: true,
		TripInviteNotifications: true,
		IsVerified:              false,
		Status:                  "active",
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		LastActive:              time.Now(),
	}

	fmt.Printf("DEBUG: Created user object: %+v\n", user)
	fmt.Printf("DEBUG: Calling repository Create...\n")

	if err := s.repo.Create(ctx, user); err != nil {
		fmt.Printf("DEBUG: Repository Create failed: %v\n", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("DEBUG: Repository Create succeeded\n")
	return user, nil
}

func (s *postgresService) GetByEmail(ctx context.Context, email string) (*User, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
}

func (s *postgresService) Update(ctx context.Context, id string, input *UpdateUserInput) (*User, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
}

func (s *postgresService) Delete(ctx context.Context, id string) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) Login(ctx context.Context, input *LoginInput) (*LoginResponse, error) {
	fmt.Printf("DEBUG: Service.Login called with input: Email=%s, Password=%s\n", input.Email, input.Password)

	// Try to find user by email
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		fmt.Printf("DEBUG: Login - Failed to find user by email: %v\n", err)
		return nil, errors.New("invalid credentials")
	}

	fmt.Printf("DEBUG: Login - Found user: ID=%s, Email=%s, PasswordHash=%s\n", user.ID, user.Email, user.PasswordHash)

	// Check password
	fmt.Printf("DEBUG: Login - Checking password: input='%s' vs stored='%s'\n", input.Password, user.PasswordHash)
	passwordMatch := utils.CheckPassword(input.Password, user.PasswordHash)
	fmt.Printf("DEBUG: Login - Password match result: %t\n", passwordMatch)
	
	if !passwordMatch {
		fmt.Printf("DEBUG: Login - Password check failed\n")
		return nil, errors.New("invalid credentials")
	}

	fmt.Printf("DEBUG: Login - Password check succeeded, generating tokens\n")

	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		fmt.Printf("DEBUG: Login - Failed to generate tokens: %v\n", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	fmt.Printf("DEBUG: Login - Tokens generated successfully\n")

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}

func (s *postgresService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
}

func (s *postgresService) ChangePassword(ctx context.Context, userID string, input *ChangePasswordInput) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) ResetPassword(ctx context.Context, input *ResetPasswordInput) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) SendPasswordResetEmail(ctx context.Context, email string) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) ResendVerificationEmail(ctx context.Context, email string) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) Search(ctx context.Context, query string, limit, offset int) ([]*User, int64, error) {
	// TODO: Implement
	return nil, 0, errors.New("not implemented")
}

// Rename existing GetFriends to avoid conflict
func (s *postgresService) getFriendsInternal(ctx context.Context, userID string) ([]*User, error) {
	return s.GetFriendsOriginal(ctx, userID)
}

// GetFriends with pagination to match interface
func (s *postgresService) GetFriends(ctx context.Context, userID string, limit, offset int) ([]*User, int64, error) {
	// TODO: Implement - convert existing GetFriends method
	users, err := s.getFriendsInternal(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	
	// Apply pagination manually
	start := offset
	end := offset + limit
	if start > len(users) {
		return []*User{}, int64(len(users)), nil
	}
	if end > len(users) {
		end = len(users)
	}
	
	return users[start:end], int64(len(users)), nil
}

func (s *postgresService) SendFriendRequest(ctx context.Context, fromUserID, toUserID string) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) AcceptFriendRequest(ctx context.Context, userID, requestID string) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) RejectFriendRequest(ctx context.Context, userID, requestID string) error {
	// TODO: Implement
	return errors.New("not implemented")
}

func (s *postgresService) GetFriendRequests(ctx context.Context, userID string, incoming bool, limit, offset int) ([]*FriendRequest, int64, error) {
	// TODO: Implement
	return nil, 0, errors.New("not implemented")
}