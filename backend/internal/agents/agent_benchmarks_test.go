package agents

import (
	"testing"
)

func BenchmarkAgents_MessageBusPublish(b *testing.B) {
	bus := NewLocalAgentBus()
	_ = bus.Subscribe(AgentFraudInvestigator, func(msg *AgentMessage) error {
		return nil
	})

	msg := &AgentMessage{
		MessageID:   "msg_bench",
		TargetAgent: AgentFraudInvestigator,
		Type:        MsgTask,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = bus.Publish(msg)
	}
}

func BenchmarkAgents_FraudReasoning(b *testing.B) {
	engine := NewFraudReasoningEngine()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = engine.Reason(0.92, 5, []string{"Velocity surge 10x"}, []string{"Tor exit node IP"}, 1)
	}
}

func BenchmarkAgents_DecisionExplainer(b *testing.B) {
	engine := NewFraudReasoningEngine()
	explainer := NewDecisionExplainer()
	hypo := engine.Reason(0.92, 5, []string{"Velocity surge 10x"}, []string{"Tor exit node IP"}, 1)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = explainer.Explain(hypo)
	}
}

func BenchmarkAgents_AutonomousInvestigator(b *testing.B) {
	mem := NewAgentMemory()
	investigator := NewAutonomousInvestigatorAgent(mem)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = investigator.InvestigateIncident(
			"trc_bench", "usr_bench_01",
			0.94, 8,
			[]string{"Device fingerprint mismatch"},
			[]string{"Compromised proxy list match"},
		)
	}
}

func BenchmarkAgents_ResponsePlanner(b *testing.B) {
	planner := NewAutonomousResponsePlanner(nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = planner.PlanOptimalResponse(5000.0, 0.95, 2)
	}
}
