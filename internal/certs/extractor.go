package certs

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// ExportInput carries the raw certificate and key material needed to write
// a certificate to disk. Unlike Input/Metadata, this includes private key
// material and must never be logged or exposed through the API.
type ExportInput struct {
	Domain      string
	Certificate string // base64-encoded PEM certificate
	Key         string // base64-encoded PEM private key
}

// Extractor writes certificates to disk as fullchain.pem/privkey.pem, one
// directory per domain, under a base export directory.
type Extractor struct {
	baseDir string
}

// NewExtractor creates an extractor that writes into baseDir.
func NewExtractor(baseDir string) *Extractor {
	return &Extractor{baseDir: baseDir}
}

// Export writes every entry in inputs to
// <baseDir>/<domain>/{fullchain,privkey}.pem. Entries that fail to decode
// or write are skipped and reported rather than aborting the whole batch.
func (e *Extractor) Export(inputs []ExportInput) []error {
	var errs []error

	for _, in := range inputs {
		if err := e.exportOne(in); err != nil {
			errs = append(errs, fmt.Errorf("export %s: %w", in.Domain, err))
		}
	}

	return errs
}

func (e *Extractor) exportOne(in ExportInput) error {
	cert, err := base64.StdEncoding.DecodeString(in.Certificate)
	if err != nil {
		return fmt.Errorf("decode certificate: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(in.Key)
	if err != nil {
		return fmt.Errorf("decode key: %w", err)
	}

	dir := filepath.Join(e.baseDir, in.Domain)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	if err := writeFileAtomic(filepath.Join(dir, "fullchain.pem"), cert, 0o644); err != nil {
		return fmt.Errorf("write fullchain.pem: %w", err)
	}

	if err := writeFileAtomic(filepath.Join(dir, "privkey.pem"), key, 0o600); err != nil {
		return fmt.Errorf("write privkey.pem: %w", err)
	}

	return nil
}

// writeFileAtomic writes data to a temp file in the same directory as path
// and renames it into place, so readers never observe a partially written
// certificate or key.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}
