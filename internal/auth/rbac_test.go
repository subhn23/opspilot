package auth

import (
	"net/http"
	"net/http/httptest"
	"opspilot/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Permission{}, &models.Role{}, &models.User{})

	// Setup Permissions and Roles
	perm := models.Permission{Slug: "test:write", Module: "Test"}
	db.Create(&perm)

	adminRole := models.Role{ID: uuid.New(), Name: "Master Admin"}
	db.Create(&adminRole)

	userRole := models.Role{ID: uuid.New(), Name: "User", Permissions: []models.Permission{perm}}
	db.Create(&userRole)

	otherRole := models.Role{ID: uuid.New(), Name: "Other"}
	db.Create(&otherRole)

	checker := NewPermissionChecker(db)

	tests := []struct {
		name           string
		user           models.User
		requiredSlug   string
		expectedStatus int
	}{
		{
			name:           "Master Admin Bypass",
			user:           models.User{RoleID: adminRole.ID},
			requiredSlug:   "any:action",
			expectedStatus: 200,
		},
		{
			name:           "User Has Permission",
			user:           models.User{RoleID: userRole.ID},
			requiredSlug:   "test:write",
			expectedStatus: 200,
		},
		{
			name:           "User Lacks Permission",
			user:           models.User{RoleID: otherRole.ID},
			requiredSlug:   "test:write",
			expectedStatus: 403,
		},
		{
			name:           "No User in Context",
			user:           models.User{}, // Will be handled by if !exists
			requiredSlug:   "test:write",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.Use(func(c *gin.Context) {
				if tt.user.RoleID != uuid.Nil {
					c.Set("role_id", tt.user.RoleID)
				}
				c.Next()
			})
			r.GET("/test", checker.RequirePermission(tt.requiredSlug), func(c *gin.Context) {
				c.Status(200)
			})

			c.Request, _ = http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, c.Request)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
