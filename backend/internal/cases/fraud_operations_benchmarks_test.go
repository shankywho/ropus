package cases

import (
	"testing"
)

func BenchmarkCases_CreateCase(b *testing.B) {
	cm := NewCaseManager(nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = cm.CreateCase("tenant_bench", "usr_bench_01", []string{"txn_1"}, 1500.0, 0.85, 3, 2, false)
	}
}

func BenchmarkCases_InvestigationGeneration(b *testing.B) {
	agent := NewAutonomousInvestigationAgent(nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = agent.Investigate(
			"txn_bench_01", "usr_bench_01", "dev_iphone_15", "10.0.0.1",
			1200.0, 0.88, 4,
			[]string{"Spending spike 5x", "New device"},
			[]string{"Proxy flagged"},
		)
	}
}

func BenchmarkCases_ResponseExecution(b *testing.B) {
	re := NewResponseEngine(nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = re.ExecuteAction(ActionFreezeAccount, "acc_bench_01", "ACCOUNT", "Syndicate member", "SYSTEM", 0.95, 0.92)
	}
}

func BenchmarkCases_EvidenceRetrieval(b *testing.B) {
	cm := NewCaseManager(nil)
	c, _ := cm.CreateCase("tenant_bench", "usr_bench_01", []string{"txn_1"}, 1500.0, 0.85, 3, 2, false)
	_ = cm.AddEvidence(c.CaseID, EvidenceItem{EvidenceID: "ev_1", Summary: "Forensic test artifact"})

	copilot := NewAnalystCopilot(cm)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = copilot.GetCaseEvidence(c.CaseID)
	}
}
