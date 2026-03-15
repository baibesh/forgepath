"use client";

import { useState } from "react";
import { SrsIndicator } from "./SrsIndicator";
import { ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";

interface WordItem {
  id: number;
  word: string;
  definition: string | null;
  example: string | null;
  level: string;
  intervalDays: number;
  repetitions: number;
  srsStatus: "learned" | "learning" | "due";
}

export function WordCard({ item }: { item: WordItem }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="bg-surface rounded-xl p-3" onClick={() => setOpen(!open)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <SrsIndicator status={item.srsStatus} />
          <span className="font-medium">{item.word}</span>
          <span className="text-xs text-text-muted">{item.level}</span>
        </div>
        <ChevronDown
          size={16}
          className={cn(
            "text-text-muted transition-transform",
            open && "rotate-180"
          )}
        />
      </div>
      {open && (
        <div className="mt-2 pt-2 border-t border-border text-sm space-y-1">
          {item.definition && <p className="text-text-muted">{item.definition}</p>}
          {item.example && <p className="italic text-text-muted">{item.example}</p>}
          <p className="text-xs text-text-muted">
            Review in {item.intervalDays}d &middot; {item.repetitions} reps
          </p>
        </div>
      )}
    </div>
  );
}
