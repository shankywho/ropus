import { create } from "zustand";
import { CANONICAL_APPROVED_DECISION, CANONICAL_BLOCKED_DECISION, RiskDecisionPayload } from "./fixtures";

export type DemoStage =
  | "IDLE"
  | "NORMAL_ACTIVITY"
  | "ATTACK_TRIGGERED"
  | "SIGNALS_ARRIVING"
  | "GRAPH_REVEALED"
  | "SCORE_ESCALATED"
  | "DOSSIER_GENERATED"
  | "CASE_CREATED"
  | "ANALYST_CONFIRMED"
  | "COMPLETE";

export interface DemoEvent {
  id: string;
  time: string;
  type: "TRANSACTION" | "SIGNAL" | "GRAPH" | "DECISION" | "DOSSIER" | "CASE" | "ACTION";
  summary: string;
  status: "OK" | "WARN" | "CRITICAL";
}

interface DemoState {
  stage: DemoStage;
  stepIndex: number;
  isPlaying: boolean;
  activeDecision: RiskDecisionPayload;
  events: DemoEvent[];
  analystActionTaken: string | null;
  
  // Actions
  startDemo: () => void;
  pauseDemo: () => void;
  resumeDemo: () => void;
  resetDemo: () => void;
  nextStep: () => void;
  takeAnalystAction: (action: string) => void;
}

const STAGES: DemoStage[] = [
  "IDLE",
  "NORMAL_ACTIVITY",
  "ATTACK_TRIGGERED",
  "SIGNALS_ARRIVING",
  "GRAPH_REVEALED",
  "SCORE_ESCALATED",
  "DOSSIER_GENERATED",
  "CASE_CREATED",
  "ANALYST_CONFIRMED",
  "COMPLETE",
];

export const useDemoStore = create<DemoState>((set, get) => ({
  stage: "IDLE",
  stepIndex: 0,
  isPlaying: false,
  activeDecision: CANONICAL_APPROVED_DECISION,
  analystActionTaken: null,
  events: [
    {
      id: "evt_001",
      time: "14:10:00",
      type: "TRANSACTION",
      summary: "Baseline purchase $42.50 POS organic transaction",
      status: "OK",
    },
  ],

  startDemo: () => {
    set({
      stage: "NORMAL_ACTIVITY",
      stepIndex: 1,
      isPlaying: true,
      activeDecision: CANONICAL_APPROVED_DECISION,
      analystActionTaken: null,
      events: [
        {
          id: "evt_101",
          time: "14:10:00",
          type: "TRANSACTION",
          summary: "Baseline purchase $42.50 from New York US — Score: 0.04 (APPROVE)",
          status: "OK",
        },
      ],
    });
  },

  pauseDemo: () => set({ isPlaying: false }),
  resumeDemo: () => set({ isPlaying: true }),

  resetDemo: () => {
    set({
      stage: "IDLE",
      stepIndex: 0,
      isPlaying: false,
      activeDecision: CANONICAL_APPROVED_DECISION,
      analystActionTaken: null,
      events: [
        {
          id: "evt_001",
          time: "14:10:00",
          type: "TRANSACTION",
          summary: "System idle — awaiting live risk evaluation stream",
          status: "OK",
        },
      ],
    });
  },

  nextStep: () => {
    const currentIndex = get().stepIndex;
    const nextIndex = Math.min(currentIndex + 1, STAGES.length - 1);
    const nextStage = STAGES[nextIndex];

    const currentEvents = [...get().events];
    let newDecision = get().activeDecision;
    const newEvents = currentEvents;

    switch (nextStage) {
      case "ATTACK_TRIGGERED":
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:00",
          type: "TRANSACTION",
          summary: "Inbound $14,500 Wire Transfer initiated for usr_sarah_connor",
          status: "WARN",
        });
        break;

      case "SIGNALS_ARRIVING":
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:00.210",
          type: "SIGNAL",
          summary: "Impossible Travel: Limassol CY (8,420 km/h) & Datacenter VPN proxy detected",
          status: "CRITICAL",
        });
        break;

      case "GRAPH_REVEALED":
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:00.640",
          type: "GRAPH",
          summary: "Graph 3-hop traversal linked hardware canvas to 14 synthetic mule accounts",
          status: "CRITICAL",
        });
        break;

      case "SCORE_ESCALATED":
        newDecision = CANONICAL_BLOCKED_DECISION;
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:01.420",
          type: "DECISION",
          summary: "Risk Score escalated 0.04 -> 0.96 (BLOCK) in 1.42ms",
          status: "CRITICAL",
        });
        break;

      case "DOSSIER_GENERATED":
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:02.100",
          type: "DOSSIER",
          summary: "AI Investigator Council synthesized tripartite evidentiary dossier",
          status: "OK",
        });
        break;

      case "CASE_CREATED":
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:02.500",
          type: "CASE",
          summary: "Priority P0 Critical Case #CASE-88419 opened in analyst review queue",
          status: "CRITICAL",
        });
        break;

      case "ANALYST_CONFIRMED":
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:15.000",
          type: "ACTION",
          summary: "Analyst confirmed fraud block; wire halted, session revoked, SAR drafted",
          status: "OK",
        });
        break;

      case "COMPLETE":
        newEvents.unshift({
          id: `evt_${Date.now()}`,
          time: "17:42:20.000",
          type: "ACTION",
          summary: "Closed-loop feedback recorded; model retraining drift signal updated",
          status: "OK",
        });
        break;
    }

    set({
      stage: nextStage,
      stepIndex: nextIndex,
      activeDecision: newDecision,
      events: newEvents,
    });
  },

  takeAnalystAction: (action: string) => {
    set((state) => ({
      analystActionTaken: action,
      events: [
        {
          id: `evt_${Date.now()}`,
          time: "17:42:18",
          type: "ACTION",
          summary: `Analyst Action Executed: ${action} — Recorded in SHA-256 audit ledger`,
          status: "OK",
        },
        ...state.events,
      ],
    }));
  },
}));
