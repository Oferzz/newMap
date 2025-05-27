package places

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

func (m *MockService) Create(ctx context.Context, userID string, input *CreatePlaceInput) (*Place, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Place), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, userID, placeID string) (*Place, error) {
	args := m.Called(ctx, userID, placeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Place), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, userID, placeID string, input *UpdatePlaceInput) (*Place, error) {
	args := m.Called(ctx, userID, placeID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Place), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, userID, placeID string) error {
	args := m.Called(ctx, userID, placeID)
	return args.Error(0)
}

func (m *MockService) GetUserPlaces(ctx context.Context, userID string, limit, offset int) ([]*Place, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Place), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetChildPlaces(ctx context.Context, userID, parentID string) ([]*Place, error) {
	args := m.Called(ctx, userID, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Place), args.Error(1)
}

func (m *MockService) Search(ctx context.Context, userID string, input *SearchPlacesInput) ([]*Place, int64, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Place), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetNearby(ctx context.Context, userID string, input *NearbyPlacesInput) ([]*Place, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Place), args.Error(1)
}

func (m *MockService) AddCollaborator(ctx context.Context, userID, placeID, collaboratorID, role string) error {
	args := m.Called(ctx, userID, placeID, collaboratorID, role)
	return args.Error(0)
}

func (m *MockService) RemoveCollaborator(ctx context.Context, userID, placeID, collaboratorID string) error {
	args := m.Called(ctx, userID, placeID, collaboratorID)
	return args.Error(0)
}

func (m *MockService) UpdateCollaboratorRole(ctx context.Context, userID, placeID, collaboratorID, role string) error {
	args := m.Called(ctx, userID, placeID, collaboratorID, role)
	return args.Error(0)
}

func (m *MockService) List(ctx context.Context, userID string, filter *PlaceFilter, limit, offset int) ([]*Place, int64, error) {
	args := m.Called(ctx, userID, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Place), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetTripPlaces(ctx context.Context, userID, tripID string) ([]*Place, error) {
	args := m.Called(ctx, userID, tripID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Place), args.Error(1)
}

func (m *MockService) AddToTrip(ctx context.Context, userID, placeID, tripID string) error {
	args := m.Called(ctx, userID, placeID, tripID)
	return args.Error(0)
}

func (m *MockService) RemoveFromTrip(ctx context.Context, userID, placeID, tripID string) error {
	args := m.Called(ctx, userID, placeID, tripID)
	return args.Error(0)
}

func (m *MockService) UpdateVisitStatus(ctx context.Context, userID, placeID string, visited bool, visitDate *time.Time) error {
	args := m.Called(ctx, userID, placeID, visited, visitDate)
	return args.Error(0)
}

func (m *MockService) AddImages(ctx context.Context, userID, placeID string, images []string) error {
	args := m.Called(ctx, userID, placeID, images)
	return args.Error(0)
}

func (m *MockService) RemoveImage(ctx context.Context, userID, placeID string, imageURL string) error {
	args := m.Called(ctx, userID, placeID, imageURL)
	return args.Error(0)
}

func (m *MockService) UpdateRating(ctx context.Context, userID, placeID string, rating float32) error {
	args := m.Called(ctx, userID, placeID, rating)
	return args.Error(0)
}

func (m *MockService) AddNote(ctx context.Context, userID, placeID, note string) error {
	args := m.Called(ctx, userID, placeID, note)
	return args.Error(0)
}

func TestHandler_CreatePlace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userID       string
		input        *CreatePlaceInput
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:   "successful place creation",
			userID: "user123",
			input: &CreatePlaceInput{
				Name:        "Eiffel Tower",
				Description: "Iconic tower in Paris",
				Type:        "poi",
				Location: &LocationInput{
					Latitude:  48.8584,
					Longitude: 2.2945,
				},
				City:     "Paris",
				Country:  "France",
				Category: []string{"attraction", "landmark"},
				Tags:     []string{"tourist", "historic"},
			},
			mockSetup: func(ms *MockService) {
				place := &Place{
					ID:          "place123",
					Name:        "Eiffel Tower",
					Description: "Iconic tower in Paris",
					Type:        "poi",
					Location: &GeoPoint{
						Type:        "Point",
						Coordinates: []float64{2.2945, 48.8584},
					},
					City:      "Paris",
					Country:   "France",
					Category:  []string{"attraction", "landmark"},
					Tags:      []string{"tourist", "historic"},
					CreatedBy: "user123",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				ms.On("Create", mock.Anything, "user123", mock.Anything).Return(place, nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:   "validation error - missing name",
			userID: "user123",
			input: &CreatePlaceInput{
				Description: "A place without name",
				Type:        "poi",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
		{
			name:   "service error",
			userID: "user123",
			input: &CreatePlaceInput{
				Name: "Failed Place",
				Type: "poi",
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
			router.POST("/places", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.CreatePlace(c)
			})

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/places", bytes.NewBuffer(body))
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

func TestHandler_GetPlace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userID       string
		placeID      string
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:    "successful get place",
			userID:  "user123",
			placeID: "place123",
			mockSetup: func(ms *MockService) {
				place := &Place{
					ID:          "place123",
					Name:        "Eiffel Tower",
					Description: "Iconic tower in Paris",
					Type:        "poi",
					City:        "Paris",
					Country:     "France",
					CreatedBy:   "user123",
					Privacy:     "public",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				ms.On("GetByID", mock.Anything, "user123", "place123").Return(place, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:    "place not found",
			userID:  "user123",
			placeID: "nonexistent",
			mockSetup: func(ms *MockService) {
				ms.On("GetByID", mock.Anything, "user123", "nonexistent").Return(nil, ErrPlaceNotFound)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
		{
			name:    "unauthorized access",
			userID:  "user456",
			placeID: "place123",
			mockSetup: func(ms *MockService) {
				ms.On("GetByID", mock.Anything, "user456", "place123").Return(nil, ErrUnauthorized)
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
			router.GET("/places/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetPlace(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/places/"+tt.placeID, nil)
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

func TestHandler_SearchPlaces(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userID       string
		query        string
		mockSetup    func(*MockService)
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:   "successful search",
			userID: "user123",
			query:  "paris",
			mockSetup: func(ms *MockService) {
				places := []*Place{
					{
						ID:          "place1",
						Name:        "Eiffel Tower",
						Description: "In Paris",
						City:        "Paris",
					},
					{
						ID:          "place2",
						Name:        "Louvre Museum",
						Description: "Also in Paris",
						City:        "Paris",
					},
				}
				ms.On("Search", mock.Anything, "user123", mock.AnythingOfType("*places.SearchPlacesInput")).Return(places, int64(2), nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
			},
		},
		{
			name:   "empty search results",
			userID: "user123",
			query:  "nonexistent",
			mockSetup: func(ms *MockService) {
				ms.On("Search", mock.Anything, "user123", mock.AnythingOfType("*places.SearchPlacesInput")).Return([]*Place{}, int64(0), nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
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
			router.GET("/places/search", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.SearchPlaces(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/places/search?q="+tt.query+"&limit=20&offset=0", nil)
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