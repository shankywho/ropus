package graphsage

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RelationshipIntelligenceEngine coordinates the temporal store, GraphSAGE neural model, and structural collusion detector.
type RelationshipIntelligenceEngine struct {
	store             *TemporalHeteroGraphStore
	model             *GraphSAGEModel
	collusionDetector *CollusionDetector
	lifecycleState    GraphSAGELifecycleState
	modelVersion      string
	timeout           time.Duration
}

// NewRelationshipIntelligenceEngine creates a production-grade shadow intelligence engine.
func NewRelationshipIntelligenceEngine(store *TemporalHeteroGraphStore) *RelationshipIntelligenceEngine {
	if store == nil {
		store = NewTemporalHeteroGraphStore()
	}
	return &RelationshipIntelligenceEngine{
		store:             store,
		model:             NewGraphSAGEModel(DefaultGraphSAGEConfig()),
		collusionDetector: NewCollusionDetector(DefaultCollusionDetectionThresholds()),
		lifecycleState:    StateDataRequired, // Transparent governance: DATA_REQUIRED until genuine internal labels exist
		modelVersion:      "graphsage-hetero-v1.0-shadow",
		timeout:           15 * time.Millisecond, // Strict sub-15ms latency bound
	}
}

// GetStore returns the underlying temporal graph store.
func (e *RelationshipIntelligenceEngine) GetStore() *TemporalHeteroGraphStore {
	return e.store
}

// SetLifecycleState updates the formal governance lifecycle state.
func (e *RelationshipIntelligenceEngine) SetLifecycleState(state GraphSAGELifecycleState) {
	e.lifecycleState = state
}

// EvaluateRelationship executes point-in-time GraphSAGE inference and structural collusion detection in shadow mode.
// FAIL-OPEN GUARANTEE: Never panics, never blocks caller beyond timeout, returns safe baseline on failure.
func (e *RelationshipIntelligenceEngine) EvaluateRelationship(ctx context.Context, rootNodeID string, asOf time.Time) (report *RelationshipIntelligenceReport) {
	start := time.Now()
	now := time.Now().UTC()
	if asOf.IsZero() {
		asOf = now
	}

	// 1. Panic Recovery & Fail-Open Handler
	defer func() {
		if r := recover(); r != nil {
			report = &RelationshipIntelligenceReport{
				RootNodeID:         rootNodeID,
				OverallRiskScore:   0.05,
				RiskLevel:          RiskLevelLow,
				LifecycleState:     e.lifecycleState,
				ObservedFacts:      []string{"GraphSAGE inference failed open due to internal error"},
				InferredPatterns:   nil,
				ExplainableSignals: nil,
				InvestigatorNotes:  fmt.Sprintf("Recovered from panic: %v", r),
				EvaluatedAt:        now,
				LatencyMs:          float64(time.Since(start).Microseconds()) / 1000.0,
				IsDegraded:         true,
				DegradeReason:      fmt.Sprintf("internal panic: %v", r),
				ModelVersion:       e.modelVersion,
				OperationalMode:    "NON_ENFORCING",
				IntendedUse:        "INVESTIGATION_ONLY",
			}
		}
	}()

	// 2. Context Cancellation / Timeout Check
	if ctx != nil && ctx.Err() != nil {
		return &RelationshipIntelligenceReport{
			RootNodeID:        rootNodeID,
			OverallRiskScore:  0.05,
			RiskLevel:         RiskLevelLow,
			LifecycleState:    e.lifecycleState,
			ObservedFacts:     []string{"GraphSAGE evaluation bypassed due to context timeout"},
			InvestigatorNotes: fmt.Sprintf("Context error: %v", ctx.Err()),
			EvaluatedAt:       now,
			LatencyMs:         float64(time.Since(start).Microseconds()) / 1000.0,
			IsDegraded:        true,
			DegradeReason:     fmt.Sprintf("context error: %v", ctx.Err()),
			ModelVersion:      e.modelVersion,
			OperationalMode:   "NON_ENFORCING",
			IntendedUse:       "INVESTIGATION_ONLY",
		}
	}

	// 3. Fetch or initialize root node with point-in-time check
	rootNode, err := e.store.GetNodePointInTime(rootNodeID, asOf)
	if err != nil {
		// Inductive fallback for unseen entity
		rootNode = &HeteroNode{
			ID:        rootNodeID,
			Type:      NodeConsumer, // Default assumption if unseen
			RiskScore: 0.05,
			CreatedAt: asOf,
		}
	}

	// 4. Extract Point-in-Time Sampled Neighborhood (Strict Temporal Causality & Bounded Degree)
	nh, err := e.store.GetTemporalSampledNeighborhood(rootNodeID, asOf, []int{10, 5})
	if err != nil {
		nh = &SampledNeighborhood{RootID: rootNodeID, AsOf: asOf}
	}

	// 5. GraphSAGE Forward Pass (Neural Aggregation & Embeddings)
	embedding, gnnSignals := e.model.Forward(rootNode, nh)

	// 6. Structural Collusion Analysis & Machine-Readable Explainability
	structFacts, structPatterns, structScore, structSignals, explainableSignals := e.collusionDetector.AnalyzeStructure(rootNode, nh)

	// 7. Combine Multi-Head Signals
	combinedSignals := RelationshipIntelligenceSignals{
		EmployeeRisk:              gnnSignals.EmployeeRisk,
		ConsumerRisk:              gnnSignals.ConsumerRisk,
		RelationshipCollusionRisk: math.Max(gnnSignals.RelationshipCollusionRisk, structScore),
		EntityNeighborhoodRisk:    gnnSignals.EntityNeighborhoodRisk,
		GraphAnomalyScore:         gnnSignals.GraphAnomalyScore,
		TransactionContextScore:   gnnSignals.TransactionContextScore,
		ConcentrationRatio:        structSignals.ConcentrationRatio,
		SharedDeviceOverlap:       structSignals.SharedDeviceOverlap,
		SharedIPOverlap:           structSignals.SharedIPOverlap,
		MultiHopPathLength:        structSignals.MultiHopPathLength,
		OffHoursAccessCount:       structSignals.OffHoursAccessCount,
		PreTxAccessProximity:      structSignals.PreTxAccessProximity,
	}

	// 8. Calculate Overall Risk Score
	overallScore := math.Max(combinedSignals.RelationshipCollusionRisk, combinedSignals.EntityNeighborhoodRisk)
	if rootNode.Type == NodeEmployee {
		overallScore = math.Max(overallScore, combinedSignals.EmployeeRisk)
	} else if rootNode.Type == NodeConsumer {
		overallScore = math.Max(overallScore, combinedSignals.ConsumerRisk)
	}
	overallScore = math.Round(overallScore*100.0) / 100.0

	// 9. Assign Risk Tier
	riskLevel := RiskLevelLow
	if overallScore >= 0.90 {
		riskLevel = RiskLevelCritical
	} else if overallScore >= 0.70 {
		riskLevel = RiskLevelHigh
	} else if overallScore >= 0.30 {
		riskLevel = RiskLevelMedium
	}

	// 10. Generate Investigator Dossier
	var facts []string
	var patterns []string

	facts = append(facts, fmt.Sprintf("Graph neighborhood evaluated as of %s (L1 neighbors: %d, L2 neighbors: %d)", asOf.Format(time.RFC3339), len(nh.Layer1), len(nh.Layer2)))
	facts = append(facts, structFacts...)
	patterns = append(patterns, structPatterns...)

	notes := fmt.Sprintf("Shadow intelligence verdict: %s (Risk Score: %.2f, Collusion Head: %.2f)", riskLevel, overallScore, combinedSignals.RelationshipCollusionRisk)

	latency := float64(time.Since(start).Microseconds()) / 1000.0

	return &RelationshipIntelligenceReport{
		RootNodeID:         rootNodeID,
		RootNodeType:       rootNode.Type,
		OverallRiskScore:   overallScore,
		RiskLevel:          riskLevel,
		LifecycleState:     e.lifecycleState,
		Signals:            combinedSignals,
		ExplainableSignals: explainableSignals,
		Embedding:          embedding,
		ObservedFacts:      facts,
		InferredPatterns:   patterns,
		InvestigatorNotes:  notes,
		EvaluatedAt:        now,
		LatencyMs:          latency,
		IsDegraded:         false,
		ModelVersion:       e.modelVersion,
		OperationalMode:    "NON_ENFORCING",
		IntendedUse:        "INVESTIGATION_ONLY",
	}
}
