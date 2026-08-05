// Package agent implements the Certificate Agent: it polls a CDS manager
// for certificate bundles, installs them atomically on disk, and reloads
// Nginx when a certificate actually changes.
package agent

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the agent's runtime configuration, loaded from environment
// variables via Load.
type Config struct {
	ManagerURL     string
	APIKey         string
	Domains        []string
	PollInterval   time.Duration
	InstallDir     string
	NginxReloadCmd string

	// TLS client identity for talking to a manager that requires mTLS, and
	// optionally a CA to trust the manager's own certificate (e.g. a
	// private/self-signed CA rather than a public one).
	TLSClientCert string
	TLSClientKey  string
	TLSCA         string
}

var App Config

// Load reads agent configuration from the environment.
func Load() {
	viper.SetDefault("MANAGER_URL", "http://localhost:8080")
	viper.SetDefault("API_KEY", "")
	viper.SetDefault("DOMAINS", "")
	viper.SetDefault("POLL_INTERVAL_SECONDS", 60)
	viper.SetDefault("INSTALL_DIR", "/etc/cds-agent/certs")
	viper.SetDefault("NGINX_RELOAD_CMD", "nginx -s reload")
	viper.SetDefault("TLS_CLIENT_CERT", "")
	viper.SetDefault("TLS_CLIENT_KEY", "")
	viper.SetDefault("TLS_CA", "")

	viper.AutomaticEnv()

	App = Config{
		ManagerURL:     strings.TrimRight(viper.GetString("MANAGER_URL"), "/"),
		APIKey:         viper.GetString("API_KEY"),
		Domains:        splitDomains(viper.GetString("DOMAINS")),
		PollInterval:   time.Duration(viper.GetInt("POLL_INTERVAL_SECONDS")) * time.Second,
		InstallDir:     viper.GetString("INSTALL_DIR"),
		NginxReloadCmd: viper.GetString("NGINX_RELOAD_CMD"),
		TLSClientCert:  viper.GetString("TLS_CLIENT_CERT"),
		TLSClientKey:   viper.GetString("TLS_CLIENT_KEY"),
		TLSCA:          viper.GetString("TLS_CA"),
	}
}

func splitDomains(raw string) []string {
	var domains []string

	for _, d := range strings.Split(raw, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			domains = append(domains, d)
		}
	}

	return domains
}
