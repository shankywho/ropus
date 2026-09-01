package chaos

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler exposes chaos engineering drills over HTTP for live demonstration and verification.
type Handler struct {
	engine *ChaosEngine
}

// NewHandler initializes the HTTP handler for chaos drills.
func NewHandler(engine *ChaosEngine) *Handler {
	return &Handler{engine: engine}
}

// HandleExecuteDrill handles POST /admin/chaos/drill or POST /v1/chaos/drill.
func (h *Handler) HandleExecuteDrill(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Scenario string `json:"scenario"`
	}

	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	scenarioStr := strings.ToUpper(strings.TrimSpace(req.Scenario))
	if scenarioStr == "" {
		scenarioStr = string(ChaosRedisFailure)
	}

	scenario := ChaosScenario(scenarioStr)
	result := h.engine.ExecuteDrill(scenario)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
