package graph

import (
	"encoding/json"
	"net/http"
)

// Handler provides HTTP endpoints for knowledge graph visualization and queries.
type Handler struct {
	engine *GraphEngine
}

// NewHandler initializes a new graph HTTP handler.
func NewHandler(engine *GraphEngine) *Handler {
	return &Handler{engine: engine}
}

// GetGraph handles GET /v1/graph, returning the active fraud graph topology.
func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decisionID := r.URL.Query().Get("decisionId")
	rootID := r.URL.Query().Get("rootId")
	if rootID == "" {
		rootID = r.URL.Query().Get("accountId")
	}
	if rootID == "" {
		rootID = r.URL.Query().Get("deviceId")
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	graphData := h.engine.ExportTenantFraudGraph(tenantID, decisionID, rootID)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(graphData)
}
