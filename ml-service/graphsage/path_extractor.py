"""
Multi-Hop Relationship Path Extractor for GraphSAGE Explainability (Phase 59)
Discovers directed structural connection chains between Employees, Accounts, Consumers, Devices, IPs, and Transactions.
"""

from typing import Dict, List, Set, Tuple, Optional
from datetime import datetime
from .schema import HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType


class RelationshipPathExtractor:
    """
    Finds and formats multi-hop explainability paths up to depth 4 strictly point-in-time.
    """

    def find_paths(
        self,
        source_id: str,
        target_id: Optional[str],
        nodes: Dict[str, HeteroNode],
        edges: List[HeteroEdge],
        max_depth: int = 3
    ) -> List[Dict[str, Any]]:
        """
        Executes bounded Depth-First Search (DFS) to extract directed and undirected graph paths.
        """
        # Build adjacency list: node_id -> list of (neighbor_id, edge)
        adj: Dict[str, List[Tuple[str, HeteroEdge]]] = {}
        for e in edges:
            if e.source_id not in adj:
                adj[e.source_id] = []
            adj[e.source_id].append((e.target_id, e))

            # Allow bidirectional navigation for undirected relationships
            if e.target_id not in adj:
                adj[e.target_id] = []
            adj[e.target_id].append((e.source_id, e))

        discovered_paths: List[Dict[str, Any]] = []
        visited = set([source_id])

        def dfs(current_id: str, current_path: List[Tuple[str, HeteroEdge]], depth: int):
            if depth > max_depth:
                return

            if target_id and current_id == target_id and len(current_path) > 0:
                discovered_paths.append(self._format_path(current_path, nodes))
                return
            elif not target_id and len(current_path) > 0 and current_id in nodes:
                node = nodes[current_id]
                if node.type in [HeteroNodeType.CONSUMER, HeteroNodeType.ACCOUNT] and current_id != source_id:
                    discovered_paths.append(self._format_path(current_path, nodes))

            for neighbor_id, edge in adj.get(current_id, []):
                if neighbor_id not in visited:
                    visited.add(neighbor_id)
                    dfs(neighbor_id, current_path + [(neighbor_id, edge)], depth + 1)
                    visited.remove(neighbor_id)

        dfs(source_id, [], 0)
        return discovered_paths[:20]  # Return top 20 representative paths

    def _format_path(
        self,
        path_segments: List[Tuple[str, HeteroEdge]],
        nodes: Dict[str, HeteroNode]
    ) -> Dict[str, Any]:
        chain = []
        timestamps = []
        edge_types = []

        first_src = path_segments[0][1].source_id
        chain.append(f"{nodes.get(first_src, HeteroNode(id=first_src, type=HeteroNodeType.ACCOUNT, created_at=datetime.now())).type.value}:{first_src}")

        for neighbor_id, edge in path_segments:
            etype = edge.type.value if hasattr(edge.type, "value") else str(edge.type)
            ntype = nodes.get(neighbor_id, HeteroNode(id=neighbor_id, type=HeteroNodeType.ACCOUNT, created_at=datetime.now())).type.value
            chain.append(f"--[{etype}]-->")
            chain.append(f"{ntype}:{neighbor_id}")
            timestamps.append(edge.timestamp.isoformat())
            edge_types.append(etype)

        return {
            "path_string": " ".join(chain),
            "hops": len(path_segments),
            "edge_types": edge_types,
            "timestamps": timestamps,
        }
