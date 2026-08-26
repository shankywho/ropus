package bot_defense

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// ReplayProtectorConfig configures the nonce window and timestamp tolerance.
type ReplayProtectorConfig struct {
	MaxTimestampDriftSec float64 // Max allowed drift between request timestamp and current server time (default: 300s)
	NonceRetentionWindow time.Duration // Time to retain seen nonces in cache (default: 15m)
	MaxCacheEntries      int     // Cap on memory usage (default: 50,000 entries)
}

// DefaultReplayProtectorConfig returns default security parameters.
func DefaultReplayProtectorConfig() ReplayProtectorConfig {
	return ReplayProtectorConfig{
		MaxTimestampDriftSec: 300.0,
		NonceRetentionWindow: 15 * time.Minute,
		MaxCacheEntries:      50000,
	}
}

// ReplayProtector maintains a sliding window of observed nonces and validates time drift.
type ReplayProtector struct {
	mu         sync.Mutex
	seenNonces map[string]time.Time
	cfg        ReplayProtectorConfig
}

// NewReplayProtector constructs a new replay and time drift protector.
func NewReplayProtector(cfg ReplayProtectorConfig) *ReplayProtector {
	return &ReplayProtector{
		seenNonces: make(map[string]time.Time),
		cfg:        cfg,
	}
}

// ValidateReplay checks for duplicate transaction/nonce identifiers and clock drift.
func (p *ReplayProtector) ValidateReplay(nonceKey string, reqTimestamp time.Time) (signals BotSignals, rules []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().UTC()

	// 1. Clock Drift Check
	if !reqTimestamp.IsZero() {
		driftSec := math.Abs(now.Sub(reqTimestamp).Seconds())
		signals.TimeDriftSec = driftSec
		if driftSec > p.cfg.MaxTimestampDriftSec {
			signals.ReplayDetected = true
			rules = append(rules, fmt.Sprintf("Request timestamp drift (%.1fs) exceeds tolerance (%.0fs)", driftSec, p.cfg.MaxTimestampDriftSec))
		}
	}

	// 2. Nonce / Transaction ID Duplicate Check
	if nonceKey != "" {
		if seenAt, exists := p.seenNonces[nonceKey]; exists {
			// Check if within active retention window
			if now.Sub(seenAt) <= p.cfg.NonceRetentionWindow {
				signals.ReplayDetected = true
				rules = append(rules, fmt.Sprintf("Duplicate request nonce/transaction_id detected: '%s' seen %v ago", nonceKey, now.Sub(seenAt).Round(time.Millisecond)))
				return signals, rules
			}
		}

		// Clean up if cache exceeds capacity
		if len(p.seenNonces) >= p.cfg.MaxCacheEntries {
			// Evict expired entries
			for k, v := range p.seenNonces {
				if now.Sub(v) > p.cfg.NonceRetentionWindow {
					delete(p.seenNonces, k)
				}
			}
		}

		// Store fresh nonce
		p.seenNonces[nonceKey] = now
	}

	return signals, rules
}
