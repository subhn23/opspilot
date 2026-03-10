package auth

import (
	"log"
	"opspilot/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// VerifyTOTP validates the 6-digit MFA code against the user's secret
func VerifyTOTP(passcode string, secret string) bool {
	return totp.Validate(passcode, secret)
}

// LogAction records a mutation event to the database
func LogAction(db *gorm.DB, userID uuid.UUID, action string, target string, ip string) {
	logEntry := models.AuditLog{
		UserID:    userID,
		Action:    action,
		Target:    target,
		IPAddress: ip,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&logEntry).Error; err != nil {
		log.Printf("Failed to record audit log: %v", err)
	}
}

// AuthMiddleware is a placeholder for session/JWT verification
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for Phase 1 Track 3:
		// Check for session cookie or JWT
		// If valid, c.Next(), else c.AbortWithStatus(401)
		c.Next()
	}
}
