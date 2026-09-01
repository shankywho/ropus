package chaos

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChaos_FailureDrills(t *testing.T) {
	chaos := NewChaosEngine()

	scenarios := []ChaosScenario{
		ChaosKafkaOutage,
		ChaosDBUnavailable,
		ChaosModelTimeout,
		ChaosNetworkLatency,
		ChaosRedisFailure,
	}

	for _, sc := range scenarios {
		res := chaos.ExecuteDrill(sc)
		assert.NotEmpty(t, res.DrillID)
		assert.True(t, res.DetectedBySystem, "Drill must trip the circuit breaker")
		assert.True(t, res.FallbackActivated, "Fallback fast-fail must be verified")
		assert.Equal(t, "OPEN", res.BreakerStateAfter, "Breaker must be in OPEN state")
		assert.True(t, res.FastFailVerified, "ErrCircuitOpen must be returned on subsequent call")
		assert.Equal(t, 0, res.CustomerImpactDrops, "Zero customer impact required under failure drills")
		assert.Less(t, res.RecoveryTimeMs, 100.0)
	}
}

func TestChaos_HTTPHandler(t *testing.T) {
	chaos := NewChaosEngine()
	handler := NewHandler(chaos)

	reqBody := []byte(`{"scenario": "REDIS_CACHE_FAILURE"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chaos/drill", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.HandleExecuteDrill(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp ChaosExecutionResult
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, ChaosRedisFailure, resp.Scenario)
	assert.True(t, resp.DetectedBySystem)
	assert.True(t, resp.FallbackActivated)
	assert.Equal(t, "OPEN", resp.BreakerStateAfter)
}
