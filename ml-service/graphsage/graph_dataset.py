"""
Temporal Heterogeneous Graph Dataset & High-Performance Point-in-Time Sampler
"""

import os
import csv
import glob
import bisect
import numpy as np
from datetime import datetime, timezone
from typing import Dict, List, Tuple, Optional, Set, Any
from .schema import HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType


class TemporalHeteroGraphDataset:
    """
    Manages in-memory temporal heterogeneous graph with strict point-in-time edge filtering.
    Optimized with binary-search sorted incidence arrays for microsecond neighbor sampling.
    """
    def __init__(self):
        self.nodes: Dict[str, HeteroNode] = {}
        self.edges: Dict[str, HeteroEdge] = {}
        # node_id -> sorted list of (timestamp_dt, neighbor_id, edge_type)
        self.incident_adj: Dict[str, List[Tuple[datetime, str, str]]] = {}
        self.out_edges: Dict[str, List[str]] = {}
        self.in_edges: Dict[str, List[str]] = {}

    def add_node(self, node: HeteroNode):
        self.nodes[node.id] = node

    def add_edge(self, edge: HeteroEdge):
        self.edges[edge.id] = edge
        if edge.source_id not in self.out_edges:
            self.out_edges[edge.source_id] = []
        self.out_edges[edge.source_id].append(edge.id)
        if edge.target_id not in self.in_edges:
            self.in_edges[edge.target_id] = []
        self.in_edges[edge.target_id].append(edge.id)

        etype_str = str(edge.type.value) if hasattr(edge.type, "value") else str(edge.type)
        for u, v in ((edge.source_id, edge.target_id), (edge.target_id, edge.source_id)):
            if u not in self.incident_adj:
                self.incident_adj[u] = []
            self.incident_adj[u].append((edge.timestamp, v, etype_str))

    def finalize_indices(self):
        """Sorts incident edges by timestamp for O(log N) point-in-time slicing."""
        for nid in self.incident_adj:
            self.incident_adj[nid].sort(key=lambda x: x[0])

    def get_point_in_time_neighbors(
        self,
        node_id: str,
        as_of: datetime,
        sample_sizes: Tuple[int, int] = (10, 5),
        seed: int = 42
    ) -> Tuple[List[HeteroNode], List[HeteroNode]]:
        """
        Extracts 2-hop sampled neighbors strictly where edge.timestamp <= as_of.
        Guarantees ZERO future data leakage.
        """
        rng = np.random.RandomState(seed)
        l1_nodes: List[HeteroNode] = []
        l2_nodes: List[HeteroNode] = []
        visited: Set[str] = {node_id}

        # 1. Hop 1 Neighbors
        valid_l1_ids = self._get_valid_neighbor_ids(node_id, as_of)
        if valid_l1_ids:
            if len(valid_l1_ids) > sample_sizes[0]:
                idx = rng.choice(len(valid_l1_ids), size=sample_sizes[0], replace=False)
                sampled_ids = [valid_l1_ids[i] for i in idx]
            else:
                sampled_ids = valid_l1_ids

            for nid in sampled_ids:
                if nid not in visited and nid in self.nodes:
                    visited.add(nid)
                    l1_nodes.append(self.nodes[nid])

        # 2. Hop 2 Neighbors
        l1_ids = [n.id for n in l1_nodes]
        for l1_id in l1_ids:
            valid_l2_ids = self._get_valid_neighbor_ids(l1_id, as_of)
            if valid_l2_ids:
                if len(valid_l2_ids) > sample_sizes[1]:
                    idx = rng.choice(len(valid_l2_ids), size=sample_sizes[1], replace=False)
                    sampled_l2_ids = [valid_l2_ids[i] for i in idx]
                else:
                    sampled_l2_ids = valid_l2_ids

                for nid in sampled_l2_ids:
                    if nid not in visited and nid in self.nodes:
                        visited.add(nid)
                        l2_nodes.append(self.nodes[nid])

        return l1_nodes, l2_nodes

    def _get_valid_neighbor_ids(self, node_id: str, as_of: datetime) -> List[str]:
        if node_id not in self.incident_adj:
            return []
        adj = self.incident_adj[node_id]
        if not adj:
            return []
        # Binary search for timestamp cutoff
        cutoff = bisect.bisect_right(adj, (as_of, chr(255), chr(255)))
        return [nid for _, nid, _ in adj[:cutoff]]

    @classmethod
    def load_from_directory(cls, data_dir: str) -> "TemporalHeteroGraphDataset":
        """
        Fast loader reading all CSV tables directly via Python csv.DictReader.
        """
        ds = cls()
        print(f"[GraphSAGE Dataset] Loading nodes from {os.path.join(data_dir, 'nodes')}", flush=True)

        def read_csv_fast(path):
            with open(path, newline="", encoding="utf-8") as f:
                reader = csv.DictReader(f)
                return list(reader)

        # 1. Employees
        emp_file = os.path.join(data_dir, "nodes", "employee.csv")
        if os.path.exists(emp_file):
            for r in read_csv_fast(emp_file):
                ds.add_node(HeteroNode(
                    id=r["employee_id"],
                    type=HeteroNodeType.EMPLOYEE,
                    risk_score=0.05,
                    created_at=datetime(2025, 1, 1, tzinfo=timezone.utc),
                    properties={"role": r.get("role", "unknown")}
                ))

        # 2. Consumers
        cons_file = os.path.join(data_dir, "nodes", "consumer.csv")
        if os.path.exists(cons_file):
            for r in read_csv_fast(cons_file):
                ds.add_node(HeteroNode(
                    id=r["consumer_id"],
                    type=HeteroNodeType.CONSUMER,
                    risk_score=0.05,
                    created_at=datetime(2025, 1, 1, tzinfo=timezone.utc)
                ))

        # 3. Accounts
        acct_file = os.path.join(data_dir, "nodes", "account.csv")
        if os.path.exists(acct_file):
            for r in read_csv_fast(acct_file):
                ds.add_node(HeteroNode(
                    id=r["account_id"],
                    type=HeteroNodeType.ACCOUNT,
                    risk_score=0.05,
                    created_at=datetime(2025, 1, 1, tzinfo=timezone.utc),
                    properties={"consumer_id": r.get("consumer_id", "")}
                ))

        # 4. Devices
        dev_file = os.path.join(data_dir, "nodes", "device.csv")
        if os.path.exists(dev_file):
            for r in read_csv_fast(dev_file):
                ds.add_node(HeteroNode(
                    id=r["device_id"],
                    type=HeteroNodeType.DEVICE,
                    risk_score=0.05,
                    created_at=datetime(2025, 1, 1, tzinfo=timezone.utc),
                    properties={"device_type": r.get("device_type", "unknown")}
                ))

        # 5. IPs
        ip_file = os.path.join(data_dir, "nodes", "ip.csv")
        if os.path.exists(ip_file):
            for r in read_csv_fast(ip_file):
                ds.add_node(HeteroNode(
                    id=r["ip_id"],
                    type=HeteroNodeType.IP,
                    risk_score=0.05,
                    created_at=datetime(2025, 1, 1, tzinfo=timezone.utc),
                    properties={"ip_type": r.get("ip_type", "unknown")}
                ))

        # 6. Payment Tokens
        tok_file = os.path.join(data_dir, "nodes", "payment_token.csv")
        if os.path.exists(tok_file):
            for r in read_csv_fast(tok_file):
                ds.add_node(HeteroNode(
                    id=r["payment_token_id"],
                    type=HeteroNodeType.PAYMENT_TOKEN,
                    risk_score=0.05,
                    created_at=datetime(2025, 1, 1, tzinfo=timezone.utc)
                ))

        # 7. Merchants
        merch_file = os.path.join(data_dir, "nodes", "merchant.csv")
        if os.path.exists(merch_file):
            for r in read_csv_fast(merch_file):
                ds.add_node(HeteroNode(
                    id=r["merchant_id"],
                    type=HeteroNodeType.MERCHANT,
                    risk_score=0.02,
                    created_at=datetime(2025, 1, 1, tzinfo=timezone.utc),
                    properties={"category": r.get("category", "unknown")}
                ))

        # 8. Cases
        case_file = os.path.join(data_dir, "nodes", "case.csv")
        if os.path.exists(case_file):
            for r in read_csv_fast(case_file):
                open_ts = datetime.fromisoformat(r["opened_at"])
                if open_ts.tzinfo is None:
                    open_ts = open_ts.replace(tzinfo=timezone.utc)
                ds.add_node(HeteroNode(
                    id=r["case_id"],
                    type=HeteroNodeType.CASE,
                    risk_score=0.20,
                    created_at=open_ts,
                    properties={"status": r.get("status", "unknown")}
                ))

        print(f"[GraphSAGE Dataset] Loaded {len(ds.nodes)} nodes. Loading edges...", flush=True)

        # 9. Edges
        edge_files = glob.glob(os.path.join(data_dir, "edges", "*.csv"))
        eid_counter = 0
        default_ts = datetime(2025, 1, 1, tzinfo=timezone.utc)

        for ef in sorted(edge_files):
            etype_str = os.path.basename(ef).replace(".csv", "").upper()
            etype = getattr(HeteroEdgeType, etype_str, HeteroEdgeType.ACCESSES)

            with open(ef, newline="", encoding="utf-8") as f:
                reader = csv.reader(f)
                header = next(reader, None)
                if not header or len(header) < 2:
                    continue
                has_ts = "edge_timestamp" in header
                ts_idx = header.index("edge_timestamp") if has_ts else -1

                for row in reader:
                    if len(row) < 2:
                        continue
                    src_id = row[0]
                    dst_id = row[1]
                    if has_ts and ts_idx < len(row) and row[ts_idx]:
                        ts = datetime.fromisoformat(row[ts_idx])
                        if ts.tzinfo is None:
                            ts = ts.replace(tzinfo=timezone.utc)
                    else:
                        ts = default_ts

                    eid = f"e_{eid_counter}"
                    eid_counter += 1
                    ds.add_edge(HeteroEdge(
                        id=eid,
                        source_id=src_id,
                        target_id=dst_id,
                        type=etype,
                        timestamp=ts
                    ))

        ds.finalize_indices()
        print(f"[GraphSAGE Dataset] Loaded {len(ds.edges)} edges successfully.", flush=True)
        return ds
