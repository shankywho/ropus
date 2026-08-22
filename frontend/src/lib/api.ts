// Type definitions matching Go backend contracts

export type RecommendedAction =
  | "ALLOW_RECOMMENDATION"
  | "STEP_UP_RECOMMENDATION"
  | "MANUAL_REVIEW"
  | "HOLD_RECOMMENDATION"
  | "DECLINE_RECOMMENDATION"
  | "SHADOW_ONLY"
  | "INSUFFICIENT_CONTEXT";

export type RuleStatus =
  | "DRAFT"
  | "PENDING_APPROVAL"
  | "SHADOW"
  | "ACTIVE"
  | "ARCHIVED";

export type CaseStatus =
  | "OPEN"
  | "UNDER_REVIEW"
  | "RESOLVED_ALLOW"
  | "RESOLVED_DECLINE"
  | "CLOSED";

export interface RiskEvaluationRequest {
  transaction_id?: string;
  amount: number;
  currency: string;
  payment_method: {
    type: string;
    token: string;
  };
  device_fingerprint: string;
  ip_address?: string;
}

export interface RiskEvaluationResponse {
  decision_id: string;
  transaction_id: string;
  recommended_action: RecommendedAction;
  risk_score: number;
  reason_codes: string[];
  feature_snapshot_ref: string;
  features: Record<string, any>;
  evaluated_at: string;
  is_degraded?: boolean;
  latency_ms: number;
}

export interface Rule {
  rule_id: string;
  tenant_id: string;
  name: string;
  description: string;
  dsl_ast: Record<string, any>;
  status: RuleStatus;
  version: number;
  created_by: string;
  approved_by?: string;
  created_at: string;
  updated_at: string;
}

export interface CaseItem {
  case_id: string;
  tenant_id: string;
  decision_id: string;
  transaction_id: string;
  status: CaseStatus;
  priority: string;
  assigned_to?: string;
  resolution_reason?: string;
  resolved_at?: string;
  sla_expires_at: string;
  created_at: string;
  updated_at: string;
}

export interface CaseDetail extends CaseItem {
  amount: number;
  currency: string;
  risk_score: number;
  recommended_action: RecommendedAction;
  reason_codes: string[];
  feature_snapshot: Record<string, any>;
  raw_payload: Record<string, any>;
}

export interface HealthStatus {
  status: string;
  database: string;
  redis: string;
  clickhouse?: string;
  kms?: string;
  ml_service?: string;
  kafka_brokers?: string[];
  kafka_topic?: string;
  time: string;
}

// Base URL resolution (browser client vs server component)
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-Tenant-ID": "00000000-0000-0000-0000-000000000001",
      ...(options.headers as Record<string, string>),
    };

    const response = await fetch(url, {
      ...options,
      headers,
      cache: "no-store",
    });

    if (!response.ok) {
      let errorMsg = `HTTP Error ${response.status}`;
      try {
        const errorData = await response.json();
        errorMsg = errorData.message || errorData.error || errorMsg;
      } catch {
        // use default error message
      }
      throw new Error(errorMsg);
    }

    return response.json();
  }

  // 1. Health Status
  async getHealth(): Promise<HealthStatus> {
    return this.request<HealthStatus>("/health");
  }

  // 2. Risk Evaluation
  async evaluateRisk(
    payload: RiskEvaluationRequest
  ): Promise<RiskEvaluationResponse> {
    return this.request<RiskEvaluationResponse>("/v1/risk-evaluations", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  }

  // 3. Rules Engine API
  async getRules(status?: RuleStatus): Promise<{ rules: Rule[]; count: number }> {
    const query = status ? `?status=${status}` : "";
    return this.request<{ rules: Rule[]; count: number }>(`/v1/rules${query}`);
  }

  async getRule(id: string): Promise<Rule> {
    return this.request<Rule>(`/v1/rules/${id}`);
  }

  async createRule(payload: {
    name: string;
    description: string;
    dsl_ast: Record<string, any>;
    actorId?: string;
  }): Promise<Rule> {
    return this.request<Rule>("/v1/rules", {
      method: "POST",
      headers: payload.actorId ? { "X-Actor-ID": payload.actorId } : {},
      body: JSON.stringify({
        name: payload.name,
        description: payload.description,
        dsl_ast: payload.dsl_ast,
      }),
    });
  }

  async updateRule(
    id: string,
    payload: {
      name: string;
      description: string;
      dsl_ast: Record<string, any>;
      actorId?: string;
    }
  ): Promise<Rule> {
    return this.request<Rule>(`/v1/rules/${id}`, {
      method: "PUT",
      headers: payload.actorId ? { "X-Actor-ID": payload.actorId } : {},
      body: JSON.stringify(payload),
    });
  }

  async transitionRule(
    id: string,
    status: RuleStatus,
    actorId: string = "analyst_2"
  ): Promise<Rule> {
    return this.request<Rule>(`/v1/rules/${id}/status`, {
      method: "PUT",
      headers: { "X-Actor-ID": actorId },
      body: JSON.stringify({ status }),
    });
  }

  // 4. Case Management API
  async getCases(status?: CaseStatus): Promise<{ cases: CaseItem[]; count: number }> {
    const query = status ? `?status=${status}` : "";
    return this.request<{ cases: CaseItem[]; count: number }>(`/v1/cases${query}`);
  }

  async getCase(id: string): Promise<CaseDetail> {
    return this.request<CaseDetail>(`/v1/cases/${id}`);
  }

  async claimCase(id: string, analystId: string = "analyst_sarah"): Promise<CaseItem> {
    return this.request<CaseItem>(`/v1/cases/${id}/claim`, {
      method: "PUT",
      headers: { "X-Actor-ID": analystId },
    });
  }

  async resolveCase(
    id: string,
    action: "ALLOW" | "DECLINE",
    reason: string,
    analystId: string = "analyst_sarah"
  ): Promise<CaseItem> {
    return this.request<CaseItem>(`/v1/cases/${id}/resolve`, {
      method: "PUT",
      headers: { "X-Actor-ID": analystId },
      body: JSON.stringify({ action, reason }),
    });
  }
}

export const api = new ApiClient(API_BASE_URL);
