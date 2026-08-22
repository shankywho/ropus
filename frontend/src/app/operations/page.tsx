"use client";

import React, { useState } from "react";
import { Activity, Server, AlertTriangle, Cpu, CheckCircle2, RefreshCw } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

export default function OperationsPage() {
  const [isRefreshing, setIsRefreshing] = useState(false);

  const services = [
    { name: "Risk Evaluation Engine", status: "HEALTHY", latency: "0.61ms", uptime: "99.999%", load: "14.2k req/s" },
    { name: "Fraud Knowledge Graph 3.0", status: "HEALTHY", latency: "2.14ms", uptime: "99.995%", load: "8.4k qps" },
    { name: "Apache Kafka Streaming Fabric", status: "HEALTHY", latency: "1.05ms", uptime: "100.0%", load: "45.0k msg/s" },
    { name: "PostgreSQL Primary Cluster (Multi-AZ)", status: "HEALTHY", latency: "3.20ms", uptime: "99.999%", load: "420 conn" },
    { name: "ONNX / XGBoost ML Model Runtime", status: "HEALTHY", latency: "0.42ms", uptime: "99.998%", load: "18.5k inf/s" },
    { name: "Multi-Agent Intelligence Council", status: "HEALTHY", latency: "18.4ms", uptime: "99.990%", load: "23 active" },
  ];

  const handleRefresh = () => {
    setIsRefreshing(true);
    setTimeout(() => setIsRefreshing(false), 800);
  };

  return (
    <div className="flex-1 p-6 md:p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <Server className="size-6 text-indigo-400" />
            <span>Fintech Infrastructure & Operations Control Plane</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Real-time cluster health, sub-millisecond latency telemetry, and dependency circuit breaker statuses.
          </p>
        </div>
        <Button onClick={handleRefresh} variant="outline" className="border-slate-800 bg-slate-900 text-xs">
          <RefreshCw className={`size-3.5 mr-1.5 ${isRefreshing ? "animate-spin" : ""}`} />
          Refresh Operations Telemetry
        </Button>
      </div>

      {/* Hero Stat Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card className="bg-slate-900/80 border-slate-800">
          <CardContent className="p-4">
            <p className="text-[11px] font-mono text-slate-400 uppercase">System Availability</p>
            <p className="text-2xl font-black text-emerald-400 mt-1">99.995%</p>
            <p className="text-[10px] text-slate-500 font-mono mt-1">Contractual SLA: 99.99%</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-900/80 border-slate-800">
          <CardContent className="p-4">
            <p className="text-[11px] font-mono text-slate-400 uppercase">P99 Decision Latency</p>
            <p className="text-2xl font-black text-white mt-1">6.8 ms</p>
            <p className="text-[10px] text-slate-500 font-mono mt-1">Target SLA: &lt; 50ms</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-900/80 border-slate-800">
          <CardContent className="p-4">
            <p className="text-[11px] font-mono text-slate-400 uppercase">Live Event Throughput</p>
            <p className="text-2xl font-black text-indigo-400 mt-1">104.2k /s</p>
            <p className="text-[10px] text-slate-500 font-mono mt-1">Peak Capacity: 2.61M/s</p>
          </CardContent>
        </Card>

        <Card className="bg-slate-900/80 border-slate-800">
          <CardContent className="p-4">
            <p className="text-[11px] font-mono text-slate-400 uppercase">Disaster Recovery RPO</p>
            <p className="text-2xl font-black text-amber-400 mt-1">&lt; 1 min</p>
            <p className="text-[10px] text-slate-500 font-mono mt-1">RTO: 12 mins (Target &lt; 30m)</p>
          </CardContent>
        </Card>
      </div>

      {/* Services Health Table */}
      <Card className="bg-slate-900/80 border-slate-800">
        <CardHeader>
          <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
            <Activity className="size-4 text-emerald-400" />
            <span>Subsystem Availability & Health Matrix</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="border-b border-slate-800 text-slate-400 uppercase font-mono">
                <tr>
                  <th className="pb-3">Subsystem</th>
                  <th className="pb-3">Health State</th>
                  <th className="pb-3">Response Latency</th>
                  <th className="pb-3">30-Day Uptime</th>
                  <th className="pb-3 text-right">Current Load</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 font-mono">
                {services.map((s, idx) => (
                  <tr key={idx} className="hover:bg-slate-800/30">
                    <td className="py-3.5 font-sans font-medium text-slate-200">{s.name}</td>
                    <td className="py-3.5">
                      <Badge className="bg-emerald-500/20 text-emerald-300 font-bold">
                        <CheckCircle2 className="size-3 mr-1" />
                        {s.status}
                      </Badge>
                    </td>
                    <td className="py-3.5 text-indigo-300 font-bold">{s.latency}</td>
                    <td className="py-3.5 text-slate-300">{s.uptime}</td>
                    <td className="py-3.5 text-right text-slate-400">{s.load}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
