package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMFAEnrollTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.LoadHTMLGlob("../../ui/templates/*")

	r.GET("/auth/mfa/enroll", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mfa_enroll.html", gin.H{
			"QRCode": "test-qr-code",
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
		"test-qr-code",
		"TESTSECRET123",
		"hx-post=\"/auth/mfa/verify\"",
	}

	for _, s := range expectedSubstrings {
		if !contains(body, s) {
			t.Errorf("Expected body to contain %q", s)
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
	if !contains(body, "System Audit Logs") {
		t.Error("Expected body to contain 'System Audit Logs'")
	}
	if !contains(body, "hx-get=\"/api/audit\"") {
		t.Error("Expected body to contain hx-get for audit api")
	}
}

func TestEnvWizardTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.LoadHTMLGlob("../../ui/templates/*")

	r.GET("/environments/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "env_wizard.html", nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/environments/new", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !contains(body, "Provision New Environment") {
		t.Error("Expected body to contain 'Provision New Environment'")
	}
	if !contains(body, "hx-post=\"/api/environments\"") {
		t.Error("Expected body to contain hx-post for environments api")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || find(s, substr))
}

func find(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
