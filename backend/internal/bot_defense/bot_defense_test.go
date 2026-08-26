package bot_defense

import (
	"testing"
	"time"
)

func TestBotDefense_NormalHumanTraffic(t *testing.T) {
	engine := NewBotDefenseEngine()

	// Simulate human interactions with realistic intervals (3s, 5s, 12s)
	baseTime := time.Now().UTC().Add(-30 * time.Second)

	intervals := []time.Duration{0, 3 * time.Second, 8 * time.Second, 20 * time.Second}
	var lastResult *BotDefenseResult

	for i, offset := range intervals {
		ctx := &BotDefenseContext{
			TenantID:          "org_tenant_1",
			PlanTier:          "ENTERPRISE",
			TransactionID:     "tx_human_" + string(rune('a'+i)),
			AccountID:         "usr_human_alice",
			DeviceFingerprint: "fp_iphone_15_pro",
			IPAddress:         "192.0.2.10",
			UserAgent:         "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1",
			CardHash:          "card_hash_alice_9988",
			Amount:            125.50,
			Currency:          "USD",
			Timestamp:         baseTime.Add(offset),
		}
		lastResult = engine.Evaluate(ctx)
	}

	if lastResult.AutomationRiskScore >= 0.30 {
		t.Fatalf("Expected normal human traffic to have low automation score (<0.30), got %.2f", lastResult.AutomationRiskScore)
	}
	if lastResult.RiskLevel != RiskLevelLow {
		t.Fatalf("Expected RiskLevelLow, got %s", lastResult.RiskLevel)
	}
	if lastResult.RecommendedAction != ActionAllow {
		t.Fatalf("Expected ActionAllow, got %s", lastResult.RecommendedAction)
	}
}

func TestBotDefense_DeterministicScriptedLoop(t *testing.T) {
	engine := NewBotDefenseEngine()

	// Simulate exact 500ms fixed timer bot loop
	baseTime := time.Now().UTC().Add(-10 * time.Second)

	var lastResult *BotDefenseResult
	for i := 0; i < 6; i++ {
		ctx := &BotDefenseContext{
			TenantID:          "org_tenant_1",
			PlanTier:          "GROWTH",
			TransactionID:     "tx_bot_loop_" + string(rune('0'+i)),
			AccountID:         "usr_bot_victim",
			DeviceFingerprint: "fp_scripted_bot_01",
			IPAddress:         "198.51.100.50",
			UserAgent:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			CardHash:          "card_hash_bot_1",
			Amount:            49.99,
			Currency:          "USD",
			Timestamp:         baseTime.Add(time.Duration(i*500) * time.Millisecond),
		}
		lastResult = engine.Evaluate(ctx)
	}

	if !lastResult.Signals.IsDeterministicCadence {
		t.Fatalf("Expected IsDeterministicCadence to be true for fixed 500ms loop")
	}
	if lastResult.AutomationRiskScore < 0.30 {
		t.Fatalf("Expected elevated automation risk score, got %.2f", lastResult.AutomationRiskScore)
	}
}

func TestBotDefense_CardTestingProbing(t *testing.T) {
	engine := NewBotDefenseEngine()

	baseTime := time.Now().UTC().Add(-5 * time.Minute)

	// Simulate attacker probing 4 different card hashes with micro-amounts ($1.00 - $2.50) from single device
	var lastResult *BotDefenseResult
	for i := 0; i < 4; i++ {
		ctx := &BotDefenseContext{
			TenantID:          "org_tenant_1",
			PlanTier:          "GROWTH",
			TransactionID:     "tx_carding_" + string(rune('0'+i)),
			AccountID:         "usr_carding_target",
			DeviceFingerprint: "fp_carding_emulator",
			IPAddress:         "203.0.113.88",
			UserAgent:         "Mozilla/5.0",
			CardHash:          "card_stolen_token_" + string(rune('A'+i)),
			Amount:            1.50, // Micro-amount
			Currency:          "USD",
			Timestamp:         baseTime.Add(time.Duration(i*10) * time.Second),
		}
		lastResult = engine.Evaluate(ctx)
	}

	if !lastResult.Signals.IsCardTestingProbing {
		t.Fatalf("Expected IsCardTestingProbing to be true for multi-card micro-amount sequence")
	}
	if lastResult.AutomationRiskScore < 0.40 {
		t.Fatalf("Expected elevated score for card testing, got %.2f", lastResult.AutomationRiskScore)
	}
}

func TestBotDefense_CredentialStuffingFanOut(t *testing.T) {
	engine := NewBotDefenseEngine()

	baseTime := time.Now().UTC().Add(-2 * time.Minute)

	// Single device attempting 5 different user accounts in rapid succession
	var lastResult *BotDefenseResult
	for i := 0; i < 5; i++ {
		ctx := &BotDefenseContext{
			TenantID:          "org_tenant_1",
			PlanTier:          "GROWTH",
			TransactionID:     "tx_stuffing_" + string(rune('0'+i)),
			AccountID:         "usr_victim_account_" + string(rune('1'+i)),
			DeviceFingerprint: "fp_stuffing_farm_device",
			IPAddress:         "192.0.2.77",
			UserAgent:         "Mozilla/5.0",
			CardHash:          "card_common",
			Amount:            100.0,
			Currency:          "USD",
			Timestamp:         baseTime.Add(time.Duration(i*5) * time.Second),
		}
		lastResult = engine.Evaluate(ctx)
	}

	if lastResult.Signals.AccountFanOut1h < 4 {
		t.Fatalf("Expected AccountFanOut1h >= 4, got %d", lastResult.Signals.AccountFanOut1h)
	}
	if lastResult.AutomationRiskScore < 0.35 {
		t.Fatalf("Expected elevated risk score for credential stuffing, got %.2f", lastResult.AutomationRiskScore)
	}
}

func TestBotDefense_ReplayAttackDetection(t *testing.T) {
	engine := NewBotDefenseEngine()

	now := time.Now().UTC()
	ctx1 := &BotDefenseContext{
		TenantID:          "org_tenant_1",
		PlanTier:          "ENTERPRISE",
		TransactionID:     "tx_unique_nonce_12345",
		Nonce:             "nonce_secure_99",
		AccountID:         "usr_alice",
		DeviceFingerprint: "dev_alice",
		IPAddress:         "192.0.2.1",
		Amount:            50.0,
		Timestamp:         now,
	}

	res1 := engine.Evaluate(ctx1)
	if res1.Signals.ReplayDetected {
		t.Fatalf("Initial request should not be flagged as replay")
	}

	// Immediate duplicate replay attempt with same nonce
	ctx2 := &BotDefenseContext{
		TenantID:          "org_tenant_1",
		PlanTier:          "ENTERPRISE",
		TransactionID:     "tx_replay_attempt_67890",
		Nonce:             "nonce_secure_99", // DUPLICATE NONCE
		AccountID:         "usr_alice",
		DeviceFingerprint: "dev_alice",
		IPAddress:         "192.0.2.1",
		Amount:            50.0,
		Timestamp:         now,
	}

	res2 := engine.Evaluate(ctx2)
	if !res2.Signals.ReplayDetected {
		t.Fatalf("Expected ReplayDetected to be true for duplicate nonce")
	}
	if res2.AutomationRiskScore < 0.50 {
		t.Fatalf("Expected risk score >= 0.50 for replay attack, got %.2f", res2.AutomationRiskScore)
	}
}

func TestBotDefense_HeadlessBrowserAutomation(t *testing.T) {
	engine := NewBotDefenseEngine()

	ctx := &BotDefenseContext{
		TenantID:          "org_tenant_1",
		PlanTier:          "GROWTH",
		TransactionID:     "tx_selenium_1",
		AccountID:         "usr_bob",
		DeviceFingerprint: "dev_headless_vm",
		IPAddress:         "198.51.100.22",
		UserAgent:         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/120.0.0.0 Safari/537.36",
		Amount:            75.0,
		Timestamp:         time.Now().UTC(),
	}

	res := engine.Evaluate(ctx)
	if !res.Signals.IsHeadlessUA {
		t.Fatalf("Expected IsHeadlessUA to be true for HeadlessChrome")
	}
	if res.AutomationRiskScore < 0.30 {
		t.Fatalf("Expected elevated risk score for headless browser, got %.2f", res.AutomationRiskScore)
	}
}

func TestBotDefense_FailOpenResilience(t *testing.T) {
	engine := NewBotDefenseEngine()

	// Corrupt context or pass zero values to test panic recovery
	ctx := &BotDefenseContext{
		Timestamp: time.Time{}, // Zero time
	}

	res := engine.Evaluate(ctx)
	if res == nil {
		t.Fatalf("Engine must never return nil result")
	}
	if res.AutomationRiskScore > 0.30 {
		t.Fatalf("Fail-open must return safe low baseline score, got %.2f", res.AutomationRiskScore)
	}
	if res.RecommendedAction != ActionAllow {
		t.Fatalf("Fail-open must return ActionAllow")
	}
}
