package graphsage

import (
	"context"
	"sync"
	"time"
)

// StreamEventEnvelope represents an incoming stream message in Go.
type StreamEventEnvelope struct {
	EventID        string                 `json:"event_id"`
	Topic          string                 `json:"topic"`
	Payload        map[string]interface{} `json:"payload"`
	EventTimestamp time.Time              `json:"event_timestamp"`
}

// TelemetryStatus represents the runtime health and metrics snapshot in Go.
type TelemetryStatus struct {
	IsRunning                     bool             `json:"is_running"`
	RealDataConnectivity         string           `json:"real_data_connectivity"`
	BufferedEventsCount           int              `json:"buffered_events_count"`
	TotalEventsProcessed          int              `json:"total_events_processed"`
	TotalDossiersGenerated        int              `json:"total_dossiers_generated"`
	IngestionMetrics              IngestionMetrics `json:"ingestion_metrics"`
	OperationalMode               string           `json:"operational_mode"`
	CustomerEnforcementAuthority  string           `json:"customer_enforcement_authority"`
	GovernanceState               string           `json:"governance_state"`
}

// ShadowTelemetryManager manages the non-blocking background shadow telemetry pipeline in Go.
type ShadowTelemetryManager struct {
	mu             sync.RWMutex
	eventChan      chan StreamEventEnvelope
	store          *TemporalHeteroGraphStore
	adapter        *PassiveShadowIngestionAdapter
	matEngine      *LabelMaturationEngine
	evidenceLedger *ShadowEvidenceLedger
	driftMonitor   *GraphDriftMonitor
	baselineEngine *EmployeeBaselineEngine
	pathExtractor  *PathExtractor
	recordedScores []float64

	isRunning      bool
	cancelFunc     context.CancelFunc
	totalProcessed int
	totalDossiers  int
	isLiveActive   bool
}

// NewShadowTelemetryManager initializes the Go shadow telemetry manager.
func NewShadowTelemetryManager(queueSize int, isLiveActive bool) *ShadowTelemetryManager {
	if queueSize <= 0 {
		queueSize = 10000
	}
	store := NewTemporalHeteroGraphStore()
	return &ShadowTelemetryManager{
		eventChan:      make(chan StreamEventEnvelope, queueSize),
		store:          store,
		adapter:        NewPassiveShadowIngestionAdapter(store, isLiveActive),
		matEngine:      NewLabelMaturationEngine(90),
		evidenceLedger: NewShadowEvidenceLedger(),
		driftMonitor:   &GraphDriftMonitor{},
		baselineEngine: &EmployeeBaselineEngine{},
		pathExtractor:  &PathExtractor{},
		recordedScores: make([]float64, 0),
		isLiveActive:   isLiveActive,
	}
}

// Start begins the background ingestion loop.
func (m *ShadowTelemetryManager) Start(ctx context.Context) {
	m.mu.Lock()
	if m.isRunning {
		m.mu.Unlock()
		return
	}
	m.isRunning = true
	ctx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel
	m.mu.Unlock()

	go m.processLoop(ctx)
}

// Stop terminates the background ingestion worker.
func (m *ShadowTelemetryManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isRunning && m.cancelFunc != nil {
		m.cancelFunc()
		m.isRunning = false
	}
}

// EnqueueEvent non-blockingly pushes an event into the shadow buffer.
func (m *ShadowTelemetryManager) EnqueueEvent(env StreamEventEnvelope) bool {
	select {
	case m.eventChan <- env:
		return true
	default:
		// Queue full (backpressure): drop non-blocking to protect latency
		return false
	}
}

func (m *ShadowTelemetryManager) processLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case env := <-m.eventChan:
			ok, _ := m.adapter.ConsumeEvent(ctx, env.EventID, env.Topic, env.Payload, env.EventTimestamp)
			if ok {
				m.mu.Lock()
				m.totalProcessed++
				m.mu.Unlock()
			}
		}
	}
}

// EvaluateShadowInvestigation evaluates an employee point-in-time in Go.
func (m *ShadowTelemetryManager) EvaluateShadowInvestigation(
	employeeID string,
	evalTS time.Time,
	role string,
) *CollusionInvestigationDossier {
	m.mu.Lock()
	defer m.mu.Unlock()

	pitEdges := m.store.GetAllPointInTimeEdges(evalTS)
	nodeMap := make(map[string]*HeteroNode)

	// 1. Point-in-time paths
	paths := m.pathExtractor.FindPaths(employeeID, "", nodeMap, pitEdges, 3)

	// 2. Point-in-time baseline
	baseline := m.baselineEngine.ComputeBaselineAndDeviation(employeeID, nil, pitEdges, evalTS)

	riskScore := 0.15
	riskLevel := RiskLevelLow

	if baseline.IsUnusualHour || baseline.IsNewDevice || baseline.VolumeZScore > 2.0 {
		riskScore = 0.88
		riskLevel = RiskLevelHigh
	}

	dossier := &CollusionInvestigationDossier{
		DossierID:        "dos_go_" + evalTS.Format("20060102150405") + "_" + employeeID,
		OverallRiskScore: riskScore,
		RiskLevel:        riskLevel,
		EmployeeID:       employeeID,
		EmployeeRole:     role,
		RelationshipPaths: paths,
		GovernanceNotice: "INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING",
	}

	m.evidenceLedger.RecordDossier(dossier)
	m.recordedScores = append(m.recordedScores, riskScore)
	m.totalDossiers++

	return dossier
}

// GetStatus returns the current runtime metrics and governance status in Go.
func (m *ShadowTelemetryManager) GetStatus() TelemetryStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn := "NOT_CONNECTED"
	if m.isLiveActive {
		conn = "CONNECTED_SHADOW"
	}

	return TelemetryStatus{
		IsRunning:                    m.isRunning,
		RealDataConnectivity:        conn,
		BufferedEventsCount:          len(m.eventChan),
		TotalEventsProcessed:         m.totalProcessed,
		TotalDossiersGenerated:       m.totalDossiers,
		IngestionMetrics:             m.adapter.GetMetrics(),
		OperationalMode:              "STRICTLY_NON_ENFORCING_SHADOW",
		CustomerEnforcementAuthority: "0% (Baseline Dynamic BMR 100% Authoritative)",
		GovernanceState:              "REAL_DATA_REQUIRED (0 / 50 Real Collusion Cases)",
	}
}
