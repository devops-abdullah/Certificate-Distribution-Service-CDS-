package storage

import (
	"sync"

	"github.com/devops-abdullah/cds/internal/certs"
)

// Store is a thread-safe in-memory holder for certificate metadata.
type Store struct {
	mu    sync.RWMutex
	certs map[string]certs.Metadata
}

// New creates an empty certificate metadata store.
func New() *Store {
	return &Store{certs: make(map[string]certs.Metadata)}
}

// Replace atomically replaces the store's contents, keyed by domain.
func (s *Store) Replace(items []certs.Metadata) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.certs = make(map[string]certs.Metadata, len(items))
	for _, item := range items {
		s.certs[item.Domain] = item
	}
}

// List returns all certificate metadata currently held by the store.
func (s *Store) List() []certs.Metadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]certs.Metadata, 0, len(s.certs))
	for _, item := range s.certs {
		items = append(items, item)
	}

	return items
}

// Get returns the certificate metadata for the given domain.
func (s *Store) Get(domain string) (certs.Metadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.certs[domain]
	return item, ok
}
