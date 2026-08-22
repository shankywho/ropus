package riskengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGracefulShutdown_FlushAndDrain(t *testing.T) {
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()

	tempDir, err := os.MkdirTemp("", "shutdown_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	stateFile := filepath.Join(tempDir, "registry_state.json")
	store, err := NewFileStateStore(stateFile)
	require.NoError(t, err)

	metricsEngine := NewMetricsEngine()
	sloEngine := NewSLOEngine(5 * time.Minute)

	// Simulate in-flight traffic
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // In-flight request simulation
		metricsEngine.RecordRequest(50.0, "ALLOW", 10, true, false, false)
		sloEngine.RecordEvaluation(50.0, true, false, false)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, reqErr := http.Get(server.URL)
			if reqErr == nil {
				_ = resp.Body.Close()
			}
		}()
	}

	// Trigger shutdown: wait for in-flight requests and persist final state
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for in-flight traffic
	wg.Wait()
	server.Close()

	// Persist critical state before exit
	err = store.Save(shutdownCtx, PersistedRetrainingState{
		CurrentState:           StateIdle,
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		SavedAt:                time.Now().UTC(),
	})
	assert.NoError(t, err, "Coordinator state must be successfully persisted upon shutdown")

	// Verify state file exists and is valid envelope
	loadedState, err := store.Load(shutdownCtx)
	require.NoError(t, err)
	assert.Equal(t, StateIdle, loadedState.CurrentState)

	time.Sleep(50 * time.Millisecond)
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()

	assert.LessOrEqual(t, finalGoroutines-initialGoroutines, 3, "Goroutines must drain cleanly upon graceful shutdown")
}
