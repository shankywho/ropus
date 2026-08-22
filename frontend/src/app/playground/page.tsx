"use client";

import React, { useState, useEffect } from "react";
import fpPromise from "@fingerprintjs/fingerprintjs";
import {
  ShieldAlert,
  ShieldCheck,
  AlertTriangle,
  Zap,
  Clock,
  Cpu,
  Sparkles,
  RefreshCw,
  Code2,
  Lock,
  ArrowRight,
  Info,
  CheckCircle2,
  XCircle,
  Fingerprint,
  Radio,
} from "lucide-react";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { api, RiskEvaluationResponse, RecommendedAction } from "@/lib/api";

export default function PlaygroundPage() {
  const [amount, setAmount] = useState<number>(1250);
  const [currency, setCurrency] = useState<string>("INR");
  const [paymentType, setPaymentType] = useState<string>("card");
  const [paymentToken, setPaymentToken] = useState<string>("tok_visa_gold_4242");
  const [ipAddress, setIpAddress] = useState<string>("192.168.1.45");
  const [deviceFingerprint, setDeviceFingerprint] = useState<string>("fp_telemetry_detecting...");
  const [liveVisitorId, setLiveVisitorId] = useState<string>("");
  const [fpLoading, setFpLoading] = useState<boolean>(true);

  // Live Device FingerprintJS Telemetry Sensor
  useEffect(() => {
    let isMounted = true;
    const captureFingerprint = async () => {
      try {
        setFpLoading(true);
        const fp = await fpPromise.load();
        const result = await fp.get();
        if (isMounted) {
          setDeviceFingerprint(result.visitorId);
          setLiveVisitorId(result.visitorId);
        }
      } catch (err) {
        console.error("Failed to capture browser fingerprint:", err);
        if (isMounted) {
          setDeviceFingerprint("fp_fallback_hw_" + Math.random().toString(36).substring(2, 10));
        }
      } finally {
        if (isMounted) {
          setFpLoading(false);
        }
      }
    };

    captureFingerprint();
    return () => {
      isMounted = false;
    };
  }, []);

  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [decision, setDecision] = useState<RiskEvaluationResponse | null>(null);
  const [showJson, setShowJson] = useState<boolean>(false);

  // Preset Handlers
  const fillSafeBaseline = () => {
    setAmount(1200);
    setCurrency("INR");
    setPaymentType("card");
    setPaymentToken("tok_visa_clean_8821");
    setIpAddress("49.207.198.12");
    if (liveVisitorId) {
      setDeviceFingerprint(liveVisitorId);
    } else {
      setDeviceFingerprint("fp_known_iphone15_pro");
    }
    setError(null);
  };

  const fillSyntheticFraud = () => {
    setAmount(95000);
    setCurrency("INR");
    setPaymentType("card");
    setPaymentToken("tok_card_stolen_velocity99");
    setIpAddress("192.168.1.100");
    setDeviceFingerprint("new_device_synthetic_fraud");
    setError(null);
  };

  const fillManualReview = () => {
    setAmount(48000);
    setCurrency("INR");
    setPaymentType("upi");
    setPaymentToken("tok_upi_suspicious_vpa");
    setIpAddress("192.168.1.100");
    setDeviceFingerprint("fp_untrusted_linux_emu");
    setError(null);
  };

  const restoreLiveFingerprint = () => {
    if (liveVisitorId) {
      setDeviceFingerprint(liveVisitorId);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const resp = await api.evaluateRisk({
        amount: Number(amount),
        currency: currency.trim() || "INR",
        payment_method: {
          type: paymentType,
          token: paymentToken.trim() || "tok_default_123",
        },
        device_fingerprint: deviceFingerprint.trim() || "fp_default_dev",
        ip_address: ipAddress.trim() || "127.0.0.1",
      });
      setDecision(resp);
    } catch (err: any) {
      console.error("Evaluation failed:", err);
      setError(err.message || "Failed to connect to Risk Engine API (:8080)");
    } finally {
      setLoading(false);
    }
  };

  // Helper for Outcome Badge Styling
  const renderOutcomeBadge = (action: RecommendedAction) => {
    switch (action) {
      case "ALLOW_RECOMMENDATION":
        return (
          <Badge
            variant="success"
            className="text-sm px-3 py-1 gap-1.5 bg-emerald-500/20 text-emerald-400 border-emerald-500/40"
          >
            <ShieldCheck className="size-4" />
            ALLOW_RECOMMENDATION
          </Badge>
        );
      case "STEP_UP_RECOMMENDATION":
        return (
          <Badge
            variant="info"
            className="text-sm px-3 py-1 gap-1.5 bg-blue-500/20 text-blue-400 border-blue-500/40"
          >
            <Info className="size-4" />
            STEP_UP_RECOMMENDATION (3DS)
          </Badge>
        );
      case "MANUAL_REVIEW":
      case "HOLD_RECOMMENDATION":
        return (
          <Badge
            variant="warning"
            className="text-sm px-3 py-1 gap-1.5 bg-amber-500/20 text-amber-400 border-amber-500/40"
          >
            <AlertTriangle className="size-4" />
            MANUAL_REVIEW (24h SLA)
          </Badge>
        );
      case "DECLINE_RECOMMENDATION":
        return (
          <Badge
            variant="danger"
            className="text-sm px-3 py-1 gap-1.5 bg-rose-500/20 text-rose-400 border-rose-500/40"
          >
            <ShieldAlert className="size-4" />
            DECLINE_RECOMMENDATION
          </Badge>
        );
      default:
        return (
          <Badge variant="outline" className="text-sm px-3 py-1">
            {action}
          </Badge>
        );
    }
  };

  // Helper for Score Color
  const getScoreColor = (score: number) => {
    if (score < 40) return "text-emerald-400";
    if (score < 70) return "text-amber-400";
    return "text-rose-400";
  };

  const getScoreProgressBg = (score: number) => {
    if (score < 40) return "bg-emerald-500";
    if (score < 70) return "bg-amber-500";
    return "bg-rose-500";
  };

  return (
    <div className="p-6 md:p-10 max-w-7xl mx-auto w-full space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <Zap className="size-6 text-indigo-400" />
              Risk Evaluation Playground
            </h1>
            <Badge variant="info">Live Engine</Badge>
          </div>
          <p className="text-sm text-slate-400">
            Real-time multi-stage transaction evaluation: Redis sliding window &rarr; Pre-rules &rarr; ONNX inference &rarr; Post-rules.
          </p>
        </div>

        {/* Preset Quick Fill Buttons */}
        <div className="flex flex-wrap items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={fillSafeBaseline}
            className="text-xs bg-slate-900 border-slate-700 hover:bg-slate-800 text-slate-300"
          >
            <CheckCircle2 className="size-3 text-emerald-400 mr-1" />
            Safe Baseline
          </Button>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={fillManualReview}
            className="text-xs bg-slate-900 border-slate-700 hover:bg-slate-800 text-slate-300"
          >
            <AlertTriangle className="size-3 text-amber-400 mr-1" />
            Trigger Review
          </Button>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={fillSyntheticFraud}
            className="text-xs bg-slate-900 border-rose-900/50 hover:bg-rose-950/40 text-rose-300"
          >
            <Sparkles className="size-3 text-rose-400 mr-1" />
            Fill Synthetic Fraud
          </Button>
        </div>
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left Column: Transaction Input Form (5 cols) */}
        <div className="lg:col-span-5 space-y-6">
          <Card className="bg-slate-900/70 border-slate-800 shadow-xl backdrop-blur">
            <CardHeader className="pb-4">
              <CardTitle className="text-base font-semibold text-white flex items-center justify-between">
                <span>Transaction Payload</span>
                <Badge variant="outline" className="font-mono text-[10px] text-slate-400">
                  POST /v1/risk-evaluations
                </Badge>
              </CardTitle>
              <CardDescription className="text-xs text-slate-400">
                Input transaction parameters for synchronous fraud evaluation.
              </CardDescription>
            </CardHeader>

            <form onSubmit={handleSubmit}>
              <CardContent className="space-y-4">
                {/* Amount and Currency */}
                <div className="grid grid-cols-3 gap-3">
                  <div className="col-span-2 space-y-1.5">
                    <label className="text-xs font-medium text-slate-300">
                      Amount (₹ Base Units)
                    </label>
                    <Input
                      type="number"
                      value={amount}
                      onChange={(e) => setAmount(Number(e.target.value))}
                      required
                      min={1}
                      placeholder="e.g. 5000"
                      className="bg-slate-950 border-slate-800 text-white font-mono"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-slate-300">
                      Currency
                    </label>
                    <Input
                      type="text"
                      value={currency}
                      onChange={(e) => setCurrency(e.target.value)}
                      required
                      placeholder="INR"
                      className="bg-slate-950 border-slate-800 text-white font-mono uppercase"
                    />
                  </div>
                </div>

                {/* Payment Method & Token */}
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-slate-300">
                    Payment Method Type
                  </label>
                  <select
                    value={paymentType}
                    onChange={(e) => setPaymentType(e.target.value)}
                    className="flex h-9 w-full rounded-md border border-slate-800 bg-slate-950 px-3 py-1 text-sm text-white shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-slate-300"
                  >
                    <option value="card">Card (Credit/Debit)</option>
                    <option value="upi">UPI (Virtual Payment Address)</option>
                    <option value="netbanking">Net Banking</option>
                    <option value="wallet">Digital Wallet</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-slate-300">
                    Payment Token / Instrument ID
                  </label>
                  <Input
                    type="text"
                    value={paymentToken}
                    onChange={(e) => setPaymentToken(e.target.value)}
                    required
                    placeholder="tok_visa_xyz123"
                    className="bg-slate-950 border-slate-800 text-white font-mono text-xs"
                  />
                </div>

                {/* IP Address */}
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-slate-300 flex items-center justify-between">
                    <span>IP Address</span>
                    <span className="text-[10px] text-slate-500 font-normal">
                      Sliding window 1h
                    </span>
                  </label>
                  <Input
                    type="text"
                    value={ipAddress}
                    onChange={(e) => setIpAddress(e.target.value)}
                    required
                    placeholder="192.168.1.1"
                    className="bg-slate-950 border-slate-800 text-white font-mono text-xs"
                  />
                </div>

                {/* Device Fingerprint (Live Hardware Hash Telemetry Sensor) */}
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <label className="text-xs font-medium text-slate-300 flex items-center gap-1.5">
                      <Fingerprint className="size-3.5 text-indigo-400" />
                      <span>Device Fingerprint</span>
                    </label>
                    <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-950/60 border border-emerald-500/30 text-[10px] font-mono text-emerald-400">
                      <span className="relative flex size-1.5">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                        <span className="relative inline-flex rounded-full size-1.5 bg-emerald-500"></span>
                      </span>
                      <Lock className="size-2.5 text-emerald-400" />
                      <span>Live Hardware Hash</span>
                    </div>
                  </div>

                  <div className="relative">
                    <Input
                      type="text"
                      value={deviceFingerprint}
                      readOnly
                      placeholder="fp_device_hash"
                      className="bg-slate-950/90 border-emerald-500/30 text-emerald-400 font-mono text-xs cursor-default pr-20 selection:bg-emerald-500/20"
                    />
                    <div className="absolute right-2.5 top-1/2 -translate-y-1/2 flex items-center gap-1.5">
                      {fpLoading ? (
                        <RefreshCw className="size-3 text-slate-500 animate-spin" />
                      ) : (
                        <Badge variant="outline" className="text-[9px] px-1.5 py-0 bg-slate-900 border-emerald-500/40 text-emerald-300 font-mono">
                          LOCKED
                        </Badge>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-[10px] text-slate-400 font-mono pt-0.5">
                    <span className="flex items-center gap-1 text-emerald-400/80">
                      <Radio className="size-3 text-emerald-400 shrink-0" />
                      Captured automatically via Browser Entropy (KMS encrypted)
                    </span>
                    {liveVisitorId && deviceFingerprint !== liveVisitorId && (
                      <button
                        type="button"
                        onClick={restoreLiveFingerprint}
                        className="text-[10px] text-indigo-400 hover:text-indigo-300 underline underline-offset-2 transition-colors"
                      >
                        Reset Live Hash
                      </button>
                    )}
                  </div>
                </div>

                {error && (
                  <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs flex items-start gap-2">
                    <AlertTriangle className="size-4 shrink-0 mt-0.5" />
                    <span>{error}</span>
                  </div>
                )}
              </CardContent>

              <CardFooter className="pt-2">
                <Button
                  type="submit"
                  disabled={loading}
                  className="w-full bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white shadow-lg shadow-indigo-500/25 gap-2 h-10 font-medium"
                >
                  {loading ? (
                    <>
                      <RefreshCw className="size-4 animate-spin" />
                      Evaluating Decision...
                    </>
                  ) : (
                    <>
                      <Zap className="size-4" />
                      Evaluate Risk Recommendation
                    </>
                  )}
                </Button>
              </CardFooter>
            </form>
          </Card>
        </div>

        {/* Right Column: Live Decision Inspector (7 cols) */}
        <div className="lg:col-span-7 space-y-6">
          {decision ? (
            <div className="space-y-6 animate-in fade-in-50 duration-300">
              {/* Primary Decision Card */}
              <Card className="bg-slate-900/80 border-slate-800 shadow-2xl backdrop-blur relative overflow-hidden">
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-indigo-500 via-purple-500 to-emerald-500" />

                <CardHeader className="pb-4">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div>
                      <CardDescription className="text-xs font-mono text-slate-400">
                        DECISION ID: {decision.decision_id}
                      </CardDescription>
                      <CardTitle className="text-xl font-bold text-white mt-1">
                        Real-Time Evaluation Result
                      </CardTitle>
                    </div>
                    {renderOutcomeBadge(decision.recommended_action)}
                  </div>
                </CardHeader>

                <CardContent className="space-y-6">
                  {/* Risk Score Gauge & Latency Header */}
                  <div className="p-5 rounded-xl bg-slate-950/80 border border-slate-800/80 flex flex-col sm:flex-row sm:items-center justify-between gap-6">
                    <div className="space-y-1">
                      <p className="text-xs font-medium text-slate-400">
                        Calculated Risk Score
                      </p>
                      <div className="flex items-baseline gap-2">
                        <span
                          className={`text-5xl font-black font-mono tracking-tight ${getScoreColor(
                            decision.risk_score
                          )}`}
                        >
                          {decision.risk_score}
                        </span>
                        <span className="text-sm text-slate-500 font-mono">/ 100</span>
                      </div>
                      <div className="w-48 h-2 bg-slate-800 rounded-full overflow-hidden mt-2">
                        <div
                          className={`h-full transition-all duration-500 ${getScoreProgressBg(
                            decision.risk_score
                          )}`}
                          style={{ width: `${Math.min(100, decision.risk_score)}%` }}
                        />
                      </div>
                    </div>

                    {/* Latency & Telemetry */}
                    <div className="sm:border-l sm:border-slate-800 sm:pl-6 space-y-2 text-xs">
                      <div className="flex items-center gap-2 text-slate-300">
                        <Clock className="size-4 text-emerald-400" />
                        <span className="font-mono font-semibold text-emerald-400">
                          {decision.latency_ms} ms
                        </span>
                        <span className="text-slate-500">(Target: &lt;100ms)</span>
                      </div>

                      <div className="flex items-center gap-2 text-slate-300">
                        <Cpu className="size-4 text-indigo-400" />
                        <span className="font-mono text-slate-200">ONNX Runtime</span>
                      </div>

                      <div className="flex items-center gap-2 text-slate-300">
                        <Lock className="size-4 text-purple-400" />
                        <span className="text-slate-400">AES-GCM Shredding Ready</span>
                      </div>

                      {decision.is_degraded && (
                        <Badge variant="warning" className="text-[10px] py-0 px-2 mt-1">
                          Degraded Heuristic Fallback
                        </Badge>
                      )}
                    </div>
                  </div>

                  {/* Reason Codes (SHAP Attribution) */}
                  <div className="space-y-2">
                    <p className="text-xs font-semibold text-slate-300 tracking-wide uppercase">
                      Contributing Reason Codes (SHAP &amp; Rules)
                    </p>
                    {decision.reason_codes && decision.reason_codes.length > 0 ? (
                      <div className="flex flex-wrap gap-2">
                        {decision.reason_codes.map((code, idx) => (
                          <Badge
                            key={idx}
                            variant="outline"
                            className="bg-slate-950 border-slate-700 text-slate-200 font-mono text-xs px-2.5 py-1"
                          >
                            <span className="size-1.5 rounded-full bg-indigo-400 mr-1.5" />
                            {code}
                          </Badge>
                        ))}
                      </div>
                    ) : (
                      <p className="text-xs text-slate-500 italic">
                        No anomaly flags triggered. Transaction matches baseline trust envelope.
                      </p>
                    )}
                  </div>

                  {/* Velocity & Feature Snapshot Inspection */}
                  {decision.features && (
                    <div className="space-y-2 pt-2 border-t border-slate-800">
                      <p className="text-xs font-semibold text-slate-300 tracking-wide uppercase">
                        Real-Time Aggregated Velocity
                      </p>
                      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                        <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800/60">
                          <p className="text-[11px] text-slate-400">IP Txns (1 Hour)</p>
                          <p className="text-lg font-bold font-mono text-white mt-0.5">
                            {decision.features["velocity.ip.1hr"] ?? 0}
                          </p>
                        </div>

                        <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800/60">
                          <p className="text-[11px] text-slate-400">Token Txns (24 Hour)</p>
                          <p className="text-lg font-bold font-mono text-white mt-0.5">
                            {decision.features["velocity.token.24hr"] ?? 0}
                          </p>
                        </div>

                        <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800/60 col-span-2 sm:col-span-1">
                          <p className="text-[11px] text-slate-400">Snapshot Ref</p>
                          <p className="text-xs font-mono text-slate-300 truncate mt-1">
                            {decision.feature_snapshot_ref}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Raw JSON Debug Toggle */}
                  <div className="pt-2 border-t border-slate-800">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowJson(!showJson)}
                      className="text-xs text-slate-400 hover:text-slate-200 p-0 h-auto gap-1.5"
                    >
                      <Code2 className="size-3.5" />
                      {showJson ? "Hide JSON Contract" : "Inspect Raw Decision Contract JSON"}
                    </Button>

                    {showJson && (
                      <pre className="mt-3 p-4 rounded-lg bg-slate-950 border border-slate-800 text-[11px] text-emerald-400 font-mono overflow-auto max-h-64 leading-relaxed">
                        {JSON.stringify(decision, null, 2)}
                      </pre>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>
          ) : (
            /* Empty State Prompt */
            <Card className="bg-slate-900/40 border-dashed border-slate-800 h-full flex flex-col items-center justify-center p-12 text-center min-h-[420px]">
              <div className="size-16 rounded-full bg-slate-800/60 flex items-center justify-center mb-4 text-slate-400">
                <Sparkles className="size-8 text-indigo-400 animate-pulse" />
              </div>
              <CardTitle className="text-lg font-semibold text-white">
                Awaiting Transaction Submission
              </CardTitle>
              <CardDescription className="text-xs text-slate-400 max-w-sm mt-2">
                Click <strong>Evaluate Risk Recommendation</strong> or use one of the quick presets above to test synchronous decisioning and SHAP reason attribution.
              </CardDescription>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
