package cases

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getTenantID(r *http.Request) string {
	if tid := r.Header.Get("X-Tenant-ID"); tid != "" {
		return tid
	}
	return "00000000-0000-0000-0000-000000000001"
}

func getActorID(r *http.Request) string {
	if aid := r.Header.Get("X-Actor-ID"); aid != "" {
		return aid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	return "analyst_default"
}

// ListCases handles GET /v1/cases
func (h *Handler) ListCases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)

	var statusFilter *CaseStatus
	statusParam := r.URL.Query().Get("status")
	if statusParam != "" {
		st := CaseStatus(statusParam)
		statusFilter = &st
	}

	casesList, err := h.service.ListCases(r.Context(), tenantID, statusFilter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "internal_error",
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"cases": casesList,
		"count": len(casesList),
	})
}

// GetCase handles GET /v1/cases/{id}
func (h *Handler) GetCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)
	caseID := chi.URLParam(r, "id")

	caseDetail, err := h.service.GetCase(r.Context(), tenantID, caseID)
	if err != nil {
		if errors.Is(err, ErrCaseNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Case not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "internal_error",
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(caseDetail)
}

// ClaimCase handles PUT /v1/cases/{id}/claim
func (h *Handler) ClaimCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)
	actorID := getActorID(r)
	caseID := chi.URLParam(r, "id")

	claimedCase, err := h.service.ClaimCase(r.Context(), tenantID, caseID, actorID)
	if err != nil {
		if errors.Is(err, ErrCaseNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Case not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "internal_error",
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(claimedCase)
}

// ResolveCase handles PUT /v1/cases/{id}/resolve
func (h *Handler) ResolveCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)
	actorID := getActorID(r)
	caseID := chi.URLParam(r, "id")

	var req struct {
		Action string `json:"action"` // "ALLOW" or "DECLINE"
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": fmt.Sprintf("Failed to parse request JSON: %v", err),
		})
		return
	}

	if req.Action == "" || req.Reason == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "validation_error",
			"message": "both 'action' and 'reason' are required",
		})
		return
	}

	resolvedCase, err := h.service.ResolveCase(r.Context(), tenantID, caseID, req.Action, req.Reason, actorID)
	if err != nil {
		if errors.Is(err, ErrCaseNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Case not found",
			})
			return
		}
		if errors.Is(err, ErrInvalidStatus) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "invalid_status",
				"message": err.Error(),
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "internal_error",
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resolvedCase)
}
