package auth

import (
	"net/http"
	"opspilot/internal/models"

	"github.com/gin-gonic/gin"
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
		// 1. Get User from Context (set by Auth middleware)
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}
		user := val.(models.User)

		// 2. Check if User is Master Admin (Bypass checks)
		var role models.Role
		pc.DB.Preload("Permissions").First(&role, user.RoleID)
		if role.Name == "Master Admin" {
			c.Next()
			return
		}

		// 3. Check for specific permission slug
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
