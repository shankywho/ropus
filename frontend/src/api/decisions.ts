import { HttpClient, defaultHttpClient } from "./client";

export type RecommendedAction =
  | "ALLOW_RECOMMENDATION"
  | "STEP_UP_RECOMMENDATION"
  | "MANUAL_REVIEW"
  | "HOLD_RECOMMENDATION"
  | "DECLINE_RECOMMENDATION"
  | "SHADOW_ONLY"
  | "INSUFFICIENT_CONTEXT";

export interface RiskEvaluationRequest {
  transaction_id?: string;
  amount: number;
  currency: string;
  payment_method?: {
    type: string;
    token: string;
  };
  device_fingerprint?: string;
  ip_address?: string;
  account_id?: string;
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

export class DecisionsApi {
  constructor(private client: HttpClient = defaultHttpClient) {}

  async evaluate(payload: RiskEvaluationRequest): Promise<RiskEvaluationResponse> {
    return this.client.request<RiskEvaluationResponse>("/v1/risk-evaluations", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  }
}

export const decisionsApi = new DecisionsApi();
