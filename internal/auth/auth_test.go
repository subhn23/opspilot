package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
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

func TestTOTP(t *testing.T) {
	email := "test@example.com"

	// 1. Generate Secret
	secret, err := GenerateTOTPSecret(email)
	if err != nil {
		t.Fatalf("Failed to generate TOTP secret: %v", err)
	}
	if secret == "" {
		t.Fatal("Generated TOTP secret is empty")
	}

	// 2. Verify Valid Passcode
	passcode, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("Failed to generate passcode: %v", err)
	}

	valid := VerifyTOTP(passcode, secret)
	if !valid {
		t.Errorf("Expected passcode %s to be valid for secret %s", passcode, secret)
	}

	// 3. Verify Invalid Passcode
	invalid := VerifyTOTP("000000", secret)
	if invalid {
		t.Error("Expected invalid passcode to be rejected")
	}
}
