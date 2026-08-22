package streaming

import (
	"context"
	"testing"
)

func BenchmarkStreaming_EventBusPublish(b *testing.B) {
	ctx := context.Background()
	bus := NewLocalEventBus()

	evt := &StreamingEvent{
		EventID:  "evt_bench",
		TenantID: "tenant_01",
		Type:     EventTransactionCreated,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = bus.Publish(ctx, "bench.events", evt)
	}
}

func BenchmarkStreaming_StreamFraudDetector(b *testing.B) {
	detector := NewStreamFraudDetector()

	evt := &StreamingEvent{
		EventID: "evt_bench",
		Type:    EventTransactionCreated,
		Payload: map[string]interface{}{
			"device_fingerprint": "dev_bench_01",
			"ip_address":         "10.0.0.1",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = detector.ProcessTransactionEvent(evt)
	}
}

func BenchmarkStreaming_CampaignDetection(b *testing.B) {
	cd := NewCampaignDetector()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cd.IngestAlert("bank_1", "CREDENTIAL_STUFFING", "ip_100_1")
	}
}

func BenchmarkStreaming_ImpactAnalysis(b *testing.B) {
	analyzer := NewImpactAnalyzer()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = analyzer.AnalyzeImpact("DEVICE_BLOCK", "dev_123", 50, 10000.0, 0.95)
	}
}

func BenchmarkStreaming_DigitalTwin(b *testing.B) {
	twin := NewFraudDigitalTwin()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = twin.SimulatePolicyChange("Simulate Subnet Freeze", 500, 200.0, 0.95)
	}
}
