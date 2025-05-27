package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite is the main test suite for integration tests
type IntegrationTestSuite struct {
	suite.Suite
	router       *gin.Engine
	accessToken  string
	refreshToken string
	userID       string
	tripID       string
	placeID      string
}

// SetupSuite runs once before all tests
func (suite *IntegrationTestSuite) SetupSuite() {
	// Set test environment
	os.Setenv("APP_ENV", "test")
	os.Setenv("JWT_SECRET", "test-secret-key-for-integration-tests")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/trip_platform_test?sslmode=disable")
	os.Setenv("REDIS_URL", "redis://localhost:6379/1")
	
	gin.SetMode(gin.TestMode)
}

// TearDownSuite runs once after all tests
func (suite *IntegrationTestSuite) TearDownSuite() {
	// Clean up test data if needed
}

// Test user registration flow
func (suite *IntegrationTestSuite) TestUserRegistration() {
	input := map[string]interface{}{
		"email":        "integration@test.com",
		"username":     "integrationtest",
		"password":     "TestPassword123!",
		"display_name": "Integration Test User",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotEmpty(suite.T(), response["data"])
}

// Test user login flow
func (suite *IntegrationTestSuite) TestUserLogin() {
	input := map[string]interface{}{
		"email":    "integration@test.com",
		"password": "TestPassword123!",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	suite.accessToken = data["access_token"].(string)
	suite.refreshToken = data["refresh_token"].(string)
	
	user := data["user"].(map[string]interface{})
	suite.userID = user["id"].(string)
}

// Test trip creation flow
func (suite *IntegrationTestSuite) TestTripCreation() {
	input := map[string]interface{}{
		"title":       "Integration Test Trip",
		"description": "A trip created during integration testing",
		"privacy":     "private",
		"tags":        []string{"test", "integration"},
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	suite.tripID = data["id"].(string)
}

// Test place creation flow
func (suite *IntegrationTestSuite) TestPlaceCreation() {
	input := map[string]interface{}{
		"name":        "Test Location",
		"description": "A test place for integration testing",
		"type":        "poi",
		"location": map[string]float64{
			"latitude":  40.7128,
			"longitude": -74.0060,
		},
		"city":     "New York",
		"country":  "USA",
		"category": []string{"test"},
		"tags":     []string{"integration", "test"},
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/places", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	suite.placeID = data["id"].(string)
}

// Test adding place to trip
func (suite *IntegrationTestSuite) TestAddPlaceToTrip() {
	input := map[string]interface{}{
		"place_id":       suite.placeID,
		"order_position": 1,
		"notes":          "First stop on our trip",
	}

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+suite.tripID+"/waypoints", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
}

// Test trip collaboration
func (suite *IntegrationTestSuite) TestTripCollaboration() {
	// First create another user
	registerInput := map[string]interface{}{
		"email":        "collaborator@test.com",
		"username":     "collaborator",
		"password":     "TestPassword123!",
		"display_name": "Collaborator User",
	}

	body, _ := json.Marshal(registerInput)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusCreated, rec.Code)

	var registerResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &registerResp)
	collaboratorID := registerResp["data"].(map[string]interface{})["id"].(string)

	// Now invite the collaborator
	inviteInput := map[string]interface{}{
		"user_id":   collaboratorID,
		"role":      "editor",
		"can_edit":  true,
		"can_invite": false,
	}

	body, _ = json.Marshal(inviteInput)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+suite.tripID+"/collaborators", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec = httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
}

// Test search functionality
func (suite *IntegrationTestSuite) TestSearchFunctionality() {
	// Search for trips
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/search?q=integration", nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	// Search for places
	req = httptest.NewRequest(http.MethodGet, "/api/v1/places/search?q=test&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec = httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
}

// Test nearby places
func (suite *IntegrationTestSuite) TestNearbyPlaces() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/places/nearby?lat=40.7128&lng=-74.0060&radius=5000", nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)

	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
}

// Test cleanup - delete created resources
func (suite *IntegrationTestSuite) TestCleanup() {
	// Delete trip (which should cascade to waypoints)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+suite.tripID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec := httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)

	// Delete place
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/places/"+suite.placeID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)
	rec = httptest.NewRecorder()

	suite.router.ServeHTTP(rec, req)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
}

// TestIntegrationTestSuite runs the test suite
func TestIntegrationTestSuite(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	suite.Run(t, new(IntegrationTestSuite))
}