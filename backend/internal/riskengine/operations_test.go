package riskengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperations_SafetyControls(t *testing.T) {
	tmpState := fmt.Sprintf("/tmp/test_ops_state_%d.json", time.Now().UnixNano())
	defer os.Remove(tmpState)

	store, err := NewFileStateStore(tmpState)
	require.NoError(t, err)

	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 10 * time.Millisecond
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	coordinator.SetStateStore(store)

	ctx := context.Background()

	// 1. Maintenance Mode
	err = coordinator.SetMaintenanceMode(ctx, true, "sre_admin", "Scheduled database migration")
	require.NoError(t, err)

	controls := coordinator.GetOperationalControls()
	assert.Equal(t, true, controls["maintenance_mode"])

	// Triggering manual retraining must be rejected while in maintenance mode
	_, err = coordinator.TriggerManual(ctx, "sre_admin", "Test trigger")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MAINTENANCE_MODE")

	// Disable maintenance mode
	err = coordinator.SetMaintenanceMode(ctx, false, "sre_admin", "Migration complete")
	require.NoError(t, err)

	// 2. Retraining Pause
	err = coordinator.SetRetrainingPaused(ctx, true, "ml_eng", "Investigating feature pipeline")
	require.NoError(t, err)

	_, err = coordinator.TriggerManual(ctx, "ml_eng", "Test trigger")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PAUSED")

	err = coordinator.SetRetrainingPaused(ctx, false, "ml_eng", "Pipeline restored")
	require.NoError(t, err)

	// 3. Model Freeze
	err = coordinator.SetModelFrozen(ctx, true, "lead_risk", "Freezing promotions before holiday peak")
	require.NoError(t, err)

	controls = coordinator.GetOperationalControls()
	assert.Equal(t, true, controls["model_frozen"])

	// 4. Persistence across process restart
	time.Sleep(50 * time.Millisecond)
	persistedState, err := store.Load(ctx)
	require.NoError(t, err)
	assert.True(t, persistedState.ModelFrozen)
	assert.False(t, persistedState.MaintenanceMode)

	// Reconcile onto new coordinator instance
	newCoordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	recoveryMgr := NewRecoveryManager(store, nil)
	reconRes, err := recoveryMgr.ReconcileOnStartup(ctx, newCoordinator.GetModelRegistry(), newCoordinator)
	require.NoError(t, err)
	require.NotNil(t, reconRes)

	restoredControls := newCoordinator.GetOperationalControls()
	assert.True(t, restoredControls["model_frozen"].(bool))
}

func TestOperations_HTTPControlPlane(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	metricsEngine := NewMetricsEngine()
	sloEngine := NewSLOEngine(5 * time.Minute)
	incidentEngine := NewIncidentEngine(NewAlertManager(&LogAlertSink{}), nil)
	healthAggregator := NewHealthAggregator(nil, nil, nil, nil, nil, coordinator, nil, sloEngine)
	defer healthAggregator.Close()

	adminKey := "test-secret-key-12345"

	r := chi.NewRouter()

	// Register test operations routes
	r.Get("/v1/operations/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(healthAggregator.GetHealthReport())
	})
	r.Get("/v1/operations/slo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sloEngine.Evaluate(time.Now().UTC()))
	})
	r.Get("/v1/operations/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metricsEngine.GetSnapshot())
	})
	r.Get("/v1/operations/incidents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(incidentEngine.ListIncidents())
	})
	r.Post("/v1/operations/maintenance/enable", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Admin-API-Key") != adminKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var req struct {
			Reason string `json:"reason"`
			Actor  string `json:"actor"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = coordinator.SetMaintenanceMode(r.Context(), true, req.Actor, req.Reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]bool{"maintenance_mode": true})
	})

	// 1. GET /v1/operations/health
	reqHealth := httptest.NewRequest("GET", "/v1/operations/health", nil)
	recHealth := httptest.NewRecorder()
	r.ServeHTTP(recHealth, reqHealth)
	assert.Equal(t, http.StatusOK, recHealth.Code)

	// 2. GET /v1/operations/slo
	reqSLO := httptest.NewRequest("GET", "/v1/operations/slo", nil)
	recSLO := httptest.NewRecorder()
	r.ServeHTTP(recSLO, reqSLO)
	assert.Equal(t, http.StatusOK, recSLO.Code)

	// 3. GET /v1/operations/incidents
	reqInc := httptest.NewRequest("GET", "/v1/operations/incidents", nil)
	recInc := httptest.NewRecorder()
	r.ServeHTTP(recInc, reqInc)
	assert.Equal(t, http.StatusOK, recInc.Code)

	// 4. POST /v1/operations/maintenance/enable (Unauthorized)
	reqBadAuth := httptest.NewRequest("POST", "/v1/operations/maintenance/enable", bytes.NewBufferString(`{"reason":"test","actor":"admin"}`))
	recBadAuth := httptest.NewRecorder()
	r.ServeHTTP(recBadAuth, reqBadAuth)
	assert.Equal(t, http.StatusUnauthorized, recBadAuth.Code)

	// 4. POST /v1/operations/maintenance/enable (Authorized)
	reqMaint := httptest.NewRequest("POST", "/v1/operations/maintenance/enable", bytes.NewBufferString(`{"reason":"maintenance window","actor":"lead_sre"}`))
	reqMaint.Header.Set("X-Admin-API-Key", adminKey)
	recMaint := httptest.NewRecorder()
	r.ServeHTTP(recMaint, reqMaint)
	assert.Equal(t, http.StatusOK, recMaint.Code)
	assert.True(t, coordinator.GetOperationalControls()["maintenance_mode"].(bool))
}
