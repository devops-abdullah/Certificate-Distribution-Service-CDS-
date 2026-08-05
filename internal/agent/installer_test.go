package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstaller_Install_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	installer := NewInstaller(dir)

	changed, err := installer.Install(Bundle{Domain: "example.com", Certificate: []byte("CERT-1"), Key: []byte("KEY-1")})
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if !changed {
		t.Error("expected first install to report changed=true")
	}

	certPath := filepath.Join(dir, "example.com", "fullchain.pem")
	keyPath := filepath.Join(dir, "example.com", "privkey.pem")

	certData, err := os.ReadFile(certPath)
	if err != nil || string(certData) != "CERT-1" {
		t.Fatalf("unexpected fullchain.pem content: %q, err=%v", certData, err)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat privkey.pem: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("expected privkey.pem to be 0600, got %o", perm)
	}
}

func TestInstaller_Install_ReportsUnchanged(t *testing.T) {
	dir := t.TempDir()
	installer := NewInstaller(dir)

	bundle := Bundle{Domain: "example.com", Certificate: []byte("CERT-1"), Key: []byte("KEY-1")}

	if _, err := installer.Install(bundle); err != nil {
		t.Fatalf("first install: %v", err)
	}

	changed, err := installer.Install(bundle)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if changed {
		t.Error("expected identical re-install to report changed=false")
	}
}

func TestInstaller_Install_DetectsChange(t *testing.T) {
	dir := t.TempDir()
	installer := NewInstaller(dir)

	if _, err := installer.Install(Bundle{Domain: "example.com", Certificate: []byte("CERT-1"), Key: []byte("KEY-1")}); err != nil {
		t.Fatalf("first install: %v", err)
	}

	changed, err := installer.Install(Bundle{Domain: "example.com", Certificate: []byte("CERT-2"), Key: []byte("KEY-2")})
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if !changed {
		t.Error("expected a renewed certificate to report changed=true")
	}

	certData, _ := os.ReadFile(filepath.Join(dir, "example.com", "fullchain.pem"))
	if string(certData) != "CERT-2" {
		t.Errorf("expected updated content, got %q", certData)
	}
}
