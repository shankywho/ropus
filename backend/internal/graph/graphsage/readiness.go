package graphsage

import (
	"context"
	"fmt"
	"time"
)

// ReadinessChecker audits graph data sources for real-data readiness.
type ReadinessChecker struct {
	dataSource RealGraphDataSource
}

// NewReadinessChecker creates a readiness checker.
func NewReadinessChecker(ds RealGraphDataSource) *ReadinessChecker {
	return &ReadinessChecker{dataSource: ds}
}

// Audit evaluates dataset completeness, causality, and label maturity.
func (c *ReadinessChecker) Audit(ctx context.Context, asOf time.Time) *ReadinessAuditReport {
	auditedAt := time.Now().UTC()
	dsType := c.dataSource.GetDatasetType()
	var blockingReasons []string

	nodes, _ := c.dataSource.GetNodes(ctx, nil, asOf)
	edges, _ := c.dataSource.GetEdges(ctx, nil, asOf)
	labels, _ := c.dataSource.GetLabels(ctx, asOf, false)

	// Check referential integrity
	nodeMap := make(map[string]bool)
	for _, n := range nodes {
		nodeMap[n.ID] = true
	}

	invalidEdgeRefs := 0
	for _, e := range edges {
		if !nodeMap[e.SourceID] || !nodeMap[e.TargetID] {
			invalidEdgeRefs++
		}
	}
	if invalidEdgeRefs > 0 {
		blockingReasons = append(blockingReasons, fmt.Sprintf("%d edges reference missing nodes", invalidEdgeRefs))
	}

	// Count confirmed labels
	confirmedFraud := 0
	confirmedCollusion := 0
	for _, l := range labels {
		mat, _ := l["maturity_state"].(LabelMaturityState)
		if mat == LabelConfirmedFraud {
			confirmedFraud++
			if l["is_collusion"] == true || l["label_type"] == "INTERNAL_COLLUSION" {
				confirmedCollusion++
			}
		}
	}

	if dsType == DatasetSynthetic {
		blockingReasons = append(blockingReasons, "Dataset is SYNTHETIC. Real data required for production promotion.")
	} else if confirmedFraud < 50 {
		blockingReasons = append(blockingReasons, fmt.Sprintf("Insufficient real confirmed fraud cases (%d/50)", confirmedFraud))
	} else if confirmedCollusion < 50 {
		blockingReasons = append(blockingReasons, fmt.Sprintf("Insufficient real confirmed collusion cases (%d/50)", confirmedCollusion))
	}

	isReady := len(blockingReasons) == 0
	govState := StateRealDataRequired
	if isReady {
		govState = StateShadowValidated
	}

	return &ReadinessAuditReport{
		AuditedAt:                             auditedAt,
		DatasetType:                           dsType,
		IsReadyForRealValidation:              isReady,
		GovernanceState:                       govState,
		BlockingReasons:                       blockingReasons,
		TotalNodes:                            len(nodes),
		TotalEdges:                            len(edges),
		ConfirmedFraudLabelsCount:             confirmedFraud,
		ConfirmedInternalCollusionLabelsCount: confirmedCollusion,
	}
}
