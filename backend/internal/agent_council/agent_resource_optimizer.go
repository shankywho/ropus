package agent_council

// AgentResourceAllocation represents compute and concurrency limits dynamically assigned to an agent role.
type AgentResourceAllocation struct {
	AgentRole         string  `json:"agent_role"`
	PriorityScore     int     `json:"priority_score"` // 1 (lowest) to 10 (highest)
	MaxConcurrentRuns int     `json:"max_concurrent_runs"`
	TimeoutSeconds    int     `json:"timeout_seconds"`
}

// AgentResourceOptimizer dynamically tunes agent compute allocations according to incident severity.
type AgentResourceOptimizer struct{}

// NewAgentResourceOptimizer initializes the resource optimizer.
func NewAgentResourceOptimizer() *AgentResourceOptimizer {
	return &AgentResourceOptimizer{}
}

// AllocateResources assigns compute weight based on incident severity.
func (o *AgentResourceOptimizer) AllocateResources(severity string) map[string]AgentResourceAllocation {
	alloc := make(map[string]AgentResourceAllocation)

	if severity == "SEV1" || severity == "CRITICAL" {
		alloc["INVESTIGATOR"] = AgentResourceAllocation{AgentRole: "INVESTIGATOR", PriorityScore: 10, MaxConcurrentRuns: 100, TimeoutSeconds: 5}
		alloc["RESPONSE"] = AgentResourceAllocation{AgentRole: "RESPONSE", PriorityScore: 10, MaxConcurrentRuns: 50, TimeoutSeconds: 2}
		alloc["COMPLIANCE"] = AgentResourceAllocation{AgentRole: "COMPLIANCE", PriorityScore: 9, MaxConcurrentRuns: 50, TimeoutSeconds: 2}
		alloc["THREAT_HUNTER"] = AgentResourceAllocation{AgentRole: "THREAT_HUNTER", PriorityScore: 8, MaxConcurrentRuns: 30, TimeoutSeconds: 10}
	} else {
		alloc["INVESTIGATOR"] = AgentResourceAllocation{AgentRole: "INVESTIGATOR", PriorityScore: 5, MaxConcurrentRuns: 20, TimeoutSeconds: 10}
		alloc["RESPONSE"] = AgentResourceAllocation{AgentRole: "RESPONSE", PriorityScore: 5, MaxConcurrentRuns: 10, TimeoutSeconds: 5}
		alloc["COMPLIANCE"] = AgentResourceAllocation{AgentRole: "COMPLIANCE", PriorityScore: 5, MaxConcurrentRuns: 10, TimeoutSeconds: 5}
		alloc["THREAT_HUNTER"] = AgentResourceAllocation{AgentRole: "THREAT_HUNTER", PriorityScore: 3, MaxConcurrentRuns: 10, TimeoutSeconds: 30}
	}

	return alloc
}
