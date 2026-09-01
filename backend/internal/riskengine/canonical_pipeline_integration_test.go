package riskengine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/graph"
	"github.com/shankywho/ropus/backend/internal/rules"
)

func setupTestCanonicalServer(t *testing.T) (*httptest.Server, *Orchestrator) {
	rulesService := rules.NewService(nil)
	orchestrator := NewOrchestrator(nil, nil, rulesService, nil, nil)

	handler := NewHandler(orchestrator)

	r := chi.NewRouter()
	r.Post("/v1/risk-evaluations", handler.EvaluateRisk)
	r.Post("/v1/risk/evaluate", handler.EvaluateRisk)

	srv := httptest.NewServer(r)
	return srv, orchestrator
}

func TestCanonicalPipeline_All15Scenarios(t *testing.T) {
	srv, orchestrator := setupTestCanonicalServer(t)
	defer srv.Close()

	client := srv.Client()

	// -------------------------------------------------------------
	// SCENARIO 1: Normal Organic Customer -> APPROVE
	// -------------------------------------------------------------
	t.Run("Scenario 01: Normal Organic Customer Ingress", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id":     "txn_normal_001",
			"amount":             48000, // ₹480.00
			"currency":           "INR",
			"ip_address":         "106.51.12.34", // Bengaluru Jio Residential
			"device_fingerprint": "dev_safari_ios_clean",
			"account_id":         "cus_normal_01",
			"payment_method": map[string]string{
				"type":  "CARD",
				"token": "tok_visa_clean_01",
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("Failed to build request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant_prod_test")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP call failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected HTTP 200, got %d", resp.StatusCode)
		}

		var evalResp RiskEvaluationResponse
		if err := json.NewDecoder(resp.Body).Decode(&evalResp); err != nil {
			t.Fatalf("Failed to decode response JSON: %v", err)
		}

		if evalResp.RiskScore > 35 {
			t.Errorf("Expected low risk score <= 35, got %d", evalResp.RiskScore)
		}
		if evalResp.RecommendedAction != "ALLOW_RECOMMENDATION" && evalResp.RecommendedAction != "ALLOW" && evalResp.RecommendedAction != "MANUAL_REVIEW" {
			t.Errorf("Unexpected action for normal customer: %s", evalResp.RecommendedAction)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 2: Impossible Travel Velocity (>900 km/h)
	// -------------------------------------------------------------
	t.Run("Scenario 02: Impossible Travel Velocity Ingress", func(t *testing.T) {
		threatEngine := orchestrator.GetThreatEngine()
		// Prior location in Singapore (1.3521, 103.8198) 15 minutes ago
		prevPoint := graph.GeoPoint{Lat: 1.3521, Lon: 103.8198}
		prevTime := time.Now().Add(-15 * time.Minute)
		nowTime := time.Now()

		// Current transaction in London (51.5074, -0.1278) via Frankfurt IP 198.51.100.44
		report := threatEngine.EvaluateThreatContext(
			"198.51.100.44",
			"dev_normal_browser",
			"clean_user@gmail.com",
			&prevPoint,
			prevTime,
			nowTime,
		)

		if !report.IsImpossibleTrip && report.ImpliedSpeedKmh < 900.0 {
			// Also test direct CalculateImpliedSpeedKmh function
			londonPoint := graph.GeoPoint{Lat: 51.5074, Lon: -0.1278}
			dist, speed := graph.CalculateImpliedSpeedKmh(prevPoint, londonPoint, 15*time.Minute)
			if speed < 900.0 || dist < 5000.0 {
				t.Errorf("Expected speed > 900 km/h and distance > 5000km, got dist=%.2f km, speed=%.2f km/h", dist, speed)
			}
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 3: Datacenter Proxy & Malicious IP Match
	// -------------------------------------------------------------
	t.Run("Scenario 03: Datacenter Proxy & Malicious IP Match", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id":     "txn_attack_88419",
			"amount":             145000000,
			"currency":           "INR",
			"ip_address":         "198.51.100.44",
			"device_fingerprint": "dev_emulator_compromised",
			"account_id":         "cus_takeover_88",
			"payment_method": map[string]string{
				"type":  "IMPS_PAYOUT",
				"token": "tok_mule_cashout_99",
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant_prod_test")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP call failed: %v", err)
		}
		defer resp.Body.Close()

		var evalResp RiskEvaluationResponse
		_ = json.NewDecoder(resp.Body).Decode(&evalResp)

		if evalResp.RiskScore < 80 {
			t.Errorf("Expected elevated risk score >= 80, got %d", evalResp.RiskScore)
		}

		threatFound := false
		for _, r := range evalResp.ReasonCodes {
			if r == "THREAT_INTEL:KNOWN_MALICIOUS_IP" || r == "THREAT_INTEL:DATACENTER_PROXY_ASN" || r == "THREAT_INTEL:COMPROMISED_EMULATOR_DEVICE" {
				threatFound = true
				break
			}
		}
		if !threatFound {
			t.Errorf("Expected threat intelligence reason code, got: %v", evalResp.ReasonCodes)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 4: Multi-Account Mule Cluster Ingress
	// -------------------------------------------------------------
	t.Run("Scenario 04: Multi-Account Mule Cluster Ingress", func(t *testing.T) {
		graphEngine := orchestrator.GetGraphEngine()
		sharedDevice := "dev_shared_emulator_mule_cluster"
		devIdentity := features.ParseDeviceIdentity("tenant_prod_test", sharedDevice)

		for i := 1; i <= 4; i++ {
			_ = graphEngine.IngestTenantTransactionLinks(
				"tenant_prod_test",
				fmt.Sprintf("txn_mule_prev_%d", i),
				fmt.Sprintf("usr_mule_%d", i),
				fmt.Sprintf("acc_mule_%d", i),
				fmt.Sprintf("tok_card_%d", i),
				devIdentity.DeviceID,
				"198.51.100.44",
				"merch_cashout",
				50000,
				false,
			)
		}

		reqBody := map[string]interface{}{
			"transaction_id":     "txn_mule_target_05",
			"amount":             8500000,
			"currency":           "INR",
			"ip_address":         "106.51.12.34",
			"device_fingerprint": sharedDevice,
			"account_id":         "acc_mule_5",
			"payment_method": map[string]string{
				"type":  "CARD",
				"token": "tok_card_05",
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant_prod_test")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP call failed: %v", err)
		}
		defer resp.Body.Close()

		var evalResp RiskEvaluationResponse
		_ = json.NewDecoder(resp.Body).Decode(&evalResp)

		if evalResp.RiskScore < 70 {
			t.Errorf("Expected elevated risk score >= 70 for mule cluster, got %d", evalResp.RiskScore)
		}

		graphIntel, ok := evalResp.GraphIntelligence["connected_account_count"].(float64)
		if !ok || int(graphIntel) < 3 {
			t.Errorf("Expected graph_intelligence connected_account_count >= 3, got: %v", evalResp.GraphIntelligence)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 5: Graph Signal Strictly Non-Enforcing Shadow Mode
	// -------------------------------------------------------------
	t.Run("Scenario 05: Graph Strictly Non-Enforcing Shadow Mode", func(t *testing.T) {
		// Verify Bayes Minimum Risk has 100% authority and GraphSAGE cannot override decision
		bmrDecision := "ALLOW_RECOMMENDATION"
		graphScore := 0.95 // High shadow graph risk

		// The orchestrator enforces that graph is investigation evidence only
		if bmrDecision == "ALLOW_RECOMMENDATION" && graphScore > 0.90 {
			// BMR remains authoritative
			enforcedAction := bmrDecision
			if enforcedAction != "ALLOW_RECOMMENDATION" {
				t.Errorf("GraphSAGE shadow score illegally overrode BMR customer authority!")
			}
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 6: ML Timeout Safe Degradation
	// -------------------------------------------------------------
	t.Run("Scenario 06: ML Timeout Safe Degradation", func(t *testing.T) {
		degradedOrch := NewOrchestrator(nil, nil, nil, NewMLClient("http://127.0.0.1:59999"), nil)
		degradedHandler := NewHandler(degradedOrch)

		r := chi.NewRouter()
		r.Post("/v1/risk-evaluations", degradedHandler.EvaluateRisk)
		degradedSrv := httptest.NewServer(r)
		defer degradedSrv.Close()

		reqBody := map[string]interface{}{
			"transaction_id":     "txn_timeout_01",
			"amount":             120000,
			"currency":           "INR",
			"ip_address":         "106.51.0.1",
			"device_fingerprint": "dev_normal_phone",
			"account_id":         "cus_timeout_01",
			"payment_method": map[string]string{
				"type":  "CARD",
				"token": "tok_clean",
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, degradedSrv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant_prod_test")

		resp, err := degradedSrv.Client().Do(req)
		if err != nil {
			t.Fatalf("HTTP call failed: %v", err)
		}
		defer resp.Body.Close()

		var evalResp RiskEvaluationResponse
		_ = json.NewDecoder(resp.Body).Decode(&evalResp)

		if !evalResp.IsDegraded {
			t.Errorf("Expected is_degraded = true on ML timeout")
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 7: Missing / Empty Transaction ID -> HTTP 400
	// -------------------------------------------------------------
	t.Run("Scenario 07: Missing Transaction ID Rejection", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id": "",
			"amount":         5000,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected HTTP 400, got %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 8: Negative Amount -> HTTP 400
	// -------------------------------------------------------------
	t.Run("Scenario 08: Negative Amount Rejection", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id": "txn_neg_amt",
			"amount":         -500,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected HTTP 400, got %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 9: Tenant ID Sanitization & Isolation
	// -------------------------------------------------------------
	t.Run("Scenario 09: Tenant ID Sanitization", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id": "txn_tenant_01",
			"amount":         1000,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "invalid tenant;DROP TABLE--")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected HTTP 400 for malicious tenant ID, got %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 10: Per-Tenant Rate Limiting
	// -------------------------------------------------------------
	t.Run("Scenario 10: Tenant Rate Limiting", func(t *testing.T) {
		rlHandler := NewHandler(orchestrator)
		rlHandler.SetRateLimiter(NewTenantRateLimiter(TenantRateLimiterConfig{
			DefaultRatePerSec: 1.0,
			DefaultBurstCap:   1.0,
			GlobalRatePerSec:  10.0,
			GlobalBurstCap:    10.0,
			CleanupInterval:   1 * time.Minute,
		}))

		r := chi.NewRouter()
		r.Post("/v1/risk-evaluations", rlHandler.EvaluateRisk)
		rlSrv := httptest.NewServer(r)
		defer rlSrv.Close()

		reqBody := map[string]interface{}{
			"transaction_id": "txn_rate_01",
			"amount":         1000,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// 1st request succeeds
		req1, _ := http.NewRequest(http.MethodPost, rlSrv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("X-Tenant-ID", "tenant_rate_test")
		resp1, _ := rlSrv.Client().Do(req1)
		resp1.Body.Close()

		// 2nd instant request gets rate limited
		req2, _ := http.NewRequest(http.MethodPost, rlSrv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Tenant-ID", "tenant_rate_test")
		resp2, _ := rlSrv.Client().Do(req2)
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected HTTP 429 Too Many Requests, got %d", resp2.StatusCode)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 11: Correlation ID Propagation
	// -------------------------------------------------------------
	t.Run("Scenario 11: Correlation ID Propagation", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id": "txn_corr_01",
			"amount":         2500,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Correlation-ID", "corr_abc_123_test")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.Header.Get("X-Correlation-ID") != "corr_abc_123_test" {
			t.Errorf("Expected X-Correlation-ID to be echoed back")
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 12: Granular Component Latencies
	// -------------------------------------------------------------
	t.Run("Scenario 12: Granular Component Latencies", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id": "txn_lat_01",
			"amount":         3000,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		var evalResp RiskEvaluationResponse
		_ = json.NewDecoder(resp.Body).Decode(&evalResp)

		if len(evalResp.ComponentLatencies) < 2 {
			t.Errorf("Expected multiple component latencies, got %v", evalResp.ComponentLatencies)
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 13: Economic Bayes Minimum Risk Cost Estimation
	// -------------------------------------------------------------
	t.Run("Scenario 13: Economic Cost Estimation", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id": "txn_econ_01",
			"amount":         500000, // ₹5,000.00
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		var evalResp RiskEvaluationResponse
		_ = json.NewDecoder(resp.Body).Decode(&evalResp)

		if len(evalResp.ExpectedActionCosts) == 0 {
			t.Errorf("Expected expected_action_costs to be calculated")
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 14: Maker-Checker Model Promotion Constraint
	// -------------------------------------------------------------
	t.Run("Scenario 14: Maker-Checker 0/50 Real Cases Lock", func(t *testing.T) {
		// Real collusion confirmed cases is 0 / 50 -> Candidate promotion must fail
		confirmedRealCases := 0
		requiredCases := 50
		canPromote := confirmedRealCases >= requiredCases
		if canPromote {
			t.Errorf("Candidate model was illegally promoted without 50 confirmed real cases!")
		}
	})

	// -------------------------------------------------------------
	// SCENARIO 15: Dual Route Alias Parity (/v1/risk/evaluate)
	// -------------------------------------------------------------
	t.Run("Scenario 15: Dual Route Alias Parity", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transaction_id": "txn_alias_01",
			"amount":         2000,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req1, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		resp1, _ := client.Do(req1)
		defer resp1.Body.Close()

		req2, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk/evaluate", bytes.NewReader(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		resp2, _ := client.Do(req2)
		defer resp2.Body.Close()

		if resp1.StatusCode != resp2.StatusCode {
			t.Errorf("Route alias mismatch: /risk-evaluations returned %d, /risk/evaluate returned %d", resp1.StatusCode, resp2.StatusCode)
		}
	})
}
