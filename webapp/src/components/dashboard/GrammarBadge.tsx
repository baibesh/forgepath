"use client";

import { useRef, useEffect } from "react";
import { scaleIn } from "@/lib/animations";

interface Props {
  weekNum: number;
  tenseName: string;
  family: string;
  focus: string;
}

export function GrammarBadge({ weekNum, tenseName, family, focus }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) scaleIn(ref.current, 0.2);
  }, []);

  return (
    <div ref={ref} className="bg-surface rounded-xl p-4" style={{ opacity: 0 }}>
      <div className="text-xs text-text-muted mb-1">Current Grammar</div>
      <div className="flex items-baseline gap-2">
        <span className="text-sm font-medium text-accent">Week {weekNum}</span>
        <span className="text-lg font-semibold">{tenseName}</span>
      </div>
      <div className="text-sm text-text-muted mt-1">
        {family} &middot; {focus}
      </div>
    </div>
  );
}
