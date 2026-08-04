package certs

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"
)

// Metadata describes a certificate's identity and validity window. It never
// carries private key material, so it is safe to expose through the API.
type Metadata struct {
	Resolver     string    `json:"resolver"`
	Store        string    `json:"store"`
	Domain       string    `json:"domain"`
	SANs         []string  `json:"sans"`
	Issuer       string    `json:"issuer,omitempty"`
	SerialNumber string    `json:"serialNumber,omitempty"`
	NotBefore    time.Time `json:"notBefore,omitempty"`
	NotAfter     time.Time `json:"notAfter,omitempty"`
	Expired      bool      `json:"expired"`
	// NotYetValid is true when the certificate's validity period hasn't
	// started yet (clock skew or a pre-staged certificate).
	NotYetValid bool `json:"notYetValid"`
	// ExpiringSoon is true when the certificate is still valid but will
	// expire within the configured warning window.
	ExpiringSoon bool `json:"expiringSoon"`
}

// Input is a provider-agnostic description of a stored certificate. ACME
// providers (e.g. the Traefik store reader) translate their own formats into
// Input so this package never depends on a provider-specific shape.
type Input struct {
	Resolver    string
	Store       string
	Domain      string
	SANs        []string
	Certificate string // base64-encoded PEM certificate data
}

// FromInput builds certificate metadata by decoding just enough of the
// certificate to read its identity fields. expiryWarningWindow controls how
// far ahead of NotAfter a still-valid certificate is flagged ExpiringSoon.
func FromInput(in Input, expiryWarningWindow time.Duration) (Metadata, error) {

	meta := Metadata{
		Resolver: in.Resolver,
		Store:    in.Store,
		Domain:   in.Domain,
		SANs:     in.SANs,
	}

	cert, err := decodeCertificate(in.Certificate)
	if err != nil {
		return meta, fmt.Errorf("decode certificate for %s: %w", meta.Domain, err)
	}

	now := time.Now()

	meta.Issuer = cert.Issuer.CommonName
	meta.SerialNumber = cert.SerialNumber.String()
	meta.NotBefore = cert.NotBefore
	meta.NotAfter = cert.NotAfter
	meta.Expired = now.After(cert.NotAfter)
	meta.NotYetValid = now.Before(cert.NotBefore)
	meta.ExpiringSoon = !meta.Expired && now.Add(expiryWarningWindow).After(cert.NotAfter)

	return meta, nil
}

func decodeCertificate(encoded string) (*x509.Certificate, error) {

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		block = &pem.Block{Bytes: raw}
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509 parse: %w", err)
	}

	return cert, nil
}
