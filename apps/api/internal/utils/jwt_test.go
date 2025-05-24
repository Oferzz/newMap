package utils

import (
	"testing"
	"time"

	"github.com/Oferzz/newMap/apps/api/internal/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestJWTManager_GenerateAndValidateTokens(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:        "test-secret-key",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test-issuer",
	}

	jwtManager := NewJWTManager(cfg)
	userID := primitive.NewObjectID()
	email := "test@example.com"

	// Test token generation
	accessToken, refreshToken, err := jwtManager.GenerateTokenPair(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	if accessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if refreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	// Test access token validation
	claims, err := jwtManager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID.Hex() {
		t.Errorf("Expected user ID %s, got %s", userID.Hex(), claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	// Test refresh token validation
	refreshClaims, err := jwtManager.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.UserID != userID.Hex() {
		t.Errorf("Expected user ID %s, got %s", userID.Hex(), refreshClaims.UserID)
	}

	// Test invalid token
	_, err = jwtManager.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestJWTManager_RefreshAccessToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:        "test-secret-key",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test-issuer",
	}

	jwtManager := NewJWTManager(cfg)
	userID := primitive.NewObjectID()
	email := "test@example.com"

	// Generate initial tokens
	_, refreshToken, err := jwtManager.GenerateTokenPair(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Test refreshing access token
	newAccessToken, err := jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	if newAccessToken == "" {
		t.Error("Expected new access token to be generated")
	}

	// Validate new access token
	claims, err := jwtManager.ValidateToken(newAccessToken)
	if err != nil {
		t.Fatalf("Failed to validate new access token: %v", err)
	}

	if claims.UserID != userID.Hex() {
		t.Errorf("Expected user ID %s, got %s", userID.Hex(), claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:        "test-secret-key",
		AccessExpiry:  -1 * time.Hour, // Already expired
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test-issuer",
	}

	jwtManager := NewJWTManager(cfg)
	userID := primitive.NewObjectID()
	email := "test@example.com"

	// Generate expired token
	accessToken, _, err := jwtManager.GenerateTokenPair(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Try to validate expired token
	_, err = jwtManager.ValidateToken(accessToken)
	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}