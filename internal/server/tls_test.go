package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestCA(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Client CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create CA certificate: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		t.Fatalf("write CA file: %v", err)
	}

	return path
}

func TestBuildTLSConfig_NoClientCA(t *testing.T) {
	cfg, err := BuildTLSConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ClientAuth != tls.NoClientCert {
		t.Errorf("expected no client cert requirement, got %v", cfg.ClientAuth)
	}
	if cfg.ClientCAs != nil {
		t.Errorf("expected nil ClientCAs, got non-nil")
	}
}

func TestBuildTLSConfig_ValidClientCA(t *testing.T) {
	path := writeTestCA(t)

	cfg, err := BuildTLSConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Errorf("expected RequireAndVerifyClientCert, got %v", cfg.ClientAuth)
	}
	if cfg.ClientCAs == nil {
		t.Fatal("expected non-nil ClientCAs")
	}
}

func TestBuildTLSConfig_MissingFile(t *testing.T) {
	_, err := BuildTLSConfig(filepath.Join(t.TempDir(), "does-not-exist.pem"))
	if err == nil {
		t.Fatal("expected error for missing CA file")
	}
}

func TestBuildTLSConfig_InvalidPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.pem")
	if err := os.WriteFile(path, []byte("not a certificate"), 0o600); err != nil {
		t.Fatalf("write bad CA file: %v", err)
	}

	_, err := BuildTLSConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid PEM content")
	}
}
