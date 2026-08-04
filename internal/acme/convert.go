package acme

import "github.com/devops-abdullah/cds/internal/certs"

// ToCertificateInputs flattens every resolver's certificates in the store
// into a provider-agnostic list the certs package can turn into metadata.
func (s Store) ToCertificateInputs() []certs.Input {

	var inputs []certs.Input

	for resolver, r := range s {
		for _, c := range r.Certificates {
			inputs = append(inputs, certs.Input{
				Resolver:    resolver,
				Store:       c.Store,
				Domain:      c.Domain.Main,
				SANs:        c.Domain.SANs,
				Certificate: c.Certificate,
			})
		}
	}

	return inputs
}
