"use client";

import { useRef, useEffect } from "react";
import { staggerFadeIn } from "@/lib/animations";
import { Flame, BookOpen, PenLine } from "lucide-react";

interface Props {
  streak: number;
  words: number;
  writings: number;
}

export function StatsCards({ streak, words, writings }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      const cards = Array.from(ref.current.children) as HTMLElement[];
      staggerFadeIn(cards, 0.1);
    }
  }, []);

  const stats = [
    { label: "Streak", value: streak, icon: Flame, color: "text-yellow" },
    { label: "Words", value: words, icon: BookOpen, color: "text-green" },
    { label: "Writings", value: writings, icon: PenLine, color: "text-accent-light" },
  ];

  return (
    <div ref={ref} className="grid grid-cols-3 gap-3">
      {stats.map(({ label, value, icon: Icon, color }) => (
        <div key={label} className="bg-surface rounded-xl p-3 text-center">
          <Icon size={20} className={`${color} mx-auto mb-1`} />
          <div className="text-2xl font-bold">{value}</div>
          <div className="text-xs text-text-muted">{label}</div>
        </div>
      ))}
    </div>
  );
}
