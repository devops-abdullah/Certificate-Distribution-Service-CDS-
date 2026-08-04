package certs

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractor_Export_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	e := NewExtractor(dir)

	certB64 := base64.StdEncoding.EncodeToString([]byte("FAKE CERT PEM"))
	keyB64 := base64.StdEncoding.EncodeToString([]byte("FAKE KEY PEM"))

	errs := e.Export([]ExportInput{
		{Domain: "example.com", Certificate: certB64, Key: keyB64},
	})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}

	certPath := filepath.Join(dir, "example.com", "fullchain.pem")
	keyPath := filepath.Join(dir, "example.com", "privkey.pem")

	certData, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("read fullchain.pem: %v", err)
	}
	if string(certData) != "FAKE CERT PEM" {
		t.Errorf("unexpected cert content: %s", certData)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read privkey.pem: %v", err)
	}
	if string(keyData) != "FAKE KEY PEM" {
		t.Errorf("unexpected key content: %s", keyData)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat privkey.pem: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("expected privkey.pem to be 0600, got %o", perm)
	}
}

func TestExtractor_Export_InvalidBase64(t *testing.T) {
	dir := t.TempDir()
	e := NewExtractor(dir)

	errs := e.Export([]ExportInput{
		{Domain: "example.com", Certificate: "not-base64!!", Key: "also-not-base64!!"},
	})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestExtractor_Export_Overwrites(t *testing.T) {
	dir := t.TempDir()
	e := NewExtractor(dir)

	first := base64.StdEncoding.EncodeToString([]byte("FIRST"))
	second := base64.StdEncoding.EncodeToString([]byte("SECOND"))

	e.Export([]ExportInput{{Domain: "example.com", Certificate: first, Key: first}})
	e.Export([]ExportInput{{Domain: "example.com", Certificate: second, Key: second}})

	data, err := os.ReadFile(filepath.Join(dir, "example.com", "fullchain.pem"))
	if err != nil {
		t.Fatalf("read fullchain.pem: %v", err)
	}
	if string(data) != "SECOND" {
		t.Errorf("expected re-export to overwrite content, got %s", data)
	}
}
