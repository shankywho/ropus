package saas

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TenantUsageSnapshot records current billing cycle resource consumption.
type TenantUsageSnapshot struct {
	OrgID           string    `json:"org_id"`
	BillingPeriod   string    `json:"billing_period"` // e.g. "2026-08"
	RiskChecksTotal uint64    `json:"risk_checks_total"`
	CasesCreated    uint64    `json:"cases_created"`
	ModelInferences uint64    `json:"model_inferences"`
	AgentCalls      uint64    `json:"agent_calls"`
	StorageUsageGB  float64   `json:"storage_usage_gb"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UsageMeterEngine provides thread-safe atomic resource metering per tenant.
type UsageMeterEngine struct {
	mu     sync.RWMutex
	meters map[string]*TenantUsageSnapshot // "orgID:period" -> Snapshot
}

// NewUsageMeterEngine initializes the metering engine.
func NewUsageMeterEngine() *UsageMeterEngine {
	return &UsageMeterEngine{
		meters: make(map[string]*TenantUsageSnapshot),
	}
}

func getPeriodKey(orgID string) string {
	period := time.Now().UTC().Format("2006-01")
	return fmt.Sprintf("%s:%s", orgID, period)
}

// RecordRiskCheck increments risk evaluation counter.
func (u *UsageMeterEngine) RecordRiskCheck(orgID string, count uint64) {
	u.mu.Lock()
	defer u.mu.Unlock()

	key := getPeriodKey(orgID)
	s, exists := u.meters[key]
	if !exists {
		s = &TenantUsageSnapshot{
			OrgID:         orgID,
			BillingPeriod: time.Now().UTC().Format("2006-01"),
			UpdatedAt:     time.Now().UTC(),
		}
		u.meters[key] = s
	}

	atomic.AddUint64(&s.RiskChecksTotal, count)
	atomic.AddUint64(&s.ModelInferences, count)
	s.UpdatedAt = time.Now().UTC()
}

// RecordCaseCreation increments case management counter.
func (u *UsageMeterEngine) RecordCaseCreation(orgID string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	key := getPeriodKey(orgID)
	s, exists := u.meters[key]
	if !exists {
		s = &TenantUsageSnapshot{
			OrgID:         orgID,
			BillingPeriod: time.Now().UTC().Format("2006-01"),
			UpdatedAt:     time.Now().UTC(),
		}
		u.meters[key] = s
	}

	atomic.AddUint64(&s.CasesCreated, 1)
	s.UpdatedAt = time.Now().UTC()
}

// RecordAgentCall increments LLM investigator invocation counter.
func (u *UsageMeterEngine) RecordAgentCall(orgID string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	key := getPeriodKey(orgID)
	s, exists := u.meters[key]
	if !exists {
		s = &TenantUsageSnapshot{
			OrgID:         orgID,
			BillingPeriod: time.Now().UTC().Format("2006-01"),
			UpdatedAt:     time.Now().UTC(),
		}
		u.meters[key] = s
	}

	atomic.AddUint64(&s.AgentCalls, 1)
	s.UpdatedAt = time.Now().UTC()
}

// GetTenantUsage retrieves the current month usage for a tenant.
func (u *UsageMeterEngine) GetTenantUsage(orgID string) *TenantUsageSnapshot {
	u.mu.RLock()
	defer u.mu.RUnlock()

	key := getPeriodKey(orgID)
	s, exists := u.meters[key]
	if !exists {
		return &TenantUsageSnapshot{
			OrgID:          orgID,
			BillingPeriod:  time.Now().UTC().Format("2006-01"),
			StorageUsageGB: 24.5,
			UpdatedAt:      time.Now().UTC(),
		}
	}

	return &TenantUsageSnapshot{
		OrgID:           s.OrgID,
		BillingPeriod:   s.BillingPeriod,
		RiskChecksTotal: atomic.LoadUint64(&s.RiskChecksTotal),
		CasesCreated:    atomic.LoadUint64(&s.CasesCreated),
		ModelInferences: atomic.LoadUint64(&s.ModelInferences),
		AgentCalls:      atomic.LoadUint64(&s.AgentCalls),
		StorageUsageGB:  24.5,
		UpdatedAt:       s.UpdatedAt,
	}
}
