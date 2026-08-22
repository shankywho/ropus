package riskengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MLPredictRequest matches the schema expected by the FastAPI ML service.
type MLPredictRequest struct {
	Amount                 float64  `json:"amount"`
	IPVelocity1h           float64  `json:"ip_velocity_1h"`
	IPVelocity24h          *float64 `json:"ip_velocity_24h,omitempty"`
	TokenVelocity24h       float64  `json:"token_velocity_24h"`
	IsNewDevice            int      `json:"is_new_device"`
	DeviceSeenBefore       *int     `json:"device_seen_before,omitempty"`
	HourOfDay              *int     `json:"hour_of_day,omitempty"`
	DayOfWeek              *int     `json:"day_of_week,omitempty"`
	ProductCD              *string  `json:"product_cd,omitempty"`
	CardType               *string  `json:"card_type,omitempty"`
	CardCategory           *string  `json:"card_category,omitempty"`
	EmailDomain            *string  `json:"email_domain,omitempty"`
	Dist1Missing           *int     `json:"dist1_missing,omitempty"`
	DeviceTypeMobile       *int     `json:"device_type_mobile,omitempty"`
	DeviceInfoMissing      *int     `json:"device_info_missing,omitempty"`
	AmountToMeanRatio      *float64 `json:"amount_to_mean_ratio,omitempty"`
	FeatureContractVersion string   `json:"feature_contract_version,omitempty"`
}

// MLPredictResponse represents the response returned by the FastAPI ML service.
type MLPredictResponse struct {
	RiskScore           int                `json:"risk_score"`
	Probability         float64            `json:"probability"`
	ReasonCodes         []string           `json:"reason_codes"`
	FeatureAttributions map[string]float64 `json:"feature_attributions"`
	LatencyMs           float64            `json:"latency_ms"`
}

// MLClient manages HTTP communication with the Python ML inference sidecar.
type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMLClient initializes an MLClient with sensible transport timeouts.
func NewMLClient(baseURL string) *MLClient {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			// Transport-level cap
			Timeout: 100 * time.Millisecond,
		},
	}
}

// Predict sends feature vectors to the ML sidecar with a strict 50ms context timeout.
func (c *MLClient) Predict(ctx context.Context, req MLPredictRequest) (*MLPredictResponse, error) {
	// Enforce 50ms strict deadline budget for ML inference
	inferenceCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ML request: %w", err)
	}

	url := fmt.Sprintf("%s/predict", c.baseURL)
	httpReq, err := http.NewRequestWithContext(inferenceCtx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create ML HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ml inference request failed (timeout/network): %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml service returned non-200 status: %d", httpResp.StatusCode)
	}

	var resp MLPredictResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode ML response: %w", err)
	}

	return &resp, nil
}

// MLShadowPredictRequest represents a 25-feature vector shadow prediction request.
type MLShadowPredictRequest struct {
	Features               []float64          `json:"features,omitempty"`
	FeaturesDict           map[string]float64 `json:"features_dict,omitempty"`
	EvaluationID           string             `json:"evaluation_id,omitempty"`
	TenantID               string             `json:"tenant_id,omitempty"`
	TransactionID          string             `json:"transaction_id,omitempty"`
	FeatureContractVersion string             `json:"feature_contract_version,omitempty"`
}

// MLShadowPredictResponse represents candidate model scoring output.
type MLShadowPredictResponse struct {
	ModelVersion           string  `json:"model_version"`
	FeatureContractVersion string  `json:"feature_contract_version"`
	FeatureCount           int     `json:"feature_count"`
	RawProbability         float64 `json:"raw_probability"`
	CalibratedProbability  float64 `json:"calibrated_probability"`
	RiskScore              int     `json:"risk_score"`
	ShadowDecision         string  `json:"shadow_decision"`
	LatencyMs              float64 `json:"latency_ms"`
	Runtime                string  `json:"runtime"`
}

// PredictShadow sends the 25-feature vector to the candidate shadow model endpoint.
func (c *MLClient) PredictShadow(ctx context.Context, req MLShadowPredictRequest) (*MLShadowPredictResponse, error) {
	inferenceCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal shadow ML request: %w", err)
	}

	url := fmt.Sprintf("%s/predict/shadow", c.baseURL)
	httpReq, err := http.NewRequestWithContext(inferenceCtx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create shadow ML HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("shadow ml request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shadow ml service returned non-200 status: %d", httpResp.StatusCode)
	}

	var resp MLShadowPredictResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode shadow ML response: %w", err)
	}

	return &resp, nil
}

// Ping checks the liveness of the ML inference service.
func (c *MLClient) Ping(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("ml client is nil")
	}
	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ml health returned non-200: %d", resp.StatusCode)
	}
	return nil
}
