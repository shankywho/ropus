package store

import (
	"fmt"
	"sync"
	"time"
)

// FeatureRegistry maintains the authoritative catalog and version history of features.
type FeatureRegistry struct {
	mu       sync.RWMutex
	features map[string]FeatureDefinition
	history  map[string][]FeatureDefinition
}

// NewFeatureRegistry initializes the feature registry with standard 25 baseline risk features.
func NewFeatureRegistry() *FeatureRegistry {
	r := &FeatureRegistry{
		features: make(map[string]FeatureDefinition),
		history:  make(map[string][]FeatureDefinition),
	}
	r.registerBaselineFeatures()
	return r
}

func (r *FeatureRegistry) registerBaselineFeatures() {
	now := time.Now().UTC()
	baseline := []FeatureDefinition{
		{Name: "amount", Version: 1, DataType: TypeFloat64, Description: "Transaction amount in base currency", SourceEntity: "transaction", CreatedAt: now},
		{Name: "user_txn_count_1h", Version: 1, DataType: TypeInt64, Description: "Transaction count for user in past 1 hour", SourceEntity: "user", CreatedAt: now},
		{Name: "user_txn_count_24h", Version: 1, DataType: TypeInt64, Description: "Transaction count for user in past 24 hours", SourceEntity: "user", CreatedAt: now},
		{Name: "user_txn_sum_1h", Version: 1, DataType: TypeFloat64, Description: "Total transaction spend for user in past 1 hour", SourceEntity: "user", CreatedAt: now},
		{Name: "user_txn_sum_24h", Version: 1, DataType: TypeFloat64, Description: "Total transaction spend for user in past 24 hours", SourceEntity: "user", CreatedAt: now},
		{Name: "card_txn_count_1h", Version: 1, DataType: TypeInt64, Description: "Card velocity count in past 1 hour", SourceEntity: "card", CreatedAt: now},
		{Name: "card_txn_count_24h", Version: 1, DataType: TypeInt64, Description: "Card velocity count in past 24 hours", SourceEntity: "card", CreatedAt: now},
		{Name: "ip_txn_count_1h", Version: 1, DataType: TypeInt64, Description: "IP address velocity in past 1 hour", SourceEntity: "ip", CreatedAt: now},
		{Name: "ip_txn_count_24h", Version: 1, DataType: TypeInt64, Description: "IP address velocity in past 24 hours", SourceEntity: "ip", CreatedAt: now},
		{Name: "device_txn_count_1h", Version: 1, DataType: TypeInt64, Description: "Device velocity in past 1 hour", SourceEntity: "device", CreatedAt: now},
		{Name: "device_txn_count_24h", Version: 1, DataType: TypeInt64, Description: "Device velocity in past 24 hours", SourceEntity: "device", CreatedAt: now},
		{Name: "device_age_days", Version: 1, DataType: TypeFloat64, Description: "Days since device first seen", SourceEntity: "device", CreatedAt: now},
		{Name: "is_vpn", Version: 1, DataType: TypeBool, Description: "Whether transaction originated from VPN", SourceEntity: "ip", CreatedAt: now},
		{Name: "is_tor", Version: 1, DataType: TypeBool, Description: "Whether transaction originated from Tor node", SourceEntity: "ip", CreatedAt: now},
		{Name: "risk_score_legacy", Version: 1, DataType: TypeFloat64, Description: "15F baseline model risk score", SourceEntity: "model", CreatedAt: now},
	}

	for _, f := range baseline {
		_ = r.RegisterFeature(f)
	}
}

// RegisterFeature adds or updates a feature definition in the catalog with version tracking.
func (r *FeatureRegistry) RegisterFeature(def FeatureDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("feature name cannot be empty")
	}
	if def.CreatedAt.IsZero() {
		def.CreatedAt = time.Now().UTC()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.features[def.Name] = def
	r.history[def.Name] = append(r.history[def.Name], def)
	return nil
}

// GetFeature retrieves the active feature definition.
func (r *FeatureRegistry) GetFeature(name string) (FeatureDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, exists := r.features[name]
	if !exists {
		return FeatureDefinition{}, fmt.Errorf("feature '%s' not found in registry", name)
	}
	return def, nil
}

// ListFeatures returns all registered features.
func (r *FeatureRegistry) ListFeatures() []FeatureDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]FeatureDefinition, 0, len(r.features))
	for _, f := range r.features {
		res = append(res, f)
	}
	return res
}

// GetHistory returns the version change log for a specific feature.
func (r *FeatureRegistry) GetHistory(name string) ([]FeatureDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hist, exists := r.history[name]
	if !exists {
		return nil, fmt.Errorf("feature '%s' not found", name)
	}
	return hist, nil
}
