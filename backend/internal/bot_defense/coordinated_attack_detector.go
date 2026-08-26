package bot_defense

import (
	"fmt"
	"sync"
	"time"
)

// CoordinatedEntityTracker tracks sliding window relationships between entities.
type CoordinatedEntityTracker struct {
	DeviceToAccounts map[string]map[string]time.Time // device -> set of accountIDs
	AccountToDevices map[string]map[string]time.Time // account -> set of devices
	IPToDevices      map[string]map[string]time.Time // IP -> set of devices
	DeviceToCards    map[string]map[string]time.Time // device -> set of cardHashes
}

// CoordinatedAttackDetector detects multi-account stuffing, card testing probing, and syndicate fan-out.
type CoordinatedAttackDetector struct {
	mu      sync.Mutex
	tracker *CoordinatedEntityTracker
	window  time.Duration
}

// NewCoordinatedAttackDetector initializes the coordinated attack detector.
func NewCoordinatedAttackDetector(window time.Duration) *CoordinatedAttackDetector {
	if window <= 0 {
		window = 15 * time.Minute
	}
	return &CoordinatedAttackDetector{
		tracker: &CoordinatedEntityTracker{
			DeviceToAccounts: make(map[string]map[string]time.Time),
			AccountToDevices: make(map[string]map[string]time.Time),
			IPToDevices:      make(map[string]map[string]time.Time),
			DeviceToCards:    make(map[string]map[string]time.Time),
		},
		window: window,
	}
}

// AnalyzeCoordinatedRisk updates the relationship mesh and returns fan-out counts and rule triggers.
func (d *CoordinatedAttackDetector) AnalyzeCoordinatedRisk(ctx *BotDefenseContext) (signals BotSignals, rules []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()

	// 1. Device -> Account Fan-Out (Credential Stuffing / Synthetic Account Farms)
	if ctx.DeviceFingerprint != "" && ctx.AccountID != "" {
		if _, exists := d.tracker.DeviceToAccounts[ctx.DeviceFingerprint]; !exists {
			d.tracker.DeviceToAccounts[ctx.DeviceFingerprint] = make(map[string]time.Time)
		}
		d.tracker.DeviceToAccounts[ctx.DeviceFingerprint][ctx.AccountID] = now

		// Count active unique accounts in window
		var accCount int64
		for _, t := range d.tracker.DeviceToAccounts[ctx.DeviceFingerprint] {
			if now.Sub(t) <= d.window {
				accCount++
			}
		}
		signals.AccountFanOut1h = accCount

		if accCount >= 4 {
			rules = append(rules, fmt.Sprintf("High device account fan-out (%d distinct accounts accessed from device %s in %v)", accCount, ctx.DeviceFingerprint, d.window))
		}
	}

	// 2. Account -> Device Fan-Out (Account Takeover / Distributed Session Abuse)
	if ctx.AccountID != "" && ctx.DeviceFingerprint != "" {
		if _, exists := d.tracker.AccountToDevices[ctx.AccountID]; !exists {
			d.tracker.AccountToDevices[ctx.AccountID] = make(map[string]time.Time)
		}
		d.tracker.AccountToDevices[ctx.AccountID][ctx.DeviceFingerprint] = now

		var devCount int64
		for _, t := range d.tracker.AccountToDevices[ctx.AccountID] {
			if now.Sub(t) <= d.window {
				devCount++
			}
		}
		signals.DeviceFanOut1h = devCount

		if devCount >= 4 {
			rules = append(rules, fmt.Sprintf("High account device fan-out (%d distinct hardware devices accessed account %s in %v)", devCount, ctx.AccountID, d.window))
		}
	}

	// 3. IP -> Device Fan-Out (Distributed Botnet / IP Proxy Swarm)
	if ctx.IPAddress != "" && ctx.DeviceFingerprint != "" {
		if _, exists := d.tracker.IPToDevices[ctx.IPAddress]; !exists {
			d.tracker.IPToDevices[ctx.IPAddress] = make(map[string]time.Time)
		}
		d.tracker.IPToDevices[ctx.IPAddress][ctx.DeviceFingerprint] = now

		var ipDevCount int64
		for _, t := range d.tracker.IPToDevices[ctx.IPAddress] {
			if now.Sub(t) <= d.window {
				ipDevCount++
			}
		}
		signals.IPDeviceFanOut5m = ipDevCount

		if ipDevCount >= 6 {
			rules = append(rules, fmt.Sprintf("IP device cluster fan-out (%d distinct devices from IP %s in %v)", ipDevCount, ctx.IPAddress, d.window))
		}
	}

	// 4. Device -> Card Testing Probing (Carding Attacks)
	if ctx.DeviceFingerprint != "" && ctx.CardHash != "" {
		if _, exists := d.tracker.DeviceToCards[ctx.DeviceFingerprint]; !exists {
			d.tracker.DeviceToCards[ctx.DeviceFingerprint] = make(map[string]time.Time)
		}
		d.tracker.DeviceToCards[ctx.DeviceFingerprint][ctx.CardHash] = now

		var cardCount int64
		for _, t := range d.tracker.DeviceToCards[ctx.DeviceFingerprint] {
			if now.Sub(t) <= d.window {
				cardCount++
			}
		}
		signals.CardTestingTokens5m = cardCount

		// Trigger if >= 3 cards in short window OR >= 2 cards with micro-amounts (< $3.00)
		if cardCount >= 3 || (cardCount >= 2 && ctx.Amount > 0 && ctx.Amount <= 3.00) {
			signals.IsCardTestingProbing = true
			rules = append(rules, fmt.Sprintf("Automated card testing sequence detected (%d unique card instruments tested on device %s)", cardCount, ctx.DeviceFingerprint))
		}
	}

	return signals, rules
}
