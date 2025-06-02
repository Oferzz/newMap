package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/config"
	"github.com/Oferzz/newMap/apps/api/internal/utils"
	"github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) Search(ctx context.Context, query string) ([]*User, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*User), args.Error(1)
}

func (m *MockRepository) AddFriend(ctx context.Context, userID, friendID string) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockRepository) RemoveFriend(ctx context.Context, userID, friendID string) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockRepository) GetFriends(ctx context.Context, userID string) ([]*User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*User), args.Error(1)
}

func TestServicePG_Register(t *testing.T) {
	mockRepo := new(MockRepository)
	mockConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			AccessExpiry: 15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
	service := NewPostgreSQLService(mockRepo, mockConfig)
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		username := "testuser"
		email := "test@example.com"
		password := "password123"

		// Mock that user doesn't exist
		mockRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("not found")).Once()
		mockRepo.On("GetByUsername", ctx, username).Return(nil, errors.New("not found")).Once()
		
		// Mock successful creation
		mockRepo.On("Create", ctx, mock.MatchedBy(func(user *User) bool {
			return user.Username == username && user.Email == email
		})).Return(nil).Once()

		user, err := service.Register(ctx, username, email, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, email, user.Email)
		assert.NotEqual(t, password, user.Password) // Should be hashed
		mockRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		username := "newuser"
		email := "existing@example.com"
		password := "password123"

		existingUser := &User{
			ID:    uuid.New().String(),
			Email: email,
		}

		mockRepo.On("GetByEmail", ctx, email).Return(existingUser, nil).Once()

		user, err := service.Register(ctx, username, email, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		username := "existinguser"
		email := "new@example.com"
		password := "password123"

		existingUser := &User{
			ID:       uuid.New().String(),
			Username: username,
		}

		mockRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("not found")).Once()
		mockRepo.On("GetByUsername", ctx, username).Return(existingUser, nil).Once()

		user, err := service.Register(ctx, username, email, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username already exists")
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestServicePG_Login(t *testing.T) {
	mockRepo := new(MockRepository)
	mockConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			AccessExpiry: 15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
	service := NewPostgreSQLService(mockRepo, mockConfig)
	ctx := context.Background()

	t.Run("successful login with email", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password)

		user := &User{
			ID:       uuid.New().String(),
			Email:    email,
			Username: "testuser",
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil).Once()

		result, err := service.Login(ctx, &LoginInput{Email: email, Password: password})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.User.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		email := "test@example.com"
		correctPassword := "password123"
		wrongPassword := "wrongpassword"
		hashedPassword, _ := utils.HashPassword(correctPassword)

		user := &User{
			ID:       uuid.New().String(),
			Email:    email,
			PasswordHash: hashedPassword,
		}

		mockRepo.On("GetByEmail", ctx, email).Return(user, nil).Once()

		result, err := service.Login(ctx, &LoginInput{Email: email, Password: wrongPassword})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		email := "nonexistent@example.com"
		password := "password123"

		mockRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("not found")).Once()
		mockRepo.On("GetByUsername", ctx, email).Return(nil, errors.New("not found")).Once()

		result, err := service.Login(ctx, &LoginInput{Email: email, Password: password})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestServicePG_UpdateProfile(t *testing.T) {
	mockRepo := new(MockRepository)
	mockConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			AccessExpiry: 15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
	service := NewPostgreSQLService(mockRepo, mockConfig)
	ctx := context.Background()

	t.Run("successful profile update", func(t *testing.T) {
		userID := uuid.New().String()
		existingUser := &User{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
			Profile: Profile{
				Name: "Old Name",
				Bio:  "Old bio",
			},
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		updates := ProfileUpdate{
			Name:     "New Name",
			Bio:      "New bio",
			Location: "New York",
		}

		mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(user *User) bool {
			return user.Profile.Name == updates.Name &&
				user.Profile.Bio == updates.Bio &&
				user.Profile.Location == updates.Location
		})).Return(nil).Once()

		err := service.UpdateProfile(ctx, userID, updates)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := uuid.New().String()
		updates := ProfileUpdate{Name: "New Name"}

		mockRepo.On("GetByID", ctx, userID).Return(nil, errors.New("not found")).Once()

		err := service.UpdateProfile(ctx, userID, updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestServicePG_AddFriend(t *testing.T) {
	mockRepo := new(MockRepository)
	mockConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			AccessExpiry: 15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
	service := NewPostgreSQLService(mockRepo, mockConfig)
	ctx := context.Background()

	t.Run("successful add friend", func(t *testing.T) {
		userID := uuid.New().String()
		friendID := uuid.New().String()

		user := &User{
			ID:      userID,
			Roles: pq.StringArray{"user"},
		}

		friend := &User{
			ID:      friendID,
			Roles: pq.StringArray{"user"},
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil).Once()
		mockRepo.On("GetByID", ctx, friendID).Return(friend, nil).Once()
		mockRepo.On("AddFriend", ctx, userID, friendID).Return(nil).Once()
		mockRepo.On("AddFriend", ctx, friendID, userID).Return(nil).Once()

		err := service.AddFriend(ctx, userID, friendID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("friend not found", func(t *testing.T) {
		userID := uuid.New().String()
		friendID := uuid.New().String()

		user := &User{
			ID:      userID,
			Roles: pq.StringArray{"user"},
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil).Once()
		mockRepo.On("GetByID", ctx, friendID).Return(nil, errors.New("not found")).Once()

		err := service.AddFriend(ctx, userID, friendID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "friend not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("already friends", func(t *testing.T) {
		userID := uuid.New().String()
		friendID := uuid.New().String()

		user := &User{
			ID:      userID,
			Roles: pq.StringArray{"user"}, // Already friends
		}

		friend := &User{
			ID:      friendID,
			Roles: pq.StringArray{"user"},
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil).Once()
		mockRepo.On("GetByID", ctx, friendID).Return(friend, nil).Once()

		err := service.AddFriend(ctx, userID, friendID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already friends")
		mockRepo.AssertExpectations(t)
	})
}

func TestServicePG_RemoveFriend(t *testing.T) {
	mockRepo := new(MockRepository)
	mockConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			AccessExpiry: 15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
	service := NewPostgreSQLService(mockRepo, mockConfig)
	ctx := context.Background()

	t.Run("successful remove friend", func(t *testing.T) {
		userID := uuid.New().String()
		friendID := uuid.New().String()

		user := &User{
			ID:      userID,
			Roles: pq.StringArray{"user"},
		}

		friend := &User{
			ID:      friendID,
			Roles: pq.StringArray{"user"},
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil).Once()
		mockRepo.On("GetByID", ctx, friendID).Return(friend, nil).Once()
		mockRepo.On("RemoveFriend", ctx, userID, friendID).Return(nil).Once()
		mockRepo.On("RemoveFriend", ctx, friendID, userID).Return(nil).Once()

		err := service.RemoveFriend(ctx, userID, friendID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not friends", func(t *testing.T) {
		userID := uuid.New().String()
		friendID := uuid.New().String()

		user := &User{
			ID:      userID,
			Roles: pq.StringArray{"user"}, // Not friends
		}

		friend := &User{
			ID:      friendID,
			Roles: pq.StringArray{"user"},
		}

		mockRepo.On("GetByID", ctx, userID).Return(user, nil).Once()
		mockRepo.On("GetByID", ctx, friendID).Return(friend, nil).Once()

		err := service.RemoveFriend(ctx, userID, friendID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not friends")
		mockRepo.AssertExpectations(t)
	})
}

func TestServicePG_SearchUsers(t *testing.T) {
	mockRepo := new(MockRepository)
	mockConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			AccessExpiry: 15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
	service := NewPostgreSQLService(mockRepo, mockConfig)
	ctx := context.Background()

	t.Run("successful search", func(t *testing.T) {
		query := "test"
		users := []*User{
			{
				ID:       uuid.New().String(),
				Username: "testuser1",
				Email:    "test1@example.com",
			},
			{
				ID:       uuid.New().String(),
				Username: "testuser2",
				Email:    "test2@example.com",
			},
		}

		mockRepo.On("Search", ctx, query).Return(users, nil).Once()

		result, err := service.SearchUsers(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty query", func(t *testing.T) {
		query := ""

		result, err := service.SearchUsers(ctx, query)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "search query cannot be empty")
		assert.Nil(t, result)
	})
}

func TestServicePG_GetFriends(t *testing.T) {
	mockRepo := new(MockRepository)
	mockConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
			AccessExpiry: 15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}
	service := NewPostgreSQLService(mockRepo, mockConfig)
	ctx := context.Background()

	t.Run("successful get friends", func(t *testing.T) {
		userID := uuid.New().String()
		friends := []*User{
			{
				ID:       uuid.New().String(),
				Username: "friend1",
				Email:    "friend1@example.com",
			},
			{
				ID:       uuid.New().String(),
				Username: "friend2",
				Email:    "friend2@example.com",
			},
		}

		mockRepo.On("GetFriends", ctx, userID).Return(friends, nil).Once()

		result, total, err := service.GetFriends(ctx, userID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no friends", func(t *testing.T) {
		userID := uuid.New().String()
		friends := []*User{}

		mockRepo.On("GetFriends", ctx, userID).Return(friends, nil).Once()

		result, total, err := service.GetFriends(ctx, userID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}