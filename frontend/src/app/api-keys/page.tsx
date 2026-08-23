"use client";

import React, { useState } from "react";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import { Key, Plus, Copy, Check, RotateCw, Trash2, ShieldCheck } from "lucide-react";

interface KeyRecord {
  id: string;
  name: string;
  prefix: string;
  environment: "live" | "test";
  created: string;
  lastUsed: string;
  status: "ACTIVE" | "REVOKED";
}

export default function APIKeysPage() {
  const [keys, setKeys] = useState<KeyRecord[]>([
    {
      id: "key_88a91c2b",
      name: "Production Payment Gateway",
      prefix: "rop_live_8a19bc...",
      environment: "live",
      created: "2026-08-01",
      lastUsed: "Just now",
      status: "ACTIVE",
    },
    {
      id: "key_33f481e0",
      name: "Staging Test Sandbox",
      prefix: "rop_test_99f412...",
      environment: "test",
      created: "2026-08-10",
      lastUsed: "2 hours ago",
      status: "ACTIVE",
    },
  ]);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleRotate = (id: string) => {
    setKeys(
      keys.map((k) =>
        k.id === id
          ? { ...k, prefix: "rop_live_rot_" + Math.random().toString(36).substring(2, 8) + "..." }
          : k
      )
    );
  };

  const handleRevoke = (id: string) => {
    setKeys(keys.map((k) => (k.id === id ? { ...k, status: "REVOKED" } : k)));
  };

  return (
    <div className="space-y-5">
      {/* 1. Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              API Key Management
            </h1>
            <DataProvenanceBadge type="LIVE_BACKEND" sublabel="SHA-256 Hashed Secrets" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Cryptographically signed HMAC tokens for production synchronous risk evaluation
          </p>
        </div>

        <button className="px-3.5 py-1.5 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-mono text-xs font-bold rounded-[4px] shadow-xs active:scale-95 cursor-pointer flex items-center gap-1.5">
          <Plus className="w-3.5 h-3.5" />
          <span>Generate API Key</span>
        </button>
      </div>

      {/* 2. Keys Table */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 font-mono text-xs">
        <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Key className="w-4 h-4 text-[#0d94fb]" />
            <h2 className="font-bold uppercase tracking-wider text-[#f4f5f7] text-xs">
              Active Organization Keys ({keys.length})
            </h2>
          </div>
          <span className="text-[#5e6c84] text-[10px]">Scoped RBAC Access</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">NAME / KEY ID</th>
                <th className="pb-2 font-medium">PREFIX</th>
                <th className="pb-2 font-medium">ENVIRONMENT</th>
                <th className="pb-2 font-medium">CREATED</th>
                <th className="pb-2 font-medium">LAST ACTIVE</th>
                <th className="pb-2 font-medium">STATUS</th>
                <th className="pb-2 font-medium text-right">ACTIONS</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {keys.map((k) => (
                <tr key={k.id} className="hover:bg-[#142036] transition-colors">
                  <td className="py-3">
                    <span className="font-bold text-[#f4f5f7] block">{k.name}</span>
                    <span className="text-[10px] text-[#5e6c84]">{k.id}</span>
                  </td>
                  <td className="py-3">
                    <div className="flex items-center gap-1.5">
                      <span className="text-[#0d94fb]">{k.prefix}</span>
                      <button
                        onClick={() => handleCopy(k.id, k.prefix)}
                        className="text-[#5e6c84] hover:text-[#f4f5f7] cursor-pointer"
                      >
                        {copiedId === k.id ? (
                          <Check className="w-3 h-3 text-[#04db7c]" />
                        ) : (
                          <Copy className="w-3 h-3" />
                        )}
                      </button>
                    </div>
                  </td>
                  <td className="py-3">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                        k.environment === "live"
                          ? "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                          : "bg-[#5e6c8415] text-[#97a0af] border-[#5e6c8433]"
                      }`}
                    >
                      {k.environment.toUpperCase()}
                    </span>
                  </td>
                  <td className="py-3 text-[#97a0af]">{k.created}</td>
                  <td className="py-3 text-[#97a0af]">{k.lastUsed}</td>
                  <td className="py-3">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                        k.status === "ACTIVE"
                          ? "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                          : "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                      }`}
                    >
                      {k.status}
                    </span>
                  </td>
                  <td className="py-3 text-right">
                    <div className="flex items-center justify-end gap-1.5">
                      <button
                        onClick={() => handleRotate(k.id)}
                        className="p-1 hover:bg-[#0a1324] rounded text-[#5e6c84] hover:text-[#f4f5f7] cursor-pointer"
                        title="Rotate Secret"
                      >
                        <RotateCw className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={() => handleRevoke(k.id)}
                        className="p-1 hover:bg-[#f0525215] rounded text-[#5e6c84] hover:text-[#f05252] cursor-pointer"
                        title="Revoke Key"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
