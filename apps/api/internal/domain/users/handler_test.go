package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the Service interface
type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, input *CreateUserInput) (*User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, id string, input *UpdateUserInput) (*User, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) Login(ctx context.Context, input *LoginInput) (*LoginResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockService) ChangePassword(ctx context.Context, userID string, input *ChangePasswordInput) error {
	args := m.Called(ctx, userID, input)
	return args.Error(0)
}

func (m *MockService) ResetPassword(ctx context.Context, input *ResetPasswordInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *MockService) SendPasswordResetEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockService) VerifyEmail(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockService) ResendVerificationEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockService) Search(ctx context.Context, query string, limit, offset int) ([]*User, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*User), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetFriends(ctx context.Context, userID string, limit, offset int) ([]*User, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*User), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) SendFriendRequest(ctx context.Context, fromUserID, toUserID string) error {
	args := m.Called(ctx, fromUserID, toUserID)
	return args.Error(0)
}

func (m *MockService) AcceptFriendRequest(ctx context.Context, userID, requestID string) error {
	args := m.Called(ctx, userID, requestID)
	return args.Error(0)
}

func (m *MockService) RejectFriendRequest(ctx context.Context, userID, requestID string) error {
	args := m.Called(ctx, userID, requestID)
	return args.Error(0)
}

func (m *MockService) RemoveFriend(ctx context.Context, userID, friendID string) error {
	args := m.Called(ctx, userID, friendID)
	return args.Error(0)
}

func (m *MockService) GetFriendRequests(ctx context.Context, userID string, incoming bool, limit, offset int) ([]*FriendRequest, int64, error) {
	args := m.Called(ctx, userID, incoming, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*FriendRequest), args.Get(1).(int64), args.Error(2)
}

func TestHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		input        *CreateUserInput
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name: "successful registration",
			input: &CreateUserInput{
				Email:       "test@example.com",
				Username:    "testuser",
				Password:    "Password123!",
				DisplayName: "Test User",
			},
			mockSetup: func(ms *MockService) {
				user := &User{
					ID:          "user123",
					Email:       "test@example.com",
					Username:    "testuser",
					DisplayName: "Test User",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				ms.On("Create", mock.Anything, mock.Anything).Return(user, nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name: "validation error - invalid email",
			input: &CreateUserInput{
				Email:       "invalid-email",
				Username:    "testuser",
				Password:    "Password123!",
				DisplayName: "Test User",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
		{
			name: "duplicate email error",
			input: &CreateUserInput{
				Email:       "test@example.com",
				Username:    "testuser",
				Password:    "Password123!",
				DisplayName: "Test User",
			},
			mockSetup: func(ms *MockService) {
				ms.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("email already exists"))
			},
			expectedCode: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := gin.New()
			router.POST("/register", handler.Register)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody["success"], response["success"])

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		input        *LoginInput
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name: "successful login",
			input: &LoginInput{
				Email:    "test@example.com",
				Password: "Password123!",
			},
			mockSetup: func(ms *MockService) {
				loginResp := &LoginResponse{
					User: &User{
						ID:          "user123",
						Email:       "test@example.com",
						Username:    "testuser",
						DisplayName: "Test User",
					},
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresIn:    3600,
				}
				ms.On("Login", mock.Anything, mock.Anything).Return(loginResp, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name: "invalid credentials",
			input: &LoginInput{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(ms *MockService) {
				ms.On("Login", mock.Anything, mock.Anything).Return(nil, errors.New("invalid credentials"))
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := gin.New()
			router.POST("/login", handler.Login)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody["success"], response["success"])

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userID       string
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:   "successful get profile",
			userID: "user123",
			mockSetup: func(ms *MockService) {
				user := &User{
					ID:          "user123",
					Email:       "test@example.com",
					Username:    "testuser",
					DisplayName: "Test User",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				ms.On("GetByID", mock.Anything, "user123").Return(user, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			mockSetup: func(ms *MockService) {
				ms.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))
			},
			expectedCode: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := gin.New()
			router.GET("/profile", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetProfile(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody["success"], response["success"])

			mockService.AssertExpectations(t)
		})
	}
}