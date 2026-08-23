import { queryOptions } from "@tanstack/react-query";
import type {
  CaseRecord,
  DecisionListItem,
  FraudGraph,
  OverviewMetrics,
  RiskDecision,
} from "./contracts";
import { decisionFixtures, blockedDecision } from "./fixtures";
import { caseFixtures } from "./case-fixtures";
import { fraudGraph } from "./graph-fixture";
import { decisionList, overviewMetrics } from "./overview-fixtures";

/**
 * Single data access layer for every ROPUS surface.
 *
 * When VITE_ROPUS_API_URL is set the app reads the real backend — the backend
 * is the source of truth and its response shapes are the contracts in
 * ./contracts. Without it the app serves the deterministic demo fixtures and
 * every screen renders the DEMO DATA tag.
 */
const BASE = (import.meta.env["VITE_ROPUS_API_URL"] as string | undefined)?.replace(/\/$/, "");

export const isLiveBackend = Boolean(BASE);

async function get<T>(path: string, fallback: () => T): Promise<T> {
  if (!BASE) return fallback();
  const res = await fetch(`${BASE}${path}`, { headers: { accept: "application/json" } });
  if (!res.ok) throw new Error(`ROPUS API ${res.status} on ${path}`);
  return (await res.json()) as T;
}

/* --------------------------------------------------------------- endpoints */

export const fetchOverview = () =>
  get<OverviewMetrics>("/v1/metrics/overview", () => overviewMetrics);

export const fetchDecisions = () =>
  get<DecisionListItem[]>("/v1/risk/decisions", () => decisionList);

export const fetchDecision = (decisionId: string) =>
  get<RiskDecision>(`/v1/risk/decisions/${decisionId}`, () => decisionFixtures[decisionId] ?? blockedDecision);

export const fetchCases = () => get<CaseRecord[]>("/v1/cases", () => caseFixtures);

export const fetchCase = (caseId: string) =>
  get<CaseRecord | undefined>(`/v1/cases/${caseId}`, () => caseFixtures.find((c) => c.caseId === caseId));

export const fetchGraph = (decisionId: string) =>
  get<FraudGraph>(`/v1/graph?decisionId=${encodeURIComponent(decisionId)}`, () => fraudGraph);

/* ------------------------------------------------------------ query options */

export const overviewQuery = () =>
  queryOptions({ queryKey: ["overview"], queryFn: fetchOverview });

export const decisionsQuery = () =>
  queryOptions({ queryKey: ["decisions"], queryFn: fetchDecisions });

export const decisionQuery = (decisionId: string) =>
  queryOptions({ queryKey: ["decision", decisionId], queryFn: () => fetchDecision(decisionId) });

export const casesQuery = () => queryOptions({ queryKey: ["cases"], queryFn: fetchCases });

export const caseQuery = (caseId: string) =>
  queryOptions({ queryKey: ["case", caseId], queryFn: () => fetchCase(caseId) });

export const graphQuery = (decisionId: string) =>
  queryOptions({ queryKey: ["graph", decisionId], queryFn: () => fetchGraph(decisionId) });
