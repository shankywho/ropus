package riskengine

import (
	"crypto/tls"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// AppEnvironment represents the operating runtime environment.
type AppEnvironment string

const (
	EnvDevelopment AppEnvironment = "development"
	EnvTest        AppEnvironment = "test"
	EnvStaging     AppEnvironment = "staging"
	EnvProduction  AppEnvironment = "production"
)

// AppConfig encapsulates the platform runtime settings with security validation.
type AppConfig struct {
	Env              AppEnvironment `json:"env"`
	Port             string         `json:"port"`
	AdminAPIKey      string         `json:"admin_api_key"`
	DatabaseURL      string         `json:"database_url"`
	PostgresPassword string         `json:"postgres_password"`
	RedisHost        string         `json:"redis_host"`
	MLServiceURL     string         `json:"ml_service_url"`
	ClickHouseHost   string         `json:"clickhouse_host"`
	TLSEnabled       bool           `json:"tls_enabled"`
	TLSCertFile      string         `json:"tls_cert_file"`
	TLSKeyFile       string         `json:"tls_key_file"`
	TLSMinVersion    uint16         `json:"tls_min_version"`
}

// DefaultAppConfig creates a development-safe baseline config.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Env:           EnvDevelopment,
		Port:          "8080",
		AdminAPIKey:   "dev_admin_api_key_for_local_testing_only_123",
		DatabaseURL:   "postgres://risk_user:risk_password@localhost:5432/risk_engine?sslmode=disable",
		RedisHost:     "localhost",
		MLServiceURL:  "http://localhost:8000",
		TLSEnabled:    false,
		TLSMinVersion: tls.VersionTLS12,
	}
}

var weakKeys = map[string]bool{
	"dummy":         true,
	"test":          true,
	"secret":        true,
	"admin":         true,
	"12345":         true,
	"12345678":      true,
	"risk_secret":   true,
	"risk_password": true,
	"password":      true,
}

// Validate verifies that the configuration satisfies all security prerequisites for the specified environment.
func (c *AppConfig) Validate() error {
	if c.Port == "" {
		c.Port = "8080"
	}
	if c.Env == "" {
		c.Env = EnvDevelopment
	}

	if c.Env == EnvProduction {
		// 1. Strict Secret Validation in Production
		if strings.TrimSpace(c.AdminAPIKey) == "" {
			return fmt.Errorf("ADMIN_API_KEY is required in production")
		}
		if len(c.AdminAPIKey) < 16 {
			return fmt.Errorf("ADMIN_API_KEY must be at least 16 characters in production (got %d)", len(c.AdminAPIKey))
		}
		if weakKeys[strings.ToLower(c.AdminAPIKey)] {
			return fmt.Errorf("ADMIN_API_KEY cannot be a known default or weak secret in production")
		}

		// 2. Postgres Password Check
		if weakKeys[strings.ToLower(c.PostgresPassword)] {
			return fmt.Errorf("POSTGRES_PASSWORD cannot be default '%s' in production", c.PostgresPassword)
		}

		// 3. TLS Validation
		if c.TLSEnabled {
			if c.TLSCertFile == "" || c.TLSKeyFile == "" {
				return fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE are required when TLS_ENABLED=true")
			}
			if _, err := os.Stat(c.TLSCertFile); err != nil {
				return fmt.Errorf("TLS_CERT_FILE not accessible: %w", err)
			}
			if _, err := os.Stat(c.TLSKeyFile); err != nil {
				return fmt.Errorf("TLS_KEY_FILE not accessible: %w", err)
			}
		}
	}

	return nil
}

var dbPasswordRegex = regexp.MustCompile(`://([^:]+):([^@]+)@`)

// SanitizedString returns a log-safe representation of the config with all credentials masked.
func (c *AppConfig) SanitizedString() string {
	sanitizedDB := dbPasswordRegex.ReplaceAllString(c.DatabaseURL, "://$1:[REDACTED]@")
	maskedKey := "[REDACTED]"
	if len(c.AdminAPIKey) > 4 {
		maskedKey = c.AdminAPIKey[:2] + "..." + c.AdminAPIKey[len(c.AdminAPIKey)-2:]
	}

	return fmt.Sprintf("AppConfig[Env=%s, Port=%s, AdminKey=%s, DB=%s, Redis=%s, MLService=%s, TLS=%v]",
		c.Env, c.Port, maskedKey, sanitizedDB, c.RedisHost, c.MLServiceURL, c.TLSEnabled)
}
