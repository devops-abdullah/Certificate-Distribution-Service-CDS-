package acme

// Store represents the raw structure of a Traefik ACME store (acme.json).
// Each key is the name of a certificate resolver (e.g. "letsencrypt").
type Store map[string]Resolver

// Resolver holds the ACME account and certificates managed by a single
// certificate resolver.
type Resolver struct {
	Account      Account       `json:"Account"`
	Certificates []Certificate `json:"Certificates"`
}

// Account holds the ACME account registered with the certificate authority.
type Account struct {
	Email        string       `json:"Email"`
	Registration Registration `json:"Registration"`
	PrivateKey   string       `json:"PrivateKey"`
	KeyType      string       `json:"KeyType"`
}

// Registration holds the ACME account registration details.
type Registration struct {
	Body RegistrationBody `json:"body"`
	URI  string           `json:"uri"`
}

// RegistrationBody holds the ACME account status and contact details.
type RegistrationBody struct {
	Status  string   `json:"status"`
	Contact []string `json:"contact"`
}

// Certificate represents a single certificate entry stored by Traefik.
// Certificate and Key are base64-encoded PEM data.
type Certificate struct {
	Domain      Domain `json:"domain"`
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
	Store       string `json:"Store"`
}

// Domain holds the primary domain and subject alternative names covered by
// a certificate.
type Domain struct {
	Main string   `json:"main"`
	SANs []string `json:"sans"`
}
