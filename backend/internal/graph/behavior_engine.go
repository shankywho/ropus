package graph

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// UserBehaviorProfile models the historical baseline habits of a legitimate user.
type UserBehaviorProfile struct {
	UserID            string    `json:"user_id"`
	AverageAmount     float64   `json:"average_amount"`
	MaxAmountSeen     float64   `json:"max_amount_seen"`
	TransactionCount  int64     `json:"transaction_count"`
	KnownDevices      []string  `json:"known_devices"`
	KnownIPs          []string  `json:"known_ips"`
	LastLocation      string    `json:"last_location"`
	LastTransactionAt time.Time `json:"last_transaction_at"`
}

// BehaviorEngine detects behavioral anomalies including spending spikes and velocity surges.
type BehaviorEngine struct {
	mu       sync.RWMutex
	profiles map[string]*UserBehaviorProfile
}

// NewBehaviorEngine initializes the behavioral analytics engine.
func NewBehaviorEngine() *BehaviorEngine {
	return &BehaviorEngine{
		profiles: make(map[string]*UserBehaviorProfile),
	}
}

// EvaluateBehavior computes behavioral anomaly scores and lists specific divergence factors.
func (b *BehaviorEngine) EvaluateBehavior(userID string, amount float64, device, ip, location string) (float64, []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, exists := b.profiles[userID]
	now := time.Now().UTC()

	if !exists {
		// Bootstrap baseline
		b.profiles[userID] = &UserBehaviorProfile{
			UserID:            userID,
			AverageAmount:     amount,
			MaxAmountSeen:     amount,
			TransactionCount:  1,
			KnownDevices:      []string{device},
			KnownIPs:          []string{ip},
			LastLocation:      location,
			LastTransactionAt: now,
		}
		return 0.05, nil // Normal initial baseline
	}

	var anomalies []string
	anomalyScore := 0.0

	// 1. Spending Spike Check
	if p.TransactionCount >= 3 && amount > p.AverageAmount*5.0 && amount > 500.0 {
		spikeRatio := amount / p.AverageAmount
		anomalies = append(anomalies, fmt.Sprintf("Transaction amount is %.1fx higher than user historical average ($%.2f)", spikeRatio, p.AverageAmount))
		anomalyScore += 0.45
	}

	// 2. Unrecognized Device Check
	deviceRecognized := false
	for _, d := range p.KnownDevices {
		if d == device {
			deviceRecognized = true
			break
		}
	}
	if !deviceRecognized && device != "" {
		anomalies = append(anomalies, "Transaction initiated from previously unseen device signature")
		anomalyScore += 0.25
		p.KnownDevices = append(p.KnownDevices, device)
	}

	// 3. Update moving average baseline
	p.AverageAmount = (p.AverageAmount*float64(p.TransactionCount) + amount) / float64(p.TransactionCount+1)
	if amount > p.MaxAmountSeen {
		p.MaxAmountSeen = amount
	}
	p.TransactionCount++
	p.LastLocation = location
	p.LastTransactionAt = now

	return math.Min(1.0, anomalyScore), anomalies
}
