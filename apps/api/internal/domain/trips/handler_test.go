package trips

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

func (m *MockService) Create(ctx context.Context, userID string, input *CreateTripInput) (*Trip, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Trip), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, userID, tripID string) (*Trip, error) {
	args := m.Called(ctx, userID, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Trip), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, userID, tripID string, input *UpdateTripInput) (*Trip, error) {
	args := m.Called(ctx, userID, tripID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Trip), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, userID, tripID string) error {
	args := m.Called(ctx, userID, tripID)
	return args.Error(0)
}

func (m *MockService) List(ctx context.Context, userID string, filter *TripFilter, limit, offset int) ([]*Trip, int64, error) {
	args := m.Called(ctx, userID, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Trip), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetUserTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Trip), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetSharedTrips(ctx context.Context, userID string, limit, offset int) ([]*Trip, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Trip), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) Search(ctx context.Context, userID string, query string, limit, offset int) ([]*Trip, int64, error) {
	args := m.Called(ctx, userID, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Trip), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) AddCollaborator(ctx context.Context, userID, tripID, collaboratorID, role string) error {
	args := m.Called(ctx, userID, tripID, collaboratorID, role)
	return args.Error(0)
}

func (m *MockService) RemoveCollaborator(ctx context.Context, userID, tripID, collaboratorID string) error {
	args := m.Called(ctx, userID, tripID, collaboratorID)
	return args.Error(0)
}

func (m *MockService) UpdateCollaboratorRole(ctx context.Context, userID, tripID, collaboratorID, role string) error {
	args := m.Called(ctx, userID, tripID, collaboratorID, role)
	return args.Error(0)
}

func (m *MockService) InviteCollaborator(ctx context.Context, userID, tripID string, input *InviteCollaboratorInput) error {
	args := m.Called(ctx, userID, tripID, input)
	return args.Error(0)
}

func (m *MockService) AddWaypoint(ctx context.Context, userID, tripID string, input *AddWaypointInput) (*Waypoint, error) {
	args := m.Called(ctx, userID, tripID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Waypoint), args.Error(1)
}

func (m *MockService) UpdateWaypoint(ctx context.Context, userID, tripID, waypointID string, input *UpdateWaypointInput) (*Waypoint, error) {
	args := m.Called(ctx, userID, tripID, waypointID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Waypoint), args.Error(1)
}

func (m *MockService) RemoveWaypoint(ctx context.Context, userID, tripID, waypointID string) error {
	args := m.Called(ctx, userID, tripID, waypointID)
	return args.Error(0)
}

func (m *MockService) ReorderWaypoints(ctx context.Context, userID, tripID string, waypointIDs []string) error {
	args := m.Called(ctx, userID, tripID, waypointIDs)
	return args.Error(0)
}

func (m *MockService) GetTripStats(ctx context.Context, userID, tripID string) (*TripStats, error) {
	args := m.Called(ctx, userID, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TripStats), args.Error(1)
}

func (m *MockService) ExportTrip(ctx context.Context, userID, tripID, format string) ([]byte, error) {
	args := m.Called(ctx, userID, tripID, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockService) CloneTrip(ctx context.Context, userID, tripID string) (*Trip, error) {
	args := m.Called(ctx, userID, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Trip), args.Error(1)
}

func TestHandler_CreateTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userID       string
		input        *CreateTripInput
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:   "successful trip creation",
			userID: "user123",
			input: &CreateTripInput{
				Title:       "Summer Vacation 2024",
				Description: "A wonderful trip to Europe",
				Privacy:     "private",
				Tags:        []string{"vacation", "europe"},
			},
			mockSetup: func(ms *MockService) {
				trip := &Trip{
					ID:          "trip123",
					Title:       "Summer Vacation 2024",
					Description: "A wonderful trip to Europe",
					OwnerID:     "user123",
					Privacy:     "private",
					Status:      "planning",
					Tags:        []string{"vacation", "europe"},
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				ms.On("Create", mock.Anything, "user123", mock.Anything).Return(trip, nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:   "validation error - missing title",
			userID: "user123",
			input: &CreateTripInput{
				Description: "A trip without title",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
		{
			name:   "service error",
			userID: "user123",
			input: &CreateTripInput{
				Title:       "Failed Trip",
				Description: "This will fail",
			},
			mockSetup: func(ms *MockService) {
				ms.On("Create", mock.Anything, "user123", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedCode: http.StatusInternalServerError,
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
			router.POST("/trips", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.CreateTrip(c)
			})

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewBuffer(body))
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

func TestHandler_GetTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userID       string
		tripID       string
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:   "successful get trip",
			userID: "user123",
			tripID: "trip123",
			mockSetup: func(ms *MockService) {
				trip := &Trip{
					ID:          "trip123",
					Title:       "Summer Vacation 2024",
					Description: "A wonderful trip to Europe",
					OwnerID:     "user123",
					Privacy:     "private",
					Status:      "planning",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				ms.On("GetByID", mock.Anything, "user123", "trip123").Return(trip, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:   "trip not found",
			userID: "user123",
			tripID: "nonexistent",
			mockSetup: func(ms *MockService) {
				ms.On("GetByID", mock.Anything, "user123", "nonexistent").Return(nil, ErrTripNotFound)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
		{
			name:   "unauthorized access",
			userID: "user456",
			tripID: "trip123",
			mockSetup: func(ms *MockService) {
				ms.On("GetByID", mock.Anything, "user456", "trip123").Return(nil, ErrUnauthorized)
			},
			expectedCode: http.StatusForbidden,
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
			router.GET("/trips/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetTrip(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/trips/"+tt.tripID, nil)
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

func TestHandler_UpdateTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userID       string
		tripID       string
		input        *UpdateTripInput
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:   "successful update",
			userID: "user123",
			tripID: "trip123",
			input: &UpdateTripInput{
				Title:       stringPtr("Updated Summer Vacation"),
				Description: stringPtr("Updated description"),
				Status:      stringPtr("active"),
			},
			mockSetup: func(ms *MockService) {
				trip := &Trip{
					ID:          "trip123",
					Title:       "Updated Summer Vacation",
					Description: "Updated description",
					OwnerID:     "user123",
					Status:      "active",
					UpdatedAt:   time.Now(),
				}
				ms.On("Update", mock.Anything, "user123", "trip123", mock.Anything).Return(trip, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:   "unauthorized update",
			userID: "user456",
			tripID: "trip123",
			input: &UpdateTripInput{
				Title: stringPtr("Unauthorized Update"),
			},
			mockSetup: func(ms *MockService) {
				ms.On("Update", mock.Anything, "user456", "trip123", mock.Anything).Return(nil, ErrUnauthorized)
			},
			expectedCode: http.StatusForbidden,
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
			router.PUT("/trips/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.UpdateTrip(c)
			})

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPut, "/trips/"+tt.tripID, bytes.NewBuffer(body))
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

// Helper function
func stringPtr(s string) *string {
	return &s
}