import React from "react";

export type VerdictType = "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";

interface VerdictBadgeProps {
  verdict: VerdictType;
  size?: "sm" | "md" | "lg";
}

export function VerdictBadge({ verdict, size = "md" }: VerdictBadgeProps) {
  const styles: Record<VerdictType, { bg: string; text: string; border: string; label: string }> = {
    APPROVE: {
      bg: "bg-[#04db7c15]",
      text: "text-[#04db7c]",
      border: "border-[#04db7c44]",
      label: "APPROVE",
    },
    REVIEW: {
      bg: "bg-[#f59e0b15]",
      text: "text-[#f59e0b]",
      border: "border-[#f59e0b44]",
      label: "REVIEW",
    },
    CHALLENGE: {
      bg: "bg-[#f9731615]",
      text: "text-[#f97316]",
      border: "border-[#f9731644]",
      label: "CHALLENGE",
    },
    BLOCK: {
      bg: "bg-[#f0525215]",
      text: "text-[#f05252]",
      border: "border-[#f0525244]",
      label: "BLOCK",
    },
  };

  const sizeClasses = {
    sm: "px-2 py-0.5 text-xs",
    md: "px-2.5 py-1 text-xs font-semibold",
    lg: "px-3.5 py-1.5 text-sm font-bold tracking-wider",
  };

  const current = styles[verdict] || styles.REVIEW;

  return (
    <span
      className={`inline-flex items-center justify-center font-mono border rounded ${current.bg} ${current.text} ${current.border} ${sizeClasses[size]}`}
    >
      {current.label}
    </span>
  );
}
