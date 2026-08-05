package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Bundle is the certificate + key material fetched from the manager for a
// single domain.
type Bundle struct {
	Domain      string
	Certificate []byte
	Key         []byte
}

type bundleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Domain      string `json:"domain"`
		Certificate string `json:"certificate"`
		Key         string `json:"key"`
	} `json:"data"`
}

// Client fetches certificate bundles from a CDS manager.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient creates a Client for the given manager base URL and agent API
// key. Pass nil for httpClient to use a plain default; pass a client with a
// custom Transport/TLSClientConfig to support mTLS or a private CA.
func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{baseURL: baseURL, apiKey: apiKey, http: httpClient}
}

// FetchBundle retrieves the certificate + private key for domain from the
// manager's bundle endpoint.
func (c *Client) FetchBundle(domain string) (Bundle, error) {
	target := c.baseURL + "/api/v1/certificates/" + url.PathEscape(domain) + "/bundle"

	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return Bundle{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Bundle{}, fmt.Errorf("request bundle for %s: %w", domain, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed bundleResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return Bundle{}, fmt.Errorf("decode response for %s: %w", domain, err)
	}

	if resp.StatusCode != http.StatusOK || !parsed.Success {
		return Bundle{}, fmt.Errorf("manager returned %d for %s: %s", resp.StatusCode, domain, parsed.Message)
	}

	return Bundle{
		Domain:      parsed.Data.Domain,
		Certificate: []byte(parsed.Data.Certificate),
		Key:         []byte(parsed.Data.Key),
	}, nil
}
