import { useSyncExternalStore } from "react";

export type DemoStageInfo = {
  id: number;
  stageNumber: number;
  label: string;
  shortExplanation: string;
  technicalDetails: string;
  activeSurface: "baseline" | "transaction" | "rules" | "ml" | "graph" | "investigator" | "case";
  simulatedScore: number;
  verdict: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";
};

export const demoStages: DemoStageInfo[] = [
  {
    id: 0,
    stageNumber: 1,
    label: "Normal Customer Baseline",
    shortExplanation: "Customer cus_4471029 transacts normally within their historical baseline.",
    technicalDetails:
      "Regular $50.00 grocery spend from Singapore (103.82°E, 1.35°N) on registered Safari iOS device (dev_safari_01). 0 disputes, normal velocity (1.2 tx/day).",
    activeSurface: "baseline",
    simulatedScore: 0.02,
    verdict: "APPROVE",
  },
  {
    id: 1,
    stageNumber: 2,
    label: "Attack Transaction Ingress",
    shortExplanation:
      "A high-value outbound wire request arrives from an unrecognised device and location.",
    technicalDetails:
      "POST /v1/risk/evaluate receives $14,500.00 USD transfer from Limassol, Cyprus IP 198.51.100.44 (Datacenter ASN 13335) on a Linux emulator canvas fingerprint.",
    activeSurface: "transaction",
    simulatedScore: 0.28,
    verdict: "REVIEW",
  },
  {
    id: 2,
    stageNumber: 3,
    label: "Rules Engine Deterministic Triggers",
    shortExplanation:
      "Declarative AST rules detect velocity surges and impossible physical travel.",
    technicalDetails:
      "AST evaluation in 0.4ms fires 3 deterministic rules: Velocity Surge (+412% in 1hr), Impossible Travel (51,350 km/h from Singapore 12m ago), and Datacenter Proxy CIDR.",
    activeSurface: "rules",
    simulatedScore: 0.65,
    verdict: "CHALLENGE",
  },
  {
    id: 3,
    stageNumber: 4,
    label: "ML Inference & Calibrated Probability",
    shortExplanation: "25-feature ONNX XGBoost model evaluates non-linear attack interactions.",
    technicalDetails:
      "Python sidecar executes ONNX tree traversal in 2.1ms. Raw probability = 0.8712; Beta Calibration yields posterior P(fraud|x) = 0.9418. Expected Loss = $13,656.10.",
    activeSurface: "ml",
    simulatedScore: 0.88,
    verdict: "BLOCK",
  },
  {
    id: 4,
    stageNumber: 5,
    label: "Fraud Knowledge Graph Traversal",
    shortExplanation: "In-memory 3-hop BFS discovers multi-account syndicate linkage.",
    technicalDetails:
      "Graph traversal expands customer -> device canvas 9f8a... -> 14 synthetic mule accounts with prior chargebacks. Degree centrality = 16. Confirmed syndicate ring.",
    activeSurface: "graph",
    simulatedScore: 0.94,
    verdict: "BLOCK",
  },
  {
    id: 5,
    stageNumber: 6,
    label: "AI Investigator Synthesizes Dossier",
    shortExplanation:
      "Autonomous Agent Council structures tripartite evidence and opens Priority 0 Case.",
    technicalDetails:
      "Council agents (Threat Hunter, Graph Analyst, AML Officer, Lead) generate Tripartite Dossier (Observed Facts vs Inferred Patterns vs Recommended Action). Opens CASE-88419 (P0 Critical).",
    activeSurface: "investigator",
    simulatedScore: 0.96,
    verdict: "BLOCK",
  },
  {
    id: 6,
    stageNumber: 7,
    label: "Human Analyst Confirms Block",
    shortExplanation:
      "Fraud analyst verifies the evidentiary dossier and confirms the permanent freeze.",
    technicalDetails:
      "Analyst reviews tripartite dossier in Control Plane and executes 'CONFIRM BLOCK & FREEZE'. Action commits to SHA-256 hash-chained audit ledger and routes ground-truth label to closed-loop ML retraining.",
    activeSurface: "case",
    simulatedScore: 0.96,
    verdict: "BLOCK",
  },
];

export type DemoState = {
  stage: number;
  playing: boolean;
  intervalMs: number;
};

let state: DemoState = { stage: 0, playing: false, intervalMs: 2500 };
let timer: ReturnType<typeof setInterval> | null = null;
const listeners = new Set<() => void>();

const emit = () => listeners.forEach((l) => l());

const set = (patch: Partial<DemoState>) => {
  state = { ...state, ...patch };
  emit();
};

const stopTimer = () => {
  if (timer) clearInterval(timer);
  timer = null;
};

const tick = () => {
  if (state.stage >= demoStages.length - 1) {
    stopTimer();
    set({ playing: false });
    return;
  }
  set({ stage: state.stage + 1 });
};

export const demoControls = {
  play() {
    if (state.playing) return;
    if (state.stage >= demoStages.length - 1) set({ stage: 0 });
    set({ playing: true });
    stopTimer();
    timer = setInterval(tick, state.intervalMs);
  },
  pause() {
    stopTimer();
    set({ playing: false });
  },
  next() {
    this.pause();
    set({ stage: Math.min(state.stage + 1, demoStages.length - 1) });
  },
  back() {
    this.pause();
    set({ stage: Math.max(state.stage - 1, 0) });
  },
  reset() {
    this.pause();
    set({ stage: 0 });
  },
  goTo(stage: number) {
    this.pause();
    set({ stage: Math.max(0, Math.min(stage, demoStages.length - 1)) });
  },
  setSpeed(intervalMs: number) {
    set({ intervalMs });
    if (state.playing) {
      stopTimer();
      timer = setInterval(tick, intervalMs);
    }
  },
};

const subscribe = (listener: () => void) => {
  listeners.add(listener);
  return () => listeners.delete(listener);
};

const getSnapshot = () => state;

export function useDemoState(): DemoState {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}
