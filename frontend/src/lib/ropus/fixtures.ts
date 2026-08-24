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
  amount: 1450000,
  currency: "INR",
  verdict: "BLOCK",
  riskScore: 0.96,
  baseScore: 0.04,
  confidence: 0.93,
  model: "ropus-risk-xgb",
  modelVersion: "v4.2.1",
  policy: "pol_imps_payout_v7",
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
      detail: "Bengaluru → Limassol proxy in 12 minutes (36,250 km/h)",
    },
    {
      id: "f_ml_amount",
      label: "Amount deviates from customer baseline",
      weight: 0.17,
      source: "ML",
      detail: "₹14,50,000 INR vs customer baseline of ₹480 INR",
    },
    {
      id: "f_ti_proxy",
      label: "Login IP on datacenter proxy network",
      weight: 0.14,
      source: "THREAT_INTEL",
      detail: "ASN 13335 · reputation: malicious hosting",
    },
    {
      id: "f_device_new",
      label: "Unrecognised emulator canvas fingerprint",
      weight: 0.1,
      source: "DEVICE",
      detail: "Headless Linux emulator first seen 6 minutes before transfer",
    },
  ],
  rules: [
    {
      id: "RULE-VELOCITY-04",
      name: "Geo velocity threshold",
      outcome: "Exceeded — 2 logins, 7,250 km apart, 12 min (36,250 km/h)",
    },
    {
      id: "RULE-AMT-11",
      name: "Outbound IMPS/RTGS payout ceiling",
      outcome: "Exceeded — ₹14,50,000 INR over ₹2,00,000 tier limit",
    },
    {
      id: "RULE-PAYOUT-02",
      name: "Beneficiary payout account age",
      outcome: "Failed — beneficiary PA-77120 added 9 minutes before payout",
    },
  ],
  evidence: [
    {
      id: "e1",
      kind: "OBSERVED",
      text: "Login from Limassol datacenter IP 198.51.100.44 at 17:00:14 UTC on Linux emulator canvas.",
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
      text: "Outbound IMPS payout of ₹14,50,000 INR submitted at 17:41:09 UTC.",
      source: "RULES",
    },
    {
      id: "e4",
      kind: "OBSERVED",
      text: "Payout account PA-77120 is shared with 13 other synthetic customer accounts.",
      source: "GRAPH",
    },
    {
      id: "e5",
      kind: "INFERRED",
      text: "Session pattern matches account-takeover followed by rapid mule exfiltration.",
    },
    {
      id: "e6",
      kind: "INFERRED",
      text: "Payout account is a confirmed cashout depot for a coordinated mule cluster.",
    },
    {
      id: "e7",
      kind: "RECOMMENDED",
      text: "Block the payout and freeze customer funds pending verification.",
      action: { id: "a_block", label: "Confirm block", destructive: true },
    },
    {
      id: "e8",
      kind: "RECOMMENDED",
      text: "Open Priority P0 investigation case linked to this decision.",
      action: { id: "a_case", label: "Open case" },
    },
    {
      id: "e9",
      kind: "RECOMMENDED",
      text: "Freeze the 13 connected mule accounts sharing payout node PA-77120.",
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
    ip: "198.51.100.44",
    asn: "ASN 13335 — datacenter proxy pool",
    ipReputation: "malicious",
    proxy: true,
    vpn: false,
    tor: false,
    deviceNovelty: "headless emulator, first seen 6 min ago",
  },
  webhook: {
    endpoint: "https://api.razorpay.com/v1/risk/webhooks",
    status: 200,
    attempt: 1,
    timestamp: "2026-08-22T17:41:09.284Z",
  },
  caseId: "CASE-88419",
  rawRequest: {
    transaction_id: "txn_9f3c21ab7d",
    customer_id: "cus_4471029",
    amount: 1450000,
    currency: "INR",
    channel: "imps_payout",
    beneficiary_account: "PA-77120",
    ip: "198.51.100.44",
    device_id: "dev_emulator_linux_9f8a",
  },
};

/** The same customer's baseline activity, before the attack sequence. */
export const approvedDecision: RiskDecision = {
  decisionId: "dec_01HZK1M0PP9C3B",
  transactionId: "txn_2a71fe09c4",
  customerId: "cus_4471029",
  tenantId: DATA_SOURCE.tenantId,
  amount: 480.0,
  currency: "INR",
  verdict: "APPROVE",
  riskScore: 0.04,
  baseScore: 0.02,
  confidence: 0.97,
  model: "ropus-risk-xgb",
  modelVersion: "v4.2.1",
  policy: "pol_upi_purchase_v7",
  latencyMs: 21.7,
  evaluatedAt: "2026-08-22T09:12:44.010Z",
  factors: [
    {
      id: "f_new_merchant",
      label: "Trusted UPI merchant",
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
      text: "UPI purchase of ₹480.00 INR from a trusted device in Bengaluru, KA.",
      source: "DEVICE",
    },
    {
      id: "e2",
      kind: "OBSERVED",
      text: "Device dev_safari_ios_01 has been active on this account for 411 days.",
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
    ip: "49.207.210.14",
    asn: "ASN 55836 — Jio Fiber Residential Bengaluru",
    ipReputation: "clean",
    proxy: false,
    vpn: false,
    tor: false,
    deviceNovelty: "known device, 411 days",
  },
  webhook: {
    endpoint: "https://api.razorpay.com/v1/risk/webhooks",
    status: 200,
    attempt: 1,
    timestamp: "2026-08-22T09:12:44.061Z",
  },
  caseId: null,
  rawRequest: {
    transaction_id: "txn_2a71fe09c4",
    customer_id: "cus_4471029",
    amount: 480.0,
    currency: "INR",
    channel: "upi_purchase",
    ip: "49.207.210.14",
    device_id: "dev_safari_ios_01",
  },
};

export const decisionFixtures: Record<string, RiskDecision> = {
  [blockedDecision.decisionId]: blockedDecision,
  [approvedDecision.decisionId]: approvedDecision,
};
