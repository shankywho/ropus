package riskengine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shankywho/ropus/backend/internal/audit"
)

// IncidentState represents the lifecycle status of an operational incident.
type IncidentState string

const (
	IncidentStateOpen         IncidentState = "OPEN"
	IncidentStateAcknowledged IncidentState = "ACKNOWLEDGED"
	IncidentStateResolved     IncidentState = "RESOLVED"
)

// IncidentSeverity represents the impact severity level of an incident.
type IncidentSeverity string

const (
	IncidentSeverityInfo     IncidentSeverity = "INFO"
	IncidentSeverityWarning  IncidentSeverity = "WARNING"
	IncidentSeverityHigh     IncidentSeverity = "HIGH"
	IncidentSeverityCritical IncidentSeverity = "CRITICAL"
)

// IncidentCategory categorizes the operational domain of the failure.
type IncidentCategory string

const (
	IncidentCategorySLOBreach         IncidentCategory = "SLO_BREACH"
	IncidentCategoryCircuitBreakerTrip IncidentCategory = "CIRCUIT_BREAKER_TRIP"
	IncidentCategoryCriticalDrift     IncidentCategory = "CRITICAL_DRIFT"
	IncidentCategoryRetrainingFailure IncidentCategory = "RETRAINING_FAILURE"
	IncidentCategoryDependencyOutage  IncidentCategory = "DEPENDENCY_OUTAGE"
	IncidentCategoryCanaryRollback    IncidentCategory = "CANARY_ROLLBACK"
	IncidentCategoryModelFailure      IncidentCategory = "MODEL_FAILURE"
	IncidentCategorySecurityViolation IncidentCategory = "SECURITY_VIOLATION"
)

// Incident represents a trackable operational incident event with deduplication counters.
type Incident struct {
	IncidentID        string           `json:"incident_id"`
	Severity          IncidentSeverity `json:"severity"`
	Category          IncidentCategory `json:"category"`
	Status            IncidentState    `json:"status"`
	Reason            string           `json:"reason"`
	AffectedSubsystem string           `json:"affected_subsystem"`
	ModelVersion      string           `json:"model_version,omitempty"`
	CorrelationID     string           `json:"correlation_id,omitempty"`
	OccurrenceCount   int64            `json:"occurrence_count"`
	FirstSeen         time.Time        `json:"first_seen"`
	LastSeen          time.Time        `json:"last_seen"`
	ResolvedAt        *time.Time       `json:"resolved_at,omitempty"`
	AcknowledgedBy    string           `json:"acknowledged_by,omitempty"`
	ResolvedBy        string           `json:"resolved_by,omitempty"`
}

// IncidentEngine automatically detects, correlates, deduplicates, and resolves incidents.
type IncidentEngine struct {
	mu sync.RWMutex

	incidents      map[string]*Incident // Key: category:subsystem
	incidentList   []*Incident
	maxHistory     int
	alertManager   *AlertManager
	chClient       *audit.ClickHouseClient
	autoRemediate  AutoRemediationHandler
}

// NewIncidentEngine initializes the automated incident management subsystem.
func NewIncidentEngine(alertManager *AlertManager, chClient *audit.ClickHouseClient) *IncidentEngine {
	return &IncidentEngine{
		incidents:    make(map[string]*Incident),
		incidentList: make([]*Incident, 0),
		maxHistory:   500,
		alertManager: alertManager,
		chClient:     chClient,
	}
}

// Evaluate automatically evaluates health reports, SLO summaries, and state machines to raise or resolve incidents.
func (ie *IncidentEngine) Evaluate(
	ctx context.Context,
	health SystemHealthReport,
	slo SLOSummary,
	cbState CircuitBreakerState,
	driftStatus DriftStatus,
	retrainState JobState,
	prodModel string,
) []Incident {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	now := time.Now().UTC()

	// 1. Evaluate Dependencies & Components
	for name, comp := range health.Components {
		key := fmt.Sprintf("%s:%s", IncidentCategoryDependencyOutage, name)
		if comp.Status == HealthStatusUnhealthy {
			ie.raiseOrUpdateLocked(
				ctx,
				key,
				IncidentSeverityCritical,
				IncidentCategoryDependencyOutage,
				fmt.Sprintf("Critical component %s is UNHEALTHY: %s", name, comp.Message),
				name,
				prodModel,
				"",
				now,
			)
		} else if comp.Status == HealthStatusDegraded {
			ie.raiseOrUpdateLocked(
				ctx,
				key,
				IncidentSeverityWarning,
				IncidentCategoryDependencyOutage,
				fmt.Sprintf("Component %s is DEGRADED: %s", name, comp.Message),
				name,
				prodModel,
				"",
				now,
			)
		} else if comp.Status == HealthStatusHealthy {
			ie.resolveIfOpenLocked(ctx, key, "AUTOMATED_ENGINE", "Component restored to HEALTHY", now)
		}
	}

	// 2. Evaluate Circuit Breaker
	cbKey := fmt.Sprintf("%s:canary_router", IncidentCategoryCircuitBreakerTrip)
	if cbState == CircuitStateFailed || cbState == CircuitStateRolledBack {
		ie.raiseOrUpdateLocked(
			ctx,
			cbKey,
			IncidentSeverityCritical,
			IncidentCategoryCircuitBreakerTrip,
			fmt.Sprintf("Circuit breaker tripped into %s state", cbState),
			"canary_router",
			prodModel,
			"",
			now,
		)
	} else if cbState == CircuitStateHealthy {
		ie.resolveIfOpenLocked(ctx, cbKey, "AUTOMATED_ENGINE", "Circuit breaker restored to HEALTHY", now)
	}

	// 3. Evaluate Critical Drift
	driftKey := fmt.Sprintf("%s:drift_detector", IncidentCategoryCriticalDrift)
	if driftStatus == DriftStatusCritical {
		ie.raiseOrUpdateLocked(
			ctx,
			driftKey,
			IncidentSeverityHigh,
			IncidentCategoryCriticalDrift,
			"Critical feature drift detected exceeding safety threshold",
			"drift_detector",
			prodModel,
			"",
			now,
		)
	} else if driftStatus == DriftStatusHealthy {
		ie.resolveIfOpenLocked(ctx, driftKey, "AUTOMATED_ENGINE", "Feature drift normalized to HEALTHY", now)
	}

	// 4. Evaluate Retraining Failures
	retrainKey := fmt.Sprintf("%s:retraining_coordinator", IncidentCategoryRetrainingFailure)
	if retrainState == StateFailed {
		ie.raiseOrUpdateLocked(
			ctx,
			retrainKey,
			IncidentSeverityWarning,
			IncidentCategoryRetrainingFailure,
			"Retraining candidate generation or validation job failed",
			"retraining_coordinator",
			prodModel,
			"",
			now,
		)
	} else if retrainState == StatePromoted || retrainState == StateIdle {
		ie.resolveIfOpenLocked(ctx, retrainKey, "AUTOMATED_ENGINE", "Retraining pipeline normalized", now)
	}

	// 5. Evaluate SLO Breaches
	for sloID, m := range slo.Measurements {
		sloKey := fmt.Sprintf("%s:%s", IncidentCategorySLOBreach, sloID)
		if m.Status == SLOStatusBreached {
			ie.raiseOrUpdateLocked(
				ctx,
				sloKey,
				IncidentSeverityCritical,
				IncidentCategorySLOBreach,
				fmt.Sprintf("SLO '%s' BREACHED: Current %.4f vs Target %.4f (Burn Rate: %.2fx)", m.Name, m.CurrentValue, m.Target, m.BurnRate),
				"slo_engine",
				prodModel,
				"",
				now,
			)
		} else if m.Status == SLOStatusWarning {
			ie.raiseOrUpdateLocked(
				ctx,
				sloKey,
				IncidentSeverityWarning,
				IncidentCategorySLOBreach,
				fmt.Sprintf("SLO '%s' in WARNING: Current %.4f vs Target %.4f", m.Name, m.CurrentValue, m.Target),
				"slo_engine",
				prodModel,
				"",
				now,
			)
		} else if m.Status == SLOStatusHealthy {
			ie.resolveIfOpenLocked(ctx, sloKey, "AUTOMATED_ENGINE", fmt.Sprintf("SLO '%s' returned to compliant status", m.Name), now)
		}
	}

	return ie.listIncidentsLocked()
}

// AutoRemediationHandler defines the callback signature for autonomous disaster recovery and safety actions.
type AutoRemediationHandler func(ctx context.Context, inc Incident)

func (ie *IncidentEngine) SetAutoRemediationHandler(handler AutoRemediationHandler) {
	ie.mu.Lock()
	defer ie.mu.Unlock()
	ie.autoRemediate = handler
}

func (ie *IncidentEngine) raiseOrUpdateLocked(
	ctx context.Context,
	key string,
	severity IncidentSeverity,
	category IncidentCategory,
	reason string,
	subsystem string,
	modelVer string,
	reqID string,
	now time.Time,
) {
	if existing, found := ie.incidents[key]; found && existing.Status != IncidentStateResolved {
		// Deduplicate and update in-place
		existing.LastSeen = now
		existing.OccurrenceCount++
		existing.Reason = reason

		// AUTOMATED ESCALATION RULE:
		// 1. WARNING repeated >= 3 times escalates to HIGH
		if existing.Severity == IncidentSeverityWarning && existing.OccurrenceCount >= 3 {
			existing.Severity = IncidentSeverityHigh
		}
		// 2. HIGH persisted >= 10 minutes escalates to CRITICAL
		if existing.Severity == IncidentSeverityHigh && now.Sub(existing.FirstSeen) >= 10*time.Minute {
			existing.Severity = IncidentSeverityCritical
		}

		if severity == IncidentSeverityCritical && existing.Severity != IncidentSeverityCritical {
			existing.Severity = IncidentSeverityCritical
		}

		// Trigger auto-remediation on critical state
		if existing.Severity == IncidentSeverityCritical && ie.autoRemediate != nil {
			incCopy := *existing
			go ie.autoRemediate(context.Background(), incCopy)
		}
		return
	}

	// New Incident
	incID := fmt.Sprintf("inc_%d_%d", now.Unix(), len(ie.incidentList)+1)
	inc := &Incident{
		IncidentID:        incID,
		Severity:          severity,
		Category:          category,
		Status:            IncidentStateOpen,
		Reason:            reason,
		AffectedSubsystem: subsystem,
		ModelVersion:      modelVer,
		CorrelationID:     reqID,
		OccurrenceCount:   1,
		FirstSeen:         now,
		LastSeen:          now,
	}

	ie.incidents[key] = inc
	ie.incidentList = append(ie.incidentList, inc)
	if len(ie.incidentList) > ie.maxHistory {
		ie.incidentList = ie.incidentList[len(ie.incidentList)-ie.maxHistory:]
	}

	// Emit Alert asynchronously
	if ie.alertManager != nil {
		ie.alertManager.Emit(Alert{
			AlertID:       fmt.Sprintf("alt_%s", incID),
			Timestamp:     now,
			Severity:      severity,
			Title:         fmt.Sprintf("[%s] Incident on %s", severity, subsystem),
			Message:       reason,
			IncidentID:    incID,
			Subsystem:     subsystem,
			CorrelationID: reqID,
		})
	}

	// Trigger auto-remediation on new critical incident
	if inc.Severity == IncidentSeverityCritical && ie.autoRemediate != nil {
		incCopy := *inc
		go ie.autoRemediate(context.Background(), incCopy)
	}
}

func (ie *IncidentEngine) resolveIfOpenLocked(ctx context.Context, key, actor, reason string, now time.Time) {
	if inc, found := ie.incidents[key]; found && inc.Status != IncidentStateResolved {
		inc.Status = IncidentStateResolved
		inc.ResolvedAt = &now
		inc.ResolvedBy = actor
	}
}

// AcknowledgeIncident marks an active incident as ACKNOWLEDGED by an operator.
func (ie *IncidentEngine) AcknowledgeIncident(id, actor, reason string) error {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	for _, inc := range ie.incidentList {
		if inc.IncidentID == id {
			if inc.Status == IncidentStateResolved {
				return fmt.Errorf("incident %s is already resolved", id)
			}
			inc.Status = IncidentStateAcknowledged
			inc.AcknowledgedBy = actor
			return nil
		}
	}
	return fmt.Errorf("incident %s not found", id)
}

// ResolveIncident allows manual operator resolution of an incident.
func (ie *IncidentEngine) ResolveIncident(id, actor, reason string) error {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	now := time.Now().UTC()
	for _, inc := range ie.incidentList {
		if inc.IncidentID == id {
			inc.Status = IncidentStateResolved
			inc.ResolvedAt = &now
			inc.ResolvedBy = actor
			return nil
		}
	}
	return fmt.Errorf("incident %s not found", id)
}

// ListIncidents returns all tracked incidents in reverse chronological order.
func (ie *IncidentEngine) ListIncidents() []Incident {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	return ie.listIncidentsLocked()
}

func (ie *IncidentEngine) listIncidentsLocked() []Incident {
	res := make([]Incident, len(ie.incidentList))
	for i, idx := 0, len(ie.incidentList)-1; idx >= 0; i, idx = i+1, idx-1 {
		res[i] = *ie.incidentList[idx]
	}
	return res
}
