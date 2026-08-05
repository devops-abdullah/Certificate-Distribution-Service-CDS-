package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port               string
	LogLevel           string
	AcmeFile           string
	ExportDir          string
	CertExpiryWarnDays int

	// APIKeyReadOnly and APIKeyAdmin gate the certificate endpoints and the
	// admin-only reload endpoint. Leaving both unset disables the API
	// entirely (fails closed) rather than serving it unauthenticated.
	APIKeyReadOnly string
	APIKeyAdmin    string

	// TLSCert/TLSKey enable HTTPS. TLSClientCA additionally enables mTLS:
	// when set, clients must present a certificate signed by this CA.
	TLSCert     string
	TLSKey      string
	TLSClientCA string
}

var App Config

func Load() {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("ACME_FILE", "/data/acme.json")
	viper.SetDefault("EXPORT_DIR", "/data/export")
	viper.SetDefault("CERT_EXPIRY_WARN_DAYS", 30)
	viper.SetDefault("API_KEY_READONLY", "")
	viper.SetDefault("API_KEY_ADMIN", "")
	viper.SetDefault("TLS_CERT", "")
	viper.SetDefault("TLS_KEY", "")
	viper.SetDefault("TLS_CLIENT_CA", "")

	viper.AutomaticEnv()

	App = Config{
		Port:               viper.GetString("PORT"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		AcmeFile:           viper.GetString("ACME_FILE"),
		ExportDir:          viper.GetString("EXPORT_DIR"),
		CertExpiryWarnDays: viper.GetInt("CERT_EXPIRY_WARN_DAYS"),
		APIKeyReadOnly:     viper.GetString("API_KEY_READONLY"),
		APIKeyAdmin:        viper.GetString("API_KEY_ADMIN"),
		TLSCert:            viper.GetString("TLS_CERT"),
		TLSKey:             viper.GetString("TLS_KEY"),
		TLSClientCA:        viper.GetString("TLS_CLIENT_CA"),
	}

	log.Println("Configuration loaded")
}
