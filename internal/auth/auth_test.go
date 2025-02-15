package auth

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	username := "testuser"
	token, err := GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated token is empty")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	username := "testuser"
	token, err := GenerateToken(username)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	var tokenValidator DefaultValidator
	claims, err := tokenValidator.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	if claims.ExpiresAt < time.Now().Unix() {
		t.Error("Token expiration time is in the past")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.here"
	var tokenValidator DefaultValidator
	_, err := tokenValidator.ValidateToken(invalidToken)
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}
