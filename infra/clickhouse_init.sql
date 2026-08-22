-- infra/clickhouse_init.sql
-- ClickHouse Analytical Store for High-Throughput Risk Decisions and Feature Snapshots

CREATE TABLE IF NOT EXISTS risk_audit_log (
    transaction_id String,
    risk_score Int32,
    rule_triggered String,
    feature_snapshot String,
    tenant_id String,
    created_at DateTime
) ENGINE = MergeTree()
ORDER BY (created_at, tenant_id);

CREATE TABLE IF NOT EXISTS shadow_score_evaluations (
    evaluation_id String,
    tenant_id String,
    transaction_id String,
    timestamp DateTime,
    production_model_version String,
    shadow_model_version String,
    production_feature_contract String,
    shadow_feature_contract String,
    production_raw_score Float64,
    production_calibrated_score Float64,
    shadow_raw_score Float64,
    shadow_calibrated_score Float64,
    production_decision String,
    shadow_decision String,
    score_delta Float64,
    absolute_score_delta Float64,
    decision_changed UInt8,
    divergence_category String,
    production_latency_ms Float64,
    shadow_inference_latency_ms Float64,
) ENGINE = MergeTree()
ORDER BY (timestamp, tenant_id, evaluation_id);

CREATE TABLE IF NOT EXISTS canary_rollout_evaluations (
    evaluation_id String,
    tenant_id String,
    transaction_id String,
    timestamp DateTime,
    model_route String,
    production_model_version String,
    candidate_model_version String,
    production_score Float64,
    candidate_score Float64,
    production_decision String,
    candidate_decision String,
    score_delta Float64,
    absolute_score_delta Float64,
    decision_changed UInt8,
    candidate_latency_ms Float64,
    fallback_used UInt8,
    error String
) ENGINE = MergeTree()
ORDER BY (timestamp, tenant_id, evaluation_id);

CREATE TABLE IF NOT EXISTS canary_rollout_events (
    event_id String,
    timestamp DateTime,
    event_type String,
    previous_percentage UInt8,
    new_percentage UInt8,
    previous_model_version String,
    new_model_version String,
    trigger String,
    safety_status String,
    error_rate Float64,
    fallback_rate Float64,
    decision_change_rate Float64,
    p95_latency_ms Float64,
    p99_latency_ms Float64,
    actor String,
    reason String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_id);

CREATE TABLE IF NOT EXISTS drift_baselines (
    baseline_id String,
    model_version String,
    feature_contract String,
    calibration_version String,
    dataset_version String,
    created_at DateTime,
    feature_count UInt16,
    metadata String
) ENGINE = MergeTree()
ORDER BY (created_at, baseline_id);

CREATE TABLE IF NOT EXISTS drift_measurements (
    measurement_id String,
    timestamp DateTime,
    model_version String,
    baseline_id String,
    evaluation_window UInt32,
    sample_count UInt32,
    overall_status String,
    max_psi Float64,
    max_jsd Float64,
    max_kl Float64,
    drifted_feature_count UInt16,
    critical_feature_count UInt16
) ENGINE = MergeTree()
ORDER BY (timestamp, measurement_id);

CREATE TABLE IF NOT EXISTS drift_feature_measurements (
    measurement_id String,
    timestamp DateTime,
    feature_name String,
    sample_count UInt32,
    psi Float64,
    jsd Float64,
    kl Float64,
    baseline_mean Float64,
    live_mean Float64,
    mean_shift Float64,
    baseline_std Float64,
    live_std Float64,
    std_shift Float64,
    missing_rate Float64,
    severity String
) ENGINE = MergeTree()
ORDER BY (timestamp, measurement_id, feature_name);

CREATE TABLE IF NOT EXISTS drift_events (
    event_id String,
    timestamp DateTime,
    model_version String,
    baseline_id String,
    previous_status String,
    new_status String,
    max_psi Float64,
    max_jsd Float64,
    max_kl Float64,
    affected_feature_count UInt16,
    critical_feature_count UInt16,
    trigger String,
    reason String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_id);

-- Phase 3.16: Automated Retraining & Closed-Loop Lifecycle Tables

CREATE TABLE IF NOT EXISTS retraining_jobs (
    job_id String,
    triggered_at DateTime,
    state String,
    trigger_type String,
    trigger_reason String,
    parent_model_version String,
    candidate_model_version String,
    dataset_id String,
    sample_count UInt32,
    completed_at DateTime,
    duration_ms Float64,
    error String
) ENGINE = MergeTree()
ORDER BY (triggered_at, job_id);

CREATE TABLE IF NOT EXISTS model_candidates (
    model_id String,
    version String,
    parent_model_version String,
    feature_contract String,
    calibration_version String,
    training_job_id String,
    dataset_id String,
    created_at DateTime,
    artifact_checksum String,
    config_hash String,
    state String
) ENGINE = MergeTree()
ORDER BY (created_at, model_id);

CREATE TABLE IF NOT EXISTS model_validation_results (
    validation_id String,
    timestamp DateTime,
    model_id String,
    model_version String,
    parent_model_version String,
    roc_auc Float64,
    pr_auc Float64,
    precision Float64,
    recall Float64,
    fpr Float64,
    fnr Float64,
    brier_score Float64,
    calibration_error Float64,
    p95_latency_ms Float64,
    passed UInt8,
    gate_details String
) ENGINE = MergeTree()
ORDER BY (timestamp, validation_id);

CREATE TABLE IF NOT EXISTS model_shadow_evaluations (
    evaluation_id String,
    timestamp DateTime,
    candidate_model_version String,
    production_model_version String,
    samples_evaluated UInt32,
    score_divergence_rate Float64,
    decision_change_rate Float64,
    error_rate Float64,
    fallback_rate Float64,
    avg_score_delta Float64,
    p95_latency_ms Float64,
    passed UInt8,
    gate_details String
) ENGINE = MergeTree()
ORDER BY (timestamp, evaluation_id);

CREATE TABLE IF NOT EXISTS model_lifecycle_events (
    event_id String,
    timestamp DateTime,
    model_id String,
    model_version String,
    previous_state String,
    new_state String,
    trigger String,
    actor String,
    reason String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_id);

CREATE TABLE IF NOT EXISTS candidate_canary_metrics (
    metric_id String,
    timestamp DateTime,
    rollout_id String,
    candidate_model_version String,
    stage_percentage UInt8,
    sample_count UInt32,
    error_rate Float64,
    fallback_rate Float64,
    p95_latency_ms Float64,
    p99_latency_ms Float64,
    decision_change_rate Float64,
    passed UInt8,
    action String
) ENGINE = MergeTree()
ORDER BY (timestamp, metric_id);

-- Phase 3.19: Production Observability, SLOs & Operational Control Plane

CREATE TABLE IF NOT EXISTS operational_slo_measurements (
    measurement_id String,
    timestamp DateTime,
    slo_id String,
    slo_name String,
    current_value Float64,
    target Float64,
    status String,
    error_budget Float64,
    error_budget_remaining Float64,
    burn_rate Float64,
    sample_count UInt32
) ENGINE = MergeTree()
ORDER BY (timestamp, slo_id);

CREATE TABLE IF NOT EXISTS operational_incidents (
    incident_id String,
    timestamp DateTime,
    severity String,
    category String,
    status String,
    subsystem String,
    reason String,
    occurrence_count UInt32,
    model_version String,
    correlation_id String
) ENGINE = MergeTree()
ORDER BY (timestamp, incident_id);

CREATE TABLE IF NOT EXISTS operational_alerts (
    alert_id String,
    timestamp DateTime,
    severity String,
    title String,
    message String,
    incident_id String,
    subsystem String,
    correlation_id String
) ENGINE = MergeTree()
ORDER BY (timestamp, alert_id);

CREATE TABLE IF NOT EXISTS operational_state_events (
    event_id String,
    timestamp DateTime,
    event_type String,
    control_name String,
    enabled UInt8,
    actor String,
    reason String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_id);
