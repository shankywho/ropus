import { HttpClient, defaultHttpClient } from "./client";
import { RecommendedAction } from "./decisions";

export type CaseStatus =
  | "OPEN"
  | "UNDER_REVIEW"
  | "RESOLVED_ALLOW"
  | "RESOLVED_DECLINE"
  | "CLOSED";

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

export class CasesApi {
  constructor(private client: HttpClient = defaultHttpClient) {}

  async list(status?: CaseStatus): Promise<{ cases: CaseItem[]; count: number }> {
    const query = status ? `?status=${status}` : "";
    return this.client.request<{ cases: CaseItem[]; count: number }>(`/v1/cases${query}`);
  }

  async get(id: string): Promise<CaseDetail> {
    return this.client.request<CaseDetail>(`/v1/cases/${id}`);
  }

  async claim(id: string, analystId: string = "lead_analyst"): Promise<CaseItem> {
    return this.client.request<CaseItem>(
      `/v1/cases/${id}/claim`,
      { method: "PUT" },
      { "X-Actor-ID": analystId }
    );
  }

  async resolve(
    id: string,
    action: "ALLOW" | "DECLINE",
    reason: string,
    analystId: string = "lead_analyst"
  ): Promise<CaseItem> {
    return this.client.request<CaseItem>(
      `/v1/cases/${id}/resolve`,
      {
        method: "PUT",
        body: JSON.stringify({ action, reason }),
      },
      { "X-Actor-ID": analystId }
    );
  }
}

export const casesApi = new CasesApi();
