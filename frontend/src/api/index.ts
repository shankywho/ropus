export * from "./client";
export * from "./decisions";
export * from "./rules";
export * from "./cases";
export * from "./models";
export * from "./operations";

import { decisionsApi } from "./decisions";
import { rulesApi } from "./rules";
import { casesApi } from "./cases";
import { modelsApi } from "./models";
import { operationsApi } from "./operations";
import { defaultHttpClient } from "./client";

export const api = {
  client: defaultHttpClient,
  decisions: decisionsApi,
  rules: rulesApi,
  cases: casesApi,
  models: modelsApi,
  operations: operationsApi,

  // Direct helper methods matching legacy ApiClient
  async getHealth() {
    return operationsApi.getHealth();
  },
  async evaluateRisk(payload: any) {
    return decisionsApi.evaluate(payload);
  },
  async getRules(status?: any) {
    return rulesApi.list(status);
  },
  async getRule(id: string) {
    return rulesApi.get(id);
  },
  async createRule(payload: any) {
    return rulesApi.create(payload);
  },
  async updateRule(id: string, payload: any) {
    return rulesApi.update(id, payload);
  },
  async transitionRule(id: string, status: any, actorId?: string) {
    return rulesApi.transitionStatus(id, status, actorId);
  },
  async getCases(status?: any) {
    return casesApi.list(status);
  },
  async getCase(id: string) {
    return casesApi.get(id);
  },
  async claimCase(id: string, analystId?: string) {
    return casesApi.claim(id, analystId);
  },
  async resolveCase(id: string, action: "ALLOW" | "DECLINE", reason: string, analystId?: string) {
    return casesApi.resolve(id, action, reason, analystId);
  },
};
