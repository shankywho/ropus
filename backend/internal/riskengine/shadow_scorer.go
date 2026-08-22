package riskengine

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shankywho/ropus/backend/internal/audit"
)

// ShadowScorerConfig defines operational parameters for asynchronous shadow scoring.
type ShadowScorerConfig struct {
	Enabled                  bool
	WorkerCount              int
	QueueCapacity            int
	SampleRate               float64
	ScoreDivergenceThreshold float64
	CandidateModelVersion    string
	CandidateFeatureContract string
}

// DefaultShadowScorerConfig returns sensible production defaults.
func DefaultShadowScorerConfig() ShadowScorerConfig {
	return ShadowScorerConfig{
		Enabled:                  true,
		WorkerCount:              4,
		QueueCapacity:            1000,
		SampleRate:               1.0,
		ScoreDivergenceThreshold: 0.05,
		CandidateModelVersion:    "fraud-xgb-25f-candidate-v1",
		CandidateFeatureContract: MLFeatureContractV25,
	}
}

// ShadowScoreTask encapsulates an immutable transaction snapshot for shadow scoring.
type ShadowScoreTask struct {
	EvaluationID              string
	TenantID                  string
	TransactionID             string
	Timestamp                 time.Time
	Amount                    float64
	Canonical25Vector         *MLFeatureVector
	ProductionModelVersion    string
	ProductionFeatureContract string
	ProductionRawScore        float64
	ProductionCalibratedScore float64
	ProductionDecision        string
	ProductionLatencyMs       float64
	EnqueuedAt                time.Time
}

// ShadowScoreResult represents the complete comparison between production and candidate evaluation.
type ShadowScoreResult struct {
	EvaluationID              string    `json:"evaluation_id"`
	TenantID                  string    `json:"tenant_id"`
	TransactionID             string    `json:"transaction_id"`
	Timestamp                 time.Time `json:"timestamp"`
	ProductionModelVersion    string    `json:"production_model_version"`
	ShadowModelVersion        string    `json:"shadow_model_version"`
	ProductionFeatureContract string    `json:"production_feature_contract"`
	ShadowFeatureContract     string    `json:"shadow_feature_contract"`
	ProductionRawScore        float64   `json:"production_raw_score"`
	ProductionCalibratedScore float64   `json:"production_calibrated_score"`
	ShadowRawScore            float64   `json:"shadow_raw_score"`
	ShadowCalibratedScore     float64   `json:"shadow_calibrated_score"`
	ProductionDecision        string    `json:"production_decision"`
	ShadowDecision            string    `json:"shadow_decision"`
	ScoreDelta                float64   `json:"score_delta"`
	AbsoluteScoreDelta        float64   `json:"absolute_score_delta"`
	DecisionChanged           bool      `json:"decision_changed"`
	DivergenceCategory        string    `json:"divergence_category"`
	ProductionLatencyMs       float64   `json:"production_latency_ms"`
	ShadowInferenceLatencyMs  float64   `json:"shadow_inference_latency_ms"`
	ShadowTotalLatencyMs      float64   `json:"shadow_total_latency_ms"`
	ShadowError               string    `json:"shadow_error,omitempty"`
}

// ShadowMetrics tracks live shadow scoring operational counters atomically.
type ShadowMetrics struct {
	RequestsTotal             atomic.Int64
	SuccessTotal              atomic.Int64
	ErrorsTotal               atomic.Int64
	QueueDroppedTotal         atomic.Int64
	ScoreDivergenceTotal      atomic.Int64
	DecisionDivergenceTotal   atomic.Int64
	ProductionAllowTotal      atomic.Int64
	ProductionReviewTotal     atomic.Int64
	ProductionDeclineTotal    atomic.Int64
	CandidateAllowTotal       atomic.Int64
	CandidateReviewTotal      atomic.Int64
	CandidateDeclineTotal     atomic.Int64
	TotalInferenceLatencyUs   atomic.Int64
	TotalPipelineLatencyUs    atomic.Int64
}

// Snapshot returns a point-in-time dictionary of current shadow metrics.
func (m *ShadowMetrics) Snapshot(queueDepth int) map[string]interface{} {
	success := m.SuccessTotal.Load()
	avgInfMs := 0.0
	avgTotMs := 0.0
	if success > 0 {
		avgInfMs = float64(m.TotalInferenceLatencyUs.Load()) / (float64(success) * 1000.0)
		avgTotMs = float64(m.TotalPipelineLatencyUs.Load()) / (float64(success) * 1000.0)
	}

	return map[string]interface{}{
		"requests_total":             m.RequestsTotal.Load(),
		"success_total":              success,
		"errors_total":               m.ErrorsTotal.Load(),
		"queue_depth":                queueDepth,
		"queue_dropped_total":        m.QueueDroppedTotal.Load(),
		"score_divergence_total":     m.ScoreDivergenceTotal.Load(),
		"decision_divergence_total":  m.DecisionDivergenceTotal.Load(),
		"production_allow_total":     m.ProductionAllowTotal.Load(),
		"production_review_total":    m.ProductionReviewTotal.Load(),
		"production_decline_total":   m.ProductionDeclineTotal.Load(),
		"candidate_allow_total":      m.CandidateAllowTotal.Load(),
		"candidate_review_total":     m.CandidateReviewTotal.Load(),
		"candidate_decline_total":    m.CandidateDeclineTotal.Load(),
		"avg_inference_latency_ms":   math.Round(avgInfMs*100) / 100,
		"avg_total_latency_ms":       math.Round(avgTotMs*100) / 100,
	}
}

// ShadowScorer manages asynchronous shadow scoring worker pool and comparison persistence.
type ShadowScorer struct {
	config    ShadowScorerConfig
	mlClient  *MLClient
	chClient  *audit.ClickHouseClient
	workQueue chan ShadowScoreTask
	metrics   ShadowMetrics
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	closed    atomic.Bool
}

// NewShadowScorer instantiates and starts the bounded shadow worker pool.
func NewShadowScorer(cfg ShadowScorerConfig, mlClient *MLClient, chClient *audit.ClickHouseClient) *ShadowScorer {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.QueueCapacity <= 0 {
		cfg.QueueCapacity = 1000
	}
	if cfg.SampleRate < 0.0 || cfg.SampleRate > 1.0 {
		cfg.SampleRate = 1.0
	}
	if cfg.ScoreDivergenceThreshold <= 0.0 {
		cfg.ScoreDivergenceThreshold = 0.05
	}
	if cfg.CandidateModelVersion == "" {
		cfg.CandidateModelVersion = "fraud-xgb-25f-candidate-v1"
	}
	if cfg.CandidateFeatureContract == "" {
		cfg.CandidateFeatureContract = MLFeatureContractV25
	}

	ctx, cancel := context.WithCancel(context.Background())

	scorer := &ShadowScorer{
		config:    cfg,
		mlClient:  mlClient,
		chClient:  chClient,
		workQueue: make(chan ShadowScoreTask, cfg.QueueCapacity),
		ctx:       ctx,
		cancel:    cancel,
	}

	if cfg.Enabled {
		scorer.startWorkers()
		log.Printf("ShadowScorer initialized: workers=%d, queue_cap=%d, sample_rate=%.2f, model=%s",
			cfg.WorkerCount, cfg.QueueCapacity, cfg.SampleRate, cfg.CandidateModelVersion)
	} else {
		log.Println("ShadowScorer initialized in DISABLED mode.")
	}

	return scorer
}

// startWorkers spawns background worker routines.
func (s *ShadowScorer) startWorkers() {
	for i := 0; i < s.config.WorkerCount; i++ {
		s.wg.Add(1)
		go s.workerRoutine(i)
	}
}

// isSampled deterministically evaluates if an evaluation ID falls within the sampling threshold.
func (s *ShadowScorer) isSampled(evalID string) bool {
	if s.config.SampleRate >= 1.0 {
		return true
	}
	if s.config.SampleRate <= 0.0 {
		return false
	}
	h := sha256.Sum256([]byte(evalID))
	val := binary.BigEndian.Uint32(h[:4])
	ratio := float64(val) / float64(math.MaxUint32)
	return ratio < s.config.SampleRate
}

// Enqueue submits a point-in-time safe feature snapshot to the shadow queue non-blockingly.
func (s *ShadowScorer) Enqueue(task ShadowScoreTask) bool {
	if !s.config.Enabled || s.closed.Load() {
		return false
	}

	if !s.isSampled(task.EvaluationID) {
		return false
	}

	task.EnqueuedAt = time.Now().UTC()
	s.metrics.RequestsTotal.Add(1)

	select {
	case s.workQueue <- task:
		return true
	default:
		// Saturated queue: drop gracefully without blocking production request
		s.metrics.QueueDroppedTotal.Add(1)
		log.Printf("Warning: Shadow scoring queue saturated (depth=%d). Dropping shadow task for eval=%s",
			len(s.workQueue), task.EvaluationID)
		return false
	}
}

// workerRoutine continuously consumes shadow evaluation tasks.
func (s *ShadowScorer) workerRoutine(workerID int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			// Drain remaining tasks if any before exiting
			for {
				select {
				case task, ok := <-s.workQueue:
					if !ok {
						return
					}
					s.processTask(task)
				default:
					return
				}
			}
		case task, ok := <-s.workQueue:
			if !ok {
				return
			}
			s.processTask(task)
		}
	}
}

// processTask executes candidate model inference, evaluates divergence, and records metrics.
func (s *ShadowScorer) processTask(task ShadowScoreTask) {
	defer func() {
		if r := recover(); r != nil {
			s.metrics.ErrorsTotal.Add(1)
			log.Printf("Recovered from panic in shadow worker: %v", r)
		}
	}()

	startProcessTime := time.Now()

	req25F := MLShadowPredictRequest{
		Features:               ToCanonical25FloatSlice(task.Canonical25Vector),
		EvaluationID:           task.EvaluationID,
		TenantID:               task.TenantID,
		TransactionID:          task.TransactionID,
		FeatureContractVersion: s.config.CandidateFeatureContract,
	}

	inferenceStart := time.Now()
	var shadowRawProb float64
	var shadowCalProb float64
	var shadowDecision string
	var shadowInfLatencyMs float64
	var shadowErrStr string

	if s.mlClient != nil {
		resp, err := s.mlClient.PredictShadow(s.ctx, req25F)
		if err != nil {
			s.metrics.ErrorsTotal.Add(1)
			shadowErrStr = err.Error()
			shadowRawProb = 0.05
			shadowCalProb = 0.05
			shadowDecision = "MANUAL_REVIEW"
			shadowInfLatencyMs = float64(time.Since(inferenceStart).Milliseconds())
		} else {
			s.metrics.SuccessTotal.Add(1)
			shadowRawProb = resp.RawProbability
			shadowCalProb = resp.CalibratedProbability
			shadowDecision = resp.ShadowDecision
			shadowInfLatencyMs = resp.LatencyMs
		}
	} else {
		shadowErrStr = "ml_client_uninitialized"
		shadowRawProb = 0.05
		shadowCalProb = 0.05
		shadowDecision = "ALLOW_RECOMMENDATION"
	}

	totalPipelineLatencyMs := float64(time.Since(task.EnqueuedAt).Microseconds()) / 1000.0
	s.metrics.TotalInferenceLatencyUs.Add(time.Since(inferenceStart).Microseconds())
	s.metrics.TotalPipelineLatencyUs.Add(time.Since(task.EnqueuedAt).Microseconds())

	// Compare decisions and scores
	scoreDelta := shadowCalProb - task.ProductionCalibratedScore
	absDelta := math.Abs(scoreDelta)
	decisionChanged := task.ProductionDecision != shadowDecision

	// Track distribution counters
	switch task.ProductionDecision {
	case "ALLOW_RECOMMENDATION":
		s.metrics.ProductionAllowTotal.Add(1)
	case "MANUAL_REVIEW", "STEP_UP_RECOMMENDATION":
		s.metrics.ProductionReviewTotal.Add(1)
	case "DECLINE_RECOMMENDATION":
		s.metrics.ProductionDeclineTotal.Add(1)
	}

	switch shadowDecision {
	case "ALLOW_RECOMMENDATION":
		s.metrics.CandidateAllowTotal.Add(1)
	case "MANUAL_REVIEW", "STEP_UP_RECOMMENDATION":
		s.metrics.CandidateReviewTotal.Add(1)
	case "DECLINE_RECOMMENDATION":
		s.metrics.CandidateDeclineTotal.Add(1)
	}

	// Categorize divergence
	divergenceCategory := s.categorizeDivergence(task.ProductionDecision, shadowDecision, absDelta)
	if absDelta >= s.config.ScoreDivergenceThreshold {
		s.metrics.ScoreDivergenceTotal.Add(1)
	}
	if decisionChanged {
		s.metrics.DecisionDivergenceTotal.Add(1)
	}

	result := ShadowScoreResult{
		EvaluationID:              task.EvaluationID,
		TenantID:                  task.TenantID,
		TransactionID:             task.TransactionID,
		Timestamp:                 task.Timestamp,
		ProductionModelVersion:    task.ProductionModelVersion,
		ShadowModelVersion:        s.config.CandidateModelVersion,
		ProductionFeatureContract: task.ProductionFeatureContract,
		ShadowFeatureContract:     s.config.CandidateFeatureContract,
		ProductionRawScore:        task.ProductionRawScore,
		ProductionCalibratedScore: task.ProductionCalibratedScore,
		ShadowRawScore:            shadowRawProb,
		ShadowCalibratedScore:     shadowCalProb,
		ProductionDecision:        task.ProductionDecision,
		ShadowDecision:            shadowDecision,
		ScoreDelta:                math.Round(scoreDelta*10000) / 10000,
		AbsoluteScoreDelta:        math.Round(absDelta*10000) / 10000,
		DecisionChanged:           decisionChanged,
		DivergenceCategory:        divergenceCategory,
		ProductionLatencyMs:       task.ProductionLatencyMs,
		ShadowInferenceLatencyMs:  shadowInfLatencyMs,
		ShadowTotalLatencyMs:      totalPipelineLatencyMs,
		ShadowError:               shadowErrStr,
	}

	// Structured Observability Logging (Safe: Zero PII/PAN/Credentials logged)
	logEntry, _ := json.Marshal(map[string]interface{}{
		"event":                    "shadow_score_completed",
		"evaluation_id":            result.EvaluationID,
		"tenant_id":                result.TenantID,
		"production_decision":      result.ProductionDecision,
		"shadow_decision":          result.ShadowDecision,
		"production_cal_score":     result.ProductionCalibratedScore,
		"shadow_cal_score":         result.ShadowCalibratedScore,
		"score_delta":              result.ScoreDelta,
		"decision_changed":         result.DecisionChanged,
		"divergence_category":      result.DivergenceCategory,
		"shadow_inf_latency_ms":    result.ShadowInferenceLatencyMs,
		"shadow_total_latency_ms":  result.ShadowTotalLatencyMs,
		"duration_ms":              float64(time.Since(startProcessTime).Microseconds()) / 1000.0,
	})
	log.Println(string(logEntry))

	// Asynchronous Persistence to ClickHouse
	if s.chClient != nil {
		decisionChangedUInt := uint8(0)
		if decisionChanged {
			decisionChangedUInt = 1
		}
		chRecord := audit.ShadowScoreEvaluation{
			EvaluationID:              result.EvaluationID,
			TenantID:                  result.TenantID,
			TransactionID:             result.TransactionID,
			Timestamp:                 result.Timestamp,
			ProductionModelVersion:    result.ProductionModelVersion,
			ShadowModelVersion:        result.ShadowModelVersion,
			ProductionFeatureContract: result.ProductionFeatureContract,
			ShadowFeatureContract:     result.ShadowFeatureContract,
			ProductionRawScore:        result.ProductionRawScore,
			ProductionCalibratedScore: result.ProductionCalibratedScore,
			ShadowRawScore:            result.ShadowRawScore,
			ShadowCalibratedScore:     result.ShadowCalibratedScore,
			ProductionDecision:        result.ProductionDecision,
			ShadowDecision:            result.ShadowDecision,
			ScoreDelta:                result.ScoreDelta,
			AbsoluteScoreDelta:        result.AbsoluteScoreDelta,
			DecisionChanged:           decisionChangedUInt,
			DivergenceCategory:        result.DivergenceCategory,
			ProductionLatencyMs:       result.ProductionLatencyMs,
			ShadowInferenceLatencyMs:  result.ShadowInferenceLatencyMs,
			ShadowTotalLatencyMs:      result.ShadowTotalLatencyMs,
			ShadowError:               result.ShadowError,
		}
		if err := s.chClient.InsertShadowEvaluation(s.ctx, chRecord); err != nil {
			log.Printf("Warning: Failed to persist shadow evaluation to ClickHouse: %v", err)
		}
	}
}

// categorizeDivergence calculates directional action and score divergence categories.
func (s *ShadowScorer) categorizeDivergence(prodAction, shadowAction string, absScoreDelta float64) string {
	if prodAction == shadowAction {
		if absScoreDelta >= s.config.ScoreDivergenceThreshold {
			return "SCORE_DIVERGENCE_DECISION_AGREEMENT"
		}
		return "FULL_AGREEMENT"
	}

	switch {
	case (prodAction == "ALLOW_RECOMMENDATION" || prodAction == "ALLOW") &&
		(shadowAction == "MANUAL_REVIEW" || shadowAction == "STEP_UP_RECOMMENDATION"):
		return "ALLOW_TO_REVIEW"
	case (prodAction == "ALLOW_RECOMMENDATION" || prodAction == "ALLOW") &&
		(shadowAction == "DECLINE_RECOMMENDATION" || shadowAction == "DECLINE"):
		return "ALLOW_TO_DECLINE"
	case (prodAction == "MANUAL_REVIEW" || prodAction == "STEP_UP_RECOMMENDATION") &&
		(shadowAction == "ALLOW_RECOMMENDATION" || shadowAction == "ALLOW"):
		return "REVIEW_TO_ALLOW"
	case (prodAction == "MANUAL_REVIEW" || prodAction == "STEP_UP_RECOMMENDATION") &&
		(shadowAction == "DECLINE_RECOMMENDATION" || shadowAction == "DECLINE"):
		return "REVIEW_TO_DECLINE"
	case (prodAction == "DECLINE_RECOMMENDATION" || prodAction == "DECLINE") &&
		(shadowAction == "ALLOW_RECOMMENDATION" || shadowAction == "ALLOW"):
		return "DECLINE_TO_ALLOW"
	case (prodAction == "DECLINE_RECOMMENDATION" || prodAction == "DECLINE") &&
		(shadowAction == "MANUAL_REVIEW" || shadowAction == "STEP_UP_RECOMMENDATION"):
		return "DECLINE_TO_REVIEW"
	default:
		return fmt.Sprintf("%s_TO_%s", prodAction, shadowAction)
	}
}

// GetStatus returns the operational status and metric counters of the shadow scorer.
func (s *ShadowScorer) GetStatus() map[string]interface{} {
	queueDepth := len(s.workQueue)
	status := map[string]interface{}{
		"enabled":                     s.config.Enabled,
		"closed":                      s.closed.Load(),
		"worker_count":                s.config.WorkerCount,
		"queue_capacity":              s.config.QueueCapacity,
		"sample_rate":                 s.config.SampleRate,
		"score_divergence_threshold":  s.config.ScoreDivergenceThreshold,
		"candidate_model_version":     s.config.CandidateModelVersion,
		"candidate_feature_contract":  s.config.CandidateFeatureContract,
		"metrics":                     s.metrics.Snapshot(queueDepth),
	}
	return status
}

// Close gracefully terminates the shadow worker pool, allowing in-flight work to complete.
func (s *ShadowScorer) Close(timeout time.Duration) error {
	if s.closed.Swap(true) {
		return nil
	}

	close(s.workQueue)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.cancel()
		log.Println("ShadowScorer shut down gracefully.")
		return nil
	case <-time.After(timeout):
		s.cancel()
		log.Println("ShadowScorer shutdown timed out. Forcing termination.")
		return fmt.Errorf("shadow scorer shutdown timed out after %v", timeout)
	}
}

// Stop shuts down workers gracefully with a default 100ms timeout.
func (s *ShadowScorer) Stop() {
	_ = s.Close(100 * time.Millisecond)
}
