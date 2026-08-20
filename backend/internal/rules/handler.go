package rules

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

// CreateRule handles POST /v1/rules
func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)
	actorID := getActorID(r)

	var req struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		DSLAST      json.RawMessage `json:"dsl_ast"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	if req.Name == "" || len(req.DSLAST) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "validation_error",
			"message": "both 'name' and 'dsl_ast' are required",
		})
		return
	}

	rule, err := h.service.CreateRule(r.Context(), CreateRuleInput{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		DSLAST:      req.DSLAST,
		CreatedBy:   actorID,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidDSL) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "invalid_dsl",
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

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(rule)
}

// ListRules handles GET /v1/rules
func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)

	var statusFilter *RuleStatus
	statusParam := r.URL.Query().Get("status")
	if statusParam != "" {
		st := RuleStatus(statusParam)
		statusFilter = &st
	}

	rulesList, err := h.service.ListRules(r.Context(), tenantID, statusFilter)
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
		"rules": rulesList,
		"count": len(rulesList),
	})
}

// GetRule handles GET /v1/rules/{id}
func (h *Handler) GetRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)
	ruleID := chi.URLParam(r, "id")

	rule, err := h.service.GetRule(r.Context(), tenantID, ruleID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Rule not found",
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
	_ = json.NewEncoder(w).Encode(rule)
}

// UpdateRule handles PUT /v1/rules/{id}
func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)
	actorID := getActorID(r)
	ruleID := chi.URLParam(r, "id")

	var req struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		DSLAST      json.RawMessage `json:"dsl_ast"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	rule, err := h.service.UpdateRule(r.Context(), tenantID, ruleID, UpdateRuleInput{
		Name:        req.Name,
		Description: req.Description,
		DSLAST:      req.DSLAST,
		UpdatedBy:   actorID,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Rule not found",
			})
			return
		}
		if errors.Is(err, ErrInvalidDSL) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "invalid_dsl",
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
	_ = json.NewEncoder(w).Encode(rule)
}

// TransitionStatus handles PUT /v1/rules/{id}/status
func (h *Handler) TransitionStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tenantID := getTenantID(r)
	actorID := getActorID(r)
	ruleID := chi.URLParam(r, "id")

	var req struct {
		Status RuleStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	rule, err := h.service.TransitionStatus(r.Context(), tenantID, ruleID, req.Status, actorID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "not_found",
				"message": "Rule not found",
			})
			return
		}
		// Maker-checker dual-control violation
		if errors.Is(err, ErrMakerCheckerViolation) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "maker_checker_violation",
				"message": "Dual-control enforcement: rule creator cannot approve their own rule. A different analyst must approve.",
			})
			return
		}
		if errors.Is(err, ErrInvalidStatusChange) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "invalid_status_transition",
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
	_ = json.NewEncoder(w).Encode(rule)
}
