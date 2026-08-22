package riskengine

import (
	"fmt"
	"time"
)

// DependencyID identifies upstream and downstream platform dependencies.
type DependencyID string

const (
	DepPostgres      DependencyID = "postgres"
	DepRedis         DependencyID = "redis"
	DepClickHouse    DependencyID = "clickhouse"
	DepMLRuntime     DependencyID = "ml_runtime"
	DepArtifactStore DependencyID = "artifact_store"
	DepModelRegistry DependencyID = "model_registry"
	DepAlertSinks    DependencyID = "alert_sinks"
)

// DepState defines the operational condition of a dependency.
type DepState string

const (
	DepStateHealthy     DepState = "HEALTHY"
	DepStateDegraded    DepState = "DEGRADED"
	DepStateUnavailable DepState = "UNAVAILABLE"
)

// ImpactAssessment describes the operational effect and automated mitigation for dependency degradation.
type ImpactAssessment struct {
	Dependency         DependencyID `json:"dependency"`
	State              DepState     `json:"state"`
	InferenceImpact    string       `json:"inference_impact"`
	RetrainingImpact   string       `json:"retraining_impact"`
	MitigationAction   string       `json:"mitigation_action"`
	FailOpenForTraffic bool         `json:"fail_open_for_traffic"`
	IncidentSeverity   string       `json:"incident_severity"`
}

// DependencyMatrix defines explicit fail-safe policies for all platform dependencies.
type DependencyMatrix struct{}

// NewDependencyMatrix returns a new dependency policy manager.
func NewDependencyMatrix() *DependencyMatrix {
	return &DependencyMatrix{}
}

// Assess evaluates the platform impact and required mitigations when a dependency state changes.
func (dm *DependencyMatrix) Assess(dep DependencyID, state DepState) ImpactAssessment {
	if state == DepStateHealthy {
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "NONE: Normal operations",
			RetrainingImpact:   "NONE: Normal operations",
			MitigationAction:   "NONE",
			FailOpenForTraffic: true,
			IncidentSeverity:   "INFO",
		}
	}

	switch dep {
	case DepPostgres:
		if state == DepStateUnavailable {
			return ImpactAssessment{
				Dependency:         dep,
				State:              state,
				InferenceImpact:    "NONE: In-memory rule heuristics and local risk scoring active; zero sync blocking",
				RetrainingImpact:   "PAUSED: Transaction history querying unavailable; new retraining triggers deferred",
				MitigationAction:   "Buffer cases locally; switch to fallback heuristic mode if state lookups fail",
				FailOpenForTraffic: true,
				IncidentSeverity:   "CRITICAL",
			}
		}
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "LOW: Connection pool saturated; fast timeout fallback enabled",
			RetrainingImpact:   "THROTTLED: Non-essential retraining queries queued",
			MitigationAction:   "Limit query pool concurrency; enable short query timeouts (500ms)",
			FailOpenForTraffic: true,
			IncidentSeverity:   "WARNING",
		}

	case DepRedis:
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "NONE: Device velocity features degrade to baseline defaults (0-count velocity)",
			RetrainingImpact:   "NONE: Offline datasets used for model training",
			MitigationAction:   "Extract default velocity vectors; circuit breaker isolates latency spikes",
			FailOpenForTraffic: true,
			IncidentSeverity:   "WARNING",
		}

	case DepClickHouse:
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "NONE: Fail-open audit; telemetry buffers in memory and drops oldest on overflow",
			RetrainingImpact:   "PAUSED: Historical baseline calculation paused until audit store recovers",
			MitigationAction:   "Enable local memory ring buffer; drop non-critical metrics before risk responses",
			FailOpenForTraffic: true,
			IncidentSeverity:   "WARNING",
		}

	case DepMLRuntime:
		if state == DepStateUnavailable {
			return ImpactAssessment{
				Dependency:         dep,
				State:              state,
				InferenceImpact:    "ROUTED_TO_FALLBACK: Circuit breaker trips immediately to 15-feature model or rules",
				RetrainingImpact:   "BLOCKED: Model promotions and live validation gates blocked",
				MitigationAction:   "Trip circuit breaker to OPEN; route 100% traffic to verified fallback model",
				FailOpenForTraffic: true,
				IncidentSeverity:   "CRITICAL",
			}
		}
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "DEGRADED: Inference latency elevated; circuit breaker monitoring error rates",
			RetrainingImpact:   "MONITORED: Shadow evaluations sampled at reduced rate",
			MitigationAction:   "Enable strict timeout (150ms); trip circuit breaker if error rate exceeds 5%",
			FailOpenForTraffic: true,
			IncidentSeverity:   "WARNING",
		}

	case DepArtifactStore:
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "NONE: Active production model loaded in memory",
			RetrainingImpact:   "BLOCKED: Candidate artifact write/read disabled",
			MitigationAction:   "Prevent new retraining candidate registration; preserve current production model",
			FailOpenForTraffic: true,
			IncidentSeverity:   "ERROR",
		}

	case DepModelRegistry:
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "NONE: Production model pointer cached in memory",
			RetrainingImpact:   "BLOCKED: Model promotions disabled",
			MitigationAction:   "Freeze model promotions; rely on verified in-memory fallback",
			FailOpenForTraffic: true,
			IncidentSeverity:   "CRITICAL",
		}

	case DepAlertSinks:
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "NONE: Bounded asynchronous alert queue drops oldest events safely",
			RetrainingImpact:   "NONE",
			MitigationAction:   "Increment dropped_alerts metric; never block synchronous request threads",
			FailOpenForTraffic: true,
			IncidentSeverity:   "WARNING",
		}

	default:
		return ImpactAssessment{
			Dependency:         dep,
			State:              state,
			InferenceImpact:    "UNKNOWN",
			RetrainingImpact:   "UNKNOWN",
			MitigationAction:   fmt.Sprintf("Inspect health of dependency: %s", dep),
			FailOpenForTraffic: true,
			IncidentSeverity:   "WARNING",
		}
	}
}

// DependencyStatusReport summarizes current status across all dependencies.
type DependencyStatusReport struct {
	Timestamp   time.Time          `json:"timestamp"`
	Assessments []ImpactAssessment `json:"assessments"`
}
