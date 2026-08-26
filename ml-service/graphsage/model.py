"""
GraphSAGE NumPy Neural Model Implementation
Supports 2-Layer Neighborhood Aggregation, Inductive Projection, Backpropagation, and Multi-Head Risk Estimation
"""

import numpy as np
from typing import List, Dict, Tuple, Optional, Any
from .schema import HeteroNode, HeteroNodeType


class GraphSAGEModel:
    """
    Inductive 2-Layer GraphSAGE Model using NumPy for fast, dependency-free serving & training.
    """
    def __init__(
        self,
        input_dim: int = 32,
        hidden_dim: int = 64,
        output_dim: int = 64,
        seed: int = 42
    ):
        self.input_dim = input_dim
        self.hidden_dim = hidden_dim
        self.output_dim = output_dim
        self.seed = seed

        rng = np.random.RandomState(seed)

        # Layer 1 Projection Weights [HiddenDim x (InputDim * 2)]
        scale1 = np.sqrt(2.0 / (input_dim * 2 + hidden_dim))
        self.W1 = rng.randn(hidden_dim, input_dim * 2) * scale1
        self.b1 = np.zeros(hidden_dim, dtype=np.float64)

        # Layer 2 Projection Weights [OutputDim x (HiddenDim * 2)]
        scale2 = np.sqrt(2.0 / (hidden_dim * 2 + output_dim))
        self.W2 = rng.randn(output_dim, hidden_dim * 2) * scale2
        self.b2 = np.zeros(output_dim, dtype=np.float64)

        # Multi-Head Prediction Weights
        self.w_employee = rng.randn(output_dim) * 0.08
        self.w_consumer = rng.randn(output_dim) * 0.08
        self.w_relationship = rng.randn(output_dim) * 0.12
        self.w_neighborhood = rng.randn(output_dim) * 0.10
        self.w_anomaly = rng.randn(output_dim) * 0.09
        self.w_combined = rng.randn(output_dim) * 0.15
        self.bias = -2.5

        # Adam optimizer state
        self._m = {}
        self._v = {}
        self._t = 0

    def compute_node_features(self, node: Optional[HeteroNode]) -> np.ndarray:
        """
        Computes 32-dim feature vector for a node.
        STRICT LEAKAGE AUDIT:
        - NEVER uses node.id numeric magnitude
        - NEVER uses scenario tags or labels
        - Purely uses node type, prior risk score, and domain properties
        """
        vec = np.zeros(self.input_dim, dtype=np.float64)
        if node is None:
            return vec

        # 1. One-hot node type encoding (dims 0..9)
        type_idx_map = {
            HeteroNodeType.EMPLOYEE: 0,
            HeteroNodeType.CONSUMER: 1,
            HeteroNodeType.ACCOUNT: 2,
            HeteroNodeType.DEVICE: 3,
            HeteroNodeType.IP: 4,
            HeteroNodeType.PAYMENT_TOKEN: 5,
            HeteroNodeType.TRANSACTION: 6,
            HeteroNodeType.MERCHANT: 7,
            HeteroNodeType.CASE: 8,
            HeteroNodeType.SESSION: 9,
        }
        if node.type in type_idx_map:
            vec[type_idx_map[node.type]] = 1.0

        # 2. Risk prior and flags (dims 10..11)
        vec[10] = float(node.risk_score)
        if node.is_known_bad:
            vec[11] = 1.0

        # 3. Dynamic domain features (dims 12..31)
        if node.features:
            for i, val in enumerate(node.features):
                if 12 + i < self.input_dim:
                    vec[12 + i] = float(val)
        elif node.properties:
            # Categorical property encodings
            if "role" in node.properties:
                roles = ["customer_support", "fraud_analyst", "risk_operations", "payments_operations",
                         "merchant_support", "engineering", "finance", "compliance", "administrator"]
                role = str(node.properties["role"])
                if role in roles:
                    vec[12 + roles.index(role)] = 1.0
            if "device_type" in node.properties:
                devs = ["mobile", "desktop", "corporate_workstation", "tablet", "customer_service_terminal"]
                dtype = str(node.properties["device_type"])
                if dtype in devs:
                    vec[21 + devs.index(dtype)] = 1.0
            if "ip_type" in node.properties:
                iptypes = ["residential", "corporate_nat", "mobile_carrier", "datacenter_proxy", "unknown"]
                itype = str(node.properties["ip_type"])
                if itype in iptypes:
                    vec[26 + iptypes.index(itype)] = 1.0

        return vec

    def forward(
        self,
        root_node: HeteroNode,
        l1_neighbors: List[HeteroNode],
        l2_neighbors: List[HeteroNode]
    ) -> Tuple[np.ndarray, Dict[str, float]]:
        """
        2-Layer GraphSAGE forward pass.
        Returns:
            embedding (64-dim float vector)
            signals (dict of calibrated multi-head risk scores)
        """
        # 1. Input representation
        h0_root = self.compute_node_features(root_node)

        # 2. Layer 1 neighborhood aggregation
        if l1_neighbors:
            l1_feats = np.array([self.compute_node_features(n) for n in l1_neighbors], dtype=np.float64)
            agg_l1 = np.mean(l1_feats, axis=0)
        else:
            agg_l1 = np.zeros(self.input_dim, dtype=np.float64)

        concat_l1 = np.concatenate([h0_root, agg_l1])
        h1_root = np.maximum(0, self._layer_norm(np.dot(self.W1, concat_l1) + self.b1))

        # 3. Layer 2 neighborhood aggregation
        if l2_neighbors:
            l2_feats = np.array([self.compute_node_features(n) for n in l2_neighbors], dtype=np.float64)
            agg_l2 = np.mean(l2_feats, axis=0)
        else:
            agg_l2 = np.zeros(self.input_dim, dtype=np.float64)

        concat_l2_agg = np.concatenate([agg_l1, agg_l2])
        h1_l1 = np.maximum(0, self._layer_norm(np.dot(self.W1, concat_l2_agg) + self.b1))

        concat_l2 = np.concatenate([h1_root, h1_l1])
        h2_root = self._layer_norm(np.dot(self.W2, concat_l2) + self.b2)

        # 4. Multi-Head Risk Estimation
        bias = self.bias
        emp_risk = float(self._sigmoid(np.dot(h2_root, self.w_employee) + root_node.risk_score * 0.2 + bias))
        con_risk = float(self._sigmoid(np.dot(h2_root, self.w_consumer) + root_node.risk_score * 0.2 + bias))
        rel_risk = float(self._sigmoid(np.dot(h2_root, self.w_relationship) + bias))
        nbr_risk = float(self._sigmoid(np.dot(h2_root, self.w_neighborhood) + bias))
        anom_risk = float(self._sigmoid(np.dot(h2_root, self.w_anomaly) + bias))
        comb_risk = float(self._sigmoid(np.dot(h2_root, self.w_combined) + bias))

        ctx_score = float(np.clip(comb_risk * 0.4 + rel_risk * 0.3 + nbr_risk * 0.2 + anom_risk * 0.1, 0.0, 1.0))

        signals = {
            "employee_risk": emp_risk,
            "consumer_risk": con_risk,
            "relationship_collusion_risk": rel_risk,
            "entity_neighborhood_risk": nbr_risk,
            "graph_anomaly_score": anom_risk,
            "transaction_context_score": ctx_score,
        }

        return h2_root, signals

    def train_step(
        self,
        batch_samples: List[Tuple[HeteroNode, List[HeteroNode], List[HeteroNode], float]],
        lr: float = 0.005,
        weight_decay: float = 1e-4
    ) -> float:
        """
        Executes one mini-batch supervised gradient update using Adam.
        batch_samples: list of (root_node, l1_neighbors, l2_neighbors, target_y)
        """
        if not batch_samples:
            return 0.0

        dW1 = np.zeros_like(self.W1)
        db1 = np.zeros_like(self.b1)
        dW2 = np.zeros_like(self.W2)
        db2 = np.zeros_like(self.b2)
        dw_comb = np.zeros_like(self.w_combined)
        dw_rel = np.zeros_like(self.w_relationship)
        dw_nbr = np.zeros_like(self.w_neighborhood)
        dbias = 0.0
        total_loss = 0.0

        for root_node, l1_nodes, l2_nodes, target_y in batch_samples:
            h2, signals = self.forward(root_node, l1_nodes, l2_nodes)
            pred = signals["transaction_context_score"]
            eps = 1e-7
            loss = -(target_y * np.log(pred + eps) + (1.0 - target_y) * np.log(1.0 - pred + eps))
            total_loss += loss

            # Error gradient wrt sigmoid logit
            dlogit = (pred - target_y)
            dw_comb += dlogit * h2 * 0.4
            dw_rel += dlogit * h2 * 0.3
            dw_nbr += dlogit * h2 * 0.2
            dbias += dlogit

            # Backprop into W2
            dh2 = dlogit * (self.w_combined * 0.4 + self.w_relationship * 0.3 + self.w_neighborhood * 0.2)

            # Input to W2
            h0_root = self.compute_node_features(root_node)
            l1_feats = np.array([self.compute_node_features(n) for n in l1_nodes], dtype=np.float64) if l1_nodes else np.zeros((1, self.input_dim))
            agg_l1 = np.mean(l1_feats, axis=0) if l1_nodes else np.zeros(self.input_dim)
            concat_l1 = np.concatenate([h0_root, agg_l1])
            h1_root = np.maximum(0, self._layer_norm(np.dot(self.W1, concat_l1) + self.b1))

            l2_feats = np.array([self.compute_node_features(n) for n in l2_nodes], dtype=np.float64) if l2_nodes else np.zeros((1, self.input_dim))
            agg_l2 = np.mean(l2_feats, axis=0) if l2_nodes else np.zeros(self.input_dim)
            concat_l2_agg = np.concatenate([agg_l1, agg_l2])
            h1_l1 = np.maximum(0, self._layer_norm(np.dot(self.W1, concat_l2_agg) + self.b1))
            concat_l2 = np.concatenate([h1_root, h1_l1])

            dW2 += np.outer(dh2, concat_l2)
            db2 += dh2

            # Backprop into W1
            dconcat_l2 = np.dot(self.W2.T, dh2)
            dh1_root = dconcat_l2[:self.hidden_dim] * (h1_root > 0)
            dh1_l1 = dconcat_l2[self.hidden_dim:] * (h1_l1 > 0)

            dW1 += np.outer(dh1_root, concat_l1) + np.outer(dh1_l1, concat_l2_agg)
            db1 += (dh1_root + dh1_l1)

        n = len(batch_samples)
        dW1 = dW1 / n + weight_decay * self.W1
        db1 = db1 / n
        dW2 = dW2 / n + weight_decay * self.W2
        db2 = db2 / n
        dw_comb = dw_comb / n + weight_decay * self.w_combined
        dw_rel = dw_rel / n + weight_decay * self.w_relationship
        dw_nbr = dw_nbr / n + weight_decay * self.w_neighborhood
        dbias = dbias / n

        # Adam updates
        self._t += 1
        params = [
            ("W1", self.W1, dW1), ("b1", self.b1, db1),
            ("W2", self.W2, dW2), ("b2", self.b2, db2),
            ("w_comb", self.w_combined, dw_comb),
            ("w_rel", self.w_relationship, dw_rel),
            ("w_nbr", self.w_neighborhood, dw_nbr)
        ]
        beta1, beta2, eps = 0.9, 0.999, 1e-8
        for name, param, grad in params:
            if name not in self._m:
                self._m[name] = np.zeros_like(param)
                self._v[name] = np.zeros_like(param)
            self._m[name] = beta1 * self._m[name] + (1.0 - beta1) * grad
            self._v[name] = beta2 * self._v[name] + (1.0 - beta2) * (grad ** 2)
            m_hat = self._m[name] / (1.0 - beta1 ** self._t)
            v_hat = self._v[name] / (1.0 - beta2 ** self._t)
            param -= lr * m_hat / (np.sqrt(v_hat) + eps)

        self.bias -= lr * dbias
        return total_loss / n

    def to_dict(self) -> Dict[str, Any]:
        """Serializes weights to dictionary for persistence."""
        return {
            "input_dim": self.input_dim,
            "hidden_dim": self.hidden_dim,
            "output_dim": self.output_dim,
            "seed": self.seed,
            "W1": self.W1.tolist(),
            "b1": self.b1.tolist(),
            "W2": self.W2.tolist(),
            "b2": self.b2.tolist(),
            "w_employee": self.w_employee.tolist(),
            "w_consumer": self.w_consumer.tolist(),
            "w_relationship": self.w_relationship.tolist(),
            "w_neighborhood": self.w_neighborhood.tolist(),
            "w_anomaly": self.w_anomaly.tolist(),
            "w_combined": self.w_combined.tolist(),
            "bias": float(self.bias),
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "GraphSAGEModel":
        """Loads model from serialized dictionary."""
        model = cls(
            input_dim=data.get("input_dim", 32),
            hidden_dim=data.get("hidden_dim", 64),
            output_dim=data.get("output_dim", 64),
            seed=data.get("seed", 42)
        )
        model.W1 = np.array(data["W1"], dtype=np.float64)
        model.b1 = np.array(data["b1"], dtype=np.float64)
        model.W2 = np.array(data["W2"], dtype=np.float64)
        model.b2 = np.array(data["b2"], dtype=np.float64)
        model.w_employee = np.array(data["w_employee"], dtype=np.float64)
        model.w_consumer = np.array(data["w_consumer"], dtype=np.float64)
        model.w_relationship = np.array(data["w_relationship"], dtype=np.float64)
        model.w_neighborhood = np.array(data["w_neighborhood"], dtype=np.float64)
        model.w_anomaly = np.array(data["w_anomaly"], dtype=np.float64)
        model.w_combined = np.array(data["w_combined"], dtype=np.float64)
        model.bias = float(data.get("bias", -2.5))
        return model

    def _layer_norm(self, x: np.ndarray) -> np.ndarray:
        mean = np.mean(x)
        std = np.std(x) + 1e-6
        return (x - mean) / std

    def _sigmoid(self, x: float) -> float:
        if x > 15.0:
            return 0.9999
        if x < -15.0:
            return 0.0001
        return 1.0 / (1.0 + np.exp(-x))
