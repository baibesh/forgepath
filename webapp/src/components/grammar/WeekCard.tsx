"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { ChevronDown } from "lucide-react";

interface GrammarWeekData {
  weekNum: number;
  family: string;
  focus: string;
  tenseName: string;
  anchor: string;
  markers: string;
  formula: string;
  example: string;
}

export function WeekCard({
  week,
  isCurrent,
}: {
  week: GrammarWeekData;
  isCurrent: boolean;
}) {
  const [open, setOpen] = useState(isCurrent);

  return (
    <div
      className={cn(
        "bg-surface rounded-xl p-3 transition-colors",
        isCurrent && "ring-1 ring-accent"
      )}
      onClick={() => setOpen(!open)}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span
            className={cn(
              "w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold",
              isCurrent ? "bg-accent text-white" : "bg-surface-2 text-text-muted"
            )}
          >
            {week.weekNum}
          </span>
          <div>
            <div className="font-medium text-sm">{week.tenseName}</div>
            <div className="text-xs text-text-muted">{week.family}</div>
          </div>
        </div>
        <ChevronDown
          size={16}
          className={cn("text-text-muted transition-transform", open && "rotate-180")}
        />
      </div>
      {open && (
        <div className="mt-3 pt-3 border-t border-border text-sm space-y-2">
          <div>
            <span className="text-text-muted">Focus: </span>
            {week.focus}
          </div>
          <div>
            <span className="text-text-muted">Anchor: </span>
            {week.anchor}
          </div>
          <div>
            <span className="text-text-muted">Formula: </span>
            <code className="text-accent text-xs">{week.formula}</code>
          </div>
          <div>
            <span className="text-text-muted">Markers: </span>
            {week.markers}
          </div>
          <div>
            <span className="text-text-muted">Example: </span>
            <em>{week.example}</em>
          </div>
        </div>
      )}
    </div>
  );
}
