package graph

import (
	"fmt"
	"sync"
	"time"
)

// EntityCluster groups disparate accounts identified as belonging to the same physical person or syndicate.
type EntityCluster struct {
	ClusterID          string    `json:"cluster_id"`
	PrimaryEntityID    string    `json:"primary_entity_id"`
	MemberNodeIDs      []string  `json:"member_node_ids"`
	DeviceFingerprints []string  `json:"device_fingerprints"`
	PaymentCards       []string  `json:"payment_cards"`
	IPAddresses        []string  `json:"ip_addresses"`
	ClusterConfidence  float64   `json:"cluster_confidence"`
	IsSuspicious       bool      `json:"is_suspicious"`
	CreatedAt          time.Time `json:"created_at"`
}

// EntityResolutionEngine resolves real-world identities across disjoint identifiers.
type EntityResolutionEngine struct {
	mu       sync.RWMutex
	clusters map[string]*EntityCluster
	nodeMap  map[string]string // nodeID -> clusterID
}

// NewEntityResolutionEngine initializes the entity resolution engine.
func NewEntityResolutionEngine() *EntityResolutionEngine {
	return &EntityResolutionEngine{
		clusters: make(map[string]*EntityCluster),
		nodeMap:  make(map[string]string),
	}
}

// ResolveEntity clusters an incoming transaction profile with existing entity identity pools.
func (e *EntityResolutionEngine) ResolveEntity(userID, deviceFingerprint, cardHash, ipAddress string) *EntityCluster {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if any identifier maps to an existing cluster
	var existingClusterID string
	for _, id := range []string{userID, deviceFingerprint, cardHash, ipAddress} {
		if id == "" {
			continue
		}
		if cid, found := e.nodeMap[id]; found {
			existingClusterID = cid
			break
		}
	}

	now := time.Now().UTC()
	if existingClusterID != "" {
		c := e.clusters[existingClusterID]
		c.MemberNodeIDs = appendUnique(c.MemberNodeIDs, userID)
		if deviceFingerprint != "" {
			c.DeviceFingerprints = appendUnique(c.DeviceFingerprints, deviceFingerprint)
			e.nodeMap[deviceFingerprint] = existingClusterID
		}
		if cardHash != "" {
			c.PaymentCards = appendUnique(c.PaymentCards, cardHash)
			e.nodeMap[cardHash] = existingClusterID
		}
		if ipAddress != "" {
			c.IPAddresses = appendUnique(c.IPAddresses, ipAddress)
			e.nodeMap[ipAddress] = existingClusterID
		}
		e.nodeMap[userID] = existingClusterID

		if len(c.MemberNodeIDs) >= 3 || len(c.PaymentCards) >= 3 {
			c.IsSuspicious = true
		}
		return c
	}

	// Create new cluster
	newCID := fmt.Sprintf("cluster_%d", time.Now().UnixNano())
	newCluster := &EntityCluster{
		ClusterID:          newCID,
		PrimaryEntityID:    userID,
		MemberNodeIDs:      []string{userID},
		DeviceFingerprints: []string{deviceFingerprint},
		PaymentCards:       []string{cardHash},
		IPAddresses:        []string{ipAddress},
		ClusterConfidence:  0.95,
		IsSuspicious:       false,
		CreatedAt:          now,
	}

	e.clusters[newCID] = newCluster
	e.nodeMap[userID] = newCID
	if deviceFingerprint != "" {
		e.nodeMap[deviceFingerprint] = newCID
	}
	if cardHash != "" {
		e.nodeMap[cardHash] = newCID
	}
	if ipAddress != "" {
		e.nodeMap[ipAddress] = newCID
	}

	return newCluster
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
