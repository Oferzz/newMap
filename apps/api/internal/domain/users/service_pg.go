package users

import (
	"context"
	"errors"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserInactive      = errors.New("user account is disabled")
)

type Service interface {
	Register(ctx context.Context, input *CreateUserInput) (*AuthResponse, error)
	Login(ctx context.Context, input *LoginInput) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, id string, input *UpdateUserInput) (*User, error)
	ChangePassword(ctx context.Context, id string, input *ChangePasswordInput) error
	Delete(ctx context.Context, id string) error
	UpdateLastActive(ctx context.Context, id string) error
}

type service struct {
	repo       Repository
	jwtManager *utils.JWTManager
}

func NewService(repo Repository, jwtManager *utils.JWTManager) Service {
	return &service{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *service) Register(ctx context.Context, input *CreateUserInput) (*AuthResponse, error) {
	// Hash password
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	
	// Create user with defaults
	user := &User{
		Email:                   input.Email,
		Username:                input.Username,
		PasswordHash:            passwordHash,
		DisplayName:             input.DisplayName,
		Roles:                   []string{"user"},
		ProfileVisibility:       "public",
		LocationSharing:         false,
		TripDefaultPrivacy:      "private",
		EmailNotifications:      true,
		PushNotifications:       true,
		SuggestionNotifications: true,
		TripInviteNotifications: true,
		Status:                  "active",
	}
	
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	
	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	
	// Update last active
	_ = s.repo.UpdateLastActive(ctx, user.ID)
	
	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

func (s *service) Login(ctx context.Context, input *LoginInput) (*AuthResponse, error) {
	// Get user by email
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// Check password
	if !utils.CheckPassword(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	
	// Check if user is active
	if user.Status != "active" {
		return nil, ErrUserInactive
	}
	
	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	
	// Update last active
	_ = s.repo.UpdateLastActive(ctx, user.ID)
	
	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}
	
	// Get user
	user, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	
	// Check if user is active
	if user.Status != "active" {
		return nil, ErrUserInactive
	}
	
	// Generate new access token
	accessToken, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}
	
	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) Update(ctx context.Context, id string, input *UpdateUserInput) (*User, error) {
	updates := make(map[string]interface{})
	
	if input.DisplayName != nil {
		updates["display_name"] = *input.DisplayName
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}
	if input.AvatarURL != nil {
		updates["avatar_url"] = *input.AvatarURL
	}
	if input.Location != nil {
		updates["location"] = *input.Location
	}
	if input.ProfileVisibility != nil {
		updates["profile_visibility"] = *input.ProfileVisibility
	}
	if input.LocationSharing != nil {
		updates["location_sharing"] = *input.LocationSharing
	}
	if input.TripDefaultPrivacy != nil {
		updates["trip_default_privacy"] = *input.TripDefaultPrivacy
	}
	if input.EmailNotifications != nil {
		updates["email_notifications"] = *input.EmailNotifications
	}
	if input.PushNotifications != nil {
		updates["push_notifications"] = *input.PushNotifications
	}
	if input.SuggestionNotifications != nil {
		updates["suggestion_notifications"] = *input.SuggestionNotifications
	}
	if input.TripInviteNotifications != nil {
		updates["trip_invite_notifications"] = *input.TripInviteNotifications
	}
	
	if len(updates) == 0 {
		return s.repo.GetByID(ctx, id)
	}
	
	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}
	
	return s.repo.GetByID(ctx, id)
}

func (s *service) ChangePassword(ctx context.Context, id string, input *ChangePasswordInput) error {
	// Get user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Verify current password
	if !utils.CheckPassword(input.CurrentPassword, user.PasswordHash) {
		return ErrInvalidPassword
	}
	
	// Hash new password
	newPasswordHash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}
	
	// Update password
	updates := map[string]interface{}{
		"password_hash": newPasswordHash,
	}
	
	return s.repo.Update(ctx, id, updates)
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) UpdateLastActive(ctx context.Context, id string) error {
	return s.repo.UpdateLastActive(ctx, id)
}