# Component 07: Autonomous AI Investigation Agents & Agent Council

---

## 1. Why It Exists
Human fraud operations teams are overwhelmed by thousands of high-risk alerts per day. Analysts waste hours manually copying IP addresses, searching transaction databases, inspecting device graphs, and drafting Suspicious Activity Reports (SARs).

The **AI Investigation Subsystem** (`backend/internal/agents/`, `backend/internal/agent_council/`, `backend/internal/llm/`) deploys **autonomous LLM agents** to act as Tier-1 financial crime investigators. Within seconds of an elevated risk verdict, agents synthesize comprehensive evidentiary dossiers that cross-correlate graph topologies, velocity logs, and threat intelligence.

---

## 2. Multi-Persona Agent Council Architecture

```text
               [ High-Risk Event Trigger (Score >= 0.80) ]
                                   │
                                   ▼
┌────────────────────────────────────────────────────────────────────────┐
│ Multi-Persona Agent Council                                            │
│                                                                        │
│  ┌───────────────────────┐              ┌───────────────────────────┐  │
│  │ 1. Threat Hunter      │              │ 2. Graph Analyst          │  │
│  │ Analyzes IP subnets,  │              │ Analyzes 3-hop multi-mule │  │
│  │ proxies & impossible  │              │ linkages & shared hardware│  │
│  │ travel velocity.      │              │ canvas hashes.            │  │
│  └───────────┬───────────┘              └─────────────┬─────────────┘  │
│              │                                        │                │
│              ├───────────────────┬────────────────────┤                │
│              ▼                   ▼                    ▼                │
│  ┌───────────────────────┐              ┌───────────────────────────┐  │
│  │ 3. AML Officer        │              │ 4. Senior Fraud Lead      │  │
│  │ Checks SAR thresholds │              │ Synthesizes consensus     │  │
│  │ & structuring patterns│              │ evidentiary dossier.      │  │
│  └───────────────────────┘              └───────────────────────────┘  │
└──────────────────────────────────┬─────────────────────────────────────┘
                                   │
                                   ▼
               [ Structured Evidentiary Dossier #CASE-88419 ]
```

---

## 3. Strict Evidentiary Separation (Preventing Hallucinations)

To guarantee institutional regulatory credibility, all AI-generated dossiers are constrained to strictly separate:
1. **OBSERVED FACTS**: Verifiable telemetry directly present in database logs (IP addresses, transaction amounts, timestamps, exact geo-coordinates).
2. **INFERRED ATTACK PATTERNS**: Analytical hypotheses deduced from the facts (e.g. "Suspected credential stuffing account takeover followed by liquidity drain").
3. **RECOMMENDED ACTIONS**: Concrete operational steps (e.g. "Freeze wire settlement", "Invalidate active sessions", "Dispatch SAR filing").

---

## 4. Key Data Structures (Go)

```go
type InvestigationDossier struct {
    CaseID             string    `json:"case_id"`
    TransactionID      string    `json:"transaction_id"`
    ExecutiveSummary   string    `json:"executive_summary"`
    ObservedFacts      []string  `json:"observed_facts"`
    InferredPatterns   []string  `json:"inferred_patterns"`
    RecommendedActions []string  `json:"recommended_actions"`
    CouncilConsensus   string    `json:"council_consensus"`
    TokenUsage         int       `json:"token_usage"`
    ModelUsed          string    `json:"model_used"`
    CompiledAt         time.Time `json:"compiled_at"`
}
```

---

## 5. Source Code Map
- [`backend/internal/llm/investigation_agent.go`](backend/internal/llm/investigation_agent.go): Prompt templating, reasoning framework, and evidence synthesis.
- [`backend/internal/llm/llm_client.go`](backend/internal/llm/llm_client.go): Real HTTP client targeting OpenAI/Anthropic `/chat/completions` with deterministic fallback.
- [`backend/internal/agent_council/debate_engine.go`](backend/internal/agent_council/debate_engine.go): Multi-persona consensus engine and debate orchestration.

---

## 6. Cross-Component Links
- [Component 01: Product API](docs/components/01-product-api.md) — Triggers investigation upon elevated risk.
- [Component 08: Case Management](docs/components/08-cases-governance.md) — Embeds generated dossier into persistent analyst case queue.
