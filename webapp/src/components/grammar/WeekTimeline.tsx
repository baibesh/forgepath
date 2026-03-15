"use client";

import { useRef, useEffect } from "react";
import { WeekCard } from "./WeekCard";
import { staggerFadeIn } from "@/lib/animations";

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

export function WeekTimeline({
  weeks,
  currentWeek,
}: {
  weeks: GrammarWeekData[];
  currentWeek: number;
}) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      const cards = Array.from(ref.current.children) as HTMLElement[];
      staggerFadeIn(cards, 0.08);
    }
  }, [weeks]);

  return (
    <div ref={ref} className="space-y-3">
      {weeks.map((w) => (
        <WeekCard key={w.weekNum} week={w} isCurrent={w.weekNum === currentWeek} />
      ))}
    </div>
  );
}
