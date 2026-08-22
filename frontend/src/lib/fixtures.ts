export interface RiskFactor {
  name: string;
  source: "Rules" | "ML" | "Threat Intel" | "Graph" | "Device";
  contribution: number;
  description: string;
}

export interface TriggeredRule {
  id: string;
  name: string;
  priority: number;
  action: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";
  reason: string;
}

export interface WebhookDeliveryRecord {
  id: string;
  endpoint: string;
  event: string;
  status: "DELIVERED" | "RETRYING" | "FAILED";
  statusCode: number;
  latencyMs: number;
  timestamp: string;
  signature: string;
}

export interface RiskDecisionPayload {
  transactionId: string;
  customerId: string;
  amount: number;
  currency: string;
  timestamp: string;
  decisionId: string;
  tenantId: string;
  verdict: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";
  riskScore: number;
  confidence: number;
  latencyMs: number;
  modelVersion: string;
  policyVersion: string;
  recommendation: "ALLOW" | "MANUAL_REVIEW" | "STEP_UP_MFA" | "BLOCK_AND_REVIEW";
  
  riskFactors: RiskFactor[];
  triggeredRules: TriggeredRule[];
  
  evidence: {
    observed: string[];
    inferred: string[];
    recommended: string[];
  };
  
  technicalDetails: {
    rawPayload: Record<string, any>;
    mlFeatures: Record<string, number | string>;
    threatIntel: {
      ip: string;
      asn: string;
      isProxy: boolean;
      isVpn: boolean;
      country: string;
      geoVelocityKmh: number;
      distanceKm: number;
    };
    graphSnapshot: {
      rootId: string;
      connectedNodes: number;
      degreeCentrality: number;
      syndicateClusterDetected: boolean;
    };
    webhookDelivery: WebhookDeliveryRecord;
    auditHashChain: string;
  };
}

export const CANONICAL_BLOCKED_DECISION: RiskDecisionPayload = {
  transactionId: "tx_order_88419",
  customerId: "usr_sarah_connor",
  amount: 14500.0,
  currency: "USD",
  timestamp: "2026-08-22T17:42:00Z",
  decisionId: "dec_9f8a1e9c7a2b",
  tenantId: "org_prod_us_east",
  verdict: "BLOCK",
  riskScore: 0.96,
  confidence: 0.985,
  latencyMs: 1.42,
  modelVersion: "xgb_fraud_v3.9_25f",
  policyVersion: "pol_v2_strict_banking",
  recommendation: "BLOCK_AND_REVIEW",
  riskFactors: [
    {
      name: "High Value Velocity Spike",
      source: "Rules",
      contribution: 0.22,
      description: "Amount $14,500 exceeds 30-day baseline moving average by 412%",
    },
    {
      name: "Impossible Travel Velocity",
      source: "Threat Intel",
      contribution: 0.21,
      description: "Egress from Limassol (CY) 18 mins after New York session (8,420 km/h)",
    },
    {
      name: "XGBoost 25-Feature Probability",
      source: "ML",
      contribution: 0.20,
      description: "Gradient boosted ensemble computed 0.982 posterior fraud probability",
    },
    {
      name: "Bulletproof Proxy Subnet Match",
      source: "Threat Intel",
      contribution: 0.18,
      description: "IP 198.51.100.44 classified in datacenter VPN anonymizer CIDR",
    },
    {
      name: "Syndicate Graph Exposure",
      source: "Graph",
      contribution: 0.17,
      description: "Hardware canvas fingerprint hash shared with 14 previously disputed accounts",
    },
  ],
  triggeredRules: [
    {
      id: "RULE-VELOCITY-04",
      name: "Excessive Monetary Outflow Window",
      priority: 90,
      action: "BLOCK",
      reason: "Tx amount > $10,000 within 24h of device change",
    },
    {
      id: "RULE-GEO-02",
      name: "Transcontinental Velocity Threshold",
      priority: 85,
      action: "CHALLENGE",
      reason: "Speed between successive locations > 900 km/h",
    },
  ],
  evidence: {
    observed: [
      "Monetary amount: $14,500.00 USD wire transfer request.",
      "IP Address: 198.51.100.44 (ISP: Datacenter Hosting Corp, Limassol CY).",
      "Device Canvas Hash: cvs_mule_cluster_99 (WebGL / Canvas entropy 0.94).",
      "Prior Session: 2026-08-22T17:24:00Z from 72.229.28.185 (New York, US).",
    ],
    inferred: [
      "AI INFERRED: Account Takeover (ATO) attack followed by immediate liquidity drain.",
      "AI INFERRED: Hardware canvas fingerprint matches automated mule syndicate tooling.",
      "AI INFERRED: Egress proxy intentionally configured to bypass IP velocity alarms.",
    ],
    recommended: [
      "Freeze pending outbound wire settlement with clearing network.",
      "Quarantine active session and invalidate all OAuth / refresh tokens.",
      "Enforce mandatory step-up biometric verification on customer profile.",
      "Auto-generate FinCEN Suspicious Activity Report (SAR) evidentiary draft.",
    ],
  },
  technicalDetails: {
    rawPayload: {
      transaction_id: "tx_order_88419",
      customer_id: "usr_sarah_connor",
      amount: 14500.0,
      currency: "USD",
      device_id: "dev_mule_cluster_99",
      ip_address: "198.51.100.44",
      country: "CY",
      timestamp: "2026-08-22T17:42:00Z",
    },
    mlFeatures: {
      amount_usd: 14500.0,
      velocity_10m: 3,
      velocity_1h: 6,
      amount_velocity_24h: 18200.0,
      device_entropy: 0.942,
      is_emulator_flag: 1.0,
      is_vpn_proxy: 1.0,
      geo_velocity_kmh: 8420.5,
      graph_degree_centrality: 14,
      graph_pagerank: 0.892,
    },
    threatIntel: {
      ip: "198.51.100.44",
      asn: "AS13335 (Bulletproof Datacenter)",
      isProxy: true,
      isVpn: true,
      country: "CY (Cyprus)",
      geoVelocityKmh: 8420.5,
      distanceKm: 7600.2,
    },
    graphSnapshot: {
      rootId: "usr_sarah_connor",
      connectedNodes: 18,
      degreeCentrality: 14,
      syndicateClusterDetected: true,
    },
    webhookDelivery: {
      id: "wh_evt_88419",
      endpoint: "https://api.acmebank.internal/v1/fraud-alerts",
      event: "risk.decision.blocked",
      status: "DELIVERED",
      statusCode: 200,
      latencyMs: 38,
      timestamp: "2026-08-22T17:42:01.458Z",
      signature: "sha256=4f8a1e9c7a2b5d8e3f1c...e91c",
    },
    auditHashChain: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  },
};

export const CANONICAL_APPROVED_DECISION: RiskDecisionPayload = {
  transactionId: "tx_order_88418",
  customerId: "usr_sarah_connor",
  amount: 42.5,
  currency: "USD",
  timestamp: "2026-08-22T14:10:00Z",
  decisionId: "dec_7a1b2c3d4e5f",
  tenantId: "org_prod_us_east",
  verdict: "APPROVE",
  riskScore: 0.04,
  confidence: 0.994,
  latencyMs: 1.15,
  modelVersion: "xgb_fraud_v3.9_25f",
  policyVersion: "pol_v2_strict_banking",
  recommendation: "ALLOW",
  riskFactors: [
    {
      name: "Organic Amount Pattern",
      source: "Rules",
      contribution: 0.01,
      description: "Amount $42.50 matches recurring grocery merchant pattern",
    },
    {
      name: "Trusted Hardware Canvas",
      source: "Device",
      contribution: 0.01,
      description: "Hardware canvas fingerprint observed across 180 days with zero disputes",
    },
    {
      name: "Residential ISP Egress",
      source: "Threat Intel",
      contribution: 0.01,
      description: "IP verified residential Verizon Fios New York subnet",
    },
    {
      name: "Baseline ML Prior",
      source: "ML",
      contribution: 0.01,
      description: "XGBoost ensemble evaluated benign baseline probability of 0.008",
    },
  ],
  triggeredRules: [],
  evidence: {
    observed: [
      "Monetary amount: $42.50 USD point-of-sale merchant transaction.",
      "IP Address: 72.229.28.185 (ISP: Verizon Fios, New York US).",
      "Device Fingerprint: dev_macbook_sarah (Verified hardware canvas).",
    ],
    inferred: [
      "AI INFERRED: Normal organic consumer purchasing behavior.",
    ],
    recommended: [
      "Allow transaction for immediate settlement authorization.",
    ],
  },
  technicalDetails: {
    rawPayload: {
      transaction_id: "tx_order_88418",
      customer_id: "usr_sarah_connor",
      amount: 42.5,
      currency: "USD",
      device_id: "dev_macbook_sarah",
      ip_address: "72.229.28.185",
      country: "US",
      timestamp: "2026-08-22T14:10:00Z",
    },
    mlFeatures: {
      amount_usd: 42.5,
      velocity_10m: 1,
      velocity_1h: 1,
      amount_velocity_24h: 42.5,
      device_entropy: 0.12,
      is_emulator_flag: 0.0,
      is_vpn_proxy: 0.0,
      geo_velocity_kmh: 0.0,
      graph_degree_centrality: 1,
      graph_pagerank: 0.02,
    },
    threatIntel: {
      ip: "72.229.28.185",
      asn: "AS701 (Verizon Residential)",
      isProxy: false,
      isVpn: false,
      country: "US (United States)",
      geoVelocityKmh: 0.0,
      distanceKm: 0.0,
    },
    graphSnapshot: {
      rootId: "usr_sarah_connor",
      connectedNodes: 2,
      degreeCentrality: 1,
      syndicateClusterDetected: false,
    },
    webhookDelivery: {
      id: "wh_evt_88418",
      endpoint: "https://api.acmebank.internal/v1/fraud-alerts",
      event: "risk.decision.approved",
      status: "DELIVERED",
      statusCode: 200,
      latencyMs: 24,
      timestamp: "2026-08-22T14:10:01.012Z",
      signature: "sha256=1a2b3c4d5e...f9a0",
    },
    auditHashChain: "d41d8cd98f00b204e9800998ecf8427e",
  },
};
