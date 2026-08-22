package graph

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// FraudRing represents a detected organized syndicate or suspicious sub-network.
type FraudRing struct {
	RingID         string    `json:"ring_id"`
	RingType       string    `json:"ring_type"` // "ACCOUNT_FARM", "MULE_NETWORK", "SYNTHETIC_IDENTITY", "COORDINATED_ATTACK"
	MemberNodeIDs  []string  `json:"member_node_ids"`
	RingSize       int       `json:"ring_size"`
	FraudRingScore float64   `json:"fraud_ring_score"` // 0.0 to 1.0
	Description    string    `json:"description"`
	DetectedAt     time.Time `json:"detected_at"`
}

// FraudRingDetector analyzes knowledge graph topology to detect synthetic identity farms and mule rings.
type FraudRingDetector struct {
	store    GraphStore
	mu       sync.RWMutex
	rings    map[string]*FraudRing
}

// NewFraudRingDetector initializes the fraud ring detector.
func NewFraudRingDetector(store GraphStore) *FraudRingDetector {
	return &FraudRingDetector{
		store: store,
		rings: make(map[string]*FraudRing),
	}
}

// DetectRings analyzes the graph for dense clusters sharing single devices or IP subnets.
func (d *FraudRingDetector) DetectRings(minClusterSize int) []*FraudRing {
	d.mu.Lock()
	defer d.mu.Unlock()

	if minClusterSize <= 0 {
		minClusterSize = 3
	}

	var detected []*FraudRing

	// Scan local store nodes for device / IP hubs
	localStore, ok := d.store.(*LocalGraphStore)
	if !ok {
		return detected
	}

	localStore.mu.RLock()
	defer localStore.mu.RUnlock()

	for nodeID, node := range localStore.nodes {
		if node.Type == NodeDevice || node.Type == NodeIPAddress {
			neighbors, _ := localStore.QueryNeighbors(nodeID, "")
			if len(neighbors) >= minClusterSize {
				ringType := "ACCOUNT_FARM"
				if node.Type == NodeIPAddress {
					ringType = "COORDINATED_ATTACK"
				}

				score := math.Min(1.0, float64(len(neighbors))*0.15)
				ringID := fmt.Sprintf("ring_%s_%d", nodeID, time.Now().UnixNano())

				var memberIDs []string
				memberIDs = append(memberIDs, nodeID)
				for _, nb := range neighbors {
					memberIDs = append(memberIDs, nb.ID)
				}

				ring := &FraudRing{
					RingID:         ringID,
					RingType:       ringType,
					MemberNodeIDs:  memberIDs,
					RingSize:       len(memberIDs),
					FraudRingScore: score,
					Description:    fmt.Sprintf("Hub node '%s' connected to %d distinct entities", nodeID, len(neighbors)),
					DetectedAt:     time.Now().UTC(),
				}

				d.rings[ringID] = ring
				detected = append(detected, ring)
			}
		}
	}

	return detected
}

// GetRing fetches a detected fraud ring by ID.
func (d *FraudRingDetector) GetRing(ringID string) (*FraudRing, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	r, exists := d.rings[ringID]
	if !exists {
		return nil, fmt.Errorf("fraud ring '%s' not found", ringID)
	}
	return r, nil
}
