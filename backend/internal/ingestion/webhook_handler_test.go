package ingestion_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shankywho/ropus/backend/internal/ingestion"
)

func TestWebhookHandler_HMACVerification(t *testing.T) {
	secret := "test_webhook_secret_key"
	_ = os.Setenv("WEBHOOK_SECRET", secret)
	defer os.Unsetenv("WEBHOOK_SECRET")

	handler := ingestion.NewWebhookHandler(nil)

	payload := map[string]interface{}{
		"event_id":   "evt_test_001",
		"event_type": "dispute.opened",
		"data": map[string]interface{}{
			"transaction_id": "txn_dispute_123",
			"amount":         15000.0,
			"reason":         "unrecognized_transaction",
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	// Compute valid HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(bodyBytes)
	validSig := hex.EncodeToString(mac.Sum(nil))

	// Case 1: Valid Signature
	reqValid := httptest.NewRequest(http.MethodPost, "/webhooks/provider", bytes.NewReader(bodyBytes))
	reqValid.Header.Set("Content-Type", "application/json")
	reqValid.Header.Set("X-Signature", validSig)
	rrValid := httptest.NewRecorder()

	handler.HandleProviderWebhook(rrValid, reqValid)
	if rrValid.Code != http.StatusOK {
		t.Fatalf("expected status 200 with valid signature, got %d: body=%s", rrValid.Code, rrValid.Body.String())
	}

	// Case 2: Invalid Signature
	reqInvalid := httptest.NewRequest(http.MethodPost, "/webhooks/provider", bytes.NewReader(bodyBytes))
	reqInvalid.Header.Set("Content-Type", "application/json")
	reqInvalid.Header.Set("X-Signature", "invalid_hex_signature")
	rrInvalid := httptest.NewRecorder()

	handler.HandleProviderWebhook(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 with invalid signature, got %d", rrInvalid.Code)
	}
}
