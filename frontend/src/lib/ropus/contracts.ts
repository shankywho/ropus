/**
 * ROPUS API contracts.
 *
 * These types mirror the shapes returned by the existing backend
 * (POST /v1/risk/evaluate, case + graph endpoints). Fixtures must satisfy
 * these types so components never need to change when the data source swaps
 * from demo fixture to live API.
 */

export type Verdict = "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";

/** Which subsystem produced a risk factor. */
export type FactorSource = "RULES" | "ML" | "THREAT_INTEL" | "GRAPH" | "DEVICE";

/** Explainability tier — required on every piece of evidence. */
export type EvidenceKind = "OBSERVED" | "INFERRED" | "RECOMMENDED";

export interface RiskFactorRecord {
  id: string;
  label: string;
  /** Additive contribution to the risk score. */
  weight: number;
  source: FactorSource;
  detail?: string;
}

export interface TriggeredRule {
  id: string;
  name: string;
  /** Deterministic rules-engine outcome. */
  outcome: string;
}

export interface EvidenceItem {
  id: string;
  kind: EvidenceKind;
  text: string;
  /** Present on OBSERVED items: where the fact was recorded. */
  source?: FactorSource;
  /** Present on RECOMMENDED items: the action a human may authorize. */
  action?: { id: string; label: string; destructive?: boolean };
}

export interface ModelInference {
  model: string;
  version: string;
  probability: number;
  features: Array<{ name: string; contribution: number }>;
}

export interface ThreatIntelSignals {
  ip: string;
  asn: string;
  ipReputation: string;
  proxy: boolean;
  vpn: boolean;
  tor: boolean;
  deviceNovelty: string;
}

export interface WebhookDelivery {
  endpoint: string;
  status: number;
  attempt: number;
  timestamp: string;
}

export interface RiskDecision {
  decisionId: string;
  transactionId: string;
  customerId: string;
  tenantId: string;
  amount: number;
  currency: string;
  verdict: Verdict;
  riskScore: number;
  baseScore: number;
  confidence: number;
  model: string;
  modelVersion: string;
  policy: string;
  latencyMs: number;
  evaluatedAt: string;
  factors: RiskFactorRecord[];
  rules: TriggeredRule[];
  evidence: EvidenceItem[];
  inference: ModelInference;
  threatIntel: ThreatIntelSignals;
  webhook: WebhookDelivery;
  caseId: string | null;
  rawRequest: Record<string, unknown>;
}

/* ------------------------------------------------------------ fraud graph */

export type EntityType = "CUSTOMER" | "DEVICE" | "IP" | "ACCOUNT" | "TRANSACTION";
export type EntityRisk = "CLEAN" | "WATCH" | "SUSPECT" | "CONFIRMED_FRAUD";

export interface GraphEntity {
  id: string;
  type: EntityType;
  label: string;
  risk: EntityRisk;
  /** Deterministic layout coordinates in a 0-100 viewBox. */
  x: number;
  y: number;
  /** Hop distance from the root entity of the investigation. */
  hop: number;
  firstSeen: string;
  lastSeen: string;
  attributes: Array<[string, string]>;
  /** Facts recorded by the platform about this entity. */
  signals: string[];
}

export interface GraphRelationship {
  source: string;
  target: string;
  label: string;
  /** True when this edge participated in the decision under investigation. */
  onDecisionPath?: boolean;
}

export interface FraudGraph {
  rootId: string;
  decisionId: string;
  entities: GraphEntity[];
  relationships: GraphRelationship[];
}

/* ------------------------------------------------------------------ cases */

export type CaseStatus = "OPEN" | "IN_REVIEW" | "ESCALATED" | "CLOSED";
export type CasePriority = "P1" | "P2" | "P3" | "P4";
export type CaseOutcome = "CONFIRMED_FRAUD" | "RELEASED" | "CHALLENGED" | null;
/** Who wrote the timeline entry. AI entries are always labelled as such. */
export type CaseActorKind = "SYSTEM" | "ANALYST" | "AI";

export interface CaseEvent {
  id: string;
  at: string;
  actor: string;
  actorKind: CaseActorKind;
  text: string;
}

export interface CaseRecord {
  caseId: string;
  decisionId: string;
  transactionId: string;
  customerId: string;
  amount: number;
  currency: string;
  verdict: Verdict;
  riskScore: number;
  priority: CasePriority;
  status: CaseStatus;
  outcome: CaseOutcome;
  assignee: string | null;
  openedBy: string;
  openedAt: string;
  updatedAt: string;
  /** Minutes remaining against the tenant SLA; negative when breached. */
  slaMinutesRemaining: number;
  summary: string;
  timeline: CaseEvent[];
}

/* --------------------------------------------------- list + overview reads */

/** Row shape returned by the decision list endpoint. */
export interface DecisionListItem {
  decisionId: string;
  transactionId: string;
  customerId: string;
  amount: number;
  currency: string;
  verdict: Verdict;
  riskScore: number;
  primarySignal: string;
  latencyMs: number;
  caseId: string | null;
  evaluatedAt: string;
}

export type ServiceState = "HEALTHY" | "DEGRADED" | "UNAVAILABLE";

export interface ServiceHealth {
  name: string;
  state: ServiceState;
  p99Ms: number;
  detail: string;
}

export interface OverviewMetrics {
  windowLabel: string;
  evaluations: number;
  blockRate: number;
  reviewRate: number;
  p99LatencyMs: number;
  openCases: number;
  distribution: Array<{ verdict: Verdict; count: number }>;
  services: ServiceHealth[];
}
