package riskengine

import (
	"fmt"
	"math"
	"sort"
)

// CalculatePSI calculates the Population Stability Index between actual (live) and expected (baseline) distributions.
// Formula: PSI = Sum((actual_i - expected_i) * ln(actual_i / expected_i))
func CalculatePSI(actualProbs, expectedProbs []float64, epsilon float64) float64 {
	if len(actualProbs) == 0 || len(expectedProbs) == 0 || len(actualProbs) != len(expectedProbs) {
		return 0.0
	}
	if epsilon <= 0 {
		epsilon = 1e-4
	}

	// Normalize distributions if needed
	var actSum, expSum float64
	for i := 0; i < len(actualProbs); i++ {
		if !math.IsNaN(actualProbs[i]) && actualProbs[i] > 0 {
			actSum += actualProbs[i]
		}
		if !math.IsNaN(expectedProbs[i]) && expectedProbs[i] > 0 {
			expSum += expectedProbs[i]
		}
	}
	if actSum == 0 || expSum == 0 {
		return 0.0
	}

	var psi float64
	for i := 0; i < len(actualProbs); i++ {
		act := actualProbs[i] / actSum
		exp := expectedProbs[i] / expSum

		// Epsilon smoothing for zero counts/probabilities
		if act < epsilon {
			act = epsilon
		}
		if exp < epsilon {
			exp = epsilon
		}

		psi += (act - exp) * math.Log(act/exp)
	}

	if math.IsNaN(psi) || math.IsInf(psi, 0) || psi < 0 {
		return 0.0
	}

	return math.Round(psi*10000) / 10000
}

// CalculateKLDivergence calculates the Kullback-Leibler divergence KL(P || Q) = Sum(P(i) * ln(P(i) / Q(i))).
func CalculateKLDivergence(p, q []float64, epsilon float64) float64 {
	if len(p) == 0 || len(q) == 0 || len(p) != len(q) {
		return 0.0
	}
	if epsilon <= 0 {
		epsilon = 1e-4
	}

	var pSum, qSum float64
	for i := 0; i < len(p); i++ {
		if !math.IsNaN(p[i]) && p[i] > 0 {
			pSum += p[i]
		}
		if !math.IsNaN(q[i]) && q[i] > 0 {
			qSum += q[i]
		}
	}
	if pSum == 0 || qSum == 0 {
		return 0.0
	}

	var kl float64
	for i := 0; i < len(p); i++ {
		pi := p[i] / pSum
		qi := q[i] / qSum

		if pi < epsilon {
			pi = epsilon
		}
		if qi < epsilon {
			qi = epsilon
		}

		kl += pi * math.Log(pi/qi)
	}

	if math.IsNaN(kl) || math.IsInf(kl, 0) || kl < 0 {
		return 0.0
	}

	return math.Round(kl*10000) / 10000
}

// CalculateJSDivergence calculates the symmetric bounded Jensen-Shannon Divergence:
// JSD(P || Q) = 0.5 * KL(P || M) + 0.5 * KL(Q || M), where M = 0.5 * (P + Q).
// Range is strictly [0, ln(2)] ≈ [0, 0.69315].
func CalculateJSDivergence(p, q []float64) float64 {
	if len(p) == 0 || len(q) == 0 || len(p) != len(q) {
		return 0.0
	}

	var pSum, qSum float64
	for i := 0; i < len(p); i++ {
		if !math.IsNaN(p[i]) && p[i] > 0 {
			pSum += p[i]
		}
		if !math.IsNaN(q[i]) && q[i] > 0 {
			qSum += q[i]
		}
	}
	if pSum == 0 || qSum == 0 {
		return 0.0
	}

	n := len(p)
	normP := make([]float64, n)
	normQ := make([]float64, n)
	m := make([]float64, n)
	for i := 0; i < n; i++ {
		normP[i] = p[i] / pSum
		normQ[i] = q[i] / qSum
		m[i] = 0.5 * (normP[i] + normQ[i])
	}

	klPM := CalculateKLDivergence(normP, m, 1e-4)
	klQM := CalculateKLDivergence(normQ, m, 1e-4)

	jsd := 0.5*klPM + 0.5*klQM
	if math.IsNaN(jsd) || math.IsInf(jsd, 0) || jsd < 0 {
		return 0.0
	}

	return math.Round(jsd*10000) / 10000
}

// BinContinuousFeature bins a slice of continuous live values into predefined baseline histogram bins.
func BinContinuousFeature(values []float64, binEdges []float64) []float64 {
	if len(values) == 0 || len(binEdges) < 2 {
		return nil
	}

	numBins := len(binEdges) - 1
	counts := make([]float64, numBins)
	validTotal := 0.0

	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		validTotal++

		// Values below min bin edge go to first bin
		if v <= binEdges[0] {
			counts[0]++
			continue
		}
		// Values above max bin edge go to last bin
		if v >= binEdges[len(binEdges)-1] {
			counts[numBins-1]++
			continue
		}

		// Find corresponding bin index
		placed := false
		for i := 0; i < numBins; i++ {
			if v >= binEdges[i] && v < binEdges[i+1] {
				counts[i]++
				placed = true
				break
			}
		}
		if !placed {
			counts[numBins-1]++
		}
	}

	if validTotal == 0 {
		return nil
	}

	// Normalize to probability distribution
	probs := make([]float64, numBins)
	for i := 0; i < numBins; i++ {
		probs[i] = counts[i] / validTotal
	}

	return probs
}

// BinCategoricalFeature maps live values into discrete category buckets and computes category probabilities & unseen rate.
func BinCategoricalFeature(values []float64, categories []string, baseCategoryProbs map[string]float64) ([]float64, []float64, float64) {
	if len(values) == 0 {
		return nil, nil, 0.0
	}

	// Build category index lookup (bounded categories)
	catIndex := make(map[string]int, len(categories))
	for i, c := range categories {
		catIndex[c] = i
	}

	numCategories := len(categories)
	// Add an extra bucket for OTHER / unseen
	hasOther := false
	for _, c := range categories {
		if c == "__OTHER__" || c == "OTHER" {
			hasOther = true
			break
		}
	}
	if !hasOther {
		numCategories++ // last index for unseen
	}

	liveCounts := make([]float64, numCategories)
	unseenCount := 0.0
	total := float64(len(values))

	for _, v := range values {
		catStr := fmt.Sprintf("%g", v)
		if idx, exists := catIndex[catStr]; exists {
			liveCounts[idx]++
		} else {
			unseenCount++
			if !hasOther {
				liveCounts[numCategories-1]++
			} else if oIdx, oExists := catIndex["__OTHER__"]; oExists {
				liveCounts[oIdx]++
			} else if oIdx2, oExists2 := catIndex["OTHER"]; oExists2 {
				liveCounts[oIdx2]++
			}
		}
	}

	liveProbs := make([]float64, numCategories)
	for i := 0; i < numCategories; i++ {
		liveProbs[i] = liveCounts[i] / total
	}

	expectedProbs := make([]float64, numCategories)
	for i, c := range categories {
		if p, ok := baseCategoryProbs[c]; ok {
			expectedProbs[i] = p
		}
	}
	if !hasOther {
		expectedProbs[numCategories-1] = 1e-5
	}

	unseenRate := unseenCount / total
	return liveProbs, expectedProbs, unseenRate
}

// calculatePercentile extracts a percentile value from a sorted float64 slice.
func calculatePercentile(sorted []float64, pct float64) float64 {
	if len(sorted) == 0 {
		return 0.0
	}
	if pct <= 0 {
		return sorted[0]
	}
	if pct >= 1.0 {
		return sorted[len(sorted)-1]
	}
	idx := int(float64(len(sorted)) * pct)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// CalculateFeatureDrift evaluates a single feature's live values against its reference baseline.
func CalculateFeatureDrift(liveValues []float64, baseline FeatureDistribution, cfg DriftConfig) FeatureDriftResult {
	res := FeatureDriftResult{
		FeatureName:  baseline.Name,
		SampleCount:  len(liveValues),
		BaselineMean: baseline.Mean,
		BaselineStd:  baseline.Std,
		Severity:     SeverityStable,
	}

	if len(liveValues) == 0 {
		return res
	}

	// 1. Calculate live descriptive statistics and missing rate
	var sum, sumSq float64
	validValues := make([]float64, 0, len(liveValues))
	missingCount := 0

	for _, v := range liveValues {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			missingCount++
			continue
		}
		validValues = append(validValues, v)
		sum += v
		sumSq += v * v
	}

	res.MissingRate = float64(missingCount) / float64(len(liveValues))
	if len(validValues) == 0 {
		res.Severity = SeverityCritical
		res.Recommendation = "100% missing values in live evaluation stream"
		return res
	}

	n := float64(len(validValues))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	std := math.Sqrt(variance)

	res.LiveMean = math.Round(mean*10000) / 10000
	res.LiveStd = math.Round(std*10000) / 10000
	res.MeanShift = math.Round((mean-baseline.Mean)*10000) / 10000
	res.StdShift = math.Round((std-baseline.Std)*10000) / 10000

	// Compute all required moments and percentiles on sorted values
	sort.Float64s(validValues)
	res.LiveMin = validValues[0]
	res.LiveMax = validValues[len(validValues)-1]
	res.LiveP01 = calculatePercentile(validValues, 0.01)
	res.LiveP05 = calculatePercentile(validValues, 0.05)
	res.LiveP25 = calculatePercentile(validValues, 0.25)
	res.LiveP50 = calculatePercentile(validValues, 0.50)
	res.LiveP75 = calculatePercentile(validValues, 0.75)
	res.LiveP95 = calculatePercentile(validValues, 0.95)
	res.LiveP99 = calculatePercentile(validValues, 0.99)

	// 2. Binning & Divergence calculation
	if baseline.IsCategorical && len(baseline.Categories) > 0 && len(baseline.CategoryProbs) > 0 {
		liveProbs, expProbs, unseenRate := BinCategoricalFeature(validValues, baseline.Categories, baseline.CategoryProbs)
		res.UnseenCategoryRate = unseenRate
		res.PSI = CalculatePSI(liveProbs, expProbs, cfg.Epsilon)
		res.JSD = CalculateJSDivergence(liveProbs, expProbs)
		res.KL = CalculateKLDivergence(liveProbs, expProbs, cfg.Epsilon)
	} else if len(baseline.BinEdges) >= 2 && len(baseline.BinProbs) == len(baseline.BinEdges)-1 {
		liveProbs := BinContinuousFeature(validValues, baseline.BinEdges)
		res.PSI = CalculatePSI(liveProbs, baseline.BinProbs, cfg.Epsilon)
		res.JSD = CalculateJSDivergence(liveProbs, baseline.BinProbs)
		res.KL = CalculateKLDivergence(liveProbs, baseline.BinProbs, cfg.Epsilon)
	}

	// 3. Classify Severity
	if res.PSI >= cfg.PSICritThreshold || res.JSD >= cfg.JSDCritThreshold {
		res.Severity = SeverityCritical
		res.Recommendation = "Significant distribution shift detected; model review recommended"
	} else if res.PSI >= cfg.PSIHighThreshold || res.JSD >= cfg.JSDHighThreshold {
		res.Severity = SeverityHigh
		res.Recommendation = "Moderate distribution drift; monitor feature trends closely"
	} else if res.PSI >= cfg.PSIWarnThreshold || res.JSD >= cfg.JSDWarnThreshold {
		res.Severity = SeverityWarning
		res.Recommendation = "Slight distribution divergence; baseline stable"
	} else {
		res.Severity = SeverityStable
		res.Recommendation = "Distribution aligned with baseline"
	}

	return res
}
