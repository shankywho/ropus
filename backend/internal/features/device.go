package features

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	// MaxFingerprintLength defines the maximum allowable byte/character length for raw device fingerprints.
	MaxFingerprintLength = 256

	// MinFingerprintLength defines the minimum length for a non-trivial device fingerprint.
	MinFingerprintLength = 8
)

// DeviceStatus defines the classification outcome of incoming device telemetry.
type DeviceStatus string

const (
	DeviceStatusValid     DeviceStatus = "VALID"
	DeviceStatusMissing   DeviceStatus = "MISSING"
	DeviceStatusEmpty     DeviceStatus = "EMPTY"
	DeviceStatusOversized DeviceStatus = "OVERSIZED"
	DeviceStatusInvalid   DeviceStatus = "INVALID_CONTROL_CHARS"
)

var (
	ErrDeviceFingerprintMissing   = errors.New("device_fingerprint is missing")
	ErrDeviceFingerprintEmpty     = errors.New("device_fingerprint is empty after trimming whitespace")
	ErrDeviceFingerprintOversized = errors.New("device_fingerprint exceeds maximum allowed length of 256 characters")
	ErrDeviceFingerprintInvalid   = errors.New("device_fingerprint contains invalid control characters or embedded null bytes")
)

// DeviceIdentity represents a sanitized, tenant-isolated device identifier.
type DeviceIdentity struct {
	TenantID             string       `json:"tenant_id"`
	DeviceID             string       `json:"device_id"` // SHA256(tenant_id + ":" + canonical_fingerprint)
	CanonicalFingerprint string       `json:"-"`         // Hidden from JSON serialization to prevent accidental leakage
	RawFingerprint       string       `json:"-"`         // Hidden from JSON serialization
	Status               DeviceStatus `json:"status"`
	IsValid              bool         `json:"is_valid"`
	ValidationError      string       `json:"validation_error,omitempty"`
}

// String implements fmt.Stringer to ensure raw fingerprints are NEVER exposed in logs.
func (d DeviceIdentity) String() string {
	maskedDevID := d.DeviceID
	if len(maskedDevID) > 16 {
		maskedDevID = maskedDevID[:8] + "..." + maskedDevID[len(maskedDevID)-8:]
	}
	return fmt.Sprintf("DeviceIdentity{TenantID: %s, DeviceID: %s, Status: %s, IsValid: %t}",
		d.TenantID, maskedDevID, d.Status, d.IsValid)
}

// MaskedFingerprint returns a privacy-preserving representation (e.g., "fp_****_4242").
func (d DeviceIdentity) MaskedFingerprint() string {
	if !d.IsValid || len(d.CanonicalFingerprint) == 0 {
		return "fp_none"
	}
	fp := d.CanonicalFingerprint
	if len(fp) <= 8 {
		return "fp_****"
	}
	return fmt.Sprintf("fp_****_%s", fp[len(fp)-4:])
}

// HashDeviceID computes the deterministic tenant-isolated SHA-256 device identifier.
// Formula: SHA256(tenant_id + ":" + canonical_fingerprint)
func HashDeviceID(tenantID, canonicalFingerprint string) string {
	h := sha256.New()
	h.Write([]byte(tenantID))
	h.Write([]byte(":"))
	h.Write([]byte(canonicalFingerprint))
	return hex.EncodeToString(h.Sum(nil))
}

// ParseDeviceIdentity canonicalizes, validates, and hashes untrusted device telemetry.
func ParseDeviceIdentity(tenantID, rawFingerprint string) DeviceIdentity {
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000001"
	}

	// 1. Check for nil/empty input
	if rawFingerprint == "" {
		return DeviceIdentity{
			TenantID:        tenantID,
			DeviceID:        "",
			Status:          DeviceStatusMissing,
			IsValid:         false,
			ValidationError: ErrDeviceFingerprintMissing.Error(),
		}
	}

	// 2. Length check on raw input (reject oversized before processing)
	if len(rawFingerprint) > MaxFingerprintLength {
		return DeviceIdentity{
			TenantID:        tenantID,
			DeviceID:        "",
			RawFingerprint:  "", // Do not retain massive hostile payloads
			Status:          DeviceStatusOversized,
			IsValid:         false,
			ValidationError: ErrDeviceFingerprintOversized.Error(),
		}
	}

	// 3. Trim leading and trailing whitespace
	trimmed := strings.TrimSpace(rawFingerprint)
	if len(trimmed) == 0 {
		return DeviceIdentity{
			TenantID:        tenantID,
			DeviceID:        "",
			Status:          DeviceStatusEmpty,
			IsValid:         false,
			ValidationError: ErrDeviceFingerprintEmpty.Error(),
		}
	}

	// 4. Validate UTF-8 integrity and check for forbidden control characters / embedded NULL bytes
	if !utf8.ValidString(trimmed) {
		return DeviceIdentity{
			TenantID:        tenantID,
			DeviceID:        "",
			Status:          DeviceStatusInvalid,
			IsValid:         false,
			ValidationError: ErrDeviceFingerprintInvalid.Error(),
		}
	}

	for _, r := range trimmed {
		// Reject control characters (C0, C1, DEL), embedded null bytes (\x00), and non-printable control runes
		if unicode.IsControl(r) || r == 0 || r == '\uFFFD' {
			return DeviceIdentity{
				TenantID:        tenantID,
				DeviceID:        "",
				Status:          DeviceStatusInvalid,
				IsValid:         false,
				ValidationError: ErrDeviceFingerprintInvalid.Error(),
			}
		}
	}

	// 5. Generate deterministic tenant-isolated SHA-256 hash
	deviceID := HashDeviceID(tenantID, trimmed)

	return DeviceIdentity{
		TenantID:             tenantID,
		DeviceID:             deviceID,
		CanonicalFingerprint: trimmed,
		RawFingerprint:       rawFingerprint,
		Status:               DeviceStatusValid,
		IsValid:              true,
		ValidationError:      "",
	}
}
