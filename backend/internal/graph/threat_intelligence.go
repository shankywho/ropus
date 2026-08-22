package graph

import (
	"fmt"
	"math"
	"sync"
)

// ThreatIntelligenceEngine maintains dynamic feeds of malicious network IOCs and compromised devices.
type ThreatIntelligenceEngine struct {
	mu                 sync.RWMutex
	maliciousIPs       map[string]float64
	compromisedDevices map[string]float64
	riskyDomains       map[string]float64
}

// NewThreatIntelligenceEngine initializes the threat intelligence engine with baseline IOC feeds.
func NewThreatIntelligenceEngine() *ThreatIntelligenceEngine {
	t := &ThreatIntelligenceEngine{
		maliciousIPs:       make(map[string]float64),
		compromisedDevices: make(map[string]float64),
		riskyDomains:       make(map[string]float64),
	}
	t.initializeBaselineFeeds()
	return t
}

func (t *ThreatIntelligenceEngine) initializeBaselineFeeds() {
	t.maliciousIPs["198.51.100.44"] = 0.95 // Known Bulletproof proxy
	t.maliciousIPs["203.0.113.88"] = 0.99  // Tor exit node cluster
	t.compromisedDevices["dev_emul_root_89a"] = 0.90 // Cloned emulator fingerprint
	t.riskyDomains["temp-mail.org"] = 0.85
	t.riskyDomains["disposable-inbox.com"] = 0.90
}

// CheckThreat evaluates whether an incoming transaction intersects with known malicious IOCs.
func (t *ThreatIntelligenceEngine) CheckThreat(ipAddress, deviceFingerprint, emailDomain string) (float64, []string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var matches []string
	maxScore := 0.0

	if score, found := t.maliciousIPs[ipAddress]; found {
		matches = append(matches, fmt.Sprintf("IP address %s is flagged on threat intelligence blocklist (risk: %.2f)", ipAddress, score))
		maxScore = math.Max(maxScore, score)
	}

	if score, found := t.compromisedDevices[deviceFingerprint]; found {
		matches = append(matches, fmt.Sprintf("Device fingerprint %s is identified as compromised emulator/root kit", deviceFingerprint))
		maxScore = math.Max(maxScore, score)
	}

	if score, found := t.riskyDomains[emailDomain]; found {
		matches = append(matches, fmt.Sprintf("Email domain %s belongs to disposable/temporary mailbox provider", emailDomain))
		maxScore = math.Max(maxScore, score)
	}

	return maxScore, matches
}

// AddMaliciousIP allows asynchronous threat feed ingestion.
func (t *ThreatIntelligenceEngine) AddMaliciousIP(ip string, score float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maliciousIPs[ip] = score
}
