package streaming

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreaming_EventBusPublishAndReplay(t *testing.T) {
	ctx := context.Background()
	bus := NewLocalEventBus()

	var received []*StreamingEvent
	err := bus.Subscribe(ctx, "risk.events", func(ctx context.Context, e *StreamingEvent) error {
		received = append(received, e)
		return nil
	})
	require.NoError(t, err)

	for i := 1; i <= 5; i++ {
		e := &StreamingEvent{
			EventID:  fmt.Sprintf("evt_%d", i),
			TenantID: "tenant_01",
			Type:     EventTransactionCreated,
		}
		_ = bus.Publish(ctx, "risk.events", e)
	}

	assert.Equal(t, 5, len(received))

	// Test Replay capability (offsets 2 to 4)
	var replayed []*StreamingEvent
	err = bus.Replay(ctx, "risk.events", 2, 4, func(ctx context.Context, e *StreamingEvent) error {
		replayed = append(replayed, e)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, len(replayed))
	assert.Equal(t, int64(2), replayed[0].Offset)
	assert.Equal(t, int64(4), replayed[2].Offset)
}

func TestStreaming_Deduplication(t *testing.T) {
	ctx := context.Background()
	proc := NewStreamProcessor(nil)

	e1 := &StreamingEvent{EventID: "evt_1", IdempotencyKey: "idem_abc"}
	processed1, err := proc.ProcessEvent(ctx, e1)
	require.NoError(t, err)
	assert.True(t, processed1)

	// Duplicate event
	e2 := &StreamingEvent{EventID: "evt_2", IdempotencyKey: "idem_abc"}
	processed2, err := proc.ProcessEvent(ctx, e2)
	require.NoError(t, err)
	assert.False(t, processed2, "Duplicate idempotency key must be ignored")
}

func TestStreaming_DetectorVelocityAttack(t *testing.T) {
	detector := NewStreamFraudDetector()

	var alert *StreamDetectionAlert
	for i := 0; i < 20; i++ {
		evt := &StreamingEvent{
			EventID: fmt.Sprintf("evt_%d", i),
			Type:    EventTransactionCreated,
			Payload: map[string]interface{}{
				"device_fingerprint": "dev_bad_burst_01",
				"ip_address":         "10.0.0.1",
			},
		}
		alert = detector.ProcessTransactionEvent(evt)
	}

	assert.NotNil(t, alert, "Surge > 20 events in 5m must trigger alert")
	assert.Equal(t, "VELOCITY_ATTACK", alert.PatternType)
	assert.Equal(t, "dev_bad_burst_01", alert.EntityID)
	assert.GreaterOrEqual(t, alert.Confidence, 0.95)
}

func TestStreaming_GlobalGraphAndFederation(t *testing.T) {
	gg := NewGlobalFraudGraph()

	// Bank A records fraudster
	gg.RecordGlobalSignal("bank_a", "HASHED_USER", "raw_email_fraudster@attacker.com", true, 0.99)

	// Bank B checks reputation
	risk, isBad, tenants := gg.QueryGlobalReputation("raw_email_fraudster@attacker.com")
	assert.True(t, isBad)
	assert.GreaterOrEqual(t, risk, 0.90)
	assert.Equal(t, 1, tenants)

	// Federated Mesh
	fed := NewFederatedIntelligenceMesh()
	sig := fed.BroadcastSignature("pattern_bot_user_agent_v3", "BOT_HEADER_ANOMALY", 0.94)
	assert.NotEmpty(t, sig.SignatureID)

	queried, found := fed.QuerySignature("pattern_bot_user_agent_v3")
	assert.True(t, found)
	assert.Equal(t, "BOT_HEADER_ANOMALY", queried.ThreatType)
}

func TestStreaming_CampaignDetector(t *testing.T) {
	cd := NewCampaignDetector()

	_ = cd.IngestAlert("bank_1", "CREDENTIAL_STUFFING", "ip_100_1")
	_ = cd.IngestAlert("bank_2", "CREDENTIAL_STUFFING", "ip_100_2")
	camp := cd.IngestAlert("bank_3", "CREDENTIAL_STUFFING", "ip_100_3")

	assert.Equal(t, "CREDENTIAL_STUFFING", camp.AttackType)
	assert.Equal(t, "CRITICAL", camp.Severity)
	assert.Equal(t, 3, len(camp.TargetTenants))
}

func TestStreaming_OnlineLearningSafeGating(t *testing.T) {
	ol := NewOnlineLearningEngine()

	initWeight := ol.GetWeight("threat_intel_ioc")
	prop := ol.IngestFeedback("threat_intel_ioc", true)
	assert.False(t, prop.IsApproved)
	assert.Equal(t, initWeight, ol.GetWeight("threat_intel_ioc"), "Weight must not change before approval")

	// Approve update
	approved := ol.ApproveProposal(prop.UpdateID)
	assert.True(t, approved)
	assert.Equal(t, prop.ProposedWeight, ol.GetWeight("threat_intel_ioc"))
}

func TestStreaming_AutonomousDefenseAndRollback(t *testing.T) {
	analyzer := NewImpactAnalyzer()
	engine := NewAutonomousDefenseEngine(analyzer)

	// Safe action -> APPROVED
	rec, err := engine.TriggerAutonomousDefense(DefenseDeviceBlock, "dev_bot_1", "High velocity attack", 5, 5000.0, 0.98)
	require.NoError(t, err)
	assert.False(t, rec.IsRolledBack)

	// Rollback
	err = engine.RollbackDefense(rec.DefenseID, "Manual analyst clearance")
	require.NoError(t, err)

	// Action with huge collateral impact (50,000 users) -> IMPACT GUARD REJECTS
	_, err = engine.TriggerAutonomousDefense(DefenseNetworkBlock, "subnet_10", "Broad suspicion", 50000, 100000.0, 0.99)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rejected by impact guard")
}

func TestStreaming_DigitalTwinAndZeroTrust(t *testing.T) {
	twin := NewFraudDigitalTwin()
	res := twin.SimulatePolicyChange("Block Fraud Ring Subnet", 1000, 250.0, 0.98)
	assert.Greater(t, res.ProjectedFraudPrevented, 200000.0)
	assert.Greater(t, res.ROIProjectedRatio, 5.0)

	// Zero Trust Guard
	zt := NewZeroTrustSecurityGuard()
	validCtx := ZeroTrustContext{
		CallerID:             "service_risk_backend",
		CallerRole:           "ROLE_RISK_OPERATOR",
		DeviceCertThumbprint: "sha256_thumbprint_prod_cluster_ca",
		SourceNamespace:      "risk-engine",
	}
	authorized, err := zt.AuthorizeAction(validCtx)
	require.NoError(t, err)
	assert.True(t, authorized)

	// Unauthorized namespace
	invalidCtx := ZeroTrustContext{
		CallerID:             "service_untrusted",
		DeviceCertThumbprint: "sha256_thumbprint_dev",
		SourceNamespace:      "unauthorized-ns",
	}
	_, err = zt.AuthorizeAction(invalidCtx)
	assert.Error(t, err)
}
