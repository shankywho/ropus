package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shankywho/ropus/backend/internal/riskengine"
	"github.com/shankywho/ropus/backend/internal/utils"
)

// RegisterOperationsHandlers mounts all operational observability, SLO, metrics, incident, and control plane endpoints.
func RegisterOperationsHandlers(
	r chi.Router,
	coordinator *riskengine.RetrainingCoordinator,
	healthAggregator *riskengine.HealthAggregator,
	sloEngine *riskengine.SLOEngine,
	metricsEngine *riskengine.MetricsEngine,
	incidentEngine *riskengine.IncidentEngine,
	canaryRouter *riskengine.CanaryRouter,
	safetyAuditor *riskengine.SafetyAuditor,
	artifactScanner *riskengine.ArtifactHealthScanner,
	drManager *riskengine.DisasterRecoveryManager,
	adminKey string,
) {
	// -------------------------------------------------------------
	// 1. Prometheus Metrics Endpoint
	// -------------------------------------------------------------
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		if metricsEngine == nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("# Telemetry metrics engine uninitialized\n"))
			return
		}

		var sloSum *riskengine.SLOSummary
		if sloEngine != nil {
			s := sloEngine.Evaluate(time.Now().UTC())
			sloSum = &s
		}

		prodVer := "fraud-xgb-25f-v3.0"
		fbVer := "fraud-xgb-15f-v1.5"
		if coordinator != nil && coordinator.GetModelRegistry() != nil {
			if pm, err := coordinator.GetModelRegistry().GetProductionModel(); err == nil {
				prodVer = pm.Version
			}
			if fm, err := coordinator.GetModelRegistry().GetFallbackModel(); err == nil {
				fbVer = fm.Version
			}
		}

		var driftSt riskengine.DriftStatus = riskengine.DriftStatusHealthy
		var maxPSI, maxJSD float64
		if healthAggregator != nil {
			report := healthAggregator.GetHealthReport()
			if dComp, ok := report.Components["drift"]; ok {
				if dComp.Status == riskengine.HealthStatusDegraded {
					driftSt = riskengine.DriftStatusDegraded
				} else if dComp.Status == riskengine.HealthStatusUnhealthy {
					driftSt = riskengine.DriftStatusCritical
				}
			}
		}

		canaryStage := 0
		var cbState riskengine.CircuitBreakerState = riskengine.CircuitStateHealthy
		if canaryRouter != nil {
			canaryStage = canaryRouter.GetPercentage()
			if canaryRouter.GetCircuitBreaker() != nil {
				cbState = canaryRouter.GetCircuitBreaker().GetState()
			}
		}

		retrainActive := false
		if coordinator != nil {
			st := coordinator.GetStatus()
			if stateVal, ok := st["state"].(riskengine.JobState); ok {
				retrainActive = (stateVal != riskengine.StateIdle && stateVal != riskengine.StatePromoted && stateVal != riskengine.StateFailed && stateVal != riskengine.StateRejected && stateVal != riskengine.StateRolledBack)
			}
		}

		text := metricsEngine.ExportPrometheus(sloSum, prodVer, fbVer, driftSt, maxPSI, maxJSD, canaryStage, cbState, retrainActive)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(text))
	})

	// -------------------------------------------------------------
	// 2. Health & Readiness Aliases
	// -------------------------------------------------------------
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if healthAggregator == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "HEALTHY", "message": "Service operational"})
			return
		}
		report := healthAggregator.GetHealthReport()
		if report.OverallStatus == riskengine.HealthStatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(report)
	}

	r.Get("/health", healthHandler)
	r.Get("/readiness", healthHandler)
	r.Get("/healthz", healthHandler)
	r.Get("/readyz", healthHandler)

	// -------------------------------------------------------------
	// 3. Operational Read APIs
	// -------------------------------------------------------------
	r.Get("/v1/operations/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if healthAggregator == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "HEALTHY"})
			return
		}
		_ = json.NewEncoder(w).Encode(healthAggregator.GetHealthReport())
	})

	r.Get("/v1/operations/slo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if sloEngine == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "UNKNOWN"})
			return
		}
		summary := sloEngine.Evaluate(time.Now().UTC())
		_ = json.NewEncoder(w).Encode(summary)
	})

	r.Get("/v1/operations/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if metricsEngine == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}
		_ = json.NewEncoder(w).Encode(metricsEngine.GetSnapshot())
	})

	r.Get("/v1/operations/incidents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if incidentEngine == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		_ = json.NewEncoder(w).Encode(incidentEngine.ListIncidents())
	})

	r.Get("/v1/operations/summary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var healthReport interface{} = map[string]string{"status": "HEALTHY"}
		if healthAggregator != nil {
			healthReport = healthAggregator.GetHealthReport()
		}

		var sloReport interface{} = map[string]string{"status": "HEALTHY"}
		if sloEngine != nil {
			sloReport = sloEngine.Evaluate(time.Now().UTC())
		}

		var opControls interface{} = map[string]interface{}{}
		if coordinator != nil {
			opControls = coordinator.GetOperationalControls()
		}

		var incidents interface{} = []interface{}{}
		if incidentEngine != nil {
			incidents = incidentEngine.ListIncidents()
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"timestamp":            time.Now().UTC(),
			"health":               healthReport,
			"slo":                  sloReport,
			"operational_controls": opControls,
			"active_incidents":     incidents,
		})
	})

	// -------------------------------------------------------------
	// 4. Admin Control Plane Mutations (Admin Protected)
	// -------------------------------------------------------------
	parseAdminReq := func(w http.ResponseWriter, r *http.Request) (actor, reason string, ok bool) {
		var req struct {
			Reason string `json:"reason"`
			Actor  string `json:"actor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": "Invalid JSON payload in request body",
			})
			return "", "", false
		}
		reason = strings.TrimSpace(req.Reason)
		if reason == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": "A non-empty 'reason' string is required for operational control changes",
			})
			return "", "", false
		}
		if len(reason) > 500 {
			reason = reason[:500]
		}
		actor = strings.TrimSpace(req.Actor)
		if actor == "" {
			actor = "ADMIN_OPERATOR"
		}
		return actor, reason, true
	}

	// Maintenance Mode
	r.Post("/v1/operations/maintenance/enable", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetMaintenanceMode(r.Context(), true, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"maintenance_mode": true, "status": "enabled"})
	}))

	r.Post("/v1/operations/maintenance/disable", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetMaintenanceMode(r.Context(), false, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"maintenance_mode": false, "status": "disabled"})
	}))

	// Model Freeze
	r.Post("/v1/operations/model/freeze", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetModelFrozen(r.Context(), true, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"model_frozen": true, "status": "frozen"})
	}))

	r.Post("/v1/operations/model/unfreeze", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetModelFrozen(r.Context(), false, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"model_frozen": false, "status": "unfrozen"})
	}))

	// Retraining Pause
	r.Post("/v1/operations/retraining/pause", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetRetrainingPaused(r.Context(), true, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"retraining_paused": true, "status": "paused"})
	}))

	r.Post("/v1/operations/retraining/resume", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetRetrainingPaused(r.Context(), false, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"retraining_paused": false, "status": "resumed"})
	}))

	// Canary Pause
	r.Post("/v1/operations/canary/pause", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetCanaryPaused(r.Context(), true, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"canary_paused": true, "status": "paused"})
	}))

	r.Post("/v1/operations/canary/resume", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || coordinator == nil {
			return
		}
		_ = coordinator.SetCanaryPaused(r.Context(), false, actor, reason)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"canary_paused": false, "status": "resumed"})
	}))

	// Incident Acknowledge & Resolve
	r.Post("/v1/operations/incidents/{id}/acknowledge", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rawID := chi.URLParam(r, "id")
		incID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad_request", "message": err.Error()})
			return
		}
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || incidentEngine == nil {
			return
		}
		if err := incidentEngine.AcknowledgeIncident(incID, actor, reason); err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not_found", "message": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"incident_id": incID, "status": "ACKNOWLEDGED"})
	}))

	r.Post("/v1/operations/incidents/{id}/resolve", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rawID := chi.URLParam(r, "id")
		incID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad_request", "message": err.Error()})
			return
		}
		actor, reason, ok := parseAdminReq(w, r)
		if !ok || incidentEngine == nil {
			return
		}
		if err := incidentEngine.ResolveIncident(incID, actor, reason); err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not_found", "message": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"incident_id": incID, "status": "RESOLVED"})
	}))

	// -------------------------------------------------------------
	// 6. Safety Audit Endpoint (Phase 3.20)
	// -------------------------------------------------------------
	r.Get("/v1/operations/safety", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if safetyAuditor == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "SAFE", "message": "Safety auditor uninitialized"})
			return
		}
		report := safetyAuditor.Audit(r.Context())
		if report.Status == "UNSAFE" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(report)
	})

	// -------------------------------------------------------------
	// 7. Artifact Health Scanner Endpoint (Phase 3.20)
	// -------------------------------------------------------------
	r.Get("/v1/models/artifacts/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if artifactScanner == nil || coordinator == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "HEALTHY", "message": "Artifact health scanner uninitialized"})
			return
		}
		report, err := artifactScanner.ScanHealth(r.Context(), coordinator.GetModelRegistry())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "scan_failed", "message": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(report)
	})

	// -------------------------------------------------------------
	// 8. Disaster Recovery Manual Trigger (Phase 3.20)
	// -------------------------------------------------------------
	r.Post("/v1/operations/recovery/trigger", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if drManager == nil || coordinator == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unavailable", "message": "Disaster recovery manager uninitialized"})
			return
		}
		report, err := drManager.ExecuteRecovery(r.Context(), coordinator.GetModelRegistry(), coordinator, canaryRouter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "recovery_failed", "message": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(report)
	}))
}
