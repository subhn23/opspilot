package auth

import (
	"net/http"
	"opspilot/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionChecker checks if a user has a specific permission
type PermissionChecker struct {
	DB *gorm.DB
}

func NewPermissionChecker(db *gorm.DB) *PermissionChecker {
	return &PermissionChecker{DB: db}
}

// RequirePermission is a middleware that enforces RBAC
func (pc *PermissionChecker) RequirePermission(slug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get RoleID from Context (set by Auth middleware)
		val, exists := c.Get("role_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}
		roleID := val.(uuid.UUID)

		// 2. Fetch Role with Permissions
		var role models.Role
		if err := pc.DB.Preload("Permissions").First(&role, "id = ?", roleID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role permissions"})
			return
		}

		// 3. Check if User is Master Admin (Bypass checks)
		if role.Name == "Master Admin" {
			c.Next()
			return
		}

		// 4. Check for specific permission slug
		hasPermission := false
		for _, p := range role.Permissions {
			if p.Slug == slug {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Permission denied: " + slug})
			return
		}

		c.Next()
	}
}
