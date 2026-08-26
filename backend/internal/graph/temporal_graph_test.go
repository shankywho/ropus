package graph_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/graph"
)

func TestTemporalFraudGraph(t *testing.T) {
	// Fixed reference anchor timestamp (2026-08-25 12:00:00 UTC)
	refTime := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	window := 72 * time.Hour // 72 hours window

	t.Run("1. Recent Edge Is Traversed", func(t *testing.T) {
		store := graph.NewLocalGraphStore()

		// Node A (User) -> Node B (Device), Edge created 2 hours before refTime
		_ = store.AddNode(&graph.Node{ID: "usr_alice", Type: graph.NodeUser, CreatedAt: refTime})
		_ = store.AddNode(&graph.Node{ID: "dev_iphone", Type: graph.NodeDevice, CreatedAt: refTime})
		_ = store.AddEdge(&graph.Edge{
			ID:        "e_recent",
			SourceID:  "usr_alice",
			TargetID:  "dev_iphone",
			Type:      graph.EdgeUsedBy,
			CreatedAt: refTime.Add(-2 * time.Hour),
		})

		neighbors, err := store.QueryNeighborsTemporal("usr_alice", "", refTime, window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(neighbors) != 1 || neighbors[0].ID != "dev_iphone" {
			t.Errorf("expected 1 recent neighbor dev_iphone, got %v", neighbors)
		}

		evidence, err := store.Traverse3HopTemporal("usr_alice", refTime, window, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if evidence.VisitedNodesCount != 2 {
			t.Errorf("expected 2 visited nodes, got %d", evidence.VisitedNodesCount)
		}
		if evidence.TraversedEdgesCount != 1 {
			t.Errorf("expected 1 traversed edge, got %d", evidence.TraversedEdgesCount)
		}
	})

	t.Run("2. Expired Edge Is Ignored", func(t *testing.T) {
		store := graph.NewLocalGraphStore()

		// Node A -> Node B, Edge created 100 hours ago (outside 72h window)
		_ = store.AddNode(&graph.Node{ID: "usr_bob", Type: graph.NodeUser, CreatedAt: refTime})
		_ = store.AddNode(&graph.Node{ID: "dev_old_laptop", Type: graph.NodeDevice, CreatedAt: refTime})
		_ = store.AddEdge(&graph.Edge{
			ID:        "e_expired",
			SourceID:  "usr_bob",
			TargetID:  "dev_old_laptop",
			Type:      graph.EdgeUsedBy,
			CreatedAt: refTime.Add(-100 * time.Hour),
		})

		neighbors, err := store.QueryNeighborsTemporal("usr_bob", "", refTime, window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(neighbors) != 0 {
			t.Errorf("expected 0 neighbors for expired edge, got %d", len(neighbors))
		}

		evidence, err := store.Traverse3HopTemporal("usr_bob", refTime, window, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if evidence.VisitedNodesCount != 1 {
			t.Errorf("expected 1 visited node (start node only), got %d", evidence.VisitedNodesCount)
		}
		if evidence.TraversedEdgesCount != 0 {
			t.Errorf("expected 0 traversed edges, got %d", evidence.TraversedEdgesCount)
		}
	})

	t.Run("3. Mixed Recent and Expired Edges Produce Clean Evidence", func(t *testing.T) {
		store := graph.NewLocalGraphStore()

		// Node A has:
		// - Recent link to Bad Device (10h ago)
		// - Expired link to Stale Device (90h ago)
		_ = store.AddNode(&graph.Node{ID: "usr_mule", Type: graph.NodeUser, CreatedAt: refTime})
		_ = store.AddNode(&graph.Node{ID: "dev_bad", Type: graph.NodeDevice, IsKnownBad: true, CreatedAt: refTime})
		_ = store.AddNode(&graph.Node{ID: "dev_stale", Type: graph.NodeDevice, IsKnownBad: true, CreatedAt: refTime})

		_ = store.AddEdge(&graph.Edge{
			ID:        "e_recent_bad",
			SourceID:  "usr_mule",
			TargetID:  "dev_bad",
			Type:      graph.EdgeUsedBy,
			CreatedAt: refTime.Add(-10 * time.Hour),
		})
		_ = store.AddEdge(&graph.Edge{
			ID:        "e_stale_bad",
			SourceID:  "usr_mule",
			TargetID:  "dev_stale",
			Type:      graph.EdgeUsedBy,
			CreatedAt: refTime.Add(-90 * time.Hour),
		})

		neighbors, err := store.QueryNeighborsTemporal("usr_mule", "", refTime, window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(neighbors) != 1 || neighbors[0].ID != "dev_bad" {
			t.Errorf("expected only dev_bad to be returned, got %v", neighbors)
		}

		evidence, err := store.Traverse3HopTemporal("usr_mule", refTime, window, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if evidence.VisitedNodesCount != 2 {
			t.Errorf("expected 2 visited nodes (usr_mule, dev_bad), got %d", evidence.VisitedNodesCount)
		}
		if evidence.FraudNodesCount != 1 {
			t.Errorf("expected 1 fraud node (dev_bad), got %d", evidence.FraudNodesCount)
		}
	})

	t.Run("4. 3-Hop Depth Limit Is Strictly Enforced", func(t *testing.T) {
		store := graph.NewLocalGraphStore()

		// Chain: N0 -> N1 -> N2 -> N3 -> N4 -> N5 (all recent edges, 1h ago)
		nodes := []string{"n0", "n1", "n2", "n3", "n4", "n5"}
		for _, id := range nodes {
			_ = store.AddNode(&graph.Node{ID: id, Type: graph.NodeAccount, CreatedAt: refTime})
		}
		for i := 0; i < len(nodes)-1; i++ {
			_ = store.AddEdge(&graph.Edge{
				ID:        fmt.Sprintf("e_%d", i),
				SourceID:  nodes[i],
				TargetID:  nodes[i+1],
				Type:      graph.EdgeTransferredTo,
				CreatedAt: refTime.Add(-1 * time.Hour),
			})
		}

		evidence, err := store.Traverse3HopTemporal("n0", refTime, window, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Depth limit 3 means we reach n0 (depth 0), n1 (1), n2 (2), n3 (3).
		// n4 (depth 4) and n5 (depth 5) must NOT be visited.
		if evidence.MaxClusterDepth != 3 {
			t.Errorf("expected max depth 3, got %d", evidence.MaxClusterDepth)
		}
		if evidence.VisitedNodesCount != 4 {
			t.Errorf("expected exactly 4 visited nodes (n0, n1, n2, n3), got %d", evidence.VisitedNodesCount)
		}
	})

	t.Run("5. Expansion Fan-Out Cap Per Node Is Enforced", func(t *testing.T) {
		store := graph.NewLocalGraphStore()

		// Hub node connected to 100 neighbors
		_ = store.AddNode(&graph.Node{ID: "hub_device", Type: graph.NodeDevice, CreatedAt: refTime})
		for i := 0; i < 100; i++ {
			accID := fmt.Sprintf("acc_%d", i)
			_ = store.AddNode(&graph.Node{ID: accID, Type: graph.NodeAccount, CreatedAt: refTime})
			_ = store.AddEdge(&graph.Edge{
				ID:        fmt.Sprintf("e_hub_%d", i),
				SourceID:  "hub_device",
				TargetID:  accID,
				Type:      graph.EdgeConnectedTo,
				CreatedAt: refTime.Add(-1 * time.Hour),
			})
		}

		maxPerHop := 25
		evidence, err := store.Traverse3HopTemporal("hub_device", refTime, window, maxPerHop)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Visited count should be hub_device (1) + maxPerHop (25) = 26
		if evidence.VisitedNodesCount != 26 {
			t.Errorf("expected 26 visited nodes with fanout cap 25, got %d", evidence.VisitedNodesCount)
		}
	})

	t.Run("6. Exact Boundary Condition Cutoff (Deterministic Inclusivity)", func(t *testing.T) {
		store := graph.NewLocalGraphStore()

		exactCutoff := refTime.Add(-72 * time.Hour)
		justPastCutoff := refTime.Add(-72*time.Hour - time.Second)

		_ = store.AddNode(&graph.Node{ID: "usr_boundary", Type: graph.NodeUser, CreatedAt: refTime})
		_ = store.AddNode(&graph.Node{ID: "dev_exact", Type: graph.NodeDevice, CreatedAt: refTime})
		_ = store.AddNode(&graph.Node{ID: "dev_past", Type: graph.NodeDevice, CreatedAt: refTime})

		_ = store.AddEdge(&graph.Edge{
			ID:        "e_exact",
			SourceID:  "usr_boundary",
			TargetID:  "dev_exact",
			Type:      graph.EdgeUsedBy,
			CreatedAt: exactCutoff, // Exactly at -72h
		})
		_ = store.AddEdge(&graph.Edge{
			ID:        "e_past",
			SourceID:  "usr_boundary",
			TargetID:  "dev_past",
			Type:      graph.EdgeUsedBy,
			CreatedAt: justPastCutoff, // -72h - 1s (expired)
		})

		neighbors, err := store.QueryNeighborsTemporal("usr_boundary", "", refTime, window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(neighbors) != 1 || neighbors[0].ID != "dev_exact" {
			t.Errorf("expected exactly dev_exact to be included at cutoff boundary, got %v", neighbors)
		}
	})

	t.Run("7. Future Timestamps Guard (Prevents Time-Travel Anomalies)", func(t *testing.T) {
		store := graph.NewLocalGraphStore()

		futureTime := refTime.Add(10 * time.Hour)

		_ = store.AddNode(&graph.Node{ID: "usr_now", Type: graph.NodeUser, CreatedAt: refTime})
		_ = store.AddNode(&graph.Node{ID: "dev_future", Type: graph.NodeDevice, CreatedAt: refTime})
		_ = store.AddEdge(&graph.Edge{
			ID:        "e_future",
			SourceID:  "usr_now",
			TargetID:  "dev_future",
			Type:      graph.EdgeUsedBy,
			CreatedAt: futureTime, // Future timestamp relative to asOf
		})

		neighbors, err := store.QueryNeighborsTemporal("usr_now", "", refTime, window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(neighbors) != 0 {
			t.Errorf("future edge must not be included in point-in-time traversal, got %d", len(neighbors))
		}
	})
}
