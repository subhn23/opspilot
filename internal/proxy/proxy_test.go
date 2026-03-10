package proxy

import (
	"net/http"
	"net/http/httptest"
	"opspilot/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServeHTTP(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.ProxyRoute{})

	// 1. Setup Mock Backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	// 2. Setup Route in DB
	route := models.ProxyRoute{
		Domain:    "example.com",
		TargetURL: backend.URL,
		IsActive:  true,
	}
	db.Create(&route)

	proxy := NewOpsProxy(db)

	// 3. Test Successful Proxying
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "backend response" {
		t.Errorf("Expected body 'backend response', got '%s'", rr.Body.String())
	}

	// 4. Test Inactive Route
	db.Model(&route).Update("IsActive", false)
	req, _ = http.NewRequest("GET", "http://example.com", nil)
	rr = httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for inactive route, got %d", rr.Code)
	}

	// 5. Test Missing Route
	req, _ = http.NewRequest("GET", "http://notfound.com", nil)
	rr = httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for missing route, got %d", rr.Code)
	}

	// 6. Test Malformed Target URL
	route2 := models.ProxyRoute{
		Domain:    "badurl.com",
		TargetURL: " ://invalid-url",
		IsActive:  true,
	}
	db.Create(&route2)
	req, _ = http.NewRequest("GET", "http://badurl.com", nil)
	rr = httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for malformed target, got %d", rr.Code)
	}
}
