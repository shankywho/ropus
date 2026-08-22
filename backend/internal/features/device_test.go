package features_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shankywho/ropus/backend/internal/features"
)

func TestParseDeviceIdentity(t *testing.T) {
	tenantA := "00000000-0000-0000-0000-000000000001"
	tenantB := "00000000-0000-0000-0000-000000000002"

	t.Run("1. Empty string rejection", func(t *testing.T) {
		id := features.ParseDeviceIdentity(tenantA, "")
		if id.IsValid {
			t.Errorf("expected IsValid=false for empty string")
		}
		if id.Status != features.DeviceStatusMissing {
			t.Errorf("expected Status=%s, got %s", features.DeviceStatusMissing, id.Status)
		}
		if id.DeviceID != "" {
			t.Errorf("expected empty DeviceID, got %s", id.DeviceID)
		}
	})

	t.Run("2. Whitespace-only string rejection", func(t *testing.T) {
		id := features.ParseDeviceIdentity(tenantA, "   \t \n \r  ")
		if id.IsValid {
			t.Errorf("expected IsValid=false for whitespace-only string")
		}
		if id.Status != features.DeviceStatusInvalid && id.Status != features.DeviceStatusEmpty {
			t.Errorf("expected Status to be EMPTY or INVALID, got %s", id.Status)
		}
	})

	t.Run("3. Standard FingerprintJS 32-character hex identifier", func(t *testing.T) {
		fp := "9c8e1a2b3c4d5e6f708192a3b4c5d6e7"
		id := features.ParseDeviceIdentity(tenantA, fp)
		if !id.IsValid {
			t.Fatalf("expected IsValid=true, got error: %s", id.ValidationError)
		}
		if id.Status != features.DeviceStatusValid {
			t.Errorf("expected Status=%s, got %s", features.DeviceStatusValid, id.Status)
		}
		if len(id.DeviceID) != 64 {
			t.Errorf("expected 64-char SHA256 hex DeviceID, got len=%d", len(id.DeviceID))
		}
		if id.CanonicalFingerprint != fp {
			t.Errorf("expected CanonicalFingerprint=%s, got %s", fp, id.CanonicalFingerprint)
		}
	})

	t.Run("4. 64-character identifier", func(t *testing.T) {
		fp := strings.Repeat("a1b2", 16) // 64 chars
		id := features.ParseDeviceIdentity(tenantA, fp)
		if !id.IsValid {
			t.Fatalf("expected IsValid=true for 64-char string")
		}
		if len(id.DeviceID) != 64 {
			t.Errorf("expected 64-char hex DeviceID, got %s", id.DeviceID)
		}
	})

	t.Run("5. 256-character boundary identifier (Maximum Allowed)", func(t *testing.T) {
		fp := strings.Repeat("x", 256)
		id := features.ParseDeviceIdentity(tenantA, fp)
		if !id.IsValid {
			t.Fatalf("expected IsValid=true for exact 256-char string")
		}
		if len(id.CanonicalFingerprint) != 256 {
			t.Errorf("expected len=256, got %d", len(id.CanonicalFingerprint))
		}
	})

	t.Run("6. 257-character identifier (Oversized Rejection)", func(t *testing.T) {
		fp := strings.Repeat("x", 257)
		id := features.ParseDeviceIdentity(tenantA, fp)
		if id.IsValid {
			t.Errorf("expected IsValid=false for 257-char string")
		}
		if id.Status != features.DeviceStatusOversized {
			t.Errorf("expected Status=%s, got %s", features.DeviceStatusOversized, id.Status)
		}
	})

	t.Run("7. Extremely large payload (10,000 chars DOS attempt)", func(t *testing.T) {
		fp := strings.Repeat("payload_spam_", 1000)
		id := features.ParseDeviceIdentity(tenantA, fp)
		if id.IsValid {
			t.Errorf("expected IsValid=false for huge payload")
		}
		if id.Status != features.DeviceStatusOversized {
			t.Errorf("expected Status=%s, got %s", features.DeviceStatusOversized, id.Status)
		}
	})

	t.Run("8. Embedded NULL byte injection", func(t *testing.T) {
		fp := "fp_trusted_device\x00_malicious_suffix"
		id := features.ParseDeviceIdentity(tenantA, fp)
		if id.IsValid {
			t.Errorf("expected IsValid=false for embedded null byte")
		}
		if id.Status != features.DeviceStatusInvalid {
			t.Errorf("expected Status=%s, got %s", features.DeviceStatusInvalid, id.Status)
		}
	})

	t.Run("9. Control characters (Newline, Carriage Return, Tab, Escape)", func(t *testing.T) {
		badInputs := []string{
			"device\nid",
			"device\rid",
			"device\tid",
			"device\x1b[31mid",
			"device\x7fid",
		}
		for _, input := range badInputs {
			id := features.ParseDeviceIdentity(tenantA, input)
			if id.IsValid {
				t.Errorf("expected IsValid=false for control char input %q", input)
			}
			if id.Status != features.DeviceStatusInvalid {
				t.Errorf("expected Status=%s for %q, got %s", features.DeviceStatusInvalid, input, id.Status)
			}
		}
	})

	t.Run("10. SQL-looking input without control characters", func(t *testing.T) {
		// Valid printable ASCII, safely hashed without string interpolation
		fp := "' OR 1=1;--"
		id := features.ParseDeviceIdentity(tenantA, fp)
		if !id.IsValid {
			t.Fatalf("expected printable string to be parsed safely, got error: %s", id.ValidationError)
		}
		if len(id.DeviceID) != 64 {
			t.Errorf("expected 64-char hash")
		}
	})

	t.Run("11. HTML/script-looking input", func(t *testing.T) {
		fp := "<script>alert('xss')</script>"
		id := features.ParseDeviceIdentity(tenantA, fp)
		if !id.IsValid {
			t.Fatalf("expected printable string to be safely accepted and hashed, got error: %s", id.ValidationError)
		}
		if len(id.DeviceID) != 64 {
			t.Errorf("expected 64-char hash")
		}
	})

	t.Run("12. JSON-looking input", func(t *testing.T) {
		fp := `{"vendor":"Apple","model":"MacBookPro"}`
		id := features.ParseDeviceIdentity(tenantA, fp)
		if !id.IsValid {
			t.Fatalf("expected valid JSON string to parse safely, got error: %s", id.ValidationError)
		}
	})

	t.Run("13. Unicode printable characters", func(t *testing.T) {
		fp := "📱iphone_pro_日本語_15"
		id := features.ParseDeviceIdentity(tenantA, fp)
		if !id.IsValid {
			t.Fatalf("expected valid Unicode string to parse safely, got: %s", id.ValidationError)
		}
		if len(id.DeviceID) != 64 {
			t.Errorf("expected 64-char hex hash")
		}
	})

	t.Run("14. Leading and trailing whitespace trimming", func(t *testing.T) {
		raw := "   fp_client_macbook_v15   "
		id := features.ParseDeviceIdentity(tenantA, raw)
		if !id.IsValid {
			t.Fatalf("expected IsValid=true after trimming")
		}
		if id.CanonicalFingerprint != "fp_client_macbook_v15" {
			t.Errorf("expected trimmed value 'fp_client_macbook_v15', got %q", id.CanonicalFingerprint)
		}
		// Hash should match hash of trimmed string
		expectedHash := features.HashDeviceID(tenantA, "fp_client_macbook_v15")
		if id.DeviceID != expectedHash {
			t.Errorf("expected DeviceID=%s, got %s", expectedHash, id.DeviceID)
		}
	})

	t.Run("15. Determinism: Same Tenant + Same Fingerprint", func(t *testing.T) {
		fp := "device_visitor_hash_12345"
		id1 := features.ParseDeviceIdentity(tenantA, fp)
		id2 := features.ParseDeviceIdentity(tenantA, fp)
		if id1.DeviceID != id2.DeviceID {
			t.Errorf("expected deterministic DeviceIDs, got %s vs %s", id1.DeviceID, id2.DeviceID)
		}
	})

	t.Run("16. Tenant Isolation: Same Fingerprint Across Tenant A vs Tenant B", func(t *testing.T) {
		fp := "shared_hardware_visitor_hash_9999"
		idA := features.ParseDeviceIdentity(tenantA, fp)
		idB := features.ParseDeviceIdentity(tenantB, fp)

		if !idA.IsValid || !idB.IsValid {
			t.Fatalf("expected both identities to be valid")
		}
		if idA.DeviceID == idB.DeviceID {
			t.Errorf("CRITICAL SECURITY FLAW: Cross-tenant device IDs collided! TenantA=%s, TenantB=%s",
				idA.DeviceID, idB.DeviceID)
		}
	})

	t.Run("17. Collision resistance: Different fingerprints produce different DeviceIDs", func(t *testing.T) {
		id1 := features.ParseDeviceIdentity(tenantA, "device_fingerprint_alpha")
		id2 := features.ParseDeviceIdentity(tenantA, "device_fingerprint_beta")
		if id1.DeviceID == id2.DeviceID {
			t.Errorf("expected different DeviceIDs for different inputs")
		}
	})

	t.Run("18. Stringer privacy protection: Raw fingerprint is not exposed", func(t *testing.T) {
		secretFP := "super_secret_browser_hardware_uuid_9999"
		id := features.ParseDeviceIdentity(tenantA, secretFP)
		str := fmt.Sprintf("%v", id)
		if strings.Contains(str, secretFP) {
			t.Errorf("CRITICAL PRIVACY FLAW: String() leaked raw fingerprint! Got: %s", str)
		}
	})

	t.Run("19. Masked fingerprint format", func(t *testing.T) {
		fp := "fp_client_macbook_v15_4242"
		id := features.ParseDeviceIdentity(tenantA, fp)
		masked := id.MaskedFingerprint()
		if masked != "fp_****_4242" {
			t.Errorf("expected 'fp_****_4242', got %s", masked)
		}
	})
}
