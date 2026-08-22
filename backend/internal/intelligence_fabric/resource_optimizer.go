package intelligence_fabric

// SystemResourceAllocation details dynamic compute allocations across intelligence subsystems.
type SystemResourceAllocation struct {
	SubsystemName      string  `json:"subsystem_name"`
	ComputeWeight      int     `json:"compute_weight"` // 1 to 10
	BatchWindowMs      int     `json:"batch_window_ms"`
	MaxConcurrency     int     `json:"max_concurrency"`
	AllocatedMemoryMB  int     `json:"allocated_memory_mb"`
}

// AutonomousResourceOptimizer dynamically allocates hardware and worker pools across AFC-IOS pipelines.
type AutonomousResourceOptimizer struct{}

// NewAutonomousResourceOptimizer initializes the resource optimizer.
func NewAutonomousResourceOptimizer() *AutonomousResourceOptimizer {
	return &AutonomousResourceOptimizer{}
}

// BalanceResources calculates optimal allocations based on current threat level.
func (ro *AutonomousResourceOptimizer) BalanceResources(threatLevel string) map[string]SystemResourceAllocation {
	alloc := make(map[string]SystemResourceAllocation)

	if threatLevel == "CRITICAL" {
		alloc["SIGNAL_INGESTION"] = SystemResourceAllocation{SubsystemName: "SIGNAL_INGESTION", ComputeWeight: 10, BatchWindowMs: 1, MaxConcurrency: 500, AllocatedMemoryMB: 8192}
		alloc["GRAPH_EVOLUTION"] = SystemResourceAllocation{SubsystemName: "GRAPH_EVOLUTION", ComputeWeight: 9, BatchWindowMs: 5, MaxConcurrency: 200, AllocatedMemoryMB: 4096}
		alloc["STRATEGY_ENGINE"] = SystemResourceAllocation{SubsystemName: "STRATEGY_ENGINE", ComputeWeight: 10, BatchWindowMs: 1, MaxConcurrency: 100, AllocatedMemoryMB: 2048}
	} else {
		alloc["SIGNAL_INGESTION"] = SystemResourceAllocation{SubsystemName: "SIGNAL_INGESTION", ComputeWeight: 5, BatchWindowMs: 10, MaxConcurrency: 100, AllocatedMemoryMB: 2048}
		alloc["GRAPH_EVOLUTION"] = SystemResourceAllocation{SubsystemName: "GRAPH_EVOLUTION", ComputeWeight: 4, BatchWindowMs: 50, MaxConcurrency: 50, AllocatedMemoryMB: 1024}
		alloc["STRATEGY_ENGINE"] = SystemResourceAllocation{SubsystemName: "STRATEGY_ENGINE", ComputeWeight: 4, BatchWindowMs: 20, MaxConcurrency: 30, AllocatedMemoryMB: 1024}
	}

	return alloc
}
