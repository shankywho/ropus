package graphsage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var (
	panRegex  = regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`)
	cvvRegex  = regexp.MustCompile(`\b\d{3,4}\b`)
	ssnRegex  = regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$|^\d{9}$`)
	prohibitedKeys = map[string]bool{
		"raw_pan": true, "pan": true, "card_number": true, "cvv": true, "cvv2": true, "cvc": true,
		"pin": true, "ssn": true, "social_security": true, "password": true, "cleartext_password": true,
		"bank_account_number": true, "iban": true, "routing_number": true,
	}
)

// PrivacySanitizer enforces data protection rules in Go.
type PrivacySanitizer struct {
	salt []byte
}

// NewPrivacySanitizer creates a privacy sanitizer.
func NewPrivacySanitizer(salt string) *PrivacySanitizer {
	if salt == "" {
		salt = "ropus_salt_prod_2026"
	}
	return &PrivacySanitizer{salt: []byte(salt)}
}

// TokenizeIdentifier produces a salted tokenized hash for entity identifiers.
func (p *PrivacySanitizer) TokenizeIdentifier(prefix, rawID string) string {
	hasher := sha256.New()
	hasher.Write(p.salt)
	hasher.Write([]byte(rawID))
	hashHex := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("%s_%s", prefix, hashHex[:12])
}

// SanitizePayload inspects and strips prohibited fields.
func (p *PrivacySanitizer) SanitizePayload(payload map[string]interface{}) (map[string]interface{}, []string, bool) {
	sanitized := make(map[string]interface{})
	var violations []string
	isQuarantined := false

	for k, v := range payload {
		kLower := strings.ToLower(k)
		if prohibitedKeys[kLower] {
			violations = append(violations, fmt.Sprintf("Prohibited field key: %s", k))
			isQuarantined = true
			continue
		}

		if strVal, ok := v.(string); ok {
			cleanStr := strings.ReplaceAll(strings.ReplaceAll(strVal, " ", ""), "-", "")
			if len(cleanStr) >= 13 && len(cleanStr) <= 19 && panRegex.MatchString(cleanStr) {
				violations = append(violations, fmt.Sprintf("Raw PAN pattern detected in field: %s", k))
				isQuarantined = true
				continue
			}
			if ssnRegex.MatchString(strVal) {
				violations = append(violations, fmt.Sprintf("SSN pattern detected in field: %s", k))
				isQuarantined = true
				continue
			}
		}

		sanitized[k] = v
	}

	return sanitized, violations, isQuarantined
}
