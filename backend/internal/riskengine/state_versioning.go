package riskengine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// CurrentSchemaVersion denotes the active state serialization schema version.
const CurrentSchemaVersion = 1

// PersistentStateEnvelope wraps the persistent state with schema versioning, checksums, and generation numbers.
type PersistentStateEnvelope struct {
	SchemaVersion int                      `json:"schema_version"`
	Generation    uint64                   `json:"generation"`
	ChecksumSHA256 string                  `json:"checksum_sha256"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	Payload       PersistedRetrainingState `json:"payload"`
}

// ComputePayloadChecksum calculates the canonical SHA-256 checksum over the state payload.
func ComputePayloadChecksum(payload PersistedRetrainingState) (string, error) {
	// Exclude checksum field itself from payload computation
	payload.ChecksumSHA256 = ""
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload for checksum: %w", err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// WrapState builds a versioned PersistentStateEnvelope from a raw state payload.
func WrapState(payload PersistedRetrainingState, generation uint64) (*PersistentStateEnvelope, error) {
	checksum, err := ComputePayloadChecksum(payload)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &PersistentStateEnvelope{
		SchemaVersion:  CurrentSchemaVersion,
		Generation:     generation,
		ChecksumSHA256: checksum,
		CreatedAt:      payload.SavedAt,
		UpdatedAt:      now,
		Payload:        payload,
	}, nil
}

// ValidateEnvelope verifies schema version compatibility and SHA-256 data integrity.
func (env *PersistentStateEnvelope) ValidateEnvelope() error {
	if env == nil {
		return fmt.Errorf("state envelope is nil")
	}

	if env.SchemaVersion > CurrentSchemaVersion {
		return fmt.Errorf("forward-incompatible schema version %d (supported: %d)", env.SchemaVersion, CurrentSchemaVersion)
	}

	if env.SchemaVersion < 1 {
		return fmt.Errorf("invalid schema version %d", env.SchemaVersion)
	}

	expectedChecksum, err := ComputePayloadChecksum(env.Payload)
	if err != nil {
		return fmt.Errorf("failed to compute payload checksum for verification: %w", err)
	}

	if env.ChecksumSHA256 != "" && env.ChecksumSHA256 != expectedChecksum {
		return fmt.Errorf("state payload checksum mismatch: expected %s, found %s (corrupted state detected)", expectedChecksum, env.ChecksumSHA256)
	}

	return nil
}

// ParseAndMigrateState parses raw persisted bytes, handling legacy un-enveloped formats
// and envelope schema versions with forward/backward compatibility.
func ParseAndMigrateState(data []byte) (*PersistentStateEnvelope, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Try parsing as versioned envelope first
	var envelope PersistentStateEnvelope
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.SchemaVersion > 0 {
		if err := envelope.ValidateEnvelope(); err != nil {
			return nil, fmt.Errorf("invalid state envelope: %w", err)
		}
		return &envelope, nil
	}

	// Fallback: Try parsing as legacy un-enveloped PersistedRetrainingState (Schema Version 0)
	var legacyState PersistedRetrainingState
	if err := json.Unmarshal(data, &legacyState); err != nil {
		return nil, fmt.Errorf("failed to parse state data as envelope or legacy format: %w", err)
	}

	// Migrate Schema 0 -> Schema 1
	wrapped, err := WrapState(legacyState, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate legacy state to envelope: %w", err)
	}

	return wrapped, nil
}
