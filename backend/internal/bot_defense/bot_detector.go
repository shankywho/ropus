package bot_defense

import (
	"fmt"
	"math"
	"time"
)

// BotDefenseEngine is the master orchestrator for automated attack, bot, and cadence defense.
type BotDefenseEngine struct {
	layeredLimiter      *LayeredLimiter
	cadenceAnalyzer     *CadenceAnalyzer
	replayProtector     *ReplayProtector
	coordinatedDetector *CoordinatedAttackDetector
}

// NewBotDefenseEngine constructs the complete bot defense engine with production defaults.
func NewBotDefenseEngine() *BotDefenseEngine {
	return &BotDefenseEngine{
		layeredLimiter:      NewLayeredLimiter(DefaultLayeredLimiterConfig()),
		cadenceAnalyzer:     NewCadenceAnalyzer(12),
		replayProtector:     NewReplayProtector(DefaultReplayProtectorConfig()),
		coordinatedDetector: NewCoordinatedAttackDetector(15 * time.Minute),
	}
}

// Evaluate performs synchronous multi-vector bot and automated attack risk analysis with fail-open safety.
func (e *BotDefenseEngine) Evaluate(ctx *BotDefenseContext) (result *BotDefenseResult) {
	start := time.Now()
	now := time.Now().UTC()

	// Fail-open panic safety wrapper: If any internal calculation panics, fail open gracefully!
	defer func() {
		if r := recover(); r != nil {
			result = &BotDefenseResult{
				AutomationRiskScore: 0.05,
				RiskLevel:           RiskLevelLow,
				RecommendedAction:   ActionAllow,
				TriggeredRules:      nil,
				ObservedPatterns:    []string{"Bot defense failed open due to internal panic"},
				EvaluatedAt:         now,
				LatencyMs:           float64(time.Since(start).Microseconds()) / 1000.0,
				IsDegraded:          true,
				DegradeReason:       fmt.Sprintf("internal error: %v", r),
			}
		}
	}()

	var allRules []string
	var allPatterns []string
	var combinedSignals BotSignals

	// Primary Entity Key for Cadence (Device > IP > Account)
	entityKey := ctx.DeviceFingerprint
	if entityKey == "" {
		entityKey = ctx.IPAddress
	}
	if entityKey == "" {
		entityKey = ctx.AccountID
	}

	// 1. Replay & Timestamp Freshness Check
	nonceKey := ctx.Nonce
	if nonceKey == "" {
		nonceKey = ctx.TransactionID
	}
	replaySignals, replayRules := e.replayProtector.ValidateReplay(nonceKey, ctx.Timestamp)
	combinedSignals.ReplayDetected = replaySignals.ReplayDetected
	combinedSignals.TimeDriftSec = replaySignals.TimeDriftSec
	allRules = append(allRules, replayRules...)

	// 2. Layered Rate Limiting Evaluation
	rateSignals, rateRules := e.layeredLimiter.EvaluateLimits(ctx)
	combinedSignals.TenantLimited = rateSignals.TenantLimited
	combinedSignals.IPLimited = rateSignals.IPLimited
	combinedSignals.DeviceLimited = rateSignals.DeviceLimited
	combinedSignals.AccountLimited = rateSignals.AccountLimited
	combinedSignals.CardLimited = rateSignals.CardLimited
	allRules = append(allRules, rateRules...)

	// 3. Cadence & Timing Entropy Analysis
	cadenceSignals, cadenceRules := e.cadenceAnalyzer.AnalyzeCadence(entityKey, ctx.Timestamp, ctx.UserAgent)
	combinedSignals.InterArrivalTimeMeanMs = cadenceSignals.InterArrivalTimeMeanMs
	combinedSignals.InterArrivalTimeStdDev = cadenceSignals.InterArrivalTimeStdDev
	combinedSignals.CoefficientOfVariation = cadenceSignals.CoefficientOfVariation
	combinedSignals.RequestCadenceHz = cadenceSignals.RequestCadenceHz
	combinedSignals.IsDeterministicCadence = cadenceSignals.IsDeterministicCadence
	combinedSignals.IsBurstPacing = cadenceSignals.IsBurstPacing
	combinedSignals.IsHeadlessUA = cadenceSignals.IsHeadlessUA
	allRules = append(allRules, cadenceRules...)

	// 4. Coordinated Syndicate & Fan-Out Analysis
	coordSignals, coordRules := e.coordinatedDetector.AnalyzeCoordinatedRisk(ctx)
	combinedSignals.AccountFanOut1h = coordSignals.AccountFanOut1h
	combinedSignals.DeviceFanOut1h = coordSignals.DeviceFanOut1h
	combinedSignals.IPDeviceFanOut5m = coordSignals.IPDeviceFanOut5m
	combinedSignals.CardTestingTokens5m = coordSignals.CardTestingTokens5m
	combinedSignals.IsCardTestingProbing = coordSignals.IsCardTestingProbing
	allRules = append(allRules, coordRules...)

	// 5. Composite Automation Risk Scoring
	score := 0.04 // Clean baseline

	if combinedSignals.ReplayDetected {
		score += 0.50
		allPatterns = append(allPatterns, "Replay attack or duplicate transaction nonce detected")
	}
	if combinedSignals.IsDeterministicCadence {
		score += 0.35
		allPatterns = append(allPatterns, "Deterministic inter-arrival timing indicates scripted bot loop")
	}
	if combinedSignals.IsBurstPacing {
		score += 0.25
		allPatterns = append(allPatterns, "Rapid burst pacing exceeds natural human interaction thresholds")
	}
	if combinedSignals.IsHeadlessUA {
		score += 0.30
		allPatterns = append(allPatterns, "Client environment exhibits headless browser or automation driver signatures")
	}
	if combinedSignals.IsCardTestingProbing {
		score += 0.45
		allPatterns = append(allPatterns, "Rapid sequential card token testing detected across small authorization amounts")
	}
	if combinedSignals.AccountFanOut1h >= 4 {
		score += 0.35
		allPatterns = append(allPatterns, "Device is rotating across multiple distinct user account identities")
	}
	if combinedSignals.IPDeviceFanOut5m >= 6 {
		score += 0.30
		allPatterns = append(allPatterns, "IP subnet exhibiting high device fan-out indicative of proxy farm")
	}
	if combinedSignals.IPLimited || combinedSignals.DeviceLimited || combinedSignals.AccountLimited {
		score += 0.30
		allPatterns = append(allPatterns, "Request velocity exceeded layered rate limiter thresholds")
	}
	if combinedSignals.CardLimited {
		score += 0.35
		allPatterns = append(allPatterns, "Payment instrument rate limit saturated")
	}

	if score > 0.99 {
		score = 0.98
	}
	normalizedScore := math.Round(score*100) / 100.0

	// 6. Map to Graduated Action & Risk Level
	riskLevel := RiskLevelLow
	action := ActionAllow

	if normalizedScore >= 0.90 {
		riskLevel = RiskLevelCritical
		action = ActionContain
	} else if normalizedScore >= 0.70 {
		riskLevel = RiskLevelHigh
		action = ActionElevateRisk
	} else if normalizedScore >= 0.30 {
		riskLevel = RiskLevelMedium
		action = ActionScrutinize
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0

	return &BotDefenseResult{
		AutomationRiskScore: normalizedScore,
		RiskLevel:           riskLevel,
		RecommendedAction:   action,
		TriggeredRules:      allRules,
		ObservedPatterns:    allPatterns,
		Signals:             combinedSignals,
		EvaluatedAt:         now,
		LatencyMs:           latency,
		IsDegraded:          false,
	}
}
