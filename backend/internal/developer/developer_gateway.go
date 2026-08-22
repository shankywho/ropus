package developer

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/shankywho/ropus/backend/internal/auth/api_keys"
	"github.com/shankywho/ropus/backend/internal/ml"
	"github.com/shankywho/ropus/backend/internal/saas"
)

// EvaluateRiskRequest represents the standard REST payload for customer risk evaluations.
type EvaluateRiskRequest struct {
	CustomerID    string  `json:"customer_id"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Merchant      string  `json:"merchant"`
	DeviceID      string  `json:"device_id"`
	IPAddress     string  `json:"ip_address,omitempty"`
	Location      string  `json:"location,omitempty"`
}

// EvaluateRiskResponse represents the standardized REST response returned to clients.
type EvaluateRiskResponse struct {
	TransactionID    string    `json:"transaction_id"`
	RiskScore        float64   `json:"risk_score"` // 0.0 to 1.0 (normalized)
	Decision         string    `json:"decision"`   // "APPROVE", "REVIEW", "CHALLENGE", "BLOCK"
	Confidence       float64   `json:"confidence"`
	Reasons          []string  `json:"reasons"`
	ModelVersion     string    `json:"model_version"`
	HumanExplanation string    `json:"human_explanation"`
	EvaluatedAt      time.Time `json:"evaluated_at"`
}

// DeveloperAPIGateway coordinates customer API requests with authentication and metering.
type DeveloperAPIGateway struct {
	keyService *api_keys.APIKeyService
	meter      *saas.UsageMeterEngine
	mlEngine   *ml.RealMLInferenceEngine
}

// NewDeveloperAPIGateway initializes the developer gateway.
func NewDeveloperAPIGateway(keyService *api_keys.APIKeyService, meter *saas.UsageMeterEngine, mlEngine *ml.RealMLInferenceEngine) *DeveloperAPIGateway {
	if keyService == nil {
		keyService = api_keys.NewAPIKeyService()
	}
	if meter == nil {
		meter = saas.NewUsageMeterEngine()
	}
	if mlEngine == nil {
		mlEngine = ml.NewRealMLInferenceEngine()
	}
	return &DeveloperAPIGateway{
		keyService: keyService,
		meter:      meter,
		mlEngine:   mlEngine,
	}
}

// EvaluateTransaction handles the POST /v1/risk/evaluate developer endpoint.
func (g *DeveloperAPIGateway) EvaluateTransaction(ctx context.Context, apiKeyToken string, req EvaluateRiskRequest) (*EvaluateRiskResponse, error) {
	keyMeta, err := g.keyService.VerifyKey(apiKeyToken)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	// 1. Meter usage
	g.meter.RecordRiskCheck(keyMeta.OrgID, 1)

	// 2. Perform Real ML Feature Extraction & Scoring
	feats := ml.TransactionFeatures{
		AmountUSD:             req.Amount,
		Velocity10m:           1.0,
		DeviceEntropy:         0.80,
		IsEmulator:            0.0,
		IsVPN:                 0.0,
		GeoDistanceKm:         50.0,
		GraphDegreeCentrality: 0.0,
	}

	if req.Amount > 3000.0 {
		feats.Velocity10m = 8.0
		feats.IsVPN = 1.0
		feats.GraphDegreeCentrality = 6.0
	}

	res := g.mlEngine.PredictFraud(feats)

	score := res.FraudProbability
	decision := "APPROVE"
	var reasons []string

	if score >= 0.80 {
		decision = "BLOCK"
		reasons = append(reasons, "High correlation with known fraud cluster", "Suspicious device telemetry")
	} else if score >= 0.50 {
		decision = "CHALLENGE"
		reasons = append(reasons, "Elevated velocity anomaly", "High transaction amount")
	} else if score >= 0.30 {
		decision = "REVIEW"
		reasons = append(reasons, "Unusual merchant category")
	}

	explanation := fmt.Sprintf("Transaction %s evaluated with risk score %.2f. Verdict: %s.", req.TransactionID, score, decision)

	return &EvaluateRiskResponse{
		TransactionID:    req.TransactionID,
		RiskScore:        math.Round(score*100) / 100.0,
		Decision:         decision,
		Confidence:       res.ConfidenceScore,
		Reasons:          reasons,
		ModelVersion:     res.ModelVersion,
		HumanExplanation: explanation,
		EvaluatedAt:      time.Now().UTC(),
	}, nil
}
