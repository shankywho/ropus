import { HttpClient, defaultHttpClient } from "./client";

export interface ComponentHealth {
  name: string;
  status: "HEALTHY" | "DEGRADED" | "UNHEALTHY";
  latency_ms: number;
  message?: string;
  last_checked: string;
}

export interface HealthReport {
  overall_status: "HEALTHY" | "DEGRADED" | "UNHEALTHY";
  components: Record<string, ComponentHealth>;
  evaluated_at: string;
}

export interface SLOSummary {
  availability_sla: number;
  current_availability: number;
  latency_p95_sla_ms: number;
  current_latency_p95_ms: number;
  latency_p99_sla_ms: number;
  current_latency_p99_ms: number;
  error_budget_remaining_percent: number;
  burn_rate: number;
  status: "MET" | "WARNING" | "BREACHED";
}

export interface IncidentRecord {
  incident_id: string;
  category: string;
  severity: "P0" | "P1" | "P2" | "INFO";
  status: "ACTIVE" | "RESOLVED";
  reason: string;
  triggered_at: string;
  resolved_at?: string;
}

export interface OperationsSummary {
  timestamp: string;
  health: HealthReport;
  slo: SLOSummary;
  operational_controls: {
    maintenance_mode: boolean;
    model_frozen: boolean;
    retraining_paused: boolean;
    canary_paused: boolean;
  };
  active_incidents: IncidentRecord[];
}

export class OperationsApi {
  constructor(private client: HttpClient = defaultHttpClient) {}

  async getSummary(): Promise<OperationsSummary> {
    return this.client.request<OperationsSummary>("/v1/operations/summary");
  }

  async getHealth(): Promise<HealthReport> {
    return this.client.request<HealthReport>("/v1/operations/health");
  }

  async getSLO(): Promise<SLOSummary> {
    return this.client.request<SLOSummary>("/v1/operations/slo");
  }

  async getMetrics(): Promise<Record<string, any>> {
    return this.client.request<Record<string, any>>("/v1/operations/metrics");
  }

  async getIncidents(): Promise<IncidentRecord[]> {
    return this.client.request<IncidentRecord[]>("/v1/operations/incidents");
  }

  async getSafetyStatus(): Promise<any> {
    return this.client.request<any>("/v1/operations/safety");
  }

  // Admin Mutations
  async setMaintenanceMode(enable: boolean, reason: string, actor: string = "ops_admin"): Promise<any> {
    const endpoint = enable ? "/v1/operations/maintenance/enable" : "/v1/operations/maintenance/disable";
    return this.client.request<any>(endpoint, {
      method: "POST",
      body: JSON.stringify({ reason, actor }),
    });
  }

  async setModelFreeze(freeze: boolean, reason: string, actor: string = "ops_admin"): Promise<any> {
    const endpoint = freeze ? "/v1/operations/model/freeze" : "/v1/operations/model/unfreeze";
    return this.client.request<any>(endpoint, {
      method: "POST",
      body: JSON.stringify({ reason, actor }),
    });
  }

  async setRetrainingPause(pause: boolean, reason: string, actor: string = "ops_admin"): Promise<any> {
    const endpoint = pause ? "/v1/operations/retraining/pause" : "/v1/operations/retraining/resume";
    return this.client.request<any>(endpoint, {
      method: "POST",
      body: JSON.stringify({ reason, actor }),
    });
  }

  async triggerDisasterRecovery(reason: string, actor: string = "ops_admin"): Promise<any> {
    return this.client.request<any>("/v1/operations/recovery/trigger", {
      method: "POST",
      body: JSON.stringify({ reason, actor }),
    });
  }
}

export const operationsApi = new OperationsApi();
