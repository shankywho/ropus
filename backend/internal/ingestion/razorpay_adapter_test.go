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

func computeHMACSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestRazorpayWebhookAdapter_ValidSignedPayment(t *testing.T) {
	secret := "whsec_razorpay_test_secret_9988"
	_ = os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("RAZORPAY_WEBHOOK_SECRET")

	adapter := ingestion.NewRazorpayWebhookAdapter(nil, nil)

	payload := ingestion.RazorpayWebhookPayload{
		Entity:    "event",
		AccountID: "acc_RazorpayMerchant01",
		Event:     "payment.authorized",
		Contains:  []string{"payment"},
		CreatedAt: 1725200000,
		Payload: ingestion.RazorpayPayloadContent{
			Payment: &ingestion.RazorpayPaymentEntityWrapper{
				Entity: ingestion.RazorpayPaymentEntity{
					ID:            "pay_L83hd82jXm91",
					Entity:        "payment",
					Amount:        48000, // ₹480.00
					Currency:      "INR",
					Status:        "authorized",
					OrderID:       "order_K8391823",
					International: false,
					Method:        "card",
					Email:         "gaurav.kumar@example.com",
					Contact:       "+919876543210",
					Card: &ingestion.RazorpayCardDetails{
						ID:      "card_98127391",
						Entity:  "card",
						Name:    "Gaurav Kumar",
						Last4:   "4321",
						Network: "Visa",
						Type:    "debit",
						Issuer:  "HDFC",
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	sig := computeHMACSignature(bodyBytes, secret)

	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/razorpay", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Razorpay-Signature", sig)
	rr := httptest.NewRecorder()

	adapter.HandleRazorpayWebhook(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp ingestion.RazorpayRiskVerdictResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.PaymentID != "pay_L83hd82jXm91" {
		t.Errorf("expected payment ID pay_L83hd82jXm91, got %s", resp.PaymentID)
	}
	if resp.RecommendedAction != "ALLOW_RECOMMENDATION" && resp.RecommendedAction != "ALLOW" {
		t.Errorf("expected action ALLOW for normal domestic payment, got %s", resp.RecommendedAction)
	}
	if resp.RiskScore > 35 {
		t.Errorf("expected low risk score <= 35, got %d", resp.RiskScore)
	}
}

func TestRazorpayWebhookAdapter_InvalidSignature(t *testing.T) {
	secret := "whsec_razorpay_test_secret_9988"
	_ = os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("RAZORPAY_WEBHOOK_SECRET")

	adapter := ingestion.NewRazorpayWebhookAdapter(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/razorpay", bytes.NewReader([]byte(`{"event":"payment.authorized"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Razorpay-Signature", "deadbeef_invalid_signature_hex")
	rr := httptest.NewRecorder()

	adapter.HandleRazorpayWebhook(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 for invalid signature, got %d", rr.Code)
	}
}

func TestRazorpayWebhookAdapter_MissingSignature(t *testing.T) {
	secret := "whsec_razorpay_test_secret_9988"
	_ = os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("RAZORPAY_WEBHOOK_SECRET")

	adapter := ingestion.NewRazorpayWebhookAdapter(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/razorpay", bytes.NewReader([]byte(`{"event":"payment.authorized"}`)))
	req.Header.Set("Content-Type", "application/json")
	// No X-Razorpay-Signature header
	rr := httptest.NewRecorder()

	adapter.HandleRazorpayWebhook(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 for missing signature, got %d", rr.Code)
	}
}

func TestRazorpayWebhookAdapter_MalformedPayload(t *testing.T) {
	secret := "whsec_razorpay_test_secret_9988"
	_ = os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("RAZORPAY_WEBHOOK_SECRET")

	adapter := ingestion.NewRazorpayWebhookAdapter(nil, nil)

	malformedJSON := []byte(`{"event": "payment.authorized", "payload": { INVALID_JSON }}`)
	sig := computeHMACSignature(malformedJSON, secret)

	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/razorpay", bytes.NewReader(malformedJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Razorpay-Signature", sig)
	rr := httptest.NewRecorder()

	adapter.HandleRazorpayWebhook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected HTTP 400 for malformed payload, got %d", rr.Code)
	}
}

func TestRazorpayWebhookAdapter_IdempotentReplay(t *testing.T) {
	secret := "whsec_razorpay_test_secret_9988"
	_ = os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("RAZORPAY_WEBHOOK_SECRET")

	adapter := ingestion.NewRazorpayWebhookAdapter(nil, nil)

	payload := ingestion.RazorpayWebhookPayload{
		Entity:    "event",
		AccountID: "acc_RazorpayMerchant01",
		Event:     "payment.authorized",
		CreatedAt: 1725200000,
		Payload: ingestion.RazorpayPayloadContent{
			Payment: &ingestion.RazorpayPaymentEntityWrapper{
				Entity: ingestion.RazorpayPaymentEntity{
					ID:       "pay_IdempotentTest_001",
					Amount:   50000,
					Currency: "INR",
					Status:   "authorized",
					Method:   "upi",
					Email:    "buyer@example.com",
				},
			},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	sig := computeHMACSignature(bodyBytes, secret)

	// 1st request (Initial Ingestion)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/webhooks/razorpay", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Razorpay-Signature", sig)
	rr1 := httptest.NewRecorder()
	adapter.HandleRazorpayWebhook(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", rr1.Code)
	}
	var resp1 ingestion.RazorpayRiskVerdictResponse
	_ = json.NewDecoder(rr1.Body).Decode(&resp1)

	// 2nd request (Idempotent Replay)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/webhooks/razorpay", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Razorpay-Signature", sig)
	rr2 := httptest.NewRecorder()
	adapter.HandleRazorpayWebhook(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("replay request failed: %d", rr2.Code)
	}
	if rr2.Header().Get("X-ROPUS-Idempotent-Replayed") != "true" {
		t.Errorf("expected X-ROPUS-Idempotent-Replayed header on replayed event")
	}

	var resp2 ingestion.RazorpayRiskVerdictResponse
	_ = json.NewDecoder(rr2.Body).Decode(&resp2)

	if resp2.DecisionID != resp1.DecisionID {
		t.Errorf("expected identical decision ID on replay: got %s vs %s", resp2.DecisionID, resp1.DecisionID)
	}
}

func TestRazorpayWebhookAdapter_HighRiskEventRejection(t *testing.T) {
	secret := "whsec_razorpay_test_secret_9988"
	_ = os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("RAZORPAY_WEBHOOK_SECRET")

	adapter := ingestion.NewRazorpayWebhookAdapter(nil, nil)

	// High risk payload: ₹1,50,000 international card from disposable email
	payload := ingestion.RazorpayWebhookPayload{
		Entity:    "event",
		AccountID: "acc_RazorpayMerchant01",
		Event:     "payment.authorized",
		CreatedAt: 1725200000,
		Payload: ingestion.RazorpayPayloadContent{
			Payment: &ingestion.RazorpayPaymentEntityWrapper{
				Entity: ingestion.RazorpayPaymentEntity{
					ID:            "pay_HighRiskAttack_001",
					Amount:        15000000, // ₹1,50,000.00
					Currency:      "INR",
					Status:        "authorized",
					International: true,
					Method:        "card",
					Email:         "attacker@guerrillamail.com", // Disposable domain
					Contact:       "+12025550199",
					Card: &ingestion.RazorpayCardDetails{
						Network:       "Visa",
						Type:          "credit",
						Last4:         "9999",
						International: true,
					},
				},
			},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	sig := computeHMACSignature(bodyBytes, secret)

	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/razorpay", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Razorpay-Signature", sig)
	rr := httptest.NewRecorder()

	adapter.HandleRazorpayWebhook(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 verdict response, got %d", rr.Code)
	}

	var resp ingestion.RazorpayRiskVerdictResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	// Must NOT produce an unsafe ALLOW
	if resp.RecommendedAction == "ALLOW" || resp.RecommendedAction == "ALLOW_RECOMMENDATION" {
		t.Fatalf("CRITICAL DEFENSE FAILURE: high-risk payment incorrectly allowed: %+v", resp)
	}

	if resp.RiskScore < 50 {
		t.Errorf("expected elevated risk score >= 50 for international high-value disposable email, got %d", resp.RiskScore)
	}
}
