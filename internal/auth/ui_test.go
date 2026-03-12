package auth

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"opspilot/internal/models"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMFAEnrollTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.LoadHTMLGlob("../../ui/templates/*")

	r.GET("/auth/mfa/enroll", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mfa_enroll.html", gin.H{
			"QRCode": template.URL("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAMgAAADICAYAAACt..."),
			"Secret": "TESTSECRET123",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/mfa/enroll", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	expectedSubstrings := []string{
		"MFA Enrollment | OpsPilot",
		"data:image/png;base64",
		"TESTSECRET123",
		"hx-post=\"/auth/mfa/verify\"",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(body, s) {
			t.Errorf("Expected body to contain %q\nBody: %s", s, body)
		}
	}
}

func TestAuditViewerTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.LoadHTMLGlob("../../ui/templates/*")

	r.GET("/audit", func(c *gin.Context) {
		c.HTML(http.StatusOK, "audit_viewer.html", nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/audit", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "System Audit Logs") {
		t.Error("Expected body to contain 'System Audit Logs'")
	}
	if !strings.Contains(body, "hx-get=\"/api/audit\"") {
		t.Error("Expected body to contain hx-get for audit api")
	}
}

func TestEnvWizardTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.LoadHTMLGlob("../../ui/templates/*")

	r.GET("/environments/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "env_wizard.html", gin.H{
			"Hosts": []models.TargetHost{
				{Name: "Mock Host", Type: "remote_ssh"},
			},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/environments/new", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Provision New Environment") {
		t.Error("Expected body to contain 'Provision New Environment'")
	}
	if !strings.Contains(body, "Mock Host") {
		t.Error("Expected body to contain 'Mock Host' in dropdown")
	}
	if !strings.Contains(body, "hx-post=\"/api/environments\"") {
		t.Error("Expected body to contain hx-post for environments api")
	}
}

func TestHostsTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.LoadHTMLGlob("../../ui/templates/*")

	r.GET("/hosts", func(c *gin.Context) {
		c.HTML(http.StatusOK, "hosts.html", nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hosts", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Target Infrastructure Hosts") {
		t.Error("Expected body to contain 'Target Infrastructure Hosts'")
	}
	if !strings.Contains(body, "hx-get=\"/api/hosts\"") {
		t.Error("Expected body to contain hx-get for hosts api")
	}
	if !strings.Contains(body, "For Federated OpsPilot, use the API Token") {
		t.Error("Expected body to contain federation hint")
	}
}

func TestHostsAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.TargetHost{})
	
	os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	defer os.Unsetenv("ENCRYPTION_KEY")

	r := gin.Default()
	
	// API List
	r.GET("/api/hosts", func(c *gin.Context) {
		var hosts []models.TargetHost
		db.Find(&hosts)
		c.String(200, hosts[0].Name)
	})

	// API Create
	r.POST("/api/hosts", func(c *gin.Context) {
		name := c.PostForm("name")
		host := models.TargetHost{Name: name, Type: "remote_ssh"}
		db.Create(&host)
		c.Status(200)
	})

	// 1. Test POST
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/hosts", strings.NewReader("name=NewHost"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200 on POST, got %d", w.Code)
	}

	// 2. Test GET
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/hosts", nil)
	r.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "NewHost") {
		t.Errorf("Expected 'NewHost' in body, got %s", w.Body.String())
	}
}

func TestFederationAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("FEDERATION_TOKEN", "secret-token")
	defer os.Unsetenv("FEDERATION_TOKEN")

	r := gin.Default()
	r.POST("/api/federation/deploy", func(c *gin.Context) {
		token := c.GetHeader("X-Federation-Token")
		if token != "secret-token" {
			c.Status(401)
			return
		}
		var req models.FederationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Status(400)
			return
		}
		c.JSON(200, gin.H{"status": "SUCCESS"})
	})

	// 1. Unauthorized
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/federation/deploy", strings.NewReader("{}"))
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("Expected 401, got %d", w.Code)
	}

	// 2. Success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/federation/deploy", strings.NewReader(`{"environment_name":"test","commit_hash":"abc"}`))
	req.Header.Set("X-Federation-Token", "secret-token")
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
