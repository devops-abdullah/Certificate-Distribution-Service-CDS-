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
}

var App Config

func Load() {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("ACME_FILE", "/data/acme.json")
	viper.SetDefault("EXPORT_DIR", "/data/export")
	viper.SetDefault("CERT_EXPIRY_WARN_DAYS", 30)

	viper.AutomaticEnv()

	App = Config{
		Port:               viper.GetString("PORT"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		AcmeFile:           viper.GetString("ACME_FILE"),
		ExportDir:          viper.GetString("EXPORT_DIR"),
		CertExpiryWarnDays: viper.GetInt("CERT_EXPIRY_WARN_DAYS"),
	}

	log.Println("Configuration loaded")
}
