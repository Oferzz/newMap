package users

import (
	"context"
	"errors"

	"github.com/Oferzz/newMap/apps/api/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidPassword    = errors.New("invalid password")
)

type Service interface {
	Register(ctx context.Context, input *CreateUserInput) (*AuthResponse, error)
	Login(ctx context.Context, input *LoginInput) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	Update(ctx context.Context, id primitive.ObjectID, input *UpdateUserInput) (*User, error)
	ChangePassword(ctx context.Context, id primitive.ObjectID, input *ChangePasswordInput) error
	Delete(ctx context.Context, id primitive.ObjectID) error
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
	
	// Create user
	user := &User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: passwordHash,
		FullName:     input.FullName,
	}
	
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	
	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	
	// Update last login
	_ = s.repo.UpdateLastLogin(ctx, user.ID)
	
	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
	}, nil
}

func (s *service) Login(ctx context.Context, input *LoginInput) (*AuthResponse, error) {
	// Get user by email
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	
	// Check password
	if !utils.CheckPassword(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	
	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}
	
	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	
	// Update last login
	_ = s.repo.UpdateLastLogin(ctx, user.ID)
	
	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}
	
	// Get user
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, err
	}
	
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is disabled")
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
		ExpiresIn:    900, // 15 minutes in seconds
	}, nil
}

func (s *service) GetByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) Update(ctx context.Context, id primitive.ObjectID, input *UpdateUserInput) (*User, error) {
	if err := s.repo.Update(ctx, id, input); err != nil {
		return nil, err
	}
	
	return s.repo.GetByID(ctx, id)
}

func (s *service) ChangePassword(ctx context.Context, id primitive.ObjectID, input *ChangePasswordInput) error {
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
	return s.repo.UpdatePassword(ctx, id, newPasswordHash)
}

func (s *service) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.Delete(ctx, id)
}