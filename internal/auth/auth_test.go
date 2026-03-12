package auth

import (
	"net/http"
	"net/http/httptest"
	"opspilot/internal/audit"
	"opspilot/internal/config"
	"opspilot/internal/models"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	userID := uuid.New()
	roleID := uuid.New()

	// Initialize DB for audit logging and user validation in middleware
	config.DB, _ = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	config.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.AuditLog{})

	// Create a user in the DB for the success case
	user := models.User{ID: userID, Email: "test@test.com", RoleID: roleID}
	if err := config.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 1. Success Case: Valid Token
	token, _ := GenerateToken(userID, roleID, secret, 1*time.Hour)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	// Mock environment for JWT secret
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	r.Use(AuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		if uid != userID {
			t.Errorf("Expected user_id %v in context, got %v", userID, uid)
		}
		c.Status(200)
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, c.Request)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 2. Failure Case: Missing Token
	w = httptest.NewRecorder()
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != 401 {
		t.Errorf("Expected status 401 for missing token, got %d", w.Code)
	}

	// 3. Failure Case: Invalid Token
	w = httptest.NewRecorder()
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, c.Request)

	if w.Code != 401 {
		t.Errorf("Expected status 401 for invalid token, got %d", w.Code)
	}

	// 4. Failure Case: User does not exist in DB
	// Clear the DB and re-migrate to ensure it's empty
	config.DB, _ = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	config.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.AuditLog{})

	nonExistentUserID := uuid.New()
	token, _ = GenerateToken(nonExistentUserID, roleID, secret, 1*time.Hour)
	w = httptest.NewRecorder()
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, c.Request)

	if w.Code != 401 {
		t.Errorf("Expected status 401 for non-existent user, got %d", w.Code)
	}
}

func TestLogAction(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.AuditLog{})

	userID := uuid.New()
	action := "TEST_ACTION"
	target := "test-target"
	ip := "127.0.0.1"
	payload := `{"key": "value"}`

	audit.LogAction(db, userID, action, target, ip, payload)

	var logEntry models.AuditLog
	err := db.First(&logEntry).Error
	if err != nil {
		t.Fatalf("Failed to find log entry: %v", err)
	}

	if logEntry.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, logEntry.UserID)
	}
	if logEntry.Action != action {
		t.Errorf("Expected Action %s, got %s", action, logEntry.Action)
	}
	if logEntry.Payload != payload {
		t.Errorf("Expected Payload %s, got %s", payload, logEntry.Payload)
	}
}
