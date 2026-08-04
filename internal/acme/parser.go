package acme

import (
	"encoding/json"
	"fmt"
	"os"
)

// Parser reads and validates a Traefik ACME store file (acme.json).
type Parser struct{}

// NewParser creates a new ACME store parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads the file at path, validates that it is well-formed JSON, and
// unmarshals it into an acme.Store.
func (p *Parser) Parse(path string) (Store, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read acme file: %w", err)
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("acme file is not valid JSON: %s", path)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parse acme file: %w", err)
	}

	return store, nil
}
