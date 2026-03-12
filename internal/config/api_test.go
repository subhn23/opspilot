package config

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBackupAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	db, mock, _ := sqlmock.New()
	defer db.Close()

	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})

	r := gin.New()
	r.POST("/api/config/backup", func(c *gin.Context) {
		backupPath := c.PostForm("path")
		if backupPath == "" {
			c.JSON(400, gin.H{"error": "path is required"})
			return
		}

		// Mock expectations for the call
		mock.ExpectExec("ALTER SYSTEM SET archive_mode = 'on'").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("ALTER SYSTEM SET archive_command = .*").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("ALTER SYSTEM SET wal_level = 'replica'").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("SELECT pg_reload_conf\\(\\)").WillReturnResult(sqlmock.NewResult(0, 0))
		
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO \"audit_logs\"").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := ConfigureWALArchiving(gormDB, backupPath)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "SUCCESS"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/config/backup", strings.NewReader("path=/mnt/backup"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
