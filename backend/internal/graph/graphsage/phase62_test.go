package graphsage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func computeGoSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func findProjectFile(relPath string) string {
	candidates := []string{
		relPath,
		"../../../../" + relPath,
		"../../../" + relPath,
		"../../" + relPath,
		"../" + relPath,
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return relPath
}

func TestPhase62_ProductionChampionChecksumInvariant(t *testing.T) {
	champPath := findProjectFile("ml-service/model/candidates/production_model_v8_bmr.joblib")
	expectedSHA := "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

	sha, err := computeGoSHA256(champPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Production champion must remain byte-for-byte unmodified!")
}

func TestPhase62_FrozenHoldoutChecksumInvariant(t *testing.T) {
	holdoutPath := findProjectFile("ml-service/data/sample_ieee_fixture.csv")
	expectedSHA := "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44"

	sha, err := computeGoSHA256(holdoutPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Frozen 52-case holdout must remain byte-for-byte unmodified!")
}

func TestPhase62_ConnectivityChecker_HonestReporting(t *testing.T) {
	checker := NewConnectivityChecker(500 * time.Millisecond)
	report := checker.AuditEnvironment()

	assert.False(t, report.AuditedAt.IsZero())
	assert.Equal(t, "INTEGRATION_BLOCKED", report.IntegrationStatus)
	assert.False(t, report.IsLiveConnected)
	assert.NotEmpty(t, report.MissingConfigurations)
	assert.Contains(t, report.SummaryMessage, "INTEGRATION_BLOCKED")
}

func TestPhase62_ObservabilityExporter_MetricsExposition(t *testing.T) {
	mgr := NewShadowTelemetryManager(1000, false)
	mgr.Start(context.Background())
	defer mgr.Stop()

	now := time.Now().UTC()

	// Ingest sample clean event
	mgr.EnqueueEvent(StreamEventEnvelope{
		EventID:        "evt_clean_01",
		Topic:          "audit.events",
		Payload:        map[string]interface{}{"employee_id": "emp_01", "account_id": "acct_01"},
		EventTimestamp: now,
	})

	// Ingest sample privacy DLQ event
	mgr.EnqueueEvent(StreamEventEnvelope{
		EventID:        "evt_bad_01",
		Topic:          "audit.events",
		Payload:        map[string]interface{}{"employee_id": "emp_02", "raw_pan": "4111222233334444"},
		EventTimestamp: now,
	})

	time.Sleep(100 * time.Millisecond)

	exporter := NewObservabilityExporter(mgr)
	snap := exporter.GetSnapshot()

	assert.Equal(t, 0.0, snap.CustomerEnforcementPct)
	assert.Equal(t, 100.0, snap.BMREnforcementPct)
	assert.Equal(t, "NOT_CONNECTED", snap.RealDataConnectivity)

	promText := exporter.ExportPrometheusText()
	assert.Contains(t, promText, "graphsage_customer_enforcement_authority_pct 0.000000")
	assert.Contains(t, promText, "bmr_customer_enforcement_authority_pct 100.000000")
	assert.Contains(t, promText, "graphsage_shadow_events_processed_total")
	assert.Contains(t, promText, "graphsage_shadow_events_dlq_total")
}

func TestPhase62_GovernanceGate_BlocksPromotion(t *testing.T) {
	matEngine := NewLabelMaturationEngine(90)
	assert.Equal(t, 0, matEngine.GetConfirmedCollusionCount())

	realSource := NewMockRealGraphDataSource(DatasetRealShadow)
	checker := NewReadinessChecker(realSource)
	report := checker.Audit(context.Background(), time.Now().UTC())

	assert.False(t, report.IsReadyForRealValidation)
	assert.Equal(t, StateRealDataRequired, report.GovernanceState)
	assert.NotEmpty(t, report.BlockingReasons)
}
