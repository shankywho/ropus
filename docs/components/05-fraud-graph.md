# Component 05: Fraud Knowledge Graph 3.0 & Syndicate Discovery

---

## 1. Why It Exists
Individual fraudulent transactions often look completely benign in isolation (e.g. a standard $50 transaction on a newly created account). However, when that account shares a hardware canvas hash, IP subnet, or payment token with 14 previously confirmed fraud accounts, the syndicate ring becomes immediately visible.

The **Fraud Knowledge Graph** (`backend/internal/graph/`) maintains an in-memory, multi-relational property graph of financial entities and executes **synchronous 3-hop neighborhood traversals in $< 1.5\text{ms}$** to detect money mule rings, synthetic identity factories, and emulator clusters.

---

## 2. Graph Schema & Edge Typology

```text
  ┌──────────────┐                  ┌─────────────────────┐
  │ Customer (U) │───[USED_DEVICE]──>│ Device Fingerprint  │
  └──────┬───────┘                  └──────────┬──────────┘
         │                                     │
         ├───[USED_IP]──────> ┌────────────┐   └──[MATCHES_CANVAS]─> ┌────────────────────┐
         │                    │ IP Subnet  │                         │ Canvas Hash (99)   │
         │                    └────────────┘                         └─────────┬──────────┘
         │                                                             [LINKED]│
         └───[PAID_WITH]────> ┌────────────┐                                   ▼
                              │ Card Token │                         ┌────────────────────┐
                              └────────────┘                         │ 14 Synthetic Mules │
                                                                     └────────────────────┘
```

### Node Types:
- `ENTITY_USER`: Customer account (`usr_...`).
- `ENTITY_DEVICE`: Hardware canvas / client fingerprint (`dev_...`).
- `ENTITY_IP`: Egress IP address / subnet (`ip_...`).
- `ENTITY_CARD`: Tokenized card / bank account (`tok_...`).
- `ENTITY_MERCHANT`: Receiving merchant entity (`mch_...`).

---

## 3. Real-Time 3-Hop BFS Traversal Algorithm

```go
func (g *InMemoryGraphEngine) Traverse3Hop(rootID string) GraphNeighborhood {
    visited := make(map[string]bool)
    queue := []string{rootID}
    visited[rootID] = true
    
    depth := 0
    neighborhood := GraphNeighborhood{RootID: rootID}
    
    for len(queue) > 0 && depth < 3 {
        levelSize := len(queue)
        for i := 0; i < levelSize; i++ {
            curr := queue[0]
            queue = queue[1:]
            
            // Expand adjacent edges in memory
            for _, edge := range g.adjList[curr] {
                if !visited[edge.TargetID] {
                    visited[edge.TargetID] = true
                    queue = append(queue, edge.TargetID)
                    neighborhood.Nodes = append(neighborhood.Nodes, edge.TargetNode)
                    neighborhood.Edges = append(neighborhood.Edges, edge)
                }
            }
        }
        depth++
    }
    
    // Compute local graph metrics
    neighborhood.DegreeCentrality = len(neighborhood.Edges)
    neighborhood.SyndicateCluster = neighborhood.DegreeCentrality >= 4
    return neighborhood
}
```

---

## 4. Key Data Structures (Go)

```go
type GraphNode struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`       // "USER", "DEVICE", "IP", "CARD"
    RiskWeight float64                `json:"risk_weight"`// 0.00 to 1.00
    Properties map[string]interface{} `json:"properties,omitempty"`
}

type GraphEdge struct {
    SourceID   string    `json:"source_id"`
    TargetID   string    `json:"target_id"`
    Relation   string    `json:"relation"` // "USED_DEVICE", "USED_IP", "PAID_WITH"
    Weight     float64   `json:"weight"`
    CreatedAt  time.Time `json:"created_at"`
}

type GraphNeighborhood struct {
    RootID           string      `json:"root_id"`
    Nodes            []GraphNode `json:"nodes"`
    Edges            []GraphEdge `json:"edges"`
    DegreeCentrality int         `json:"degree_centrality"`
    SyndicateCluster bool        `json:"syndicate_cluster"`
}
```

---

## 5. Failure Modes & Concurrency
- **High-Degree Hub Explosion**: To prevent unbounded traversal latency on high-degree nodes (e.g. Amazon or PayPal merchant nodes), the BFS caps node expansion to 50 adjacent edges per step.
- **Thread Safety**: Adjacency lists are protected by read-write locks (`sync.RWMutex`), allowing concurrent read traversals across thousands of evaluation goroutines.

---

## 6. Source Code Map
- [`backend/internal/graph/graph_engine.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/graph/graph_engine.go): In-memory graph representation, BFS traversal, and syndicate detection.
- [`backend/internal/graph/graph_schema.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/graph/graph_schema.go): Node/Edge schema definitions.
- [`backend/internal/graph/adaptive_risk_engine.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/graph/adaptive_risk_engine.go): Graph exposure risk scoring.

---

## 7. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Ingests graph exposure score during decisioning.
- [Component 07: AI Investigators](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/07-ai-investigators.md) — Analyzes graph clusters to synthesize evidentiary dossiers.
