/**
 * DEMO DATASET — NOT LIVE BACKEND DATA.
 *
 * The ROPUS backend is the source of truth. No endpoints are called here and
 * none are invented. Every screen that renders these records must surface the
 * `DEMO DATA` marker so operators never mistake fixtures for production state.
 *
 * When the backend contract is wired in, replace the readers below with typed
 * service calls; the component layer already consumes them through this module.
 */

export const DATA_SOURCE = {
  mode: "demo" as const,
  label: "DEMO DATA",
  note: "Fixtures rendered locally. Backend integration not connected on this build.",
};

export type Verdict = "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";
export type Health = "HEALTHY" | "DEGRADED" | "UNAVAILABLE";

export interface RiskEvent {
  time: string;
  txn: string;
  customer: string;
  amount: number;
  score: number;
  verdict: Verdict;
  signal: string;
  latencyMs: number;
  caseId: string | null;
  decisionId: string;
}

export const kpis = [
  {
    label: "Transactions evaluated",
    value: "1,284,392",
    delta: "+12.4%",
    dir: "up" as const,
    window: "Last 24 hours",
  },
  {
    label: "Approval rate",
    value: "93.71%",
    delta: "-0.42pp",
    dir: "down" as const,
    window: "Last 24 hours",
  },
  {
    label: "Review rate",
    value: "3.18%",
    delta: "+0.21pp",
    dir: "up" as const,
    window: "Last 24 hours",
  },
  {
    label: "Block rate",
    value: "1.44%",
    delta: "+0.19pp",
    dir: "up" as const,
    window: "Last 24 hours",
  },
  {
    label: "Average risk score",
    value: "0.148",
    delta: "+0.006",
    dir: "up" as const,
    window: "Last 24 hours",
  },
  {
    label: "P99 decision latency",
    value: "11.4ms",
    delta: "-1.8ms",
    dir: "down" as const,
    window: "Last 24 hours",
  },
];

export const decisionDistribution = [
  { verdict: "APPROVE" as Verdict, count: 1203621, pct: 93.71 },
  { verdict: "REVIEW" as Verdict, count: 40843, pct: 3.18 },
  { verdict: "CHALLENGE" as Verdict, count: 21447, pct: 1.67 },
  { verdict: "BLOCK" as Verdict, count: 18481, pct: 1.44 },
];

export const decisionTrend = [
  { t: "00:00", approve: 48210, review: 1620, challenge: 840, block: 690 },
  { t: "03:00", approve: 39104, review: 1410, challenge: 720, block: 610 },
  { t: "06:00", approve: 42880, review: 1502, challenge: 795, block: 664 },
  { t: "09:00", approve: 61230, review: 2180, challenge: 1130, block: 902 },
  { t: "12:00", approve: 72544, review: 2640, challenge: 1348, block: 1194 },
  { t: "15:00", approve: 69811, review: 2905, challenge: 1502, block: 1461 },
  { t: "18:00", approve: 64920, review: 2411, challenge: 1288, block: 1120 },
  { t: "21:00", approve: 55302, review: 1988, challenge: 1044, block: 941 },
];

export const riskEvents: RiskEvent[] = [
  {
    time: "14:42:31",
    txn: "txn_8f29c14a",
    customer: "usr_1842",
    amount: 14500,
    score: 0.96,
    verdict: "BLOCK",
    signal: "Impossible Travel",
    latencyMs: 4.2,
    caseId: "CASE-88419",
    decisionId: "dec_01hzk4m2",
  },
  {
    time: "14:42:29",
    txn: "txn_8f29c149",
    customer: "usr_9931",
    amount: 218.4,
    score: 0.11,
    verdict: "APPROVE",
    signal: "None",
    latencyMs: 3.1,
    caseId: null,
    decisionId: "dec_01hzk4m1",
  },
  {
    time: "14:42:26",
    txn: "txn_8f29c145",
    customer: "usr_4410",
    amount: 3200,
    score: 0.62,
    verdict: "REVIEW",
    signal: "Velocity Surge",
    latencyMs: 5.8,
    caseId: "CASE-88417",
    decisionId: "dec_01hzk4kz",
  },
  {
    time: "14:42:22",
    txn: "txn_8f29c141",
    customer: "usr_2287",
    amount: 890,
    score: 0.48,
    verdict: "CHALLENGE",
    signal: "Device Novelty",
    latencyMs: 4.7,
    caseId: null,
    decisionId: "dec_01hzk4kw",
  },
  {
    time: "14:42:19",
    txn: "txn_8f29c13d",
    customer: "usr_7765",
    amount: 74.9,
    score: 0.07,
    verdict: "APPROVE",
    signal: "None",
    latencyMs: 2.9,
    caseId: null,
    decisionId: "dec_01hzk4kt",
  },
  {
    time: "14:42:15",
    txn: "txn_8f29c138",
    customer: "usr_1842",
    amount: 9600,
    score: 0.91,
    verdict: "BLOCK",
    signal: "Fraud Graph Cluster",
    latencyMs: 6.4,
    caseId: "CASE-88414",
    decisionId: "dec_01hzk4kp",
  },
  {
    time: "14:42:11",
    txn: "txn_8f29c132",
    customer: "usr_3018",
    amount: 1450,
    score: 0.39,
    verdict: "APPROVE",
    signal: "Proxy IP",
    latencyMs: 4.0,
    caseId: null,
    decisionId: "dec_01hzk4kk",
  },
  {
    time: "14:42:08",
    txn: "txn_8f29c12e",
    customer: "usr_5502",
    amount: 22800,
    score: 0.74,
    verdict: "REVIEW",
    signal: "Amount Anomaly",
    latencyMs: 7.1,
    caseId: "CASE-88411",
    decisionId: "dec_01hzk4kg",
  },
  {
    time: "14:42:04",
    txn: "txn_8f29c129",
    customer: "usr_8873",
    amount: 312.5,
    score: 0.16,
    verdict: "APPROVE",
    signal: "None",
    latencyMs: 3.3,
    caseId: null,
    decisionId: "dec_01hzk4kb",
  },
  {
    time: "14:41:58",
    txn: "txn_8f29c121",
    customer: "usr_6640",
    amount: 5400,
    score: 0.55,
    verdict: "CHALLENGE",
    signal: "Bulletproof IP",
    latencyMs: 5.2,
    caseId: null,
    decisionId: "dec_01hzk4k4",
  },
  {
    time: "14:41:52",
    txn: "txn_8f29c118",
    customer: "usr_1109",
    amount: 640,
    score: 0.24,
    verdict: "APPROVE",
    signal: "None",
    latencyMs: 3.8,
    caseId: null,
    decisionId: "dec_01hzk4jx",
  },
  {
    time: "14:41:47",
    txn: "txn_8f29c110",
    customer: "usr_2287",
    amount: 18250,
    score: 0.88,
    verdict: "BLOCK",
    signal: "Emulator Canvas",
    latencyMs: 6.9,
    caseId: "CASE-88407",
    decisionId: "dec_01hzk4jr",
  },
];

export const focusDecision = {
  decisionId: "dec_01hzk4m2v7q9ptb3",
  transactionId: "txn_8f29c14a71bd",
  customerId: "usr_1842",
  verdict: "BLOCK" as Verdict,
  score: 0.96,
  baseScore: 0.08,
  confidence: 0.982,
  model: "xgboost-v3.4",
  latencyMs: 4.8,
  timestamp: "2026-08-22T14:42:31.408Z",
  amount: 14500,
  currency: "USD",
  policy: "policy-highrisk-v9",
  factors: [
    {
      name: "Velocity Surge",
      detail: "+$14,500 across 4 attempts / 60s",
      weight: 0.22,
      kind: "observed" as const,
    },
    {
      name: "Impossible Travel",
      detail: "US → Cyprus in 11 minutes",
      weight: 0.21,
      kind: "observed" as const,
    },
    {
      name: "ML Probability",
      detail: "XGBoost p(fraud) = 0.94",
      weight: 0.2,
      kind: "inferred" as const,
    },
    {
      name: "Device Novelty",
      detail: "Emulator canvas fingerprint",
      weight: 0.18,
      kind: "observed" as const,
    },
    {
      name: "Threat Intelligence",
      detail: "Bulletproof hosting ASN 203471",
      weight: 0.18,
      kind: "observed" as const,
    },
    {
      name: "Fraud Graph",
      detail: "14 connected accounts, 3 confirmed fraud",
      weight: 0.17,
      kind: "inferred" as const,
    },
  ],
  observed: [
    "4 authorization attempts from 2 devices within 60 seconds.",
    "Geolocation of last successful session: Austin, US (14:31:12Z).",
    "Current request originates from ASN 203471 (Cyprus), flagged bulletproof.",
    "Canvas fingerprint matches known emulator signature set EMU-118.",
  ],
  inferred: [
    "Session behaviour consistent with automated card testing followed by a single high-value cash-out.",
    "Account is a probable member of ring FG-2291 based on shared bank account and device identifiers.",
  ],
  recommended: [
    "Hold the transaction and keep the block in force.",
    "Freeze payout rail for usr_1842 pending manual identity re-verification.",
    "Extend a 3-hop graph sweep across the 14 connected accounts.",
  ],
};

export const cases = [
  {
    id: "CASE-88419",
    priority: "P1",
    customer: "usr_1842",
    txn: "txn_8f29c14a",
    score: 0.96,
    verdict: "BLOCK" as Verdict,
    assignee: "m.okafor",
    status: "In review",
    created: "14:42",
    updated: "14:51",
  },
  {
    id: "CASE-88417",
    priority: "P2",
    customer: "usr_4410",
    txn: "txn_8f29c145",
    score: 0.62,
    verdict: "REVIEW" as Verdict,
    assignee: "s.reyes",
    status: "Open",
    created: "14:42",
    updated: "14:44",
  },
  {
    id: "CASE-88414",
    priority: "P1",
    customer: "usr_1842",
    txn: "txn_8f29c138",
    score: 0.91,
    verdict: "BLOCK" as Verdict,
    assignee: "m.okafor",
    status: "Escalated",
    created: "14:42",
    updated: "14:49",
  },
  {
    id: "CASE-88411",
    priority: "P3",
    customer: "usr_5502",
    txn: "txn_8f29c12e",
    score: 0.74,
    verdict: "REVIEW" as Verdict,
    assignee: "Unassigned",
    status: "Open",
    created: "14:42",
    updated: "14:42",
  },
  {
    id: "CASE-88407",
    priority: "P2",
    customer: "usr_2287",
    txn: "txn_8f29c110",
    score: 0.88,
    verdict: "BLOCK" as Verdict,
    assignee: "j.lindqvist",
    status: "Awaiting customer",
    created: "14:41",
    updated: "14:55",
  },
  {
    id: "CASE-88402",
    priority: "P4",
    customer: "usr_7311",
    txn: "txn_8f29c0f8",
    score: 0.51,
    verdict: "CHALLENGE" as Verdict,
    assignee: "s.reyes",
    status: "Closed — released",
    created: "14:39",
    updated: "14:47",
  },
];

export const caseTimeline = [
  {
    at: "14:42:31Z",
    actor: "system",
    action: "Decision dec_01hzk4m2 returned BLOCK (0.96).",
    kind: "system" as const,
  },
  {
    at: "14:42:31Z",
    actor: "system",
    action: "Case CASE-88419 opened at priority P1 by policy-highrisk-v9.",
    kind: "system" as const,
  },
  {
    at: "14:44:02Z",
    actor: "agent.investigator",
    action: "Evidence bundle collected: 6 factors, 14 graph neighbours, 3 device records.",
    kind: "ai" as const,
  },
  {
    at: "14:47:19Z",
    actor: "m.okafor",
    action: "Assigned to self. Requested 3-hop graph expansion.",
    kind: "analyst" as const,
  },
  {
    at: "14:51:44Z",
    actor: "m.okafor",
    action: "Confirmed block. Payout rail frozen pending identity re-verification.",
    kind: "analyst" as const,
  },
];

export const graphNodes = [
  { id: "usr_1842", type: "Customer", x: 50, y: 46, risk: "critical" as const },
  { id: "dev_7c31", type: "Device", x: 22, y: 22, risk: "high" as const },
  { id: "ip_203.0.113.44", type: "IP", x: 20, y: 70, risk: "high" as const },
  { id: "acct_ee18", type: "Bank Account", x: 78, y: 24, risk: "high" as const },
  { id: "mail_r***@mx7.io", type: "Email", x: 80, y: 68, risk: "medium" as const },
  { id: "usr_2287", type: "Customer", x: 52, y: 84, risk: "high" as const },
  { id: "usr_6640", type: "Customer", x: 14, y: 46, risk: "medium" as const },
  { id: "mrc_9920", type: "Merchant", x: 86, y: 46, risk: "low" as const },
  { id: "txn_8f29c14a", type: "Transaction", x: 50, y: 14, risk: "critical" as const },
];

export const graphEdges: Array<[string, string, string]> = [
  ["usr_1842", "dev_7c31", "authenticated_from"],
  ["usr_1842", "ip_203.0.113.44", "session_ip"],
  ["usr_1842", "acct_ee18", "payout_account"],
  ["usr_1842", "mail_r***@mx7.io", "contact_email"],
  ["usr_1842", "txn_8f29c14a", "initiated"],
  ["usr_2287", "dev_7c31", "authenticated_from"],
  ["usr_2287", "acct_ee18", "payout_account"],
  ["usr_6640", "ip_203.0.113.44", "session_ip"],
  ["txn_8f29c14a", "mrc_9920", "settles_to"],
];

export const rules = [
  {
    id: "RULE-VELOCITY-001",
    name: "High velocity transaction",
    description: "More than 5 authorizations from one customer inside 60 seconds.",
    status: "Enabled",
    priority: 10,
    trigger: "count(auth, 60s) > 5",
    action: "REVIEW" as Verdict,
    modified: "2026-08-19",
  },
  {
    id: "RULE-GEO-004",
    name: "Impossible travel",
    description: "Two sessions whose geodesic distance exceeds feasible travel speed.",
    status: "Enabled",
    priority: 5,
    trigger: "travel_speed_kmh > 900",
    action: "BLOCK" as Verdict,
    modified: "2026-08-21",
  },
  {
    id: "RULE-DEV-011",
    name: "Emulator fingerprint",
    description: "Canvas or WebGL fingerprint matches known emulator signature set.",
    status: "Enabled",
    priority: 20,
    trigger: "device.emulator_score > 0.8",
    action: "CHALLENGE" as Verdict,
    modified: "2026-08-12",
  },
  {
    id: "RULE-AMT-002",
    name: "Amount anomaly",
    description: "Transaction exceeds 12x the customer's trailing 90-day median.",
    status: "Enabled",
    priority: 30,
    trigger: "amount > 12 * median_90d",
    action: "REVIEW" as Verdict,
    modified: "2026-07-30",
  },
  {
    id: "RULE-TI-007",
    name: "Bulletproof hosting ASN",
    description: "Origin ASN present in bulletproof hosting reputation feed.",
    status: "Enabled",
    priority: 15,
    trigger: "ip.asn in feed:bulletproof",
    action: "BLOCK" as Verdict,
    modified: "2026-08-04",
  },
  {
    id: "RULE-CARD-019",
    name: "Card testing pattern",
    description: "Sequential low-value declines followed by a high-value approval attempt.",
    status: "Disabled",
    priority: 25,
    trigger: "declines(<$5, 300s) >= 6",
    action: "BLOCK" as Verdict,
    modified: "2026-06-18",
  },
];

export const models = [
  {
    name: "risk-core",
    version: "xgboost-v3.4",
    status: "Serving",
    accuracy: 0.971,
    precision: 0.934,
    recall: 0.902,
    trained: "2026-08-14",
    deployment: "100% production",
    features: 284,
    latency: "3.9ms",
    calibration: "Isotonic, Brier 0.031",
  },
  {
    name: "risk-core",
    version: "xgboost-v3.5",
    status: "Shadow",
    accuracy: 0.978,
    precision: 0.941,
    recall: 0.918,
    trained: "2026-08-20",
    deployment: "Shadow traffic",
    features: 301,
    latency: "4.4ms",
    calibration: "Isotonic, Brier 0.028",
  },
  {
    name: "graph-embed",
    version: "gnn-v1.2",
    status: "Serving",
    accuracy: 0.923,
    precision: 0.887,
    recall: 0.861,
    trained: "2026-07-28",
    deployment: "100% production",
    features: 96,
    latency: "8.1ms",
    calibration: "Platt, Brier 0.049",
  },
  {
    name: "device-trust",
    version: "lgbm-v2.1",
    status: "Deprecated",
    accuracy: 0.902,
    precision: 0.848,
    recall: 0.812,
    trained: "2026-04-11",
    deployment: "Retired 2026-07-01",
    features: 64,
    latency: "2.6ms",
    calibration: "None",
  },
];

export const threatPanels = [
  {
    title: "IP reputation",
    rows: [
      ["Address", "203.0.113.44"],
      ["ASN", "AS203471 — bulletproof"],
      ["Feed hits", "4 of 6 feeds"],
      ["First seen", "2026-06-02"],
      ["Classification", "Malicious infrastructure"],
    ],
  },
  {
    title: "Proxy & VPN",
    rows: [
      ["Proxy", "Detected — residential rotation"],
      ["VPN", "Not detected"],
      ["Tor exit", "No"],
      ["Hosting provider", "Yes"],
      ["Exit consistency", "0.12 (unstable)"],
    ],
  },
  {
    title: "Geo risk",
    rows: [
      ["Origin", "Limassol, CY"],
      ["Billing country", "US"],
      ["Mismatch", "Yes"],
      ["Country risk tier", "Tier 3"],
      ["Travel feasibility", "Violated (11 min)"],
    ],
  },
  {
    title: "Device intelligence",
    rows: [
      ["Device", "dev_7c31"],
      ["Emulator score", "0.91"],
      ["Fingerprint age", "38 minutes"],
      ["Accounts on device", "9"],
      ["Integrity attestation", "Failed"],
    ],
  },
  {
    title: "Behavioural signals",
    rows: [
      ["Typing cadence", "Non-human (scripted)"],
      ["Form fill time", "412ms"],
      ["Paste events", "6 of 6 fields"],
      ["Session depth", "2 pages"],
      ["Return visitor", "No"],
    ],
  },
  {
    title: "Feed status",
    rows: [
      ["abuse-registry", "Synced 4m ago"],
      ["asn-reputation", "Synced 11m ago"],
      ["device-signatures", "Synced 2m ago"],
      ["sanctions-list", "Synced 51m ago"],
      ["breach-corpus", "Stale — 3h 12m"],
    ],
  },
];

export const services: Array<{
  name: string;
  state: Health;
  detail: string;
  p50: string;
  p95: string;
  p99: string;
}> = [
  {
    name: "Risk API",
    state: "HEALTHY",
    detail: "Error rate 0.021%",
    p50: "3.1ms",
    p95: "8.4ms",
    p99: "11.4ms",
  },
  {
    name: "ML inference service",
    state: "HEALTHY",
    detail: "Queue depth 3",
    p50: "3.9ms",
    p95: "9.8ms",
    p99: "14.2ms",
  },
  {
    name: "PostgreSQL (primary)",
    state: "HEALTHY",
    detail: "Replication lag 82ms",
    p50: "1.2ms",
    p95: "4.0ms",
    p99: "9.1ms",
  },
  {
    name: "Redis (feature cache)",
    state: "HEALTHY",
    detail: "Hit rate 99.2%",
    p50: "0.4ms",
    p95: "1.1ms",
    p99: "2.8ms",
  },
  {
    name: "Kafka (decision stream)",
    state: "DEGRADED",
    detail: "Consumer lag 41,208 messages on partition 6",
    p50: "12ms",
    p95: "480ms",
    p99: "1.9s",
  },
  {
    name: "Webhook delivery",
    state: "DEGRADED",
    detail: "3.1% retry rate to 2 endpoints",
    p50: "88ms",
    p95: "620ms",
    p99: "2.4s",
  },
  {
    name: "Graph store",
    state: "HEALTHY",
    detail: "2-hop traversal budget within SLO",
    p50: "18ms",
    p95: "62ms",
    p99: "140ms",
  },
];

export const apiKeys = [
  {
    id: "key_live_9f2a",
    label: "Payments gateway (prod)",
    env: "PRODUCTION",
    prefix: "rk_live_9f2a",
    created: "2026-03-11",
    lastUsed: "2 minutes ago",
    scopes: "risk:evaluate, cases:read",
  },
  {
    id: "key_live_3b71",
    label: "Batch scoring worker",
    env: "PRODUCTION",
    prefix: "rk_live_3b71",
    created: "2026-05-02",
    lastUsed: "17 minutes ago",
    scopes: "risk:evaluate",
  },
  {
    id: "key_test_c410",
    label: "Integration tests",
    env: "SANDBOX",
    prefix: "rk_test_c410",
    created: "2026-06-24",
    lastUsed: "3 hours ago",
    scopes: "risk:evaluate, rules:write",
  },
];

export const webhooks = [
  {
    url: "https://ops.internal/ropus/decisions",
    events: "decision.blocked, decision.review",
    state: "HEALTHY" as Health,
    success: "99.94%",
    lastDelivery: "8s ago",
  },
  {
    url: "https://ledger.internal/hooks/case",
    events: "case.opened, case.closed",
    state: "HEALTHY" as Health,
    success: "99.99%",
    lastDelivery: "1m ago",
  },
  {
    url: "https://partner.example.com/ropus",
    events: "decision.*",
    state: "DEGRADED" as Health,
    success: "96.80%",
    lastDelivery: "retrying (attempt 3)",
  },
];

export const team = [
  {
    user: "m.okafor",
    email: "m.okafor@ropus.internal",
    role: "Risk Analyst",
    perms: "cases:write, decisions:read",
    mfa: "Hardware key",
    lastActive: "now",
  },
  {
    user: "s.reyes",
    email: "s.reyes@ropus.internal",
    role: "Risk Analyst",
    perms: "cases:write, decisions:read",
    mfa: "TOTP",
    lastActive: "12 minutes ago",
  },
  {
    user: "j.lindqvist",
    email: "j.lindqvist@ropus.internal",
    role: "Fraud Lead",
    perms: "cases:write, rules:write, decisions:read",
    mfa: "Hardware key",
    lastActive: "1 hour ago",
  },
  {
    user: "d.varma",
    email: "d.varma@ropus.internal",
    role: "Platform Engineer",
    perms: "keys:write, webhooks:write",
    mfa: "Hardware key",
    lastActive: "3 hours ago",
  },
  {
    user: "a.nowak",
    email: "a.nowak@ropus.internal",
    role: "Read Only",
    perms: "decisions:read",
    mfa: "TOTP",
    lastActive: "2 days ago",
  },
];

export const usage = [
  { meter: "Risk evaluations", used: 1284392, included: 2000000, unit: "evaluations" },
  { meter: "Cases", used: 4182, included: 10000, unit: "cases" },
  { meter: "Agent investigations", used: 912, included: 2500, unit: "calls" },
  { meter: "Evidence storage", used: 412, included: 1000, unit: "GB" },
];

export const invoices = [
  {
    id: "INV-2026-07",
    period: "Jul 2026",
    amount: "$18,400.00",
    status: "Paid",
    issued: "2026-08-01",
  },
  {
    id: "INV-2026-06",
    period: "Jun 2026",
    amount: "$17,120.00",
    status: "Paid",
    issued: "2026-07-01",
  },
  {
    id: "INV-2026-05",
    period: "May 2026",
    amount: "$16,880.00",
    status: "Paid",
    issued: "2026-06-01",
  },
];

export const securityControls: Array<{
  name: string;
  state: string;
  kind: "real" | "simulated" | "control" | "docs";
  detail: string;
}> = [
  {
    name: "API authentication",
    state: "Enforced",
    kind: "real",
    detail: "Key-scoped HMAC request signing on every /v1 route.",
  },
  {
    name: "Tenant isolation",
    state: "Enforced",
    kind: "real",
    detail: "Row-level tenant predicates applied at the query layer.",
  },
  {
    name: "Encryption at rest",
    state: "Enforced",
    kind: "control",
    detail: "Volume-level encryption managed by the hosting platform.",
  },
  {
    name: "Audit integrity",
    state: "Hash-chained",
    kind: "real",
    detail: "Append-only audit log, per-event SHA-256 chaining.",
  },
  {
    name: "Webhook signing",
    state: "Enforced",
    kind: "real",
    detail: "Per-endpoint secret, timestamped signature header.",
  },
  {
    name: "Rate limiting",
    state: "Enforced",
    kind: "real",
    detail: "Token bucket per API key, 429 with retry-after.",
  },
  {
    name: "Threat feed ingestion",
    state: "Deterministic simulation",
    kind: "simulated",
    detail: "Feed responses replayed from fixtures in this environment.",
  },
  {
    name: "Compliance mapping",
    state: "Documentation only",
    kind: "docs",
    detail:
      "Control descriptions written down. No external audit or certification has been performed.",
  },
];

export const securityEvents = [
  {
    at: "14:39:02Z",
    type: "auth.key_rejected",
    subject: "rk_live_0000…",
    detail: "Signature mismatch from 198.51.100.7",
    severity: "warning" as const,
  },
  {
    at: "14:12:44Z",
    type: "ratelimit.exceeded",
    subject: "rk_live_3b71",
    detail: "1,240 requests over budget in 60s window",
    severity: "warning" as const,
  },
  {
    at: "13:51:10Z",
    type: "audit.chain_verified",
    subject: "segment 0x4f21",
    detail: "Hash chain verified over 128,442 events",
    severity: "info" as const,
  },
  {
    at: "12:02:31Z",
    type: "rbac.denied",
    subject: "a.nowak",
    detail: "Attempted cases:write without grant",
    severity: "warning" as const,
  },
];
