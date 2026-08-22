package utils

import (
	"crypto/subtle"
	"fmt"
	"regexp"
	"strings"
)

var (
	// validIDRegex allows alphanumeric, underscores, hyphens, and dots.
	validIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	// panRegex identifies potential credit card PAN patterns (13-19 digits).
	panRegex = regexp.MustCompile(`\b\d{4}[ -]?\d{4}[ -]?\d{4}[ -]?\d{1,7}\b`)
	// cvvRegex identifies 3 or 4 digit CVVs.
	cvvRegex = regexp.MustCompile(`(?i)(cvv|cvc|cid)[:=\s]*\d{3,4}`)
)

// SanitizeIdentifier validates that an identifier (model version, candidate ID, job ID)
// does not contain path traversal sequences (..), slashes, null bytes, or malformed characters.
func SanitizeIdentifier(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	if len(trimmed) > 128 {
		return "", fmt.Errorf("identifier exceeds maximum length of 128 characters")
	}
	if strings.Contains(trimmed, "..") || strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") || strings.Contains(trimmed, "\x00") {
		return "", fmt.Errorf("path traversal sequences and path separators are strictly forbidden")
	}
	if !validIDRegex.MatchString(trimmed) {
		return "", fmt.Errorf("identifier contains invalid characters (must match [a-zA-Z0-9_\\-\\.]+)")
	}
	return trimmed, nil
}

// ConstantTimeCompare compares two strings in constant time to prevent timing attacks.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SanitizeLogMessage masks any inadvertent sensitive data (such as card numbers or CVVs) from logs.
func SanitizeLogMessage(msg string) string {
	masked := panRegex.ReplaceAllStringFunc(msg, func(pan string) string {
		clean := strings.ReplaceAll(strings.ReplaceAll(pan, " ", ""), "-", "")
		if len(clean) >= 10 {
			return clean[:4] + strings.Repeat("*", len(clean)-8) + clean[len(clean)-4:]
		}
		return strings.Repeat("*", len(clean))
	})
	masked = cvvRegex.ReplaceAllString(masked, "$1:***")
	return masked
}
