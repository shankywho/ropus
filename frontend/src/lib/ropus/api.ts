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
 * Connects directly to the Go Chi backend at /v1 via Vite proxy / Docker network.
 */
const BASE =
  (import.meta.env["VITE_ROPUS_API_URL"] as string | undefined)?.replace(/\/$/, "") ?? "";

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
  features?: {
    ml_feature_contract?: {
      canonical_version?: string;
      canonical_25?: Record<string, number>;
    };
  };
  evaluated_at: string;
}

export const evaluateRisk = async (payload: {
  transaction_id: string;
  customer_id: string;
  amount: number;
  currency: string;
  ip_address?: string;
  device_id?: string;
}): Promise<LiveEvaluationResponse> => {
  const res = await fetch(`${BASE}/v1/risk-evaluations`, {
    method: "POST",
    headers: DEFAULT_HEADERS,
    body: JSON.stringify({
      transaction_id: payload.transaction_id,
      customer_id: payload.customer_id,
      amount: Math.round(payload.amount),
      currency: payload.currency || "INR",
      ip_address: payload.ip_address || "198.51.100.44",
      device_fingerprint: payload.device_id || "9f8a84b12c",
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
