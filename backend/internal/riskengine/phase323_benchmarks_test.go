package riskengine

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkRBAC_AuthenticateCredential(b *testing.B) {
	am := NewAuthManager()
	_ = am.RegisterKey("admin_secret_key_123", Identity{Subject: "admin_user", Role: RoleAdmin})
	_ = am.RegisterKey("ml_operator_key_456", Identity{Subject: "ml_eng", Role: RoleMLOperator})
	_ = am.RegisterKey("risk_operator_key_789", Identity{Subject: "risk_analyst", Role: RoleRiskOperator})
	_ = am.RegisterKey("readonly_key_000", Identity{Subject: "auditor", Role: RoleReadOnly})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = am.AuthenticateCredential("admin_secret_key_123")
	}
}

func BenchmarkIdempotency_Middleware(b *testing.B) {
	store := NewIdempotencyStore(1*time.Minute, 100000)
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	payload := []byte(`{"action":"promote","model":"fraud-xgb-25f-v3.0"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/models/promote", bytes.NewReader(payload))
	req.Header.Set("X-Idempotency-Key", "bench_idempotency_key_01")

	// Pre-populate cache record
	rrInit := httptest.NewRecorder()
	handler.ServeHTTP(rrInit, req)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		reqReplay := httptest.NewRequest(http.MethodPost, "/v1/models/promote", bytes.NewReader(payload))
		reqReplay.Header.Set("X-Idempotency-Key", "bench_idempotency_key_01")
		handler.ServeHTTP(rr, reqReplay)
	}
}

func BenchmarkOutbox_Enqueue(b *testing.B) {
	dispatcher := &mockDispatcher{}
	outbox := NewDurableOutbox(dispatcher, 4, 200000)
	defer func() { _ = outbox.FlushAndClose(1 * time.Second) }()

	event := OutboxEvent{
		EventID:       "evt_bench_01",
		EventType:     "RISK_EVALUATION",
		CorrelationID: "corr_01",
		Payload:       []byte(`{"score":10}`),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = outbox.Enqueue(event)
	}
}
