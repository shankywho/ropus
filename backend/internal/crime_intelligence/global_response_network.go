package crime_intelligence

import (
	"fmt"
	"sync"
	"time"
)

// ConsortiumDefenseBroadcast encapsulates an indicator or rule deployed across the global defense network.
type ConsortiumDefenseBroadcast struct {
	BroadcastID     string    `json:"broadcast_id"`
	ActionType      string    `json:"action_type"` // "BLOCK_INFRASTRUCTURE", "SYNC_MULE_ACCOUNTS", "DEPLOY_RULE_DSL"
	PayloadSummary  string    `json:"payload_summary"`
	OriginatorNode  string    `json:"originator_node"`
	ParticipatingPeersCount int `json:"participating_peers_count"`
	DispatchedAt    time.Time `json:"dispatched_at"`
}

// GlobalThreatResponseNetwork coordinates real-time defense synchronization across financial institutions.
type GlobalThreatResponseNetwork struct {
	mu         sync.RWMutex
	broadcasts []*ConsortiumDefenseBroadcast
}

// NewGlobalThreatResponseNetwork initializes the global response coordinator.
func NewGlobalThreatResponseNetwork() *GlobalThreatResponseNetwork {
	return &GlobalThreatResponseNetwork{
		broadcasts: make([]*ConsortiumDefenseBroadcast, 0),
	}
}

// BroadcastDefenseAction disseminates a synchronized defense control across the network.
func (n *GlobalThreatResponseNetwork) BroadcastDefenseAction(actionType, summary, originator string) *ConsortiumDefenseBroadcast {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UTC()
	b := &ConsortiumDefenseBroadcast{
		BroadcastID:             fmt.Sprintf("gresp_%d", now.UnixNano()),
		ActionType:              actionType,
		PayloadSummary:          summary,
		OriginatorNode:          originator,
		ParticipatingPeersCount: 24, // 24 partner institutions
		DispatchedAt:            now,
	}

	n.broadcasts = append(n.broadcasts, b)
	return b
}

// ListBroadcasts returns recent consortium defense actions.
func (n *GlobalThreatResponseNetwork) ListBroadcasts() []*ConsortiumDefenseBroadcast {
	n.mu.RLock()
	defer n.mu.RUnlock()

	res := make([]*ConsortiumDefenseBroadcast, len(n.broadcasts))
	copy(res, n.broadcasts)
	return res
}
