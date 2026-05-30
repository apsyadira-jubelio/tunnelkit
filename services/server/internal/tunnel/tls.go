package tunnel

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
)

// TLSManager handles automatic TLS certificate management
type TLSManager struct {
	domain      string
	certDir     string
	development bool
	manager     *autocert.Manager
}

type TLSConfig struct {
	Domain      string
	CertDir     string // Directory to store certificates
	Development bool   // Use self-signed certs in dev
	Email       string // Let's Encrypt contact email
}

func NewTLSManager(cfg *TLSConfig) *TLSManager {
	if cfg.CertDir == "" {
		cfg.CertDir = filepath.Join(os.TempDir(), "tunnelkit-certs")
	}

	mgr := &TLSManager{
		domain:      cfg.Domain,
		certDir:     cfg.CertDir,
		development: cfg.Development,
	}

	if !cfg.Development {
		mgr.manager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(cfg.CertDir),
			HostPolicy: mgr.hostPolicy,
			Email:      cfg.Email,
		}
	}

	return mgr
}

func (m *TLSManager) hostPolicy(ctx context.Context, host string) error {
	// Allow any subdomain of the configured domain
	if host == m.domain || (len(host) > len(m.domain) && host[len(host)-len(m.domain)-1:] == "."+m.domain) {
		return nil
	}
	return fmt.Errorf("acme/autocert: host %q not allowed", host)
}

// GetTLSConfig returns TLS configuration for the server
func (m *TLSManager) GetTLSConfig() *tls.Config {
	if m.development {
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return &tls.Config{
		GetCertificate: m.manager.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	}
}




// IsDevelopment returns true if running in development mode
func (m *TLSManager) IsDevelopment() bool {
	return m.development
}


