package auth

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
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
		c.HTML(http.StatusOK, "env_wizard.html", nil)
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
	if !strings.Contains(body, "hx-post=\"/api/environments\"") {
		t.Error("Expected body to contain hx-post for environments api")
	}
}
