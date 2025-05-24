package users

import (
	"context"
	"testing"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/config"
	"github.com/Oferzz/newMap/apps/api/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockRepository struct {
	users map[string]*User
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users: make(map[string]*User),
	}
}

func (m *mockRepository) Create(ctx context.Context, user *User) error {
	if _, exists := m.users[user.Email]; exists {
		return ErrEmailExists
	}
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true
	user.IsVerified = false
	m.users[user.Email] = user
	m.users[user.ID.Hex()] = user
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	if user, exists := m.users[id.Hex()]; exists {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (m *mockRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockRepository) Update(ctx context.Context, id primitive.ObjectID, update *UpdateUserInput) error {
	if user, exists := m.users[id.Hex()]; exists {
		if update.FullName != nil {
			user.FullName = *update.FullName
		}
		if update.Bio != nil {
			user.Bio = *update.Bio
		}
		if update.Avatar != nil {
			user.Avatar = *update.Avatar
		}
		user.UpdatedAt = time.Now()
		return nil
	}
	return ErrUserNotFound
}

func (m *mockRepository) UpdatePassword(ctx context.Context, id primitive.ObjectID, passwordHash string) error {
	if user, exists := m.users[id.Hex()]; exists {
		user.PasswordHash = passwordHash
		user.UpdatedAt = time.Now()
		return nil
	}
	return ErrUserNotFound
}

func (m *mockRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	if user, exists := m.users[id.Hex()]; exists {
		now := time.Now()
		user.LastLoginAt = &now
		user.UpdatedAt = now
		return nil
	}
	return ErrUserNotFound
}

func (m *mockRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if user, exists := m.users[id.Hex()]; exists {
		delete(m.users, user.Email)
		delete(m.users, id.Hex())
		return nil
	}
	return ErrUserNotFound
}

func (m *mockRepository) List(ctx context.Context, filter bson.M, opts *options.FindOptions) ([]*User, error) {
	var users []*User
	for _, user := range m.users {
		if user.ID != primitive.NilObjectID {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *mockRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
	count := 0
	for _, user := range m.users {
		if user.ID != primitive.NilObjectID {
			count++
		}
	}
	return int64(count), nil
}

func TestUserService_Register(t *testing.T) {
	// Setup
	repo := newMockRepository()
	jwtManager := utils.NewJWTManager(&config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test",
	})
	service := NewService(repo, jwtManager)

	ctx := context.Background()

	// Test successful registration
	input := &CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		FullName: "Test User",
	}

	authResponse, err := service.Register(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if authResponse.User.Email != input.Email {
		t.Errorf("Expected email %s, got %s", input.Email, authResponse.User.Email)
	}

	if authResponse.AccessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if authResponse.RefreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	// Test duplicate email
	_, err = service.Register(ctx, input)
	if err != ErrEmailExists {
		t.Errorf("Expected ErrEmailExists, got %v", err)
	}
}

func TestUserService_Login(t *testing.T) {
	// Setup
	repo := newMockRepository()
	jwtManager := utils.NewJWTManager(&config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test",
	})
	service := NewService(repo, jwtManager)

	ctx := context.Background()

	// Register a user first
	registerInput := &CreateUserInput{
		Email:    "login@example.com",
		Username: "loginuser",
		Password: "password123",
		FullName: "Login User",
	}
	_, err := service.Register(ctx, registerInput)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test successful login
	loginInput := &LoginInput{
		Email:    "login@example.com",
		Password: "password123",
	}

	authResponse, err := service.Login(ctx, loginInput)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if authResponse.User.Email != loginInput.Email {
		t.Errorf("Expected email %s, got %s", loginInput.Email, authResponse.User.Email)
	}

	// Test invalid password
	loginInput.Password = "wrongpassword"
	_, err = service.Login(ctx, loginInput)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	// Test non-existent user
	loginInput.Email = "nonexistent@example.com"
	_, err = service.Login(ctx, loginInput)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	// Setup
	repo := newMockRepository()
	jwtManager := utils.NewJWTManager(&config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test",
	})
	service := NewService(repo, jwtManager)

	ctx := context.Background()

	// Register a user first
	registerInput := &CreateUserInput{
		Email:    "changepass@example.com",
		Username: "changepassuser",
		Password: "oldpassword123",
		FullName: "Change Pass User",
	}
	authResponse, err := service.Register(ctx, registerInput)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test successful password change
	changeInput := &ChangePasswordInput{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword123",
	}

	err = service.ChangePassword(ctx, authResponse.User.ID, changeInput)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify new password works
	loginInput := &LoginInput{
		Email:    "changepass@example.com",
		Password: "newpassword123",
	}
	_, err = service.Login(ctx, loginInput)
	if err != nil {
		t.Errorf("Expected successful login with new password, got %v", err)
	}

	// Test invalid current password
	changeInput.CurrentPassword = "wrongpassword"
	err = service.ChangePassword(ctx, authResponse.User.ID, changeInput)
	if err != ErrInvalidPassword {
		t.Errorf("Expected ErrInvalidPassword, got %v", err)
	}
}