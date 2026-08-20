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
	Amount           float64 `json:"amount"`
	IPVelocity1h     float64 `json:"ip_velocity_1h"`
	TokenVelocity24h float64 `json:"token_velocity_24h"`
	IsNewDevice      int     `json:"is_new_device"`
	HourOfDay        *int    `json:"hour_of_day,omitempty"`
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
