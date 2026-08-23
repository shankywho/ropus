import type { RiskDecision } from "./contracts";

/**
 * DEMO DATA — deterministic fixtures shaped exactly like the real API response.
 * Never presented as live production data; screens sourced from these render
 * the DEMO DATA tag.
 */
export const DATA_SOURCE = {
  label: "Demo data",
  note: "Deterministic fixture shaped like the /v1/risk/evaluate response. Not live production traffic.",
  tenantId: "tnt_ropus_demo",
  environment: "sandbox",
} as const;

/** The escalated wire transfer — the attack sequence in the demo narrative. */
export const blockedDecision: RiskDecision = {
  decisionId: "dec_01HZK4JR7T2QW8",
  transactionId: "txn_9f3c21ab7d",
  customerId: "cus_4471029",
  tenantId: DATA_SOURCE.tenantId,
  amount: 14500,
  currency: "USD",
  verdict: "BLOCK",
  riskScore: 0.96,
  baseScore: 0.04,
  confidence: 0.93,
  model: "ropus-risk-xgb",
  modelVersion: "v4.2.1",
  policy: "pol_wire_outbound_v7",
  latencyMs: 38.4,
  evaluatedAt: "2026-08-22T17:41:09.220Z",
  factors: [
    {
      id: "f_graph_cluster",
      label: "Shared payout account with fraud cluster",
      weight: 0.31,
      source: "GRAPH",
      detail: "14 accounts, 3 confirmed fraud",
    },
    {
      id: "f_geo_velocity",
      label: "Impossible travel velocity",
      weight: 0.24,
      source: "RULES",
      detail: "Boston → Limassol in 41 minutes",
    },
    {
      id: "f_ml_amount",
      label: "Amount deviates from customer baseline",
      weight: 0.17,
      source: "ML",
      detail: "14,500 USD vs p95 of 820 USD",
    },
    {
      id: "f_ti_proxy",
      label: "Login IP on residential proxy network",
      weight: 0.14,
      source: "THREAT_INTEL",
      detail: "AS200651 · reputation: malicious",
    },
    {
      id: "f_device_new",
      label: "Unrecognised device fingerprint",
      weight: 0.1,
      source: "DEVICE",
      detail: "First seen 6 minutes before transfer",
    },
  ],
  rules: [
    {
      id: "RULE-VELOCITY-04",
      name: "Geo velocity threshold",
      outcome: "Exceeded — 2 logins, 3,180 km apart, 41 min",
    },
    {
      id: "RULE-AMT-11",
      name: "Outbound wire ceiling",
      outcome: "Exceeded — 14,500 USD over 5,000 USD tier limit",
    },
    {
      id: "RULE-PAYOUT-02",
      name: "Payout account age",
      outcome: "Failed — beneficiary added 9 minutes before transfer",
    },
  ],
  evidence: [
    {
      id: "e1",
      kind: "OBSERVED",
      text: "Login from Limassol, CY at 17:00:14 UTC on an unrecognised device.",
      source: "DEVICE",
    },
    {
      id: "e2",
      kind: "OBSERVED",
      text: "Beneficiary payout account PA-77120 added at 17:32:02 UTC.",
      source: "RULES",
    },
    {
      id: "e3",
      kind: "OBSERVED",
      text: "Outbound wire of 14,500 USD submitted at 17:41:09 UTC.",
      source: "RULES",
    },
    {
      id: "e4",
      kind: "OBSERVED",
      text: "Payout account PA-77120 is shared with 13 other customer accounts.",
      source: "GRAPH",
    },
    {
      id: "e5",
      kind: "INFERRED",
      text: "Session pattern matches account-takeover followed by rapid exfiltration.",
    },
    {
      id: "e6",
      kind: "INFERRED",
      text: "Payout account is likely a mule endpoint for a coordinated cluster.",
    },
    {
      id: "e7",
      kind: "RECOMMENDED",
      text: "Block the transfer and hold funds pending review.",
      action: { id: "a_block", label: "Confirm block", destructive: true },
    },
    {
      id: "e8",
      kind: "RECOMMENDED",
      text: "Open an investigation case linked to this decision.",
      action: { id: "a_case", label: "Open case" },
    },
    {
      id: "e9",
      kind: "RECOMMENDED",
      text: "Freeze the 13 connected accounts sharing payout PA-77120.",
      action: { id: "a_freeze", label: "Request cluster freeze", destructive: true },
    },
  ],
  inference: {
    model: "ropus-risk-xgb",
    version: "v4.2.1",
    probability: 0.9612,
    features: [
      { name: "amount_zscore_30d", contribution: 0.29 },
      { name: "geo_velocity_kmh", contribution: 0.26 },
      { name: "payout_account_age_min", contribution: 0.21 },
      { name: "device_first_seen_min", contribution: 0.14 },
      { name: "graph_fraud_neighbour_count", contribution: 0.1 },
    ],
  },
  threatIntel: {
    ip: "185.220.101.44",
    asn: "AS200651 — residential proxy pool",
    ipReputation: "malicious",
    proxy: true,
    vpn: false,
    tor: false,
    deviceNovelty: "first seen 6 minutes ago",
  },
  webhook: {
    endpoint: "https://api.acme-bank.example/ropus/decisions",
    status: 200,
    attempt: 1,
    timestamp: "2026-08-22T17:41:09.284Z",
  },
  caseId: "CASE-88419",
  rawRequest: {
    transaction_id: "txn_9f3c21ab7d",
    customer_id: "cus_4471029",
    amount: 14500,
    currency: "USD",
    channel: "wire_outbound",
    beneficiary_account: "PA-77120",
    ip: "185.220.101.44",
    device_id: "dvc_5b91e0",
  },
};

/** The same customer's baseline activity, before the attack sequence. */
export const approvedDecision: RiskDecision = {
  decisionId: "dec_01HZK1M0PP9C3B",
  transactionId: "txn_2a71fe09c4",
  customerId: "cus_4471029",
  tenantId: DATA_SOURCE.tenantId,
  amount: 312.4,
  currency: "USD",
  verdict: "APPROVE",
  riskScore: 0.04,
  baseScore: 0.02,
  confidence: 0.97,
  model: "ropus-risk-xgb",
  modelVersion: "v4.2.1",
  policy: "pol_card_purchase_v7",
  latencyMs: 21.7,
  evaluatedAt: "2026-08-22T09:12:44.010Z",
  factors: [
    {
      id: "f_new_merchant",
      label: "Merchant not seen in prior 90 days",
      weight: 0.02,
      source: "ML",
      detail: "Category matches customer profile",
    },
  ],
  rules: [],
  evidence: [
    {
      id: "e1",
      kind: "OBSERVED",
      text: "Card purchase of 312.40 USD from a known device in Boston, US.",
      source: "DEVICE",
    },
    {
      id: "e2",
      kind: "OBSERVED",
      text: "Device dvc_1c07aa has been active on this account for 411 days.",
      source: "DEVICE",
    },
    {
      id: "e3",
      kind: "INFERRED",
      text: "Spend pattern is consistent with the customer's 90-day baseline.",
    },
    {
      id: "e4",
      kind: "RECOMMENDED",
      text: "No analyst action required. Continue monitoring.",
      action: { id: "a_ack", label: "Acknowledge" },
    },
  ],
  inference: {
    model: "ropus-risk-xgb",
    version: "v4.2.1",
    probability: 0.0388,
    features: [
      { name: "merchant_familiarity", contribution: 0.02 },
      { name: "amount_zscore_30d", contribution: 0.01 },
      { name: "device_trust_score", contribution: 0.01 },
    ],
  },
  threatIntel: {
    ip: "73.14.201.9",
    asn: "AS7922 — residential broadband",
    ipReputation: "clean",
    proxy: false,
    vpn: false,
    tor: false,
    deviceNovelty: "known device, 411 days",
  },
  webhook: {
    endpoint: "https://api.acme-bank.example/ropus/decisions",
    status: 200,
    attempt: 1,
    timestamp: "2026-08-22T09:12:44.061Z",
  },
  caseId: null,
  rawRequest: {
    transaction_id: "txn_2a71fe09c4",
    customer_id: "cus_4471029",
    amount: 312.4,
    currency: "USD",
    channel: "card_purchase",
    ip: "73.14.201.9",
    device_id: "dvc_1c07aa",
  },
};

export const decisionFixtures: Record<string, RiskDecision> = {
  [blockedDecision.decisionId]: blockedDecision,
  [approvedDecision.decisionId]: approvedDecision,
};
