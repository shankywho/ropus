package agent_council

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// AttackTechnique represents a standardized fraud tactic (e.g. T1001: Device Emulator Spoofing).
type AttackTechnique struct {
	TechniqueID   string   `json:"technique_id"`
	Name          string   `json:"name"`
	Category      string   `json:"category"` // "CREDENTIAL_ACCESS", "DEFENSE_EVASION", "LATERAL_MOVEMENT"
	ObservedCount int      `json:"observed_count"`
	Mitigations   []string `json:"mitigations"`
}

// FraudGroup represents an organized adversary syndicate profile.
type FraudGroup struct {
	GroupID            string    `json:"group_id"`
	HashedName         string    `json:"hashed_name"`
	AssociatedCampaigns []string  `json:"associated_campaigns"`
	TechniquesUsed     []string  `json:"techniques_used"`
	FirstObserved      time.Time `json:"first_observed"`
	LastActive         time.Time `json:"last_active"`
}

// GlobalIntelligenceGraph2 maintains threat groups, technique matrices, and campaign genealogies.
type GlobalIntelligenceGraph2 struct {
	mu         sync.RWMutex
	groups     map[string]*FraudGroup
	techniques map[string]*AttackTechnique
}

// NewGlobalIntelligenceGraph2 initializes the enhanced threat knowledge graph.
func NewGlobalIntelligenceGraph2() *GlobalIntelligenceGraph2 {
	return &GlobalIntelligenceGraph2{
		groups:     make(map[string]*FraudGroup),
		techniques: make(map[string]*AttackTechnique),
	}
}

// HashKey produces a deterministic SHA-256 hash.
func HashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RecordThreatGroup registers an organized fraud syndicate in the intelligence graph.
func (g *GlobalIntelligenceGraph2) RecordThreatGroup(rawGroupName string, campaignID string, techniqueIDs []string) *FraudGroup {
	hashed := HashKey(rawGroupName)
	now := time.Now().UTC()

	g.mu.Lock()
	defer g.mu.Unlock()

	grp, exists := g.groups[hashed]
	if !exists {
		grp = &FraudGroup{
			GroupID:             "grp_" + hashed[:12],
			HashedName:          hashed,
			AssociatedCampaigns: []string{campaignID},
			TechniquesUsed:      techniqueIDs,
			FirstObserved:       now,
			LastActive:          now,
		}
		g.groups[hashed] = grp
		return grp
	}

	grp.LastActive = now
	grp.AssociatedCampaigns = append(grp.AssociatedCampaigns, campaignID)
	return grp
}

// QueryGroup retrieves a fraud syndicate profile.
func (g *GlobalIntelligenceGraph2) QueryGroup(rawGroupName string) (*FraudGroup, bool) {
	hashed := HashKey(rawGroupName)

	g.mu.RLock()
	defer g.mu.RUnlock()

	grp, exists := g.groups[hashed]
	return grp, exists
}
