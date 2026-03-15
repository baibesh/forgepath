"use client";

import { cn } from "@/lib/utils";

type SrsStatus = "learned" | "learning" | "due";

export function SrsIndicator({ status }: { status: SrsStatus }) {
  return (
    <span
      className={cn(
        "inline-block w-2 h-2 rounded-full",
        status === "learned" && "bg-green",
        status === "learning" && "bg-yellow",
        status === "due" && "bg-red"
      )}
    />
  );
}
