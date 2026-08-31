package graph

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphAPI_ExportStructure(t *testing.T) {
	engine := NewGraphEngine(NewLocalGraphStore())
	engine.SeedDefaultDemoGraph()

	res := engine.ExportFraudGraph("dec_test_123", "PA-77120")

	assert.NotNil(t, res)
	assert.Equal(t, "PA-77120", res.RootID)
	assert.Equal(t, "dec_test_123", res.DecisionID)
	assert.Equal(t, "live_graph_engine", res.Source)
	assert.True(t, len(res.Entities) >= 7, "Must contain at least 7 entities from seed")
	assert.True(t, len(res.Relationships) >= 6, "Must contain at least 6 edges from seed")

	// Verify all coordinates are bounded to 0-100 viewBox
	for _, ent := range res.Entities {
		assert.NotEmpty(t, ent.ID)
		assert.NotEmpty(t, ent.Type)
		assert.NotEmpty(t, ent.Label)
		assert.NotEmpty(t, ent.Risk)
		assert.True(t, ent.X >= 0.0 && ent.X <= 100.0, "X coordinate must be within [0, 100]")
		assert.True(t, ent.Y >= 0.0 && ent.Y <= 100.0, "Y coordinate must be within [0, 100]")
	}
}

func TestGraphAPI_PIIMasking(t *testing.T) {
	engine := NewGraphEngine(NewLocalGraphStore())
	_ = engine.IngestTransactionLinks("txn_pii_1", "usr_1", "acc_1", "tok_1", "dev_1", "198.51.100.44", "merch_1", 100.0, false)

	res := engine.ExportFraudGraph("dec_pii", "acc_1")

	foundIP := false
	for _, ent := range res.Entities {
		if ent.Type == "IP" {
			foundIP = true
			assert.Contains(t, ent.Label, "***.***", "IP address label must mask octets")
			assert.False(t, strings.Contains(ent.Label, "198.51.100.44"), "Raw unmasked IP must not appear in label")
		}
	}
	assert.True(t, foundIP, "Must have IP entity in exported graph")
}

func TestGraphAPI_HTTPHandler(t *testing.T) {
	engine := NewGraphEngine(NewLocalGraphStore())
	handler := NewHandler(engine)

	req := httptest.NewRequest(http.MethodGet, "/v1/graph?decisionId=dec_http_99&rootId=PA-77120", nil)
	rr := httptest.NewRecorder()

	handler.GetGraph(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp FrontendFraudGraphResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "dec_http_99", resp.DecisionID)
	assert.Equal(t, "PA-77120", resp.RootID)
	assert.Equal(t, "live_graph_engine", resp.Source)
	assert.NotEmpty(t, resp.Entities)
	assert.NotEmpty(t, resp.Relationships)
}
