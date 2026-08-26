"""
Automated Pytest Suite: Graph Topology, Components, and Edge Validity
"""
import os
import glob
import pandas as pd
import networkx as nx
import pytest

DATA_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "data"))

@pytest.fixture(scope="module")
def loaded_graph():
    G = nx.Graph()

    # Load nodes
    for nf in glob.glob(os.path.join(DATA_DIR, "nodes", "*.csv")):
        ntype = os.path.basename(nf).replace(".csv", "")
        df = pd.read_csv(nf)
        id_col = df.columns[0]
        for nid in df[id_col].astype(str):
            G.add_node(nid, node_type=ntype)

    # Load edges
    for ef in glob.glob(os.path.join(DATA_DIR, "edges", "*.csv")):
        etype = os.path.basename(ef).replace(".csv", "")
        df = pd.read_csv(ef)
        src_col = df.columns[0]
        dst_col = df.columns[1]
        for _, r in df.iterrows():
            G.add_edge(str(r[src_col]), str(r[dst_col]), edge_type=etype)

    return G

def test_giant_component_size(loaded_graph):
    G = loaded_graph
    assert G.number_of_nodes() > 50000, f"Graph too small: {G.number_of_nodes()} nodes"

    components = list(nx.connected_components(G))
    largest_comp = max(len(c) for c in components)
    pct = largest_comp / G.number_of_nodes()

    assert pct >= 0.90, f"Largest connected component is only {pct*100:.2f}% of nodes (expected >= 90%)"

def test_flood_attack_node_degree(loaded_graph):
    G = loaded_graph
    tags_path = os.path.join(DATA_DIR, "labels", "scenario_tags.csv")
    tags = pd.read_csv(tags_path)

    flood_tags = tags[tags["scenario_type"] == "graph_poisoning_flood"]
    assert len(flood_tags) > 0, "No graph poisoning flood tags found"

    for fd in flood_tags["entity_id"]:
        deg = G.degree(str(fd))
        assert deg >= 100, f"Flood node {fd} has degree {deg} (expected >= 100)"
