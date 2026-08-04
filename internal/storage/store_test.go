package storage

import (
	"testing"

	"github.com/devops-abdullah/cds/internal/certs"
)

func TestStore_ReplaceListGet(t *testing.T) {
	s := New()

	s.Replace([]certs.Metadata{
		{Domain: "example.com"},
		{Domain: "example.org"},
	})

	if len(s.List()) != 2 {
		t.Fatalf("expected 2 items, got %d", len(s.List()))
	}

	if _, ok := s.Get("example.com"); !ok {
		t.Fatalf("expected to find example.com")
	}

	if _, ok := s.Get("missing.example.com"); ok {
		t.Fatalf("expected missing.example.com to not be found")
	}

	s.Replace([]certs.Metadata{{Domain: "example.net"}})

	if len(s.List()) != 1 {
		t.Fatalf("expected replace to reset contents, got %d items", len(s.List()))
	}
}
