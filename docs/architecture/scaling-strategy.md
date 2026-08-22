# ROPUS Scale & Capacity Engineering Strategy

Plan to scale ROPUS to **10 Million Customers**, **1 Billion Transactions**, and **100k+ Events/sec**:

---

## 1. Horizontal Worker Auto-Scaling
- **Kubernetes HPA (Horizontal Pod Autoscaler)**: Scales decision worker pods based on CPU and request queue depth (min: 6, max: 48 replicas).
- **Stateless Decisioning**: Real-time evaluation pipeline holds no local session state, enabling instant traffic routing across any healthy replica.

---

## 2. Low-Latency Feature Caching
- **Redis Cluster Sharding**: Features are sharded across a 3-node Redis ElastiCache cluster with $< 1\text{ms}$ retrieval latency.
- **Batched Graph Queries**: Graph traversal queries use partitioned memory indices to evaluate 3-hop entity neighborhoods in $< 3\text{ms}$.
