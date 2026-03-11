package audit

import (
	"log"
	"opspilot/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LogAction records a mutation event to the database
func LogAction(db *gorm.DB, userID uuid.UUID, action string, target string, ip string, payload string) {
	logEntry := models.AuditLog{
		UserID:    userID,
		Action:    action,
		Target:    target,
		Payload:   payload,
		IPAddress: ip,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&logEntry).Error; err != nil {
		log.Printf("Failed to record audit log: %v", err)
	}
}
