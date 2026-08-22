# Component 06: Threat Intelligence & Behavioral Telemetry

---

## 1. Why It Exists
Network-level and physical-layer anomalies provide critical early indicators of account takeover (ATO) and automated credential stuffing. 

The **Threat Intelligence & Behavioral Engine** (`backend/internal/graph/threat_intelligence.go`) evaluates:
1. Whether an egress IP belongs to a commercial bulletproof proxy, Tor exit node, or VPN hosting provider.
2. Whether the geographical distance between successive user sessions violates the laws of physics (**Impossible Travel Velocity**).
3. Whether the hardware canvas fingerprint deviates from the customer's established device profile.

---

## 2. Impossible Travel Velocity Algorithm

Given two consecutive session locations $L_1(\text{lat}_1, \text{lon}_1, t_1)$ and $L_2(\text{lat}_2, \text{lon}_2, t_2)$ on Earth (radius $R = 6371\text{ km}$):

### Great-Circle Haversine Formula:
$$\Delta\text{lat} = \text{lat}_2 - \text{lat}_1, \quad \Delta\text{lon} = \text{lon}_2 - \text{lon}_1$$
$$a = \sin^2\left(\frac{\Delta\text{lat}}{2}\right) + \cos(\text{lat}_1)\cos(\text{lat}_2)\sin^2\left(\frac{\Delta\text{lon}}{2}\right)$$
$$d = 2 R \cdot \arcsin\left(\sqrt{a}\right)$$
$$\text{Velocity} = \frac{d}{(t_2 - t_1)_{\text{hours}}} \quad (\text{km/h})$$

### Anomaly Rule:
$$\text{Anomaly} = \begin{cases}
\text{TRUE} (+0.21 \text{ risk weight}) & \text{if } \text{Velocity} > 900\text{ km/h} \text{ AND } d > 500\text{ km} \\
\text{FALSE} & \text{otherwise}
\end{cases}$$

---

## 3. Bulletproof IP & Subnet Matching

The engine maintains a radix tree of CIDR blocks representing known malicious subnets:

```go
type IPThreatRecord struct {
    Subnet       string    `json:"subnet"`
    RiskCategory string    `json:"risk_category"` // "BULLETPROOF_VPN", "TOR_EXIT", "DATACENTER_PROXY"
    ThreatScore  float64   `json:"threat_score"`  // 0.00 to 1.00
    LastObserved time.Time `json:"last_observed"`
}
```

When an IP matches a known proxy subnet (e.g. `198.51.100.0/24`), it contributes $+0.18$ to the composite risk score.

---

## 4. Source Code Map
- [`backend/internal/graph/threat_intelligence.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/graph/threat_intelligence.go): Haversine distance, velocity calculation, and subnet matchers.

---

## 5. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Consumes travel velocity and IP threat score.
- [Component 07: AI Investigators](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/07-ai-investigators.md) — Embeds network facts into investigation dossiers.
