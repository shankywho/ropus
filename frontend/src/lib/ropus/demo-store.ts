import { useSyncExternalStore } from "react";

/**
 * Deterministic demo playback. No randomness, no simulated network jitter —
 * the controller only advances a stage index; every surface derives what it
 * shows from that index.
 */
export const demoStages = [
  { id: 0, label: "Idle", note: "No traffic replayed yet." },
  { id: 1, label: "Baseline activity", note: "Customer cus_4471029 transacting within their normal profile." },
  { id: 2, label: "Session anomaly", note: "Login from an unrecognised device in Limassol, CY." },
  { id: 3, label: "Beneficiary added", note: "Payout account PA-77120 attached to the customer." },
  { id: 4, label: "Wire submitted", note: "Outbound wire of 14,500 USD enters the decision path." },
  { id: 5, label: "Signals evaluated", note: "Rules, model, device and threat-intel factors returned." },
  { id: 6, label: "Graph expanded", note: "Payout account resolves to two confirmed-fraud accounts." },
  { id: 7, label: "Score escalated", note: "Risk score reaches 0.96 and the policy threshold is crossed." },
  { id: 8, label: "Decision returned", note: "BLOCK returned to the payments gateway in 38.4 ms." },
  { id: 9, label: "Case opened", note: "CASE-88419 opened at P1 and routed to the analyst queue." },
  { id: 10, label: "Analyst review", note: "Human confirms the block. Nothing executed automatically." },
] as const;

export type DemoState = {
  stage: number;
  playing: boolean;
  intervalMs: number;
};

let state: DemoState = { stage: 0, playing: false, intervalMs: 1400 };
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
