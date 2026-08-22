// RegisterDriftHandlers registers drift status and history endpoints.
package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shankywho/ropus/backend/internal/riskengine"
)

// RegisterDriftHandlers wires the drift monitoring HTTP endpoints.
func RegisterDriftHandlers(r chi.Router, detector *riskengine.DriftDetector) {
	// Drift status endpoint
	r.Get("/drift/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if detector == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "UNAVAILABLE",
				"message": "drift detection subsystem is not initialized",
			})
			return
		}
		status := detector.GetStatus()
		_ = json.NewEncoder(w).Encode(status)
	})

	// Drift history endpoint
	r.Get("/drift/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if detector == nil {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		history := detector.GetHistory()
		_ = json.NewEncoder(w).Encode(history)
	})

	// On-demand drift evaluation trigger endpoint
	r.Post("/drift/evaluate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if detector == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "service_unavailable",
				"message": "Drift detection subsystem is not initialized",
			})
			return
		}
		measurement := detector.EvaluateLiveDrift(r.Context())
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(measurement)
	})
}
