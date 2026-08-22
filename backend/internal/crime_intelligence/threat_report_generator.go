package crime_intelligence

import (
	"fmt"
	"time"
)

// ExecutiveThreatReport provides high-level financial crime posture for board and executive audiences.
type ExecutiveThreatReport struct {
	ReportID              string    `json:"report_id"`
	ExecutiveSummary      string    `json:"executive_summary"`
	TotalLossesPrevented  float64   `json:"total_losses_prevented"`
	ActiveSyndicatesCount int       `json:"active_syndicates_count"`
	KeyVulnerabilities    []string  `json:"key_vulnerabilities"`
	RecommendedActions    []string  `json:"recommended_actions"`
	GeneratedAt           time.Time `json:"generated_at"`
}

// ForensicAnalystReport provides granular investigative intelligence for fraud operations teams.
type ForensicAnalystReport struct {
	ReportID           string                 `json:"report_id"`
	AdversaryGroup     string                 `json:"adversary_group"`
	ObservedTimeline   []string               `json:"observed_timeline"`
	HashedEntities     []string               `json:"hashed_entities"`
	ConfidenceScore    float64                `json:"confidence_score"`
	TechnicalDetails   map[string]interface{} `json:"technical_details"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

// ThreatReportGenerator formats executive and technical forensic reports.
type ThreatReportGenerator struct{}

// NewThreatReportGenerator initializes the report generator.
func NewThreatReportGenerator() *ThreatReportGenerator {
	return &ThreatReportGenerator{}
}

// GenerateExecutiveReport creates an executive threat report.
func (g *ThreatReportGenerator) GenerateExecutiveReport(syndicatesCount int, lossPrevented float64) *ExecutiveThreatReport {
	now := time.Now().UTC()
	return &ExecutiveThreatReport{
		ReportID:              fmt.Sprintf("exec_rep_%d", now.UnixNano()),
		ExecutiveSummary:      fmt.Sprintf("Global threat posture: %d transnational syndicates tracked; $%.2f fraud prevented", syndicatesCount, lossPrevented),
		TotalLossesPrevented:  lossPrevented,
		ActiveSyndicatesCount: syndicatesCount,
		KeyVulnerabilities:    []string{"Residential proxy rotation in EU checkout channels", "Synthetic account aging"},
		RecommendedActions:    []string{"Enforce biometric step-up for high risk ASN ranges", "Synchronize laundering signatures with partner rails"},
		GeneratedAt:           now,
	}
}

// GenerateAnalystReport creates a technical forensic report.
func (g *ThreatReportGenerator) GenerateAnalystReport(rawGroup string, entities []string, confidence float64) *ForensicAnalystReport {
	now := time.Now().UTC()
	var hashed []string
	for _, e := range entities {
		hashed = append(hashed, HashID(e))
	}

	return &ForensicAnalystReport{
		ReportID:         fmt.Sprintf("analyst_rep_%d", now.UnixNano()),
		AdversaryGroup:   "Adversary_" + HashID(rawGroup)[:8],
		ObservedTimeline: []string{"T-72h: Proxy pool reconnaissance", "T-24h: Card testing burst", "T-0h: Distributed cashout attempt"},
		HashedEntities:   hashed,
		ConfidenceScore:  confidence,
		TechnicalDetails: map[string]interface{}{
			"bot_framework": "Puppeteer-Stealth-Custom",
			"proxy_asn":     "AS12345",
		},
		GeneratedAt: now,
	}
}
