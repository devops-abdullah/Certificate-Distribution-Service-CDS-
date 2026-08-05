package agent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devops-abdullah/cds/pkg/utils"
)

// Installer writes fetched certificate bundles to disk, one directory per
// domain.
type Installer struct {
	baseDir string
}

// NewInstaller creates an installer that writes into baseDir.
func NewInstaller(baseDir string) *Installer {
	return &Installer{baseDir: baseDir}
}

// Install writes bundle to <baseDir>/<domain>/{fullchain,privkey}.pem and
// reports whether the certificate content changed since the last install,
// so callers can skip an Nginx reload when nothing actually changed.
func (i *Installer) Install(bundle Bundle) (changed bool, err error) {
	dir := filepath.Join(i.baseDir, bundle.Domain)
	certPath := filepath.Join(dir, "fullchain.pem")
	keyPath := filepath.Join(dir, "privkey.pem")

	if existing, readErr := os.ReadFile(certPath); readErr == nil && bytes.Equal(existing, bundle.Certificate) {
		return false, nil
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false, fmt.Errorf("create directory: %w", err)
	}

	if err := utils.WriteFileAtomic(certPath, bundle.Certificate, 0o644); err != nil {
		return false, fmt.Errorf("write fullchain.pem: %w", err)
	}

	if err := utils.WriteFileAtomic(keyPath, bundle.Key, 0o600); err != nil {
		return false, fmt.Errorf("write privkey.pem: %w", err)
	}

	return true, nil
}
