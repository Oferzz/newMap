package utils

import (
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword123"

	// Test password hashing
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Expected hash to be generated")
	}

	if hash == password {
		t.Error("Hash should not be the same as the password")
	}

	// Test that same password generates different hashes (due to salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password second time: %v", err)
	}

	if hash == hash2 {
		t.Error("Same password should generate different hashes due to salt")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "mysecretpassword123"
	wrongPassword := "wrongpassword123"

	// Generate hash
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	if !CheckPassword(password, hash) {
		t.Error("Expected correct password to match")
	}

	// Test wrong password
	if CheckPassword(wrongPassword, hash) {
		t.Error("Expected wrong password to not match")
	}

	// Test empty password
	if CheckPassword("", hash) {
		t.Error("Expected empty password to not match")
	}

	// Test invalid hash
	if CheckPassword(password, "invalid-hash") {
		t.Error("Expected invalid hash to not match")
	}
}

func TestPasswordHashingPerformance(t *testing.T) {
	password := "testpassword123"

	// This test ensures hashing doesn't take too long
	// bcrypt with cost 12 should complete within reasonable time
	start := time.Now()
	_, err := HashPassword(password)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Hashing should complete within 1 second
	if duration > 1*time.Second {
		t.Errorf("Password hashing took too long: %v", duration)
	}
}