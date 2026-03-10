package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTProvider(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	secret := "test-secret"

	// 1. Generate Token
	token, err := GenerateToken(userID, roleID, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// 2. Validate Token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, claims.UserID)
	}
	if claims.RoleID != roleID {
		t.Errorf("Expected RoleID %v, got %v", roleID, claims.RoleID)
	}

	// 3. Test Invalid Token
	_, err = ValidateToken("invalid-token", secret)
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}

	// 4. Test Expired Token (short duration)
	token, _ = GenerateToken(userID, roleID, secret, -1*time.Hour)
	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}
