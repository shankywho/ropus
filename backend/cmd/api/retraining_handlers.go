package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/shankywho/ropus/backend/internal/riskengine"
	"github.com/shankywho/ropus/backend/internal/utils"
)

// RegisterRetrainingHandlers registers retraining, candidate model, and model registry HTTP endpoints.
func RegisterRetrainingHandlers(r chi.Router, coordinator *riskengine.RetrainingCoordinator, adminKey string) {
	// Retraining status endpoint
	r.Get("/retraining/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"state":   "UNAVAILABLE",
				"message": "Retraining subsystem is not initialized",
			})
			return
		}
		status := coordinator.GetStatus()
		_ = json.NewEncoder(w).Encode(status)
	})

	// Retraining history endpoint
	r.Get("/retraining/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		history := coordinator.GetHistory()
		_ = json.NewEncoder(w).Encode(history)
	})

	// Single Retraining Job Status endpoint
	r.Get("/retraining/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		jobID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid job ID '%s': %v", rawID, err),
			})
			return
		}
		job, err := coordinator.GetJobByID(jobID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(job)
	})

	// Retraining Job Logs endpoint
	r.Get("/retraining/jobs/{id}/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		jobID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid job ID '%s': %v", rawID, err),
			})
			return
		}
		logs, err := coordinator.GetJobLogs(jobID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id": jobID,
			"logs":   logs,
		})
	})

	// Retraining Job Cancel endpoint (Admin Protected)
	r.Post("/retraining/jobs/{id}/cancel", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		jobID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid job ID '%s': %v", rawID, err),
			})
			return
		}
		var req struct {
			Reason string `json:"reason"`
			Actor  string `json:"actor"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		reason := strings.TrimSpace(req.Reason)
		if reason == "" {
			reason = "Cancelled by admin operator"
		}
		if len(reason) > 500 {
			reason = reason[:500]
		}
		actor := strings.TrimSpace(req.Actor)
		if actor == "" {
			actor = "ADMIN_OPERATOR"
		}

		if err := coordinator.CancelJob(r.Context(), jobID, actor, reason); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "cancel_failed",
				"message": err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "cancelled",
			"job_id": jobID,
		})
	}))

	// Manual Retraining Trigger endpoint (Admin Protected)
	r.Post("/retraining/trigger", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "service_unavailable",
				"message": "Retraining subsystem is not initialized",
			})
			return
		}

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
			return
		}

		reason := strings.TrimSpace(req.Reason)
		if reason == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": "A non-empty 'reason' string is required for manual retraining trigger",
			})
			return
		}
		if len(reason) > 500 {
			reason = reason[:500]
		}

		actor := strings.TrimSpace(req.Actor)
		if actor == "" {
			actor = "ADMIN_OPERATOR"
		}

		job, err := coordinator.TriggerManual(r.Context(), actor, reason)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "trigger_rejected",
				"message": err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	}))

	// Model Registry Collection endpoint
	r.Get("/models/registry", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil || coordinator.GetModelRegistry() == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		models := coordinator.GetModelRegistry().ListModels()
		_ = json.NewEncoder(w).Encode(models)
	})

	// Production Model endpoint
	r.Get("/models/production", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil || coordinator.GetModelRegistry() == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"production_model": "fraud-xgb-25f-v3.0",
				"fallback_model":   "fraud-xgb-15f-v1.5",
			})
			return
		}
		prodModel, err := coordinator.GetModelRegistry().GetProductionModel()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		fbModel, _ := coordinator.GetModelRegistry().GetFallbackModel()

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"production_model": prodModel,
			"fallback_model":   fbModel,
		})
	})

	// Models Candidates collection endpoint
	r.Get("/models/candidates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		candidates := coordinator.GetCandidates()
		_ = json.NewEncoder(w).Encode(candidates)
	})

	// Single Candidate endpoint
	r.Get("/models/candidates/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		candID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid candidate ID '%s': %v", rawID, err),
			})
			return
		}
		candidate, err := coordinator.GetCandidateByID(candID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(candidate)
	})

	// Candidate Validation Scorecard endpoint
	r.Get("/models/candidates/{id}/validation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		candID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid candidate ID '%s': %v", rawID, err),
			})
			return
		}
		candidate, err := coordinator.GetCandidateByID(candID)
		if err != nil || candidate.ValidationResult == nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Validation scorecard not found for candidate",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(candidate.ValidationResult)
	})

	// Candidate Shadow Scorecard endpoint
	r.Get("/models/candidates/{id}/shadow", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		candID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid candidate ID '%s': %v", rawID, err),
			})
			return
		}
		candidate, err := coordinator.GetCandidateByID(candID)
		if err != nil || candidate.ShadowResult == nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Shadow evaluation scorecard not found for candidate",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(candidate.ShadowResult)
	})

	// Candidate Approval endpoint (Admin Protected)
	r.Post("/models/candidates/{id}/approve", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		candID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid candidate ID '%s': %v", rawID, err),
			})
			return
		}
		var req struct {
			Reason string `json:"reason"`
			Actor  string `json:"actor"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		reason := strings.TrimSpace(req.Reason)
		if reason == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": "A non-empty 'reason' string is required to approve candidate",
			})
			return
		}
		if len(reason) > 500 {
			reason = reason[:500]
		}
		actor := strings.TrimSpace(req.Actor)
		if actor == "" {
			actor = "ADMIN_OPERATOR"
		}

		if err := coordinator.ApproveCandidate(r.Context(), candID, actor, reason); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "approval_failed",
				"message": err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "approved",
			"model_id": candID,
			"action":   "CANARY_ROLLOUT_INITIATED",
		})
	}))

	// Candidate Rejection endpoint (Admin Protected)
	r.Post("/models/candidates/{id}/reject", RequireAdminAuth(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "service_unavailable"})
			return
		}
		rawID := chi.URLParam(r, "id")
		candID, err := utils.SanitizeIdentifier(rawID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid candidate ID '%s': %v", rawID, err),
			})
			return
		}
		var req struct {
			Reason string `json:"reason"`
			Actor  string `json:"actor"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		reason := strings.TrimSpace(req.Reason)
		if reason == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": "A non-empty 'reason' string is required to reject candidate",
			})
			return
		}
		if len(reason) > 500 {
			reason = reason[:500]
		}
		actor := strings.TrimSpace(req.Actor)
		if actor == "" {
			actor = "ADMIN_OPERATOR"
		}

		if err := coordinator.RejectCandidate(r.Context(), candID, actor, reason); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "rejection_failed",
				"message": err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":   "rejected",
			"model_id": candID,
		})
	}))

	// Specific Model Version endpoint
	r.Get("/models/{version}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil || coordinator.GetModelRegistry() == nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "model_registry_unavailable"})
			return
		}
		rawVersion := chi.URLParam(r, "version")
		version, err := utils.SanitizeIdentifier(rawVersion)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid model version '%s': %v", rawVersion, err),
			})
			return
		}
		model, err := coordinator.GetModelRegistry().GetModel(version)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(model)
	})

	// Specific Model Version Provenance endpoint
	r.Get("/models/{version}/provenance", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if coordinator == nil || coordinator.GetModelRegistry() == nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "model_registry_unavailable"})
			return
		}
		rawVersion := chi.URLParam(r, "version")
		version, err := utils.SanitizeIdentifier(rawVersion)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "bad_request",
				"message": fmt.Sprintf("Invalid model version '%s': %v", rawVersion, err),
			})
			return
		}
		prov, err := coordinator.GetModelRegistry().GetModelProvenance(version)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(prov)
	})
}
