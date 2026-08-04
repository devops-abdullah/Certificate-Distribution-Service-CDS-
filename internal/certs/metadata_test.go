package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

const testWarningWindow = 30 * 24 * time.Hour

func generateTestCertificate(t *testing.T, notBefore, notAfter time.Time) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test CA"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		DNSNames:     []string{"example.com"},
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	return base64.StdEncoding.EncodeToString(pemBytes)
}

func TestFromInput_ValidCertificate(t *testing.T) {
	encoded := generateTestCertificate(t, time.Now().Add(-time.Hour), time.Now().Add(90*24*time.Hour))

	meta, err := FromInput(Input{
		Resolver:    "letsencrypt",
		Store:       "default",
		Domain:      "example.com",
		SANs:        []string{"www.example.com"},
		Certificate: encoded,
	}, testWarningWindow)
	if err != nil {
		t.Fatalf("FromInput returned error: %v", err)
	}

	if meta.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", meta.Domain)
	}

	if meta.Expired {
		t.Errorf("expected certificate to not be expired")
	}

	if meta.NotYetValid {
		t.Errorf("expected certificate to already be valid")
	}

	if meta.ExpiringSoon {
		t.Errorf("expected certificate to not be expiring soon")
	}

	if meta.Issuer != "Test CA" {
		t.Errorf("expected issuer Test CA, got %s", meta.Issuer)
	}
}

func TestFromInput_ExpiredCertificate(t *testing.T) {
	encoded := generateTestCertificate(t, time.Now().Add(-48*time.Hour), time.Now().Add(-time.Hour))

	meta, err := FromInput(Input{Domain: "example.com", Certificate: encoded}, testWarningWindow)
	if err != nil {
		t.Fatalf("FromInput returned error: %v", err)
	}

	if !meta.Expired {
		t.Errorf("expected certificate to be expired")
	}

	if meta.ExpiringSoon {
		t.Errorf("an already-expired certificate should not also be reported as expiring soon")
	}
}

func TestFromInput_ExpiringSoon(t *testing.T) {
	encoded := generateTestCertificate(t, time.Now().Add(-89*24*time.Hour), time.Now().Add(4*24*time.Hour))

	meta, err := FromInput(Input{Domain: "example.com", Certificate: encoded}, testWarningWindow)
	if err != nil {
		t.Fatalf("FromInput returned error: %v", err)
	}

	if meta.Expired {
		t.Errorf("expected certificate to still be valid")
	}

	if !meta.ExpiringSoon {
		t.Errorf("expected certificate expiring in 4 days to be flagged with a 30 day warning window")
	}
}

func TestFromInput_NotYetValid(t *testing.T) {
	encoded := generateTestCertificate(t, time.Now().Add(24*time.Hour), time.Now().Add(90*24*time.Hour))

	meta, err := FromInput(Input{Domain: "example.com", Certificate: encoded}, testWarningWindow)
	if err != nil {
		t.Fatalf("FromInput returned error: %v", err)
	}

	if !meta.NotYetValid {
		t.Errorf("expected certificate with a future NotBefore to be flagged NotYetValid")
	}
}

func TestFromInput_InvalidCertificate(t *testing.T) {
	_, err := FromInput(Input{Domain: "example.com", Certificate: "not-valid-base64!!"}, testWarningWindow)
	if err == nil {
		t.Fatal("expected error for invalid certificate data")
	}
}
