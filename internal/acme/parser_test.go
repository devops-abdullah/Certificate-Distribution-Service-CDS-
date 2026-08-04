package acme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_Parse_ValidFile(t *testing.T) {
	store, err := NewParser().Parse(filepath.Join("..", "..", "examples", "acme.sample.json"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	resolver, ok := store["letsencrypt"]
	if !ok {
		t.Fatalf("expected %q resolver in store", "letsencrypt")
	}

	if len(resolver.Certificates) != 2 {
		t.Fatalf("expected 2 certificates, got %d", len(resolver.Certificates))
	}

	if resolver.Certificates[0].Domain.Main != "new.example.com" {
		t.Fatalf("expected domain new.example.com, got %s", resolver.Certificates[0].Domain.Main)
	}

	if resolver.Certificates[1].Domain.Main != "expiring.example.com" {
		t.Fatalf("expected domain expiring.example.com, got %s", resolver.Certificates[1].Domain.Main)
	}
}

func TestParser_Parse_MissingFile(t *testing.T) {
	_, err := NewParser().Parse(filepath.Join(os.TempDir(), "cds-does-not-exist.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParser_Parse_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "acme.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := NewParser().Parse(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
