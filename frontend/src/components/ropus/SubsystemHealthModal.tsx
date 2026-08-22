import React from "react";
import { X, CheckCircle2, ShieldCheck } from "lucide-react";

interface SubsystemHealthModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function SubsystemHealthModal({ isOpen, onClose }: SubsystemHealthModalProps) {
  if (!isOpen) return null;

  const subsystems = [
    {
      name: "Risk Engine",
      subsystemId: "sub_risk_engine",
      breakerState: "CLOSED",
      latencyP99: "1.42ms",
      errorRate: "0.00%",
      status: "HEALTHY",
    },
    {
      name: "Rules Engine",
      subsystemId: "sub_rules_eval",
      breakerState: "CLOSED",
      latencyP99: "0.18ms",
      errorRate: "0.00%",
      status: "HEALTHY",
    },
    {
      name: "ML Inference (XGBoost 25-Feature)",
      subsystemId: "sub_ml_inference",
      breakerState: "CLOSED",
      latencyP99: "0.45ms",
      errorRate: "0.01%",
      status: "HEALTHY",
    },
    {
      name: "Fraud Knowledge Graph 3.0",
      subsystemId: "sub_graph_bfs",
      breakerState: "CLOSED",
      latencyP99: "1.10ms",
      errorRate: "0.00%",
      status: "HEALTHY",
    },
    {
      name: "Threat Intelligence & Geo Velocity",
      subsystemId: "sub_threat_intel",
      breakerState: "CLOSED",
      latencyP99: "0.22ms",
      errorRate: "0.00%",
      status: "HEALTHY",
    },
    {
      name: "Apache Kafka Event Bus",
      subsystemId: "sub_kafka_streaming",
      breakerState: "CLOSED",
      latencyP99: "0.85ms",
      errorRate: "0.00%",
      status: "HEALTHY",
    },
    {
      name: "Redis Sliding Feature Store",
      subsystemId: "sub_redis_features",
      breakerState: "CLOSED",
      latencyP99: "0.32ms",
      errorRate: "0.00%",
      status: "HEALTHY",
    },
    {
      name: "Webhook Delivery & HMAC Worker",
      subsystemId: "sub_webhook_dispatcher",
      breakerState: "CLOSED",
      latencyP99: "18.4ms",
      errorRate: "0.02%",
      status: "HEALTHY",
    },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4">
      <div className="bg-[#0b1528] border border-[#1c2b48] rounded w-full max-w-2xl overflow-hidden shadow-2xl">
        <div className="px-5 py-4 border-b border-[#1c2b48] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ShieldCheck className="w-5 h-5 text-[#04db7c]" />
            <h3 className="font-bold text-sm text-[#f4f5f7]">
              Subsystem Resilience & Circuit Breaker Registry
            </h3>
          </div>
          <button
            onClick={onClose}
            className="text-[#6b778c] hover:text-[#f4f5f7] p-1 rounded hover:bg-[#1c2b48]"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-5 max-h-[70vh] overflow-y-auto">
          <div className="mb-4 flex items-center justify-between p-3 bg-[#070e1c] border border-[#1c2b48] rounded">
            <div>
              <span className="text-xs text-[#6b778c] font-mono block">CONTRACTUAL AVAILABILITY</span>
              <span className="text-base font-mono font-bold text-[#04db7c]">99.994%</span>
            </div>
            <div>
              <span className="text-xs text-[#6b778c] font-mono block">30-DAY ERROR BUDGET</span>
              <span className="text-base font-mono font-semibold text-[#38bdf8]">92.8% Remaining</span>
            </div>
            <div>
              <span className="text-xs text-[#6b778c] font-mono block">FALLBACK BUFFER</span>
              <span className="text-base font-mono font-semibold text-[#97a0af]">0 Events Buffered</span>
            </div>
          </div>

          <table className="w-full text-left text-xs font-mono">
            <thead>
              <tr className="border-b border-[#1c2b48] text-[#6b778c]">
                <th className="pb-2 font-medium">SUBSYSTEM</th>
                <th className="pb-2 font-medium">BREAKER</th>
                <th className="pb-2 font-medium">P99</th>
                <th className="pb-2 font-medium">ERRORS</th>
                <th className="pb-2 font-medium text-right">STATUS</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2b48]">
              {subsystems.map((sub, i) => (
                <tr key={i} className="hover:bg-[#0f1c34]">
                  <td className="py-2.5 text-[#f4f5f7] font-semibold">{sub.name}</td>
                  <td className="py-2.5">
                    <span className="px-1.5 py-0.5 rounded text-[10px] bg-[#04db7c15] text-[#04db7c] border border-[#04db7c33]">
                      {sub.breakerState}
                    </span>
                  </td>
                  <td className="py-2.5 text-[#97a0af]">{sub.latencyP99}</td>
                  <td className="py-2.5 text-[#97a0af]">{sub.errorRate}</td>
                  <td className="py-2.5 text-right">
                    <span className="inline-flex items-center gap-1 text-[#04db7c]">
                      <CheckCircle2 className="w-3 h-3" /> OK
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="px-5 py-3 bg-[#070e1c] border-t border-[#1c2b48] flex justify-end">
          <button
            onClick={onClose}
            className="px-3 py-1.5 text-xs font-mono bg-[#1c2b48] hover:bg-[#2c3e66] text-[#f4f5f7] rounded border border-[#2c3e66]"
          >
            Close Inspector
          </button>
        </div>
      </div>
    </div>
  );
}
