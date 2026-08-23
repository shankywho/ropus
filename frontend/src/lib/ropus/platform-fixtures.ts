/**
 * DEMO DATA — deterministic platform-surface fixtures.
 *
 * Every identifier here is consistent with the single scenario used across
 * Demo → Decision → Fraud Graph → Case (cus_4471029, txn_9f3c21ab7d,
 * dec_01HZK4JR7T2QW8, PA-77120, dvc_5b91e0, 185.220.101.44,
 * pol_wire_outbound_v7, CASE-88419).
 */

export type RuleRecord = {
  id: string;
  name: string;
  policy: string;
  scope: string;
  action: "BLOCK" | "REVIEW" | "CHALLENGE" | "SCORE";
  weight: number;
  state: "ENABLED" | "SHADOW" | "DISABLED";
  hits24h: number;
  precision: number;
  updatedAt: string;
  updatedBy: string;
};

export const rules: RuleRecord[] = [
  {
    id: "rule_velocity_geo_07",
    name: "Impossible travel velocity",
    policy: "pol_wire_outbound_v7",
    scope: "session",
    action: "SCORE",
    weight: 0.24,
    state: "ENABLED",
    hits24h: 412,
    precision: 0.78,
    updatedAt: "2026-07-19",
    updatedBy: "r.duarte",
  },
  {
    id: "rule_new_beneficiary_02",
    name: "New beneficiary within 24h of wire",
    policy: "pol_wire_outbound_v7",
    scope: "transaction",
    action: "REVIEW",
    weight: 0.18,
    state: "ENABLED",
    hits24h: 1_884,
    precision: 0.41,
    updatedAt: "2026-06-02",
    updatedBy: "r.duarte",
  },
  {
    id: "rule_amount_baseline_11",
    name: "Amount deviates from customer baseline",
    policy: "pol_wire_outbound_v7",
    scope: "customer",
    action: "SCORE",
    weight: 0.17,
    state: "ENABLED",
    hits24h: 2_310,
    precision: 0.36,
    updatedAt: "2026-05-27",
    updatedBy: "m.okafor",
  },
  {
    id: "rule_payout_cluster_04",
    name: "Payout account in confirmed fraud cluster",
    policy: "pol_global_hardstop_v3",
    scope: "graph",
    action: "BLOCK",
    weight: 0.31,
    state: "ENABLED",
    hits24h: 37,
    precision: 0.96,
    updatedAt: "2026-08-11",
    updatedBy: "s.ibrahim",
  },
  {
    id: "rule_residential_proxy_09",
    name: "Login IP on residential proxy network",
    policy: "pol_session_v4",
    scope: "session",
    action: "SCORE",
    weight: 0.14,
    state: "ENABLED",
    hits24h: 946,
    precision: 0.52,
    updatedAt: "2026-08-03",
    updatedBy: "s.ibrahim",
  },
  {
    id: "rule_device_novel_15",
    name: "Unrecognised device fingerprint",
    policy: "pol_session_v4",
    scope: "device",
    action: "CHALLENGE",
    weight: 0.1,
    state: "ENABLED",
    hits24h: 5_120,
    precision: 0.22,
    updatedAt: "2026-04-14",
    updatedBy: "m.okafor",
  },
  {
    id: "rule_mule_fanout_01",
    name: "Payout fan-out across linked accounts",
    policy: "pol_wire_outbound_v7",
    scope: "graph",
    action: "REVIEW",
    weight: 0.12,
    state: "SHADOW",
    hits24h: 88,
    precision: 0.64,
    updatedAt: "2026-08-20",
    updatedBy: "s.ibrahim",
  },
  {
    id: "rule_dormant_reactivation_06",
    name: "Dormant account reactivated then wires out",
    policy: "pol_wire_outbound_v7",
    scope: "customer",
    action: "REVIEW",
    weight: 0.09,
    state: "DISABLED",
    hits24h: 0,
    precision: 0.31,
    updatedAt: "2026-02-09",
    updatedBy: "r.duarte",
  },
];

export type ModelRecord = {
  id: string;
  name: string;
  version: string;
  stage: "PRODUCTION" | "SHADOW" | "RETIRED";
  auc: number;
  driftPsi: number;
  p99Ms: number;
  trainedOn: string;
  features: number;
  callShare: number;
};

export const models: ModelRecord[] = [
  {
    id: "mdl_wire_risk",
    name: "Wire outbound risk",
    version: "v4.2.1",
    stage: "PRODUCTION",
    auc: 0.972,
    driftPsi: 0.04,
    p99Ms: 19.6,
    trainedOn: "2026-06-30",
    features: 148,
    callShare: 1,
  },
  {
    id: "mdl_wire_risk",
    name: "Wire outbound risk",
    version: "v4.3.0-rc2",
    stage: "SHADOW",
    auc: 0.979,
    driftPsi: 0.02,
    p99Ms: 22.4,
    trainedOn: "2026-08-12",
    features: 163,
    callShare: 0.1,
  },
  {
    id: "mdl_card_cnp",
    name: "Card-not-present risk",
    version: "v7.1.4",
    stage: "PRODUCTION",
    auc: 0.943,
    driftPsi: 0.11,
    p99Ms: 14.2,
    trainedOn: "2026-05-18",
    features: 96,
    callShare: 1,
  },
  {
    id: "mdl_mule_graph",
    name: "Mule network embedding",
    version: "v2.0.3",
    stage: "PRODUCTION",
    auc: 0.918,
    driftPsi: 0.07,
    p99Ms: 41.1,
    trainedOn: "2026-07-05",
    features: 64,
    callShare: 0.34,
  },
  {
    id: "mdl_wire_risk",
    name: "Wire outbound risk",
    version: "v4.1.7",
    stage: "RETIRED",
    auc: 0.961,
    driftPsi: 0.19,
    p99Ms: 18.9,
    trainedOn: "2026-03-22",
    features: 141,
    callShare: 0,
  },
];

/** Configuration + recent-evaluation detail shown in the rule inspector. */
export type RuleDetail = {
  logic: { if: string; and?: string; then: string };
  avgContribution: number;
  lastTriggeredAt: string | null;
  lastTriggeredBy: string | null;
  lastDecisionId: string | null;
};

export const ruleDetails: Record<string, RuleDetail> = {
  rule_velocity_geo_07: {
    logic: {
      if: "distance between successful sessions > 3,000 km",
      and: "time difference < 60 minutes",
      then: "add +0.24 risk contribution",
    },
    avgContribution: 0.24,
    lastTriggeredAt: "2026-08-22 17:41:09 UTC",
    lastTriggeredBy: "txn_9f3c21ab7d",
    lastDecisionId: "dec_01HZK4JR7T2QW8",
  },
  rule_new_beneficiary_02: {
    logic: {
      if: "beneficiary account age < 24 hours",
      and: "channel = wire_outbound",
      then: "route verdict to REVIEW",
    },
    avgContribution: 0.18,
    lastTriggeredAt: "2026-08-22 17:41:09 UTC",
    lastTriggeredBy: "txn_9f3c21ab7d",
    lastDecisionId: "dec_01HZK4JR7T2QW8",
  },
  rule_amount_baseline_11: {
    logic: {
      if: "amount z-score over trailing 30 days > 3.0",
      and: "customer tenure > 90 days",
      then: "add +0.17 risk contribution",
    },
    avgContribution: 0.17,
    lastTriggeredAt: "2026-08-22 17:41:09 UTC",
    lastTriggeredBy: "txn_9f3c21ab7d",
    lastDecisionId: "dec_01HZK4JR7T2QW8",
  },
  rule_payout_cluster_04: {
    logic: {
      if: "payout account within 2 hops of a CONFIRMED_FRAUD account",
      and: "shared beneficiary edge observed in last 30 days",
      then: "return BLOCK (hard stop)",
    },
    avgContribution: 0.31,
    lastTriggeredAt: "2026-08-22 17:41:09 UTC",
    lastTriggeredBy: "txn_9f3c21ab7d",
    lastDecisionId: "dec_01HZK4JR7T2QW8",
  },
  rule_residential_proxy_09: {
    logic: {
      if: "login IP present on residential-proxy intel feed",
      and: "ASN differs from customer 30-day baseline",
      then: "add +0.14 risk contribution",
    },
    avgContribution: 0.14,
    lastTriggeredAt: "2026-08-22 17:39:52 UTC",
    lastTriggeredBy: "txn_9f3c21ab7d",
    lastDecisionId: "dec_01HZK4JR7T2QW8",
  },
  rule_device_novel_15: {
    logic: {
      if: "device fingerprint not seen for this customer",
      and: "session age < 30 minutes",
      then: "issue step-up CHALLENGE",
    },
    avgContribution: 0.1,
    lastTriggeredAt: "2026-08-22 17:38:04 UTC",
    lastTriggeredBy: "txn_9f3c21ab7d",
    lastDecisionId: "dec_01HZK4JR7T2QW8",
  },
  rule_mule_fanout_01: {
    logic: {
      if: "payout account received funds from > 4 distinct customers in 7 days",
      and: "outbound sweep within 15 minutes of credit",
      then: "route verdict to REVIEW (shadow only)",
    },
    avgContribution: 0.12,
    lastTriggeredAt: "2026-08-22 15:22:03 UTC",
    lastTriggeredBy: "txn_5b81de20c4",
    lastDecisionId: null,
  },
  rule_dormant_reactivation_06: {
    logic: {
      if: "no activity for > 180 days",
      and: "outbound wire within 48h of reactivation",
      then: "route verdict to REVIEW",
    },
    avgContribution: 0.09,
    lastTriggeredAt: null,
    lastTriggeredBy: null,
    lastDecisionId: null,
  },
};

/** Registry lifecycle metadata shown in the model inspector. */
export type ModelDetail = {
  servingAlias: string;
  promotedOn: string | null;
  previousVersion: string | null;
  shadowCandidate: string | null;
};

export const modelDetails: Record<string, ModelDetail> = {
  "mdl_wire_risk@v4.2.1": {
    servingAlias: "ropus-risk-xgb",
    promotedOn: "2026-08-12",
    previousVersion: "v4.1.7",
    shadowCandidate: "v4.3.0-rc2",
  },
  "mdl_wire_risk@v4.3.0-rc2": {
    servingAlias: "ropus-risk-xgb",
    promotedOn: null,
    previousVersion: "v4.2.1",
    shadowCandidate: null,
  },
  "mdl_wire_risk@v4.1.7": {
    servingAlias: "ropus-risk-xgb",
    promotedOn: "2026-04-02",
    previousVersion: "v4.1.2",
    shadowCandidate: null,
  },
  "mdl_card_cnp@v7.1.4": {
    servingAlias: "ropus-cnp-gbm",
    promotedOn: "2026-05-30",
    previousVersion: "v7.0.9",
    shadowCandidate: null,
  },
  "mdl_mule_graph@v2.0.3": {
    servingAlias: "ropus-mule-gnn",
    promotedOn: "2026-07-11",
    previousVersion: "v2.0.1",
    shadowCandidate: null,
  },
};

export type InvestigationRecord = {
  id: string;
  title: string;
  entities: number;
  linkedCases: string[];
  owner: string;
  state: "ACTIVE" | "MONITORING" | "CLOSED";
  opened: string;
  lastActivity: string;
  exposure: number;
};

export const investigations: InvestigationRecord[] = [
  {
    id: "inv_2026_0184",
    title: "PA-77120 payout cluster",
    entities: 7,
    linkedCases: ["CASE-88419", "CASE-88361"],
    owner: "m.okafor",
    state: "ACTIVE",
    opened: "2026-08-22",
    lastActivity: "2026-08-22 17:44Z",
    exposure: 61_400,
  },
  {
    id: "inv_2026_0179",
    title: "Residential proxy ASN 209588 sessions",
    entities: 23,
    linkedCases: ["CASE-88407"],
    owner: "s.ibrahim",
    state: "ACTIVE",
    opened: "2026-08-19",
    lastActivity: "2026-08-22 09:12Z",
    exposure: 128_950,
  },
  {
    id: "inv_2026_0166",
    title: "Merchant velocity ring — three acquirers",
    entities: 41,
    linkedCases: ["CASE-88392"],
    owner: "r.duarte",
    state: "MONITORING",
    opened: "2026-08-04",
    lastActivity: "2026-08-21 14:31Z",
    exposure: 74_220,
  },
  {
    id: "inv_2026_0142",
    title: "Device farm dvc_5b91e0 family",
    entities: 12,
    linkedCases: [],
    owner: "s.ibrahim",
    state: "CLOSED",
    opened: "2026-07-11",
    lastActivity: "2026-08-01 11:07Z",
    exposure: 18_600,
  },
];

export type IndicatorRecord = {
  value: string;
  type: "IP" | "ASN" | "DEVICE" | "ACCOUNT" | "BIN";
  classification: "MALICIOUS" | "SUSPICIOUS" | "BENIGN";
  feed: string;
  confidence: number;
  firstSeen: string;
  lastSeen: string;
  hits24h: number;
};

export const indicators: IndicatorRecord[] = [
  {
    value: "185.220.101.44",
    type: "IP",
    classification: "MALICIOUS",
    feed: "internal-graph",
    confidence: 0.94,
    firstSeen: "2026-05-02",
    lastSeen: "2026-08-22",
    hits24h: 31,
  },
  {
    value: "AS209588",
    type: "ASN",
    classification: "SUSPICIOUS",
    feed: "proxy-registry",
    confidence: 0.71,
    firstSeen: "2025-11-18",
    lastSeen: "2026-08-22",
    hits24h: 946,
  },
  {
    value: "PA-77120",
    type: "ACCOUNT",
    classification: "MALICIOUS",
    feed: "consortium-fraud",
    confidence: 0.99,
    firstSeen: "2026-08-14",
    lastSeen: "2026-08-22",
    hits24h: 4,
  },
  {
    value: "dvc_5b91e0",
    type: "DEVICE",
    classification: "SUSPICIOUS",
    feed: "internal-device",
    confidence: 0.63,
    firstSeen: "2026-08-22",
    lastSeen: "2026-08-22",
    hits24h: 2,
  },
  {
    value: "531993",
    type: "BIN",
    classification: "BENIGN",
    feed: "issuer-registry",
    confidence: 0.12,
    firstSeen: "2024-01-09",
    lastSeen: "2026-08-22",
    hits24h: 12_884,
  },
];

export type ApiKeyRecord = {
  id: string;
  label: string;
  prefix: string;
  environment: "PRODUCTION" | "SANDBOX";
  scopes: string[];
  created: string;
  lastUsed: string;
  state: "ACTIVE" | "ROTATING" | "REVOKED";
};

export const apiKeys: ApiKeyRecord[] = [
  {
    id: "key_9f21ab",
    label: "Payments gateway (primary)",
    prefix: "rpk_live_9f21…",
    environment: "PRODUCTION",
    scopes: ["risk:evaluate", "decisions:read"],
    created: "2025-09-14",
    lastUsed: "2026-08-22 17:41Z",
    state: "ACTIVE",
  },
  {
    id: "key_44b0c1",
    label: "Case automation worker",
    prefix: "rpk_live_44b0…",
    environment: "PRODUCTION",
    scopes: ["cases:read", "cases:write"],
    created: "2026-01-22",
    lastUsed: "2026-08-22 17:39Z",
    state: "ACTIVE",
  },
  {
    id: "key_7c8de2",
    label: "Ledger reconciliation",
    prefix: "rpk_live_7c8d…",
    environment: "PRODUCTION",
    scopes: ["decisions:read"],
    created: "2025-06-03",
    lastUsed: "2026-08-22 16:02Z",
    state: "ROTATING",
  },
  {
    id: "key_1a55f0",
    label: "Integration sandbox",
    prefix: "rpk_test_1a55…",
    environment: "SANDBOX",
    scopes: ["risk:evaluate", "graph:read"],
    created: "2026-04-08",
    lastUsed: "2026-08-21 10:18Z",
    state: "ACTIVE",
  },
  {
    id: "key_0e33ba",
    label: "Deprecated batch scorer",
    prefix: "rpk_live_0e33…",
    environment: "PRODUCTION",
    scopes: ["risk:evaluate"],
    created: "2024-11-30",
    lastUsed: "2026-03-02 08:44Z",
    state: "REVOKED",
  },
];

export type WebhookEndpoint = {
  id: string;
  url: string;
  events: string[];
  state: "ACTIVE" | "PAUSED";
  successRate: number;
  p95Ms: number;
  lastDelivery: string;
};

export const webhookEndpoints: WebhookEndpoint[] = [
  {
    id: "whk_ledger_01",
    url: "https://ops.northbank.example/ropus/decisions",
    events: ["decision.returned", "decision.overridden"],
    state: "ACTIVE",
    successRate: 0.9992,
    p95Ms: 148.2,
    lastDelivery: "2026-08-22 17:41Z",
  },
  {
    id: "whk_cases_02",
    url: "https://ops.northbank.example/ropus/cases",
    events: ["case.opened", "case.resolved"],
    state: "ACTIVE",
    successRate: 0.9977,
    p95Ms: 210.5,
    lastDelivery: "2026-08-22 17:42Z",
  },
  {
    id: "whk_siem_03",
    url: "https://siem.northbank.example/collect/ropus",
    events: ["decision.returned", "threat.indicator.added"],
    state: "PAUSED",
    successRate: 0.9411,
    p95Ms: 884.0,
    lastDelivery: "2026-08-20 03:12Z",
  },
];

export type WebhookDelivery = {
  id: string;
  endpoint: string;
  event: string;
  status: number;
  attempt: number;
  at: string;
  latencyMs: number;
};

export const webhookDeliveries: WebhookDelivery[] = [
  { id: "dlv_88419a", endpoint: "whk_cases_02", event: "case.opened", status: 200, attempt: 1, at: "2026-08-22 17:42:11Z", latencyMs: 182.4 },
  { id: "dlv_9f3c21", endpoint: "whk_ledger_01", event: "decision.returned", status: 200, attempt: 1, at: "2026-08-22 17:41:09Z", latencyMs: 141.7 },
  { id: "dlv_5b81de", endpoint: "whk_ledger_01", event: "decision.returned", status: 200, attempt: 1, at: "2026-08-22 15:22:03Z", latencyMs: 133.2 },
  { id: "dlv_siem14", endpoint: "whk_siem_03", event: "decision.returned", status: 504, attempt: 4, at: "2026-08-20 03:12:44Z", latencyMs: 30_000 },
  { id: "dlv_1c04ff", endpoint: "whk_ledger_01", event: "decision.returned", status: 200, attempt: 1, at: "2026-08-22 11:04:51Z", latencyMs: 155.9 },
];

export type ApiEndpointDoc = {
  method: "POST" | "GET";
  path: string;
  summary: string;
  p99Ms: number;
  calls24h: number;
  errorRate: number;
};

export const apiEndpoints: ApiEndpointDoc[] = [
  { method: "POST", path: "/v1/risk/evaluate", summary: "Score a transaction and return a verdict synchronously.", p99Ms: 62.4, calls24h: 1_284_392, errorRate: 0.0003 },
  { method: "GET", path: "/v1/risk/decisions/{decisionId}", summary: "Fetch a decision with factors and evidence.", p99Ms: 28.1, calls24h: 41_220, errorRate: 0.0001 },
  { method: "GET", path: "/v1/graph", summary: "Entity neighbourhood for a decision or entity.", p99Ms: 411.0, calls24h: 9_804, errorRate: 0.0142 },
  { method: "GET", path: "/v1/cases", summary: "List investigation cases for the tenant.", p99Ms: 74.9, calls24h: 6_112, errorRate: 0.0002 },
  { method: "POST", path: "/v1/cases/{caseId}/resolve", summary: "Record an analyst outcome on a case.", p99Ms: 88.3, calls24h: 214, errorRate: 0 },
  { method: "GET", path: "/v1/metrics/overview", summary: "Aggregated decisioning and health metrics.", p99Ms: 96.7, calls24h: 3_400, errorRate: 0 },
];

export const evaluateSample = `curl -X POST https://api.ropus.io/v1/risk/evaluate \\
  -H "Authorization: Bearer rpk_live_9f21…" \\
  -H "Content-Type: application/json" \\
  -d '{
    "transaction_id": "txn_9f3c21ab7d",
    "customer_id": "cus_4471029",
    "channel": "wire_outbound",
    "amount": 14500.00,
    "currency": "USD",
    "beneficiary_account": "PA-77120",
    "device_id": "dvc_5b91e0",
    "ip": "185.220.101.44"
  }'`;

export const evaluateResponse = `{
  "decision_id": "dec_01HZK4JR7T2QW8",
  "verdict": "BLOCK",
  "risk_score": 0.96,
  "confidence": 0.93,
  "policy": "pol_wire_outbound_v7",
  "latency_ms": 38.4,
  "case_id": "CASE-88419"
}`;

export type OpsEvent = {
  id: string;
  at: string;
  component: string;
  severity: "INFO" | "WARN" | "CRITICAL";
  text: string;
};

export const opsEvents: OpsEvent[] = [
  { id: "ops_9931", at: "2026-08-22 17:38:02Z", component: "Graph service", severity: "WARN", text: "3-hop traversals queued; p99 411ms against a 250ms objective." },
  { id: "ops_9930", at: "2026-08-22 16:55:41Z", component: "Decision stream", severity: "WARN", text: "Consumer lag on partition 6 reached 42,118 records." },
  { id: "ops_9927", at: "2026-08-22 14:10:19Z", component: "Model serving", severity: "INFO", text: "mdl_wire_risk v4.3.0-rc2 promoted to shadow at 10% mirrored traffic." },
  { id: "ops_9921", at: "2026-08-22 09:02:07Z", component: "Rules engine", severity: "INFO", text: "rule_mule_fanout_01 deployed in shadow mode by s.ibrahim." },
  { id: "ops_9914", at: "2026-08-21 22:47:33Z", component: "Webhook delivery", severity: "CRITICAL", text: "whk_siem_03 paused after 4 consecutive 504 responses." },
];

export const slos = [
  { name: "Decision API availability", target: "99.99%", actual: "99.995%", budget: 0.42, healthy: true },
  { name: "Decision p99 latency", target: "< 80 ms", actual: "62.4 ms", budget: 0.18, healthy: true },
  { name: "Graph traversal p99", target: "< 250 ms", actual: "411.0 ms", budget: 1, healthy: false },
  { name: "Webhook delivery success", target: "> 99.5%", actual: "99.62%", budget: 0.31, healthy: true },
];

export type AccessEvent = {
  id: string;
  at: string;
  actor: string;
  action: string;
  target: string;
  ip: string;
  result: "ALLOWED" | "DENIED";
};

export const accessEvents: AccessEvent[] = [
  { id: "aud_44120", at: "2026-08-22 17:44:02Z", actor: "m.okafor", action: "case.view", target: "CASE-88419", ip: "10.14.7.22", result: "ALLOWED" },
  { id: "aud_44119", at: "2026-08-22 17:42:55Z", actor: "svc_case_worker", action: "case.create", target: "CASE-88419", ip: "10.20.1.9", result: "ALLOWED" },
  { id: "aud_44112", at: "2026-08-22 16:31:10Z", actor: "r.duarte", action: "rule.update", target: "rule_velocity_geo_07", ip: "10.14.7.51", result: "ALLOWED" },
  { id: "aud_44108", at: "2026-08-22 15:04:48Z", actor: "t.novak", action: "apikey.reveal", target: "key_9f21ab", ip: "10.14.9.3", result: "DENIED" },
  { id: "aud_44101", at: "2026-08-22 11:22:36Z", actor: "s.ibrahim", action: "model.promote", target: "mdl_wire_risk v4.3.0-rc2", ip: "10.14.7.31", result: "ALLOWED" },
];

export const securityPosture = [
  { control: "SSO (SAML 2.0)", state: "Enforced", detail: "Okta — northbank.okta.example" },
  { control: "MFA", state: "Enforced", detail: "WebAuthn required for analyst and admin roles" },
  { control: "Key rotation", state: "90 days", detail: "1 key currently rotating (key_7c8de2)" },
  { control: "Data residency", state: "eu-central-1", detail: "Decision payloads retained 400 days" },
  { control: "PII redaction", state: "Enabled", detail: "Beneficiary names masked in decision payloads" },
  { control: "Audit export", state: "Streaming", detail: "SIEM endpoint paused — whk_siem_03" },
];

export const teamMembers = [
  { user: "m.okafor", name: "M. Okafor", role: "Risk Analyst", scopes: "cases:write, decisions:read", lastActive: "2026-08-22 17:44Z" },
  { user: "s.ibrahim", name: "S. Ibrahim", role: "Fraud Strategy", scopes: "rules:write, models:write", lastActive: "2026-08-22 16:12Z" },
  { user: "r.duarte", name: "R. Duarte", role: "Risk Engineering", scopes: "rules:write, api:admin", lastActive: "2026-08-22 16:31Z" },
  { user: "t.novak", name: "T. Novak", role: "Operations", scopes: "decisions:read", lastActive: "2026-08-22 15:04Z" },
  { user: "svc_case_worker", name: "Case automation", role: "Service account", scopes: "cases:write", lastActive: "2026-08-22 17:42Z" },
];

export const tenantSettings = [
  { key: "Tenant", value: "tnt_4f19ac" },
  { key: "Legal entity", value: "Northbank Payments N.V." },
  { key: "Default policy", value: "pol_wire_outbound_v7" },
  { key: "Block threshold", value: "0.90" },
  { key: "Review threshold", value: "0.55" },
  { key: "Challenge threshold", value: "0.35" },
  { key: "Decision timeout", value: "120 ms" },
  { key: "Fallback on timeout", value: "REVIEW" },
  { key: "Case auto-open", value: "verdict = BLOCK or score ≥ 0.80" },
  { key: "Retention", value: "400 days" },
];
