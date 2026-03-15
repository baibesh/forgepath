"use client";

import { useRef, useEffect } from "react";
import { cn } from "@/lib/utils";
import { staggerFadeIn } from "@/lib/animations";

interface StreakDay {
  date: string;
  wordDone: boolean;
  writingDone: boolean;
  reviewDone: boolean;
}

export function StreakCalendar({ days }: { days: StreakDay[] }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      const cells = Array.from(ref.current.querySelectorAll("[data-cell]")) as HTMLElement[];
      staggerFadeIn(cells, 0.02);
    }
  }, [days]);

  const today = new Date();
  const cells: { date: string; active: boolean; isToday: boolean }[] = [];

  for (let i = 29; i >= 0; i--) {
    const d = new Date(today);
    d.setDate(d.getDate() - i);
    const dateStr = d.toISOString().slice(0, 10);
    const streak = days.find((s) => s.date === dateStr);
    const active = streak ? streak.wordDone || streak.writingDone || streak.reviewDone : false;
    cells.push({ date: dateStr, active, isToday: i === 0 });
  }

  return (
    <div>
      <h3 className="text-sm font-medium text-text-muted mb-2">Last 30 days</h3>
      <div ref={ref} className="grid grid-cols-10 gap-1.5">
        {cells.map((cell) => (
          <div
            key={cell.date}
            data-cell
            className={cn(
              "aspect-square rounded-sm transition-colors",
              cell.active ? "bg-accent" : "bg-surface-2",
              cell.isToday && "ring-1 ring-accent-light"
            )}
            title={cell.date}
          />
        ))}
      </div>
    </div>
  );
}
