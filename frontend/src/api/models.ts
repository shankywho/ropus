import { HttpClient, defaultHttpClient } from "./client";

export interface ModelMetadata {
  name: string;
  version: string;
  stage: string;
  feature_contract: string;
  feature_count: number;
  calibration_version?: string;
  format?: string;
  status: string;
  auc_roc?: number;
  log_loss?: number;
  brier_score?: number;
  latency_p99_ms?: number;
  training_dataset?: string;
  trained_at?: string;
  approved_by?: string;
}

export interface CanaryStatus {
  enabled: boolean;
  target_percentage: number;
  candidate_model_version: string;
  safety_gate_status: string;
  circuit_breaker: {
    state: string;
    error_count?: number;
    fallback_count?: number;
  };
}

export interface RetrainingStatus {
  state: string;
  enabled: boolean;
  active_job?: any;
  last_retraining_time?: string;
  cooldown_active?: boolean;
}

export interface DriftSummary {
  status: string;
  max_psi: number;
  max_jsd: number;
  drifted_features_count: number;
  last_evaluation_time: string;
}

export class ModelsApi {
  constructor(private client: HttpClient = defaultHttpClient) {}

  async getProductionModel(): Promise<{ production_model: ModelMetadata; fallback_model?: ModelMetadata }> {
    return this.client.request<{ production_model: ModelMetadata; fallback_model?: ModelMetadata }>("/v1/models/production");
  }

  async listRegistry(): Promise<ModelMetadata[]> {
    return this.client.request<ModelMetadata[]>("/v1/models/registry");
  }

  async getModel(version: string): Promise<ModelMetadata> {
    return this.client.request<ModelMetadata>(`/v1/models/${version}`);
  }

  async getModelProvenance(version: string): Promise<any> {
    return this.client.request<any>(`/v1/models/${version}/provenance`);
  }

  async getCanaryStatus(): Promise<CanaryStatus> {
    return this.client.request<CanaryStatus>("/v1/canary/status");
  }

  async updateCanaryControl(payload: {
    enabled?: boolean;
    percentage: number;
    reason: string;
  }): Promise<any> {
    return this.client.request<any>("/v1/canary/control", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  }

  async getRetrainingStatus(): Promise<RetrainingStatus> {
    return this.client.request<RetrainingStatus>("/v1/retraining/status");
  }

  async getRetrainingHistory(): Promise<any[]> {
    return this.client.request<any[]>("/v1/retraining/history");
  }

  async triggerRetraining(reason: string, actor: string = "ml_operator"): Promise<any> {
    return this.client.request<any>("/v1/retraining/trigger", {
      method: "POST",
      body: JSON.stringify({ reason, actor }),
    });
  }

  async getDriftStatus(): Promise<DriftSummary> {
    return this.client.request<DriftSummary>("/v1/drift/status");
  }
}

export const modelsApi = new ModelsApi();
