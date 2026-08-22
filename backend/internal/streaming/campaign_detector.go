package streaming

import (
	"fmt"
	"sync"
	"time"
)

// CampaignRecord describes an ongoing coordinated cyber-fraud campaign.
type CampaignRecord struct {
	CampaignID       string    `json:"campaign_id"`
	AttackType       string    `json:"attack_type"` // "CREDENTIAL_STUFFING", "BOT_ATTACK_WAVE", "SYNTHETIC_IDENTITY_WAVE", "MULE_NETWORK"
	TargetTenants    []string  `json:"target_tenants"`
	InvolvedEntities []string  `json:"involved_entities"`
	Confidence       float64   `json:"confidence"`
	Severity         string    `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM"
	FirstSeenAt      time.Time `json:"first_seen_at"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	TotalAttacks     int       `json:"total_attacks"`
}

// CampaignDetector aggregates real-time alerts across tenants to detect organized fraud campaigns.
type CampaignDetector struct {
	mu        sync.RWMutex
	campaigns map[string]*CampaignRecord
}

// NewCampaignDetector initializes the campaign detector.
func NewCampaignDetector() *CampaignDetector {
	return &CampaignDetector{
		campaigns: make(map[string]*CampaignRecord),
	}
}

// IngestAlert checks if a stream alert belongs to an ongoing distributed campaign.
func (d *CampaignDetector) IngestAlert(tenantID, patternType, entityID string) *CampaignRecord {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	campaignKey := fmt.Sprintf("camp_%s", patternType)

	c, exists := d.campaigns[campaignKey]
	if !exists {
		c = &CampaignRecord{
			CampaignID:       fmt.Sprintf("cmp_%d_%s", now.UnixNano(), patternType),
			AttackType:       patternType,
			TargetTenants:    []string{tenantID},
			InvolvedEntities: []string{entityID},
			Confidence:       0.92,
			Severity:         "HIGH",
			FirstSeenAt:      now,
			LastSeenAt:       now,
			TotalAttacks:     1,
		}
		d.campaigns[campaignKey] = c
		return c
	}

	c.LastSeenAt = now
	c.TotalAttacks++
	c.InvolvedEntities = appendUnique(c.InvolvedEntities, entityID)
	c.TargetTenants = appendUnique(c.TargetTenants, tenantID)

	if len(c.TargetTenants) >= 3 || len(c.InvolvedEntities) >= 10 {
		c.Severity = "CRITICAL"
		c.Confidence = 0.99
	}

	return c
}

// ListCampaigns returns all detected fraud campaigns.
func (d *CampaignDetector) ListCampaigns() []*CampaignRecord {
	d.mu.RLock()
	defer d.mu.RUnlock()

	res := make([]*CampaignRecord, 0, len(d.campaigns))
	for _, c := range d.campaigns {
		res = append(res, c)
	}
	return res
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
