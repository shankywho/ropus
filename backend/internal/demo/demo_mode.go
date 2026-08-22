package demo

import (
	"fmt"
	"sync"
	"time"

	"github.com/shankywho/ropus/backend/internal/auth/api_keys"
	"github.com/shankywho/ropus/backend/internal/product_api"
	"github.com/shankywho/ropus/backend/internal/saas"
)

// DemoState defines the current execution state of the demo session.
type DemoState string

const (
	DemoStateReady     DemoState = "READY"
	DemoStateRunning   DemoState = "RUNNING"
	DemoStatePaused    DemoState = "PAUSED"
	DemoStateCompleted DemoState = "COMPLETED"
)

// DemoStepDetail represents an individual stage in the deterministic 7-step narrative.
type DemoStepDetail struct {
	StepNumber        int       `json:"step_number"`
	Name              string    `json:"name"`
	Phase             string    `json:"phase"`
	ObservedFact      string    `json:"observed_fact"`
	InferredPattern   string    `json:"inferred_pattern"`
	RecommendedAction string    `json:"recommended_action"`
	ScoreContribution float64   `json:"score_contribution"`
	CumulativeScore   float64   `json:"cumulative_score"`
	Verdict           string    `json:"verdict"`
	Timestamp         time.Time `json:"timestamp"`
}

// DemoSession encapsulates an active interactive demonstration.
type DemoSession struct {
	SessionID      string                          `json:"session_id"`
	TenantID       string                          `json:"tenant_id"`
	State          DemoState                       `json:"state"`
	CurrentStep    int                             `json:"current_step"`
	TotalSteps     int                             `json:"total_steps"`
	Steps          []DemoStepDetail                `json:"steps"`
	DecisionResult *product_api.CanonicalRiskResponse `json:"decision_result,omitempty"`
	StartedAt      time.Time                       `json:"started_at"`
	UpdatedAt      time.Time                       `json:"updated_at"`
}

// DemoModeManager coordinates deterministic, repeatable investor/customer demonstrations.
type DemoModeManager struct {
	mu         sync.RWMutex
	keyService *api_keys.APIKeyService
	pipeline   *product_api.UnifiedRiskPipeline
	sessions   map[string]*DemoSession
}

// NewDemoModeManager initializes the deterministic demo manager.
func NewDemoModeManager() *DemoModeManager {
	keyService := api_keys.NewAPIKeyService()
	meter := saas.NewUsageMeterEngine()
	pipeline := product_api.NewUnifiedRiskPipeline(keyService, meter, nil)

	return &DemoModeManager{
		keyService: keyService,
		pipeline:   pipeline,
		sessions:   make(map[string]*DemoSession),
	}
}

// CreateSession initializes a clean demo session at Step 1 (Normal baseline).
func (m *DemoModeManager) CreateSession(tenantID string) (*DemoSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tenantID == "" {
		tenantID = "org_demo_synthetic"
	}

	sessionID := fmt.Sprintf("demo_sess_%d", time.Now().UnixNano())
	now := time.Now().UTC()

	steps := []DemoStepDetail{
		{
			StepNumber:        1,
			Name:              "Normal Customer Behavioral Baseline",
			Phase:             "BASELINE",
			ObservedFact:      "Customer usr_sarah_connor has 3-year spotless profile (NYC, $64.50 avg, MacBook Pro)",
			InferredPattern:   "Legitimate account with consistent domestic behavioral velocity",
			RecommendedAction: "Allow transaction to settle without friction",
			ScoreContribution: 0.04,
			CumulativeScore:   0.04,
			Verdict:           "APPROVE",
			Timestamp:         now,
		},
		{
			StepNumber:        2,
			Name:              "Compromise & Impossible Travel",
			Phase:             "ANOMALY",
			ObservedFact:      "Session originating from Limassol, Cyprus (7,850 km jump in 12 mins) requesting $14,500 wire",
			InferredPattern:   "Credential stuffing account takeover followed by rapid liquidity drain attempt",
			RecommendedAction: "Flag velocity anomaly and quarantine session",
			ScoreContribution: 0.43,
			CumulativeScore:   0.47,
			Verdict:           "REVIEW",
			Timestamp:         now.Add(10 * time.Millisecond),
		},
		{
			StepNumber:        3,
			Name:              "Graph Intelligence & Syndicate Discovery",
			Phase:             "CORRELATION",
			ObservedFact:      "Hardware canvas hash (dev_mule_cluster_99) is linked across 14 other customer nodes in graph",
			InferredPattern:   "Coordinated syndicate activity linking multiple synthetic money mule identities",
			RecommendedAction: "Trigger multi-entity link expansion and alert consortium rails",
			ScoreContribution: 0.17,
			CumulativeScore:   0.64,
			Verdict:           "CHALLENGE",
			Timestamp:         now.Add(20 * time.Millisecond),
		},
		{
			StepNumber:        4,
			Name:              "ML & Additive Mathematical Factor Fusion",
			Phase:             "EVALUATION",
			ObservedFact:      "XGBoost model infers 0.982 fraud probability from combined velocity + device + graph features",
			InferredPattern:   "High confidence malicious attack pattern matching known fraud vector",
			RecommendedAction: "Issue hard BLOCK verdict and halt payment settlement",
			ScoreContribution: 0.32,
			CumulativeScore:   0.96,
			Verdict:           "BLOCK",
			Timestamp:         now.Add(30 * time.Millisecond),
		},
		{
			StepNumber:        5,
			Name:              "Autonomous AI Investigator Dossier",
			Phase:             "INVESTIGATION",
			ObservedFact:      "Evidence collected: Bulletproof proxy IP 198.51.100.44 + Cyprus impossible travel + 14-node graph cluster",
			InferredPattern:   "Transnational automated mule cashout campaign",
			RecommendedAction: "Freeze linked accounts and dispatch SAR report package",
			ScoreContribution: 0.00,
			CumulativeScore:   0.96,
			Verdict:           "BLOCK",
			Timestamp:         now.Add(40 * time.Millisecond),
		},
		{
			StepNumber:        6,
			Name:              "Automated Review Case Creation",
			Phase:             "CASE_MANAGEMENT",
			ObservedFact:      "Case #CASE-88419 opened in P0 review queue with 5 attached evidence artifacts",
			InferredPattern:   "Confirmed high-severity financial crime incident requiring compliance retention",
			RecommendedAction: "Assign to Senior Financial Crime Analyst Queue",
			ScoreContribution: 0.00,
			CumulativeScore:   0.96,
			Verdict:           "BLOCK",
			Timestamp:         now.Add(50 * time.Millisecond),
		},
		{
			StepNumber:        7,
			Name:              "Human-in-the-Loop Governance",
			Phase:             "GOVERNANCE",
			ObservedFact:      "Analyst elena.r@acmebank.com confirms block and logs action to SHA-256 audit ledger",
			InferredPattern:   "Human confirmation validated; feedback dispatched to closed-loop model retraining",
			RecommendedAction: "Case closed as CONFIRMED_FRAUD; rules updated",
			ScoreContribution: 0.00,
			CumulativeScore:   0.96,
			Verdict:           "BLOCK",
			Timestamp:         now.Add(60 * time.Millisecond),
		},
	}

	session := &DemoSession{
		SessionID:   sessionID,
		TenantID:    tenantID,
		State:       DemoStateReady,
		CurrentStep: 1,
		TotalSteps:  len(steps),
		Steps:       steps,
		StartedAt:   now,
		UpdatedAt:   now,
	}

	m.sessions[sessionID] = session
	return session, nil
}

// ExecuteStep advances the demo session to a specific step or next step deterministically.
func (m *DemoModeManager) ExecuteStep(sessionID string, stepNumber int) (*DemoStepDetail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("demo session '%s' not found", sessionID)
	}

	if stepNumber < 1 || stepNumber > session.TotalSteps {
		return nil, fmt.Errorf("invalid step number %d (must be 1-%d)", stepNumber, session.TotalSteps)
	}

	session.CurrentStep = stepNumber
	session.State = DemoStateRunning
	if stepNumber == session.TotalSteps {
		session.State = DemoStateCompleted
	}
	session.UpdatedAt = time.Now().UTC()

	return &session.Steps[stepNumber-1], nil
}

// ResetSession resets the demo session back to Step 1 without data corruption.
func (m *DemoModeManager) ResetSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("demo session '%s' not found", sessionID)
	}

	session.CurrentStep = 1
	session.State = DemoStateReady
	session.UpdatedAt = time.Now().UTC()
	return nil
}

// GetSession retrieves the current state of a demo session.
func (m *DemoModeManager) GetSession(sessionID string) (*DemoSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("demo session '%s' not found", sessionID)
	}
	return session, nil
}
