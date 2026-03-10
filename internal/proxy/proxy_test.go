package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"opspilot/internal/models"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper to generate a self-signed cert for testing
func generateTestCert() (string, string) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"OpsPilot Test"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	
	return string(certPEM), string(privPEM)
}

func TestGetCertificate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Certificate{}, &models.CertTestOverride{})

	fullChain, privKey := generateTestCert()
	
	// 1. Create a Global Production Cert
	prodCert := models.Certificate{
		Label:        "Prod Cert",
		FullChain:    fullChain,
		PrivateKey:   privKey,
		IsProduction: true,
	}
	db.Create(&prodCert)

	proxy := NewOpsProxy(db)

	// 2. Test Fallback to Prod Cert
	hello := &tls.ClientHelloInfo{ServerName: "any.com"}
	cert, err := proxy.GetCertificate(hello)
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}
	if cert == nil {
		t.Fatal("Expected production certificate, got nil")
	}

	// 3. Test Override Logic
	testChain, testKey := generateTestCert()
	testCert := models.Certificate{
		Label:        "Test Cert",
		FullChain:    testChain,
		PrivateKey:   testKey,
		IsProduction: false,
	}
	db.Create(&testCert)
	
	db.Create(&models.CertTestOverride{
		Domain: "test.com",
		CertID: testCert.ID,
	})

	hello = &tls.ClientHelloInfo{ServerName: "test.com"}
	cert, err = proxy.GetCertificate(hello)
	if err != nil {
		t.Fatalf("GetCertificate override failed: %v", err)
	}
	if cert == nil {
		t.Fatal("Expected override certificate, got nil")
	}
	
	// Check if it's the test cert (simple length check or comparison could work)
	if len(cert.Certificate[0]) == 0 {
		t.Error("Returned certificate is empty")
	}

	// 4. Test Invalid Certificate Data
	invalidCert := models.Certificate{
		Label:        "Invalid Cert",
		FullChain:    "invalid-chain",
		PrivateKey:   "invalid-key",
		IsProduction: true,
	}
	db.Create(&invalidCert)
	// We need to remove or deactivate previous prod certs
	db.Where("label = ?", "Prod Cert").Delete(&models.Certificate{})
	
	// ServerName that doesn't match the override
	hello = &tls.ClientHelloInfo{ServerName: "other.com"}
	cert, err = proxy.GetCertificate(hello)
	if err == nil {
		t.Error("Expected error for invalid certificate data, got nil")
	}
	if cert != nil {
		t.Error("Expected nil cert for invalid data, got something")
	}
}

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
