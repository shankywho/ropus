package cases

import (
	"fmt"
)

// ResponseGuard enforces safety invariants and policy guardrails before executing automated containment.
type ResponseGuard struct {
	MinConfidenceForAccountFreeze float64
	MinConfidenceForAutoBlock    float64
}

// NewResponseGuard initializes response safety guardrails.
func NewResponseGuard() *ResponseGuard {
	return &ResponseGuard{
		MinConfidenceForAccountFreeze: 0.85,
		MinConfidenceForAutoBlock:     0.75,
	}
}

// ValidateSafety checks if an autonomous response action meets strict safety thresholds.
func (g *ResponseGuard) ValidateSafety(actionType, targetEntity string, confidence, riskScore float64) error {
	if targetEntity == "" {
		return fmt.Errorf("target entity cannot be empty")
	}

	switch actionType {
	case "FREEZE_ACCOUNT":
		if confidence < g.MinConfidenceForAccountFreeze {
			return fmt.Errorf("action FREEZE_ACCOUNT rejected: confidence %.2f < safety threshold %.2f", confidence, g.MinConfidenceForAccountFreeze)
		}
		if riskScore < 0.80 {
			return fmt.Errorf("action FREEZE_ACCOUNT rejected: risk score %.2f < threshold 0.80", riskScore)
		}
	case "BLOCK_TRANSACTION":
		if confidence < g.MinConfidenceForAutoBlock {
			return fmt.Errorf("action BLOCK_TRANSACTION rejected: confidence %.2f < safety threshold %.2f", confidence, g.MinConfidenceForAutoBlock)
		}
	}

	return nil
}
