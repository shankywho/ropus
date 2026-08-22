package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// AuditRecord represents a flattened audit log entry in ClickHouse.
type AuditRecord struct {
	TransactionID   string    `json:"transaction_id"`
	RiskScore       int32     `json:"risk_score"`
	RuleTriggered   string    `json:"rule_triggered"`
	FeatureSnapshot string    `json:"feature_snapshot"`
	TenantID        string    `json:"tenant_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// ClickHouseClient manages connection to ClickHouse server.
type ClickHouseClient struct {
	conn driver.Conn
}

// NewClickHouseClient initializes a native ClickHouse connection.
func NewClickHouseClient(addr, database, username, password string) (*ClickHouseClient, error) {
	if addr == "" {
		addr = "localhost:9000"
	}
	if database == "" {
		database = "default"
	}
	if username == "" {
		username = "default"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	return &ClickHouseClient{conn: conn}, nil
}

// Ping verifies connectivity to ClickHouse.
func (c *ClickHouseClient) Ping(ctx context.Context) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("clickhouse connection is uninitialized")
	}
	return c.conn.Ping(ctx)
}

// InsertAuditRecord inserts a single audit record into risk_audit_log.
func (c *ClickHouseClient) InsertAuditRecord(ctx context.Context, record AuditRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO risk_audit_log (
			transaction_id, risk_score, rule_triggered, feature_snapshot, tenant_id, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?
		)
	`
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}

	err := c.conn.Exec(ctx, query,
		record.TransactionID,
		record.RiskScore,
		record.RuleTriggered,
		record.FeatureSnapshot,
		record.TenantID,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert audit record into clickhouse: %w", err)
	}

	return nil
}

// ShadowScoreEvaluation represents a structured shadow scoring record in ClickHouse.
type ShadowScoreEvaluation struct {
	EvaluationID               string    `json:"evaluation_id"`
	TenantID                   string    `json:"tenant_id"`
	TransactionID               string    `json:"transaction_id"`
	Timestamp                  time.Time `json:"timestamp"`
	ProductionModelVersion     string    `json:"production_model_version"`
	ShadowModelVersion         string    `json:"shadow_model_version"`
	ProductionFeatureContract  string    `json:"production_feature_contract"`
	ShadowFeatureContract      string    `json:"shadow_feature_contract"`
	ProductionRawScore         float64   `json:"production_raw_score"`
	ProductionCalibratedScore  float64   `json:"production_calibrated_score"`
	ShadowRawScore             float64   `json:"shadow_raw_score"`
	ShadowCalibratedScore      float64   `json:"shadow_calibrated_score"`
	ProductionDecision         string    `json:"production_decision"`
	ShadowDecision             string    `json:"shadow_decision"`
	ScoreDelta                 float64   `json:"score_delta"`
	AbsoluteScoreDelta         float64   `json:"absolute_score_delta"`
	DecisionChanged            uint8     `json:"decision_changed"`
	DivergenceCategory         string    `json:"divergence_category"`
	ProductionLatencyMs        float64   `json:"production_latency_ms"`
	ShadowInferenceLatencyMs   float64   `json:"shadow_inference_latency_ms"`
	ShadowTotalLatencyMs       float64   `json:"shadow_total_latency_ms"`
	ShadowError                string    `json:"shadow_error"`
}

// InsertShadowEvaluation inserts a single shadow scoring evaluation record into shadow_score_evaluations.
func (c *ClickHouseClient) InsertShadowEvaluation(ctx context.Context, record ShadowScoreEvaluation) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO shadow_score_evaluations (
			evaluation_id, tenant_id, transaction_id, timestamp,
			production_model_version, shadow_model_version,
			production_feature_contract, shadow_feature_contract,
			production_raw_score, production_calibrated_score,
			shadow_raw_score, shadow_calibrated_score,
			production_decision, shadow_decision,
			score_delta, absolute_score_delta, decision_changed, divergence_category,
			production_latency_ms, shadow_inference_latency_ms, shadow_total_latency_ms,
			shadow_error
		) VALUES (
			?, ?, ?, ?,
			?, ?,
			?, ?,
			?, ?,
			?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?
		)
	`
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	}

	err := c.conn.Exec(ctx, query,
		record.EvaluationID, record.TenantID, record.TransactionID, record.Timestamp,
		record.ProductionModelVersion, record.ShadowModelVersion,
		record.ProductionFeatureContract, record.ShadowFeatureContract,
		record.ProductionRawScore, record.ProductionCalibratedScore,
		record.ShadowRawScore, record.ShadowCalibratedScore,
		record.ProductionDecision, record.ShadowDecision,
		record.ScoreDelta, record.AbsoluteScoreDelta, record.DecisionChanged, record.DivergenceCategory,
		record.ProductionLatencyMs, record.ShadowInferenceLatencyMs, record.ShadowTotalLatencyMs,
		record.ShadowError,
	)
	if err != nil {
		return fmt.Errorf("failed to insert shadow evaluation into clickhouse: %w", err)
	}

	return nil
}

// CanaryRolloutEvaluation represents a structured canary rollout evaluation record in ClickHouse.
type CanaryRolloutEvaluation struct {
	EvaluationID             string    `json:"evaluation_id"`
	TenantID                 string    `json:"tenant_id"`
	TransactionID            string    `json:"transaction_id"`
	Timestamp                time.Time `json:"timestamp"`
	ModelRoute               string    `json:"model_route"`
	ProductionModelVersion   string    `json:"production_model_version"`
	CandidateModelVersion    string    `json:"candidate_model_version"`
	ProductionScore          float64   `json:"production_score"`
	CandidateScore           float64   `json:"candidate_score"`
	ProductionDecision       string    `json:"production_decision"`
	CandidateDecision        string    `json:"candidate_decision"`
	ScoreDelta               float64   `json:"score_delta"`
	AbsoluteScoreDelta       float64   `json:"absolute_score_delta"`
	DecisionChanged          uint8     `json:"decision_changed"`
	CandidateLatencyMs       float64   `json:"candidate_latency_ms"`
	FallbackUsed             uint8     `json:"fallback_used"`
	Error                    string    `json:"error"`
}

// InsertCanaryEvaluation inserts a single canary evaluation record into canary_rollout_evaluations.
func (c *ClickHouseClient) InsertCanaryEvaluation(ctx context.Context, record CanaryRolloutEvaluation) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO canary_rollout_evaluations (
			evaluation_id, tenant_id, transaction_id, timestamp,
			model_route, production_model_version, candidate_model_version,
			production_score, candidate_score,
			production_decision, candidate_decision,
			score_delta, absolute_score_delta, decision_changed,
			candidate_latency_ms, fallback_used, error
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?,
			?, ?,
			?, ?,
			?, ?, ?,
			?, ?, ?
		)
	`
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	}

	err := c.conn.Exec(ctx, query,
		record.EvaluationID, record.TenantID, record.TransactionID, record.Timestamp,
		record.ModelRoute, record.ProductionModelVersion, record.CandidateModelVersion,
		record.ProductionScore, record.CandidateScore,
		record.ProductionDecision, record.CandidateDecision,
		record.ScoreDelta, record.AbsoluteScoreDelta, record.DecisionChanged,
		record.CandidateLatencyMs, record.FallbackUsed, record.Error,
	)
	if err != nil {
		return fmt.Errorf("failed to insert canary evaluation into clickhouse: %w", err)
	}

	return nil
}

// CanaryRolloutEvent represents an administrative or automated rollout change event in ClickHouse.
type CanaryRolloutEvent struct {
	EventID               string    `json:"event_id"`
	Timestamp             time.Time `json:"timestamp"`
	EventType             string    `json:"event_type"`
	PreviousPercentage    uint8     `json:"previous_percentage"`
	NewPercentage         uint8     `json:"new_percentage"`
	PreviousModelVersion  string    `json:"previous_model_version"`
	NewModelVersion       string    `json:"new_model_version"`
	Trigger               string    `json:"trigger"`
	SafetyStatus          string    `json:"safety_status"`
	ErrorRate             float64   `json:"error_rate"`
	FallbackRate          float64   `json:"fallback_rate"`
	DecisionChangeRate    float64   `json:"decision_change_rate"`
	P95LatencyMs          float64   `json:"p95_latency_ms"`
	P99LatencyMs          float64   `json:"p99_latency_ms"`
	Actor                 string    `json:"actor"`
	Reason                string    `json:"reason"`
}

// InsertCanaryRolloutEvent inserts a single rollout event into canary_rollout_events.
func (c *ClickHouseClient) InsertCanaryRolloutEvent(ctx context.Context, event CanaryRolloutEvent) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO canary_rollout_events (
			event_id, timestamp, event_type,
			previous_percentage, new_percentage,
			previous_model_version, new_model_version,
			trigger, safety_status,
			error_rate, fallback_rate, decision_change_rate,
			p95_latency_ms, p99_latency_ms,
			actor, reason
		) VALUES (
			?, ?, ?,
			?, ?,
			?, ?,
			?, ?,
			?, ?, ?,
			?, ?,
			?, ?
		)
	`
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	err := c.conn.Exec(ctx, query,
		event.EventID, event.Timestamp, event.EventType,
		event.PreviousPercentage, event.NewPercentage,
		event.PreviousModelVersion, event.NewModelVersion,
		event.Trigger, event.SafetyStatus,
		event.ErrorRate, event.FallbackRate, event.DecisionChangeRate,
		event.P95LatencyMs, event.P99LatencyMs,
		event.Actor, event.Reason,
	)
	if err != nil {
		return fmt.Errorf("failed to insert canary rollout event into clickhouse: %w", err)
	}

	return nil
}

// DriftBaselineRecord represents a model baseline distribution record in ClickHouse.
type DriftBaselineRecord struct {
	BaselineID         string    `json:"baseline_id"`
	ModelVersion       string    `json:"model_version"`
	FeatureContract    string    `json:"feature_contract"`
	CalibrationVersion string    `json:"calibration_version"`
	DatasetVersion     string    `json:"dataset_version"`
	CreatedAt          time.Time `json:"created_at"`
	FeatureCount       uint16    `json:"feature_count"`
	Metadata           string    `json:"metadata"`
}

// InsertDriftBaseline inserts a baseline definition record into drift_baselines.
func (c *ClickHouseClient) InsertDriftBaseline(ctx context.Context, record DriftBaselineRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO drift_baselines (
			baseline_id, model_version, feature_contract, calibration_version,
			dataset_version, created_at, feature_count, metadata
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?
		)
	`
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		record.BaselineID, record.ModelVersion, record.FeatureContract, record.CalibrationVersion,
		record.DatasetVersion, record.CreatedAt, record.FeatureCount, record.Metadata,
	)
}

// DriftMeasurementRecord represents an aggregated drift measurement snapshot in ClickHouse.
type DriftMeasurementRecord struct {
	MeasurementID        string    `json:"measurement_id"`
	Timestamp            time.Time `json:"timestamp"`
	ModelVersion         string    `json:"model_version"`
	BaselineID           string    `json:"baseline_id"`
	EvaluationWindow     uint32    `json:"evaluation_window"`
	SampleCount          uint32    `json:"sample_count"`
	OverallStatus        string    `json:"overall_status"`
	MaxPSI               float64   `json:"max_psi"`
	MaxJSD               float64   `json:"max_jsd"`
	MaxKL                float64   `json:"max_kl"`
	DriftedFeatureCount  uint16    `json:"drifted_feature_count"`
	CriticalFeatureCount uint16    `json:"critical_feature_count"`
}

// InsertDriftMeasurement inserts a drift measurement snapshot into drift_measurements.
func (c *ClickHouseClient) InsertDriftMeasurement(ctx context.Context, record DriftMeasurementRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO drift_measurements (
			measurement_id, timestamp, model_version, baseline_id,
			evaluation_window, sample_count, overall_status,
			max_psi, max_jsd, max_kl,
			drifted_feature_count, critical_feature_count
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?
		)
	`
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		record.MeasurementID, record.Timestamp, record.ModelVersion, record.BaselineID,
		record.EvaluationWindow, record.SampleCount, record.OverallStatus,
		record.MaxPSI, record.MaxJSD, record.MaxKL,
		record.DriftedFeatureCount, record.CriticalFeatureCount,
	)
}

// DriftFeatureMeasurementRecord represents a single feature drift record in ClickHouse.
type DriftFeatureMeasurementRecord struct {
	MeasurementID string    `json:"measurement_id"`
	Timestamp     time.Time `json:"timestamp"`
	FeatureName   string    `json:"feature_name"`
	SampleCount   uint32    `json:"sample_count"`
	PSI           float64   `json:"psi"`
	JSD           float64   `json:"jsd"`
	KL            float64   `json:"kl"`
	BaselineMean  float64   `json:"baseline_mean"`
	LiveMean      float64   `json:"live_mean"`
	MeanShift     float64   `json:"mean_shift"`
	BaselineStd   float64   `json:"baseline_std"`
	LiveStd       float64   `json:"live_std"`
	StdShift      float64   `json:"std_shift"`
	MissingRate   float64   `json:"missing_rate"`
	Severity      string    `json:"severity"`
}

// InsertDriftFeatureMeasurements inserts multiple feature drift records into drift_feature_measurements.
func (c *ClickHouseClient) InsertDriftFeatureMeasurements(ctx context.Context, records []DriftFeatureMeasurementRecord) error {
	if c == nil || c.conn == nil || len(records) == 0 {
		return nil
	}

	query := `
		INSERT INTO drift_feature_measurements (
			measurement_id, timestamp, feature_name, sample_count,
			psi, jsd, kl,
			baseline_mean, live_mean, mean_shift,
			baseline_std, live_std, std_shift,
			missing_rate, severity
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?
		)
	`
	for _, r := range records {
		if r.Timestamp.IsZero() {
			r.Timestamp = time.Now().UTC()
		}
		if err := c.conn.Exec(ctx, query,
			r.MeasurementID, r.Timestamp, r.FeatureName, r.SampleCount,
			r.PSI, r.JSD, r.KL,
			r.BaselineMean, r.LiveMean, r.MeanShift,
			r.BaselineStd, r.LiveStd, r.StdShift,
			r.MissingRate, r.Severity,
		); err != nil {
			return err
		}
	}

	return nil
}

// DriftEventRecord represents an operational state transition event in ClickHouse.
type DriftEventRecord struct {
	EventID              string    `json:"event_id"`
	Timestamp            time.Time `json:"timestamp"`
	ModelVersion         string    `json:"model_version"`
	BaselineID           string    `json:"baseline_id"`
	PreviousStatus       string    `json:"previous_status"`
	NewStatus            string    `json:"new_status"`
	MaxPSI               float64   `json:"max_psi"`
	MaxJSD               float64   `json:"max_jsd"`
	MaxKL                float64   `json:"max_kl"`
	AffectedFeatureCount uint16    `json:"affected_feature_count"`
	CriticalFeatureCount uint16    `json:"critical_feature_count"`
	Trigger              string    `json:"trigger"`
	Reason               string    `json:"reason"`
}

// InsertDriftEvent persists an operational drift event to drift_events.
func (c *ClickHouseClient) InsertDriftEvent(ctx context.Context, event DriftEventRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO drift_events (
			event_id, timestamp, model_version, baseline_id,
			previous_status, new_status,
			max_psi, max_jsd, max_kl,
			affected_feature_count, critical_feature_count,
			trigger, reason
		) VALUES (
			?, ?, ?, ?,
			?, ?,
			?, ?, ?,
			?, ?,
			?, ?
		)
	`
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		event.EventID, event.Timestamp, event.ModelVersion, event.BaselineID,
		event.PreviousStatus, event.NewStatus,
		event.MaxPSI, event.MaxJSD, event.MaxKL,
		event.AffectedFeatureCount, event.CriticalFeatureCount,
		event.Trigger, event.Reason,
	)
}

// -------------------------------------------------------------
// Phase 3.16: Retraining & Closed-Loop Lifecycle Records
// -------------------------------------------------------------

// RetrainingJobRecord represents a retraining execution job in ClickHouse.
type RetrainingJobRecord struct {
	JobID                 string    `json:"job_id"`
	TriggeredAt           time.Time `json:"triggered_at"`
	State                 string    `json:"state"`
	TriggerType           string    `json:"trigger_type"`
	TriggerReason         string    `json:"trigger_reason"`
	ParentModelVersion    string    `json:"parent_model_version"`
	CandidateModelVersion string    `json:"candidate_model_version"`
	DatasetID             string    `json:"dataset_id"`
	SampleCount           uint32    `json:"sample_count"`
	CompletedAt           time.Time `json:"completed_at"`
	DurationMs            float64   `json:"duration_ms"`
	Error                 string    `json:"error"`
}

// InsertRetrainingJob records a retraining job snapshot.
func (c *ClickHouseClient) InsertRetrainingJob(ctx context.Context, job RetrainingJobRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO retraining_jobs (
			job_id, triggered_at, state, trigger_type, trigger_reason,
			parent_model_version, candidate_model_version, dataset_id, sample_count,
			completed_at, duration_ms, error
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?
		)
	`
	if job.TriggeredAt.IsZero() {
		job.TriggeredAt = time.Now().UTC()
	}
	if job.CompletedAt.IsZero() {
		job.CompletedAt = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		job.JobID, job.TriggeredAt, job.State, job.TriggerType, job.TriggerReason,
		job.ParentModelVersion, job.CandidateModelVersion, job.DatasetID, job.SampleCount,
		job.CompletedAt, job.DurationMs, job.Error,
	)
}

// ModelCandidateRecord represents an immutable registered candidate model in ClickHouse.
type ModelCandidateRecord struct {
	ModelID            string    `json:"model_id"`
	Version            string    `json:"version"`
	ParentModelVersion string    `json:"parent_model_version"`
	FeatureContract    string    `json:"feature_contract"`
	CalibrationVersion string    `json:"calibration_version"`
	TrainingJobID      string    `json:"training_job_id"`
	DatasetID          string    `json:"dataset_id"`
	CreatedAt          time.Time `json:"created_at"`
	ArtifactChecksum   string    `json:"artifact_checksum"`
	ConfigHash         string    `json:"config_hash"`
	State              string    `json:"state"`
}

// InsertModelCandidate registers a newly trained candidate model.
func (c *ClickHouseClient) InsertModelCandidate(ctx context.Context, candidate ModelCandidateRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO model_candidates (
			model_id, version, parent_model_version, feature_contract, calibration_version,
			training_job_id, dataset_id, created_at, artifact_checksum, config_hash, state
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?
		)
	`
	if candidate.CreatedAt.IsZero() {
		candidate.CreatedAt = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		candidate.ModelID, candidate.Version, candidate.ParentModelVersion,
		candidate.FeatureContract, candidate.CalibrationVersion, candidate.TrainingJobID,
		candidate.DatasetID, candidate.CreatedAt, candidate.ArtifactChecksum,
		candidate.ConfigHash, candidate.State,
	)
}

// ModelValidationRecord represents offline validation scorecards in ClickHouse.
type ModelValidationRecord struct {
	ValidationID       string    `json:"validation_id"`
	Timestamp          time.Time `json:"timestamp"`
	ModelID            string    `json:"model_id"`
	ModelVersion       string    `json:"model_version"`
	ParentModelVersion string    `json:"parent_model_version"`
	ROCAUC             float64   `json:"roc_auc"`
	PRAUC              float64   `json:"pr_auc"`
	Precision          float64   `json:"precision"`
	Recall             float64   `json:"recall"`
	FPR                float64   `json:"fpr"`
	FNR                float64   `json:"fnr"`
	BrierScore         float64   `json:"brier_score"`
	CalibrationError   float64   `json:"calibration_error"`
	P95LatencyMs       float64   `json:"p95_latency_ms"`
	Passed             uint8     `json:"passed"`
	GateDetails        string    `json:"gate_details"`
}

// InsertValidationResult persists offline validation gate results.
func (c *ClickHouseClient) InsertValidationResult(ctx context.Context, res ModelValidationRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO model_validation_results (
			validation_id, timestamp, model_id, model_version, parent_model_version,
			roc_auc, pr_auc, precision, recall, fpr, fnr,
			brier_score, calibration_error, p95_latency_ms, passed, gate_details
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`
	if res.Timestamp.IsZero() {
		res.Timestamp = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		res.ValidationID, res.Timestamp, res.ModelID, res.ModelVersion, res.ParentModelVersion,
		res.ROCAUC, res.PRAUC, res.Precision, res.Recall, res.FPR, res.FNR,
		res.BrierScore, res.CalibrationError, res.P95LatencyMs, res.Passed, res.GateDetails,
	)
}

// ModelShadowEvaluationRecord represents shadow evaluation summary metrics in ClickHouse.
type ModelShadowEvaluationRecord struct {
	EvaluationID            string    `json:"evaluation_id"`
	Timestamp               time.Time `json:"timestamp"`
	CandidateModelVersion   string    `json:"candidate_model_version"`
	ProductionModelVersion  string    `json:"production_model_version"`
	SamplesEvaluated        uint32    `json:"samples_evaluated"`
	ScoreDivergenceRate     float64   `json:"score_divergence_rate"`
	DecisionChangeRate      float64   `json:"decision_change_rate"`
	ErrorRate               float64   `json:"error_rate"`
	FallbackRate            float64   `json:"fallback_rate"`
	AvgScoreDelta           float64   `json:"avg_score_delta"`
	P95LatencyMs            float64   `json:"p95_latency_ms"`
	Passed                  uint8     `json:"passed"`
	GateDetails             string    `json:"gate_details"`
}

// InsertModelShadowEvaluation records shadow evaluation gate results.
func (c *ClickHouseClient) InsertModelShadowEvaluation(ctx context.Context, eval ModelShadowEvaluationRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO model_shadow_evaluations (
			evaluation_id, timestamp, candidate_model_version, production_model_version,
			samples_evaluated, score_divergence_rate, decision_change_rate,
			error_rate, fallback_rate, avg_score_delta, p95_latency_ms, passed, gate_details
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?, ?, ?
		)
	`
	if eval.Timestamp.IsZero() {
		eval.Timestamp = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		eval.EvaluationID, eval.Timestamp, eval.CandidateModelVersion, eval.ProductionModelVersion,
		eval.SamplesEvaluated, eval.ScoreDivergenceRate, eval.DecisionChangeRate,
		eval.ErrorRate, eval.FallbackRate, eval.AvgScoreDelta, eval.P95LatencyMs, eval.Passed, eval.GateDetails,
	)
}

// ModelLifecycleEventRecord represents state machine transition events in ClickHouse.
type ModelLifecycleEventRecord struct {
	EventID       string    `json:"event_id"`
	Timestamp     time.Time `json:"timestamp"`
	ModelID       string    `json:"model_id"`
	ModelVersion  string    `json:"model_version"`
	PreviousState string    `json:"previous_state"`
	NewState      string    `json:"new_state"`
	Trigger       string    `json:"trigger"`
	Actor         string    `json:"actor"`
	Reason        string    `json:"reason"`
}

// InsertModelLifecycleEvent persists state machine transition events.
func (c *ClickHouseClient) InsertModelLifecycleEvent(ctx context.Context, event ModelLifecycleEventRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO model_lifecycle_events (
			event_id, timestamp, model_id, model_version,
			previous_state, new_state, trigger, actor, reason
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		event.EventID, event.Timestamp, event.ModelID, event.ModelVersion,
		event.PreviousState, event.NewState, event.Trigger, event.Actor, event.Reason,
	)
}

// CandidateCanaryMetricRecord represents candidate performance across canary rollout stages.
type CandidateCanaryMetricRecord struct {
	MetricID              string    `json:"metric_id"`
	Timestamp             time.Time `json:"timestamp"`
	RolloutID             string    `json:"rollout_id"`
	CandidateModelVersion string    `json:"candidate_model_version"`
	StagePercentage       uint8     `json:"stage_percentage"`
	SampleCount           uint32    `json:"sample_count"`
	ErrorRate             float64   `json:"error_rate"`
	FallbackRate          float64   `json:"fallback_rate"`
	P95LatencyMs          float64   `json:"p95_latency_ms"`
	P99LatencyMs          float64   `json:"p99_latency_ms"`
	DecisionChangeRate    float64   `json:"decision_change_rate"`
	Passed                uint8     `json:"passed"`
	Action                string    `json:"action"`
}

// InsertCandidateCanaryMetrics logs canary stage verification metrics.
func (c *ClickHouseClient) InsertCandidateCanaryMetrics(ctx context.Context, m CandidateCanaryMetricRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO candidate_canary_metrics (
			metric_id, timestamp, rollout_id, candidate_model_version,
			stage_percentage, sample_count, error_rate, fallback_rate,
			p95_latency_ms, p99_latency_ms, decision_change_rate, passed, action
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now().UTC()
	}

	return c.conn.Exec(ctx, query,
		m.MetricID, m.Timestamp, m.RolloutID, m.CandidateModelVersion,
		m.StagePercentage, m.SampleCount, m.ErrorRate, m.FallbackRate,
		m.P95LatencyMs, m.P99LatencyMs, m.DecisionChangeRate, m.Passed, m.Action,
	)
}

// Phase 3.19 Operational Telemetry Structs

type SLOMeasurementRecord struct {
	MeasurementID        string    `json:"measurement_id"`
	Timestamp            time.Time `json:"timestamp"`
	SLOID                string    `json:"slo_id"`
	SLOName              string    `json:"slo_name"`
	CurrentValue         float64   `json:"current_value"`
	Target               float64   `json:"target"`
	Status               string    `json:"status"`
	ErrorBudget          float64   `json:"error_budget"`
	ErrorBudgetRemaining float64   `json:"error_budget_remaining"`
	BurnRate             float64   `json:"burn_rate"`
	SampleCount          uint32    `json:"sample_count"`
}

type IncidentAuditRecord struct {
	IncidentID      string    `json:"incident_id"`
	Timestamp       time.Time `json:"timestamp"`
	Severity        string    `json:"severity"`
	Category        string    `json:"category"`
	Status          string    `json:"status"`
	Subsystem       string    `json:"subsystem"`
	Reason          string    `json:"reason"`
	OccurrenceCount uint32    `json:"occurrence_count"`
	ModelVersion    string    `json:"model_version"`
	CorrelationID   string    `json:"correlation_id"`
}

type AlertAuditRecord struct {
	AlertID       string    `json:"alert_id"`
	Timestamp     time.Time `json:"timestamp"`
	Severity      string    `json:"severity"`
	Title         string    `json:"title"`
	Message       string    `json:"message"`
	IncidentID    string    `json:"incident_id"`
	Subsystem     string    `json:"subsystem"`
	CorrelationID string    `json:"correlation_id"`
}

type OperationalStateEventRecord struct {
	EventID     string    `json:"event_id"`
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	ControlName string    `json:"control_name"`
	Enabled     uint8     `json:"enabled"`
	Actor       string    `json:"actor"`
	Reason      string    `json:"reason"`
}

// InsertSLOMeasurement inserts an SLO computation record into operational_slo_measurements.
func (c *ClickHouseClient) InsertSLOMeasurement(ctx context.Context, m SLOMeasurementRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}
	query := `
		INSERT INTO operational_slo_measurements (
			measurement_id, timestamp, slo_id, slo_name,
			current_value, target, status, error_budget,
			error_budget_remaining, burn_rate, sample_count
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?
		)
	`
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now().UTC()
	}
	return c.conn.Exec(ctx, query,
		m.MeasurementID, m.Timestamp, m.SLOID, m.SLOName,
		m.CurrentValue, m.Target, m.Status, m.ErrorBudget,
		m.ErrorBudgetRemaining, m.BurnRate, m.SampleCount,
	)
}

// InsertIncidentRecord inserts an incident into operational_incidents.
func (c *ClickHouseClient) InsertIncidentRecord(ctx context.Context, inc IncidentAuditRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}
	query := `
		INSERT INTO operational_incidents (
			incident_id, timestamp, severity, category,
			status, subsystem, reason, occurrence_count,
			model_version, correlation_id
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?
		)
	`
	if inc.Timestamp.IsZero() {
		inc.Timestamp = time.Now().UTC()
	}
	return c.conn.Exec(ctx, query,
		inc.IncidentID, inc.Timestamp, inc.Severity, inc.Category,
		inc.Status, inc.Subsystem, inc.Reason, inc.OccurrenceCount,
		inc.ModelVersion, inc.CorrelationID,
	)
}

// InsertAlertRecord inserts an alert record into operational_alerts.
func (c *ClickHouseClient) InsertAlertRecord(ctx context.Context, alt AlertAuditRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}
	query := `
		INSERT INTO operational_alerts (
			alert_id, timestamp, severity, title,
			message, incident_id, subsystem, correlation_id
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?
		)
	`
	if alt.Timestamp.IsZero() {
		alt.Timestamp = time.Now().UTC()
	}
	return c.conn.Exec(ctx, query,
		alt.AlertID, alt.Timestamp, alt.Severity, alt.Title,
		alt.Message, alt.IncidentID, alt.Subsystem, alt.CorrelationID,
	)
}

// InsertOperationalStateEvent inserts a control plane mutation into operational_state_events.
func (c *ClickHouseClient) InsertOperationalStateEvent(ctx context.Context, evt OperationalStateEventRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}
	query := `
		INSERT INTO operational_state_events (
			event_id, timestamp, event_type, control_name,
			enabled, actor, reason
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?
		)
	`
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now().UTC()
	}
	return c.conn.Exec(ctx, query,
		evt.EventID, evt.Timestamp, evt.EventType, evt.ControlName,
		evt.Enabled, evt.Actor, evt.Reason,
	)
}

// Close terminates the ClickHouse connection.
func (c *ClickHouseClient) Close() error {
	if c != nil && c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
