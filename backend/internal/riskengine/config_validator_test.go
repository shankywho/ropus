package riskengine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidator_ProductionRules(t *testing.T) {
	// 1. Valid Production Config
	validProd := AppConfig{
		Env:              EnvProduction,
		Port:             "8080",
		AdminAPIKey:      "strong_super_secret_production_key_2026",
		DatabaseURL:      "postgres://risk_user:strong_prod_pass_999@db.internal:5432/risk_engine",
		PostgresPassword: "strong_prod_pass_999",
	}
	assert.NoError(t, validProd.Validate())

	// 2. Missing Admin API Key in Production
	missingKey := validProd
	missingKey.AdminAPIKey = ""
	assert.Error(t, missingKey.Validate())

	// 3. Short Admin API Key (< 16 chars)
	shortKey := validProd
	shortKey.AdminAPIKey = "short_key"
	assert.Error(t, shortKey.Validate())

	// 4. Weak / Default Key
	weakKey := validProd
	weakKey.AdminAPIKey = "risk_secret"
	assert.Error(t, weakKey.Validate())

	// 5. Default Postgres Password in Production
	defaultPass := validProd
	defaultPass.PostgresPassword = "risk_password"
	assert.Error(t, defaultPass.Validate())

	// 6. TLS Enabled without Certs
	tlsMissingCerts := validProd
	tlsMissingCerts.TLSEnabled = true
	assert.Error(t, tlsMissingCerts.Validate())
}

func TestConfigValidator_DevelopmentDefaults(t *testing.T) {
	devConfig := DefaultAppConfig()
	assert.NoError(t, devConfig.Validate())
	assert.Equal(t, EnvDevelopment, devConfig.Env)
}

func TestConfigValidator_SanitizedStringRedaction(t *testing.T) {
	cfg := AppConfig{
		Env:         EnvProduction,
		Port:        "8080",
		AdminAPIKey: "super_secret_production_key_123456789",
		DatabaseURL: "postgres://risk_user:super_secret_db_password@postgres:5432/risk_engine",
	}

	sanitized := cfg.SanitizedString()
	assert.NotContains(t, sanitized, "super_secret_db_password")
	assert.NotContains(t, sanitized, "super_secret_production_key_123456789")
	assert.Contains(t, sanitized, "[REDACTED]")
}
