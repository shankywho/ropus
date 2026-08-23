import { HttpClient, defaultHttpClient } from "./client";

export type RuleStatus =
  | "DRAFT"
  | "PENDING_APPROVAL"
  | "SHADOW"
  | "ACTIVE"
  | "ARCHIVED";

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

export class RulesApi {
  constructor(private client: HttpClient = defaultHttpClient) {}

  async list(status?: RuleStatus): Promise<{ rules: Rule[]; count: number }> {
    const query = status ? `?status=${status}` : "";
    return this.client.request<{ rules: Rule[]; count: number }>(`/v1/rules${query}`);
  }

  async get(id: string): Promise<Rule> {
    return this.client.request<Rule>(`/v1/rules/${id}`);
  }

  async create(payload: {
    name: string;
    description: string;
    dsl_ast: Record<string, any>;
    actorId?: string;
  }): Promise<Rule> {
    return this.client.request<Rule>(
      "/v1/rules",
      {
        method: "POST",
        body: JSON.stringify({
          name: payload.name,
          description: payload.description,
          dsl_ast: payload.dsl_ast,
        }),
      },
      payload.actorId ? { "X-Actor-ID": payload.actorId } : undefined
    );
  }

  async update(
    id: string,
    payload: {
      name: string;
      description: string;
      dsl_ast: Record<string, any>;
      actorId?: string;
    }
  ): Promise<Rule> {
    return this.client.request<Rule>(
      `/v1/rules/${id}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
      payload.actorId ? { "X-Actor-ID": payload.actorId } : undefined
    );
  }

  async transitionStatus(
    id: string,
    status: RuleStatus,
    actorId: string = "analyst_operator"
  ): Promise<Rule> {
    return this.client.request<Rule>(
      `/v1/rules/${id}/status`,
      {
        method: "PUT",
        body: JSON.stringify({ status }),
      },
      { "X-Actor-ID": actorId }
    );
  }
}

export const rulesApi = new RulesApi();
