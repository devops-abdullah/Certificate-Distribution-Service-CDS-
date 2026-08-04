package storage

import (
	"sync"

	"github.com/devops-abdullah/cds/internal/certs"
)

// Repository is the contract certificate metadata storage backends must
// satisfy. Callers depend on this interface rather than a concrete type so
// the in-memory implementation can later be swapped for a persistent one
// (e.g. Redis, a file, a database) without touching consumers.
type Repository interface {
	// Replace atomically replaces the repository's contents, keyed by domain.
	Replace(items []certs.Metadata)
	// List returns all certificate metadata currently held.
	List() []certs.Metadata
	// Get returns the certificate metadata for the given domain.
	Get(domain string) (certs.Metadata, bool)
}

// Store is a thread-safe in-memory Repository implementation.
type Store struct {
	mu    sync.RWMutex
	certs map[string]certs.Metadata
}

var _ Repository = (*Store)(nil)

// New creates an empty in-memory certificate metadata store.
func New() *Store {
	return &Store{certs: make(map[string]certs.Metadata)}
}

func (s *Store) Replace(items []certs.Metadata) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.certs = make(map[string]certs.Metadata, len(items))
	for _, item := range items {
		s.certs[item.Domain] = item
	}
}

func (s *Store) List() []certs.Metadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]certs.Metadata, 0, len(s.certs))
	for _, item := range s.certs {
		items = append(items, item)
	}

	return items
}

func (s *Store) Get(domain string) (certs.Metadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.certs[domain]
	return item, ok
}
