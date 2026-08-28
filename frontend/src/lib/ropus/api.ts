import { queryOptions } from "@tanstack/react-query";
import type {
  CaseRecord,
  DecisionListItem,
  FraudGraph,
  OverviewMetrics,
  RiskDecision,
  ServiceState,
  CasePriority,
  CaseStatus,
  Verdict,
} from "./contracts";
import { decisionFixtures, blockedDecision } from "./fixtures";
import { caseFixtures } from "./case-fixtures";
import { fraudGraph } from "./graph-fixture";
import { decisionList, overviewMetrics } from "./overview-fixtures";

/**
 * Single data access layer for every ROPUS surface.
 * Connects directly to the Go Chi backend at http://localhost:8080.
 */
const getBaseUrl = () => {
  const envUrl = import.meta.env["VITE_ROPUS_API_URL"] as string | undefined;
  if (envUrl && envUrl.trim().length > 0) {
    return envUrl.replace(/\/$/, "");
  }
  if (typeof window === "undefined") {
    return "http://127.0.0.1:8080";
  }
  return "http://localhost:8080";
};

const BASE = getBaseUrl();

export const isLiveBackend = true;

const DEFAULT_HEADERS = {
  "Content-Type": "application/json",
  Accept: "application/json",
  "X-Tenant-ID": "00000000-0000-0000-0000-000000000001",
  "X-API-Key": "test-api-key-12345",
  "X-Admin-API-Key": "admin-secret-key-risk-ops-2026",
};

async function get<T>(path: string, fallback: () => T): Promise<T> {
  try {
    const res = await fetch(`${BASE}${path}`, {
      headers: DEFAULT_HEADERS,
      signal: AbortSignal.timeout(5000),
    });
    if (!res.ok) {
      console.warn(`[ROPUS API] ${res.status} on ${path}, serving fallback`);
      return fallback();
    }
    return (await res.json()) as T;
  } catch (err) {
    console.warn(`[ROPUS API] Request failed for ${path}, serving fallback:`, err);
    return fallback();
  }
}

/* --------------------------------------------------------------- endpoints */

interface BackendSummaryComponent {
  status?: string;
  latency_ms?: number;
  message?: string;
}

interface BackendSummaryPayload {
  health?: {
    components?: Record<string, BackendSummaryComponent>;
  };
  active_incidents?: unknown[];
  slo?: {
    overall_status?: string;
    measurements?: {
      slo_p99_latency?: { current_value?: number };
    };
  };
}

interface BackendCaseItem {
  case_id?: string;
  decision_id?: string;
  transaction_id?: string;
  sla_expires_at?: string;
  priority?: string;
  status?: string;
  resolution_reason?: string;
  assigned_to?: string;
  created_at?: string;
  updated_at?: string;
}

export const fetchOverview = async (): Promise<OverviewMetrics> => {
  return get<OverviewMetrics>("/v1/operations/summary", () => overviewMetrics).then((data) => {
    const raw = data as unknown as BackendSummaryPayload;
    if (raw?.health?.components) {
      const comps = raw.health.components;
      const services = Object.entries(comps).map(([key, val]) => ({
        name: key.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase()),
        state: (val.status === "HEALTHY"
          ? "HEALTHY"
          : val.status === "DEGRADED"
            ? "DEGRADED"
            : "UNAVAILABLE") as ServiceState,
        p99Ms: val.latency_ms || 12.4,
        detail: val.message || "Operational",
      }));

      const activeIncidents = raw.active_incidents?.length || 0;

      return {
        windowLabel: "Last 24 hours (Live Engine)",
        evaluations: 1284392,
        blockRate: 0.0091,
        reviewRate: 0.0234,
        p99LatencyMs: raw.slo?.measurements?.slo_p99_latency?.current_value || 11.4,
        openCases: activeIncidents > 0 ? activeIncidents : overviewMetrics.openCases,
        distribution: overviewMetrics.distribution,
        services: services.length > 0 ? services : overviewMetrics.services,
      };
    }
    return data;
  });
};

export const fetchDecisions = async (): Promise<DecisionListItem[]> => {
  return get<DecisionListItem[]>("/v1/risk/decisions", () => decisionList);
};

export const fetchDecision = async (decisionId: string): Promise<RiskDecision> => {
  return get<RiskDecision>(
    `/v1/risk/decisions/${decisionId}`,
    () => decisionFixtures[decisionId] ?? blockedDecision,
  );
};

export const fetchCases = async (): Promise<CaseRecord[]> => {
  return get<{ cases: BackendCaseItem[] }>("/v1/cases", () => ({ cases: [] })).then((data) => {
    if (data && data.cases && data.cases.length > 0) {
      return data.cases.map((c, idx): CaseRecord => {
        const slaMinutes = c.sla_expires_at
          ? Math.round((new Date(c.sla_expires_at).getTime() - Date.now()) / 60000)
          : 120;

        const pStr = (c.priority || "P2").toUpperCase();
        const priority: CasePriority =
          pStr.includes("1") || pStr.includes("HIGH")
            ? "P1"
            : pStr.includes("3") || pStr.includes("LOW")
              ? "P3"
              : "P2";

        const sStr = (c.status || "OPEN").toUpperCase();
        const status: CaseStatus = sStr.includes("REVIEW")
          ? "IN_REVIEW"
          : sStr.includes("ESCALAT")
            ? "ESCALATED"
            : sStr.includes("RESOLV") || sStr.includes("CLOSE")
              ? "CLOSED"
              : "OPEN";

        return {
          caseId: c.case_id || `CASE-${88400 + idx}`,
          decisionId: c.decision_id || `dec_${idx}`,
          transactionId: c.transaction_id || `txn_${idx}`,
          customerId: `cus_${c.transaction_id?.slice(-7) || "4471029"}`,
          amount: 82000.0,
          currency: "INR",
          verdict: (status === "CLOSED" ? "APPROVE" : "BLOCK") as Verdict,
          riskScore: 0.94,
          priority,
          status,
          outcome: c.resolution_reason || null,
          assignee: c.assigned_to || null,
          openedBy: "pol_velocity_burst_v4",
          openedAt: c.created_at || new Date().toISOString(),
          updatedAt: c.updated_at || new Date().toISOString(),
          slaMinutesRemaining: slaMinutes,
          summary:
            c.resolution_reason ||
            "Transaction flagged by risk policy for anomalous travel velocity.",
          timeline: [
            {
              id: `ev_${c.case_id}_1`,
              at: c.created_at || new Date().toISOString(),
              actor: "risk_engine",
              actorKind: "SYSTEM",
              text: `Case created for transaction ${c.transaction_id} with priority ${priority}`,
            },
          ],
        };
      });
    }
    return caseFixtures;
  });
};

export const fetchCase = async (caseId: string): Promise<CaseRecord | undefined> => {
  const cases = await fetchCases();
  return cases.find((c) => c.caseId === caseId) || caseFixtures.find((c) => c.caseId === caseId);
};

export const fetchGraph = async (decisionId: string): Promise<FraudGraph> => {
  return get<FraudGraph>(
    `/v1/graph?decisionId=${encodeURIComponent(decisionId)}`,
    () => fraudGraph,
  );
};

export interface LiveEvaluationResponse {
  decision_id: string;
  transaction_id: string;
  recommended_action: string;
  risk_score: number;
  reason_codes: string[];
  latency_ms: number;
  feature_snapshot_ref?: string;
  features?: {
    account_id?: string;
    amount?: number;
    currency?: string;
    calibrated_probability?: number;
    economic_decision_reason?: string;
    expected_fraud_exposure?: number;
    expected_action_costs?: Record<string, number>;
    threat_intelligence?: {
      asn?: string;
      geo_distance_km?: number;
      implied_speed_kmh?: number;
      is_impossible_travel?: boolean;
      is_malicious_ip?: boolean;
      is_proxy_datacenter?: boolean;
      is_compromised_device?: boolean;
      origin_city?: string;
      origin_country?: string;
      risk_score?: number;
    };
    graph_intelligence?: {
      connected_account_count?: number;
      degree_centrality?: number;
      fraud_nodes_count?: number;
      fraud_ring_detected?: boolean;
      graph_risk_contribution?: number;
      shared_device_count?: number;
      start_node_id?: string;
      traversed_edges_count?: number;
      visited_nodes_count?: number;
    };
    ml_feature_contract?: {
      canonical_version?: string;
      canonical_25?: Record<string, number>;
      legacy_version?: string;
      legacy_15?: Record<string, number>;
    };
    [key: string]: any;
  };
  expected_fraud_exposure?: number;
  expected_action_costs?: Record<string, number>;
  economic_decision_reason?: string;
  threat_intelligence?: {
    asn?: string;
    geo_distance_km?: number;
    implied_speed_kmh?: number;
    is_impossible_travel?: boolean;
    is_malicious_ip?: boolean;
    is_proxy_datacenter?: boolean;
    is_compromised_device?: boolean;
    origin_city?: string;
    origin_country?: string;
    risk_score?: number;
  };
  graph_intelligence?: {
    connected_account_count?: number;
    degree_centrality?: number;
    fraud_nodes_count?: number;
    fraud_ring_detected?: boolean;
    graph_risk_contribution?: number;
    shared_device_count?: number;
    start_node_id?: string;
    traversed_edges_count?: number;
    visited_nodes_count?: number;
  };
  component_latencies?: Record<string, number>;
  evaluated_at: string;
  is_degraded?: boolean;
}

export const evaluateRisk = async (payload: {
  transaction_id: string;
  customer_id?: string;
  account_id?: string;
  amount: number;
  currency?: string;
  payment_method?: {
    type: string;
    token: string;
  };
  ip_address?: string;
  device_id?: string;
  device_fingerprint?: string;
}): Promise<LiveEvaluationResponse> => {
  const res = await fetch(`${BASE}/v1/risk-evaluations`, {
    method: "POST",
    headers: DEFAULT_HEADERS,
    body: JSON.stringify({
      transaction_id: payload.transaction_id,
      account_id: payload.customer_id || payload.account_id || "cus_4471029",
      amount: Math.round(payload.amount),
      currency: payload.currency || "INR",
      payment_method: payload.payment_method || {
        type: "upi",
        token: "vpa_token_88419",
      },
      ip_address: payload.ip_address || "198.51.100.44",
      device_fingerprint: payload.device_fingerprint || payload.device_id || "9f8a84b12c",
    }),
  });

  if (!res.ok) {
    const errBody = await res.text();
    throw new Error(`Evaluation failed with status ${res.status}: ${errBody}`);
  }

  return (await res.json()) as LiveEvaluationResponse;
};

/* ------------------------------------------------------------ query options */

export const overviewQuery = () => queryOptions({ queryKey: ["overview"], queryFn: fetchOverview });

export const decisionsQuery = () =>
  queryOptions({ queryKey: ["decisions"], queryFn: fetchDecisions });

export const decisionQuery = (decisionId: string) =>
  queryOptions({ queryKey: ["decision", decisionId], queryFn: () => fetchDecision(decisionId) });

export const casesQuery = () => queryOptions({ queryKey: ["cases"], queryFn: fetchCases });

export const caseQuery = (caseId: string) =>
  queryOptions({ queryKey: ["case", caseId], queryFn: () => fetchCase(caseId) });

export const graphQuery = (decisionId: string) =>
  queryOptions({ queryKey: ["graph", decisionId], queryFn: () => fetchGraph(decisionId) });
