package proxy

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"opspilot/internal/models"

	"gorm.io/gorm"
)

type OpsProxy struct {
	DB *gorm.DB
}

func NewOpsProxy(db *gorm.DB) *OpsProxy {
	return &OpsProxy{DB: db}
}

// Start launches the HTTPS reverse proxy
func (p *OpsProxy) Start(addr string) {
	tlsConfig := &tls.Config{
		GetCertificate: p.GetCertificate,
	}

	server := &http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
		Handler:   http.HandlerFunc(p.ServeHTTP),
	}

	log.Printf("OpsProxy starting on %s", addr)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("OpsProxy failed: %v", err)
	}
}

// ServeHTTP handles the actual request proxying
func (p *OpsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Basic Health Check
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	var route models.ProxyRoute
	err := p.DB.Where("domain = ? AND is_active = ?", r.Host, true).First(&route).Error
	if err != nil {
		log.Printf("Proxy: No active route found for domain %s", r.Host)
		http.Error(w, "Service Not Found", http.StatusNotFound)
		return
	}

	target, err := url.Parse(route.TargetURL)
	if err != nil {
		log.Printf("Proxy: Malformed target URL for domain %s: %v", r.Host, err)
		http.Error(w, "Internal Configuration Error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Update headers for proxying
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Header.Set("X-Forwarded-Host", r.Host)

	log.Printf("Proxying request: %s -> %s", r.Host, route.TargetURL)
	proxy.ServeHTTP(w, r)
}

// GetCertificate implements the "Test-then-Deploy" SSL logic
func (p *OpsProxy) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var cert models.Certificate

	log.Printf("SSL: Handshake for domain %s", hello.ServerName)

	// 1. Check for Test Override first
	var override models.CertTestOverride
	if err := p.DB.Where("domain = ?", hello.ServerName).First(&override).Error; err == nil {
		log.Printf("SSL: Found test override for domain %s", hello.ServerName)
		if err := p.DB.First(&cert, override.CertID).Error; err == nil {
			return p.parseCert(cert)
		}
	}

	// 2. Fallback to Global Production Certificate
	if err := p.DB.Where("is_production = ?", true).First(&cert).Error; err == nil {
		log.Printf("SSL: Using global production certificate for domain %s", hello.ServerName)
		return p.parseCert(cert)
	}

	log.Printf("SSL: No certificate found for domain %s", hello.ServerName)
	return nil, nil // No certificate found
}

func (p *OpsProxy) parseCert(c models.Certificate) (*tls.Certificate, error) {
	tlsCert, err := tls.X509KeyPair([]byte(c.FullChain), []byte(c.PrivateKey))
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}
