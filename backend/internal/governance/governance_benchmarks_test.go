package governance

import (
	"testing"
)

func BenchmarkGovernance_ExplainDecision(b *testing.B) {
	expEngine := NewExplainabilityEngine()
	features := map[string]float64{
		"amount":            15000.0,
		"user_txn_count_1h": 6.0,
		"device_age_days":   0.2,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = expEngine.ExplainDecision("dec_bench_01", "BLOCK", 0.94, features)
	}
}

func BenchmarkGovernance_AuditAppend(b *testing.B) {
	trail := NewDecisionAuditTrail()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = trail.AppendDecision("dec_bench_01", "req_hash_1", "v3.0", "contract_v2", "ALLOW", "exp_01", "pol_v1", 0.12)
	}
}

func BenchmarkGovernance_PolicyEvaluation(b *testing.B) {
	pe := NewPolicyEngine()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = pe.EvaluatePolicies(0.96, 50.0)
	}
}

func BenchmarkGovernance_DashboardOverview(b *testing.B) {
	mgr := NewModelRiskManager()
	audit := NewDecisionAuditTrail()
	fairness := NewFairnessMonitor()
	reviews := NewHumanReviewSystem()
	pe := NewPolicyEngine()

	dash := NewGovernanceDashboardAggregator(mgr, audit, fairness, reviews, pe)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = dash.GetOverview()
	}
}
