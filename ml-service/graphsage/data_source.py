"""
Real-Data Ingestion Adapter Interface, Offline Fixture Adapter, and Point-in-Time Replay Harness (Phase 58)
"""

import os
import csv
from abc import ABC, abstractmethod
from datetime import datetime, timezone
from typing import Dict, List, Optional, Any, Iterator, Tuple

from .schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    LabelMaturityState, DatasetType
)


class RealGraphDataSource(ABC):
    """
    Abstract Ingestion Adapter Interface for Real/Shadow Graph Telemetry.
    Decouples GraphSAGE from concrete storage backends (Kafka, Snowflake, Postgres, Neo4j).
    """

    @abstractmethod
    def get_dataset_type(self) -> DatasetType:
        """Returns the formal dataset classification."""
        pass

    @abstractmethod
    def get_nodes(
        self,
        node_types: Optional[List[HeteroNodeType]] = None,
        as_of: Optional[datetime] = None
    ) -> List[HeteroNode]:
        """Retrieves nodes created as of timestamp T."""
        pass

    @abstractmethod
    def get_edges(
        self,
        edge_types: Optional[List[HeteroEdgeType]] = None,
        as_of: Optional[datetime] = None
    ) -> List[HeteroEdge]:
        """Retrieves directed relationships created as of timestamp T."""
        pass

    @abstractmethod
    def get_transactions(
        self,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None
    ) -> List[Dict[str, Any]]:
        """Retrieves transaction events within the temporal window."""
        pass

    @abstractmethod
    def get_labels(
        self,
        as_of: Optional[datetime] = None,
        mature_only: bool = True
    ) -> List[Dict[str, Any]]:
        """Retrieves labels mature as of timestamp T."""
        pass

    @abstractmethod
    def get_investigation_events(
        self,
        as_of: Optional[datetime] = None
    ) -> List[Dict[str, Any]]:
        """Retrieves case and investigation events as of timestamp T."""
        pass


class MockRealGraphDataSource(RealGraphDataSource):
    """
    Offline/Mock Data Source Adapter.
    Consumes point-in-time test fixtures safely without network calls or production access.
    """

    def __init__(
        self,
        nodes: Optional[List[HeteroNode]] = None,
        edges: Optional[List[HeteroEdge]] = None,
        transactions: Optional[List[Dict[str, Any]]] = None,
        labels: Optional[List[Dict[str, Any]]] = None,
        dataset_type: DatasetType = DatasetType.REAL_SHADOW
    ):
        self.nodes_list = nodes or []
        self.edges_list = edges or []
        self.txns_list = transactions or []
        self.labels_list = labels or []
        self.dataset_type = dataset_type

    def add_node(self, n: HeteroNode):
        self.nodes_list.append(n)

    def add_edge(self, e: HeteroEdge):
        self.edges_list.append(e)

    def add_label(self, l: Dict[str, Any]):
        self.labels_list.append(l)

    def add_transaction(self, t: Dict[str, Any]):
        self.txns_list.append(t)

    # Aliases for compatibility
    AddNode = add_node
    AddEdge = add_edge
    AddLabel = add_label
    AddTransaction = add_transaction

    @property
    def nodes(self) -> Dict[str, HeteroNode]:
        return {n.id: n for n in self.nodes_list}

    @property
    def edges(self) -> List[HeteroEdge]:
        return self.edges_list

    def get_point_in_time_subgraph(
        self,
        center_node_id: str,
        as_of: datetime,
        depth: int = 2
    ):
        from .schema import HeteroSubGraph
        valid_edges = [e for e in self.edges_list if e.timestamp <= as_of]
        visited_nodes: Dict[str, HeteroNode] = {}
        all_nodes = {n.id: n for n in self.nodes_list}

        if center_node_id in all_nodes:
            visited_nodes[center_node_id] = all_nodes[center_node_id]
        else:
            ntype = HeteroNodeType.EMPLOYEE if center_node_id.startswith("emp_") else (
                HeteroNodeType.ACCOUNT if center_node_id.startswith("acct_") else HeteroNodeType.CONSUMER
            )
            visited_nodes[center_node_id] = HeteroNode(
                id=center_node_id,
                type=ntype,
                created_at=as_of
            )

        curr_frontier = {center_node_id}
        subgraph_edges: List[HeteroEdge] = []

        for _ in range(depth):
            next_frontier = set()
            for e in valid_edges:
                if e.source_id in curr_frontier and e.target_id not in visited_nodes:
                    if e.target_id in all_nodes:
                        visited_nodes[e.target_id] = all_nodes[e.target_id]
                    else:
                        ntype = HeteroNodeType.ACCOUNT if e.target_id.startswith("acct_") else HeteroNodeType.DEVICE
                        visited_nodes[e.target_id] = HeteroNode(id=e.target_id, type=ntype, created_at=e.timestamp)
                    next_frontier.add(e.target_id)
                    subgraph_edges.append(e)
                elif e.target_id in curr_frontier and e.source_id not in visited_nodes:
                    if e.source_id in all_nodes:
                        visited_nodes[e.source_id] = all_nodes[e.source_id]
                    else:
                        ntype = HeteroNodeType.EMPLOYEE if e.source_id.startswith("emp_") else HeteroNodeType.CONSUMER
                        visited_nodes[e.source_id] = HeteroNode(id=e.source_id, type=ntype, created_at=e.timestamp)
                    next_frontier.add(e.source_id)
                    subgraph_edges.append(e)
                elif e.source_id in curr_frontier or e.target_id in curr_frontier:
                    if e not in subgraph_edges:
                        subgraph_edges.append(e)
            curr_frontier = next_frontier

        return HeteroSubGraph(
            center_node_id=center_node_id,
            as_of=as_of,
            nodes=visited_nodes,
            edges=subgraph_edges
        )

    def get_dataset_type(self) -> DatasetType:
        return self.dataset_type

    def get_nodes(
        self,
        node_types: Optional[List[HeteroNodeType]] = None,
        as_of: Optional[datetime] = None
    ) -> List[HeteroNode]:
        res = []
        for n in self.nodes_list:
            if node_types and n.type not in node_types:
                continue
            if as_of and n.created_at > as_of:
                continue
            res.append(n)
        return res

    def get_edges(
        self,
        edge_types: Optional[List[HeteroEdgeType]] = None,
        as_of: Optional[datetime] = None
    ) -> List[HeteroEdge]:
        res = []
        for e in self.edges_list:
            if edge_types and e.type not in edge_types:
                continue
            if as_of and e.timestamp > as_of:
                continue
            res.append(e)
        return res

    def get_transactions(
        self,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None
    ) -> List[Dict[str, Any]]:
        res = []
        for t in self.txns_list:
            ts = t.get("event_timestamp")
            if isinstance(ts, str):
                ts = datetime.fromisoformat(ts)
            if start_time and ts and ts < start_time:
                continue
            if end_time and ts and ts > end_time:
                continue
            res.append(t)
        return res

    def get_labels(
        self,
        as_of: Optional[datetime] = None,
        mature_only: bool = True
    ) -> List[Dict[str, Any]]:
        res = []
        for l in self.labels_list:
            ts = l.get("label_timestamp")
            if isinstance(ts, str):
                ts = datetime.fromisoformat(ts)
            if as_of and ts and ts > as_of:
                continue
            if mature_only and l.get("maturity_state") not in [
                LabelMaturityState.CONFIRMED_FRAUD.value,
                LabelMaturityState.CONFIRMED_LEGITIMATE.value
            ]:
                continue
            res.append(l)
        return res

    def get_investigation_events(
        self,
        as_of: Optional[datetime] = None
    ) -> List[Dict[str, Any]]:
        return []


class PointInTimeReplayHarness:
    """
    Point-in-Time Replay Harness for historical shadow evaluation.
    Guarantees:
    - Zero production BMR decision changes
    - Zero payment gateway mutation
    - Zero future edge lookahead (edge.timestamp <= T)
    - Zero immature label exposure (label.timestamp <= T)
    """

    def __init__(self, data_source: RealGraphDataSource):
        self.data_source = data_source

    def replay_stream(
        self,
        start_time: datetime,
        end_time: datetime
    ) -> Iterator[Tuple[Dict[str, Any], List[HeteroNode], List[HeteroEdge]]]:
        """
        Streams transactions chronologically, yielding the point-in-time neighborhood
        strictly available at each transaction's timestamp T.
        """
        txns = self.data_source.get_transactions(start_time=start_time, end_time=end_time)
        txns_sorted = sorted(txns, key=lambda x: x["event_timestamp"])

        for txn in txns_sorted:
            t_eval = txn["event_timestamp"]
            if isinstance(t_eval, str):
                t_eval = datetime.fromisoformat(t_eval)
            if t_eval.tzinfo is None:
                t_eval = t_eval.replace(tzinfo=timezone.utc)

            # Strict point-in-time slice
            nodes_as_of = self.data_source.get_nodes(as_of=t_eval)
            edges_as_of = self.data_source.get_edges(as_of=t_eval)

            yield txn, nodes_as_of, edges_as_of
