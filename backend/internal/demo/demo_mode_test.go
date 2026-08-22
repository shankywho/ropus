package demo

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDemoMode_DeterministicLifecycle(t *testing.T) {
	mgr := NewDemoModeManager()

	// 1. Create Session
	session, err := mgr.CreateSession("org_test_investor")
	require.NoError(t, err)
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, DemoStateReady, session.State)
	assert.Equal(t, 1, session.CurrentStep)
	assert.Equal(t, 7, session.TotalSteps)

	// Step 1: Normal Baseline
	step1 := session.Steps[0]
	assert.Equal(t, "APPROVE", step1.Verdict)
	assert.Equal(t, 0.04, step1.CumulativeScore)
	assert.NotEmpty(t, step1.ObservedFact)
	assert.NotEmpty(t, step1.InferredPattern)
	assert.NotEmpty(t, step1.RecommendedAction)

	// 2. Advance through Steps 2 to 7
	for s := 2; s <= 7; s++ {
		step, err := mgr.ExecuteStep(session.SessionID, s)
		require.NoError(t, err)
		assert.Equal(t, s, step.StepNumber)
	}

	// Verify Final Step State
	finalSession, err := mgr.GetSession(session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, DemoStateCompleted, finalSession.State)
	assert.Equal(t, 7, finalSession.CurrentStep)
	assert.Equal(t, 0.96, finalSession.Steps[6].CumulativeScore)
	assert.Equal(t, "BLOCK", finalSession.Steps[6].Verdict)

	// 3. Mathematical Factor Sum Verification
	sum := 0.0
	for _, step := range finalSession.Steps {
		sum += step.ScoreContribution
	}
	assert.InDelta(t, 0.96, math.Round(sum*100)/100.0, 0.01, "Sum of score contributions must equal final composite score")

	// 4. Reset Session
	require.NoError(t, mgr.ResetSession(session.SessionID))
	resetSession, err := mgr.GetSession(session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, DemoStateReady, resetSession.State)
	assert.Equal(t, 1, resetSession.CurrentStep)
}
