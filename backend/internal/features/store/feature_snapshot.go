package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// FeatureVector maps feature names to numerical / categorical values.
type FeatureVector map[string]float64

// FeatureSnapshot represents a point-in-time snapshot of training features with an immutable SHA-256 checksum.
type FeatureSnapshot struct {
	SnapshotID    string                   `json:"snapshot_id"`
	CreatedAt     time.Time                `json:"created_at"`
	EntityKey     string                   `json:"entity_key"`
	FeatureNames  []string                 `json:"feature_names"`
	Records       []map[string]interface{} `json:"records"`
	SampleCount   int                      `json:"sample_count"`
	ChecksumSHA256 string                  `json:"checksum_sha256"`
}

// ComputeSnapshotChecksum computes the deterministic SHA-256 hash over the snapshot records.
func ComputeSnapshotChecksum(records []map[string]interface{}) (string, error) {
	data, err := json.Marshal(records)
	if err != nil {
		return "", fmt.Errorf("failed to marshal records: %w", err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// NewFeatureSnapshot builds an immutable, checksummed point-in-time feature snapshot.
func NewFeatureSnapshot(snapshotID, entityKey string, featureNames []string, records []map[string]interface{}) (*FeatureSnapshot, error) {
	checksum, err := ComputeSnapshotChecksum(records)
	if err != nil {
		return nil, err
	}

	return &FeatureSnapshot{
		SnapshotID:     snapshotID,
		CreatedAt:      time.Now().UTC(),
		EntityKey:      entityKey,
		FeatureNames:   featureNames,
		Records:        records,
		SampleCount:    len(records),
		ChecksumSHA256: checksum,
	}, nil
}
