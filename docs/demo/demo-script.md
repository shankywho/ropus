# AI Risk Manager — Interactive Demo Script & Walkthrough

This script provides a complete presentation guide for demonstrating **AI Risk Manager (Ropus)** to prospective clients, executives, and fraud operations teams.

---

## Act 1: Executive Overview & Real-Time Posture
1. Navigate to `/` (Overview Dashboard).
2. **Key Talking Point**: "Welcome to AI Risk Manager. You're looking at live telemetry processing thousands of decisions per second with a sub-10ms P99 SLA. Notice our $4.5M+ in fraud prevented and global cluster health."
3. Highlight the decision distribution pie chart and latency timeline curves.

---

## Act 2: Live Transaction Stream & Explainability
1. Navigate to `/transactions` (Transaction Risk Explorer).
2. Click on transaction `tx_9981_mule_burst`.
3. **Key Talking Point**: "Why did our system block this $14,500 transaction? Look at the AI Summary and Factor Attribution bars. It wasn't just a simple rule — our Fraud Graph detected links to an active syndicate cluster (92% weight), and Threat Intel flagged an emulator proxy (95% weight)."

---

## Act 3: Fraud Graph & Hidden Entity Lineage
1. Navigate to `/graph` (Fraud Graph Explorer).
2. Click on the central node `dev_emulator_root_pool`.
3. **Key Talking Point**: "Here is our dynamic Knowledge Graph. What appeared to be 3 unrelated customer signups are actually sharing the exact same cloned emulator fingerprint. Notice how funds route through mule accounts into crypto cashout rails. One click allows the analyst to freeze all 7 connected entities simultaneously."

---

## Act 4: Interactive Live Attack Simulation (10/10 WOW Factor)
1. Navigate to `/demo` (Demo Simulator).
2. Select **Scenario 3: Distributed Card Testing & Mule Laundering Ring**.
3. Click **"Execute Scenario Run"**.
4. Watch the 6-stage narrative timeline execute live:
   `ATTACK -> DETECTION -> REASONING -> INVESTIGATION -> RESPONSE -> LEARNING`.
5. **Key Talking Point**: "In under 200 milliseconds, Ropus detected the distributed campaign, consulted the AI Agent Council, created an evidentiary case, synchronized blocks across 24 consortium banks, and generated a canary countermeasure rule."

---

## Act 5: Model Governance & Regulatory Audit Readiness
1. Navigate to `/governance` (Model Governance).
2. Show the active ensemble models, Kolmogorov-Smirnov statistics, PSI feature drift tracking, and fairness parity ratios ($>90\%$).
3. **Closing Point**: "Every decision is auditable, compliant with SR 11-7 model risk management standards, and protected against data poisoning."
