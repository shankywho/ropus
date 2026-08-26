package graphsage

import (
	"fmt"
)

// PathExtractor searches for multi-hop explainability paths up to maxDepth.
type PathExtractor struct{}

// FindPaths executes bounded DFS to extract relationship chains.
func (p *PathExtractor) FindPaths(
	sourceID string,
	targetID string,
	nodes map[string]*HeteroNode,
	edges []*HeteroEdge,
	maxDepth int,
) []string {
	if maxDepth <= 0 {
		maxDepth = 3
	}

	adj := make(map[string][]*HeteroEdge)
	for _, e := range edges {
		adj[e.SourceID] = append(adj[e.SourceID], e)
		adj[e.TargetID] = append(adj[e.TargetID], e)
	}

	var discovered []string
	visited := make(map[string]bool)
	visited[sourceID] = true

	var dfs func(currID string, path []*HeteroEdge, depth int)
	dfs = func(currID string, path []*HeteroEdge, depth int) {
		if depth > maxDepth {
			return
		}
		if targetID != "" && currID == targetID && len(path) > 0 {
			discovered = append(discovered, p.formatPath(sourceID, path, nodes))
			return
		} else if targetID == "" && len(path) > 0 {
			if n, ok := nodes[currID]; ok && (n.Type == NodeConsumer || n.Type == NodeAccount) && currID != sourceID {
				discovered = append(discovered, p.formatPath(sourceID, path, nodes))
			}
		}

		for _, e := range adj[currID] {
			neighbor := e.TargetID
			if e.TargetID == currID {
				neighbor = e.SourceID
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				dfs(neighbor, append(path, e), depth+1)
				visited[neighbor] = false
			}
		}
	}

	dfs(sourceID, nil, 0)
	if len(discovered) > 10 {
		return discovered[:10]
	}
	return discovered
}

func (p *PathExtractor) formatPath(sourceID string, path []*HeteroEdge, nodes map[string]*HeteroNode) string {
	res := fmt.Sprintf("(%s)", sourceID)
	for _, e := range path {
		target := e.TargetID
		if target == sourceID {
			target = e.SourceID
		}
		res += fmt.Sprintf(" -[%s]-> (%s)", e.Type, target)
	}
	return res
}
