import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-slate-900 text-white shadow hover:bg-slate-800 dark:bg-slate-50 dark:text-slate-900 dark:hover:bg-slate-200",
        secondary:
          "border-transparent bg-slate-100 text-slate-900 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-50 dark:hover:bg-slate-700",
        destructive:
          "border-transparent bg-red-500 text-white shadow hover:bg-red-600 dark:bg-red-900 dark:text-slate-50 dark:hover:bg-red-800",
        outline: "text-slate-950 dark:text-slate-50 border-slate-200 dark:border-slate-800",
        success:
          "border-transparent bg-emerald-500/15 text-emerald-700 border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-400",
        warning:
          "border-transparent bg-amber-500/15 text-amber-700 border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-400",
        danger:
          "border-transparent bg-rose-500/15 text-rose-700 border-rose-500/30 dark:bg-rose-500/10 dark:text-rose-400",
        info:
          "border-transparent bg-blue-500/15 text-blue-700 border-blue-500/30 dark:bg-blue-500/10 dark:text-blue-400",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
