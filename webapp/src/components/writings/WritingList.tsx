"use client";

import { useRef, useEffect } from "react";
import { WritingCard } from "./WritingCard";
import { staggerFadeIn } from "@/lib/animations";

interface WritingData {
  id: number;
  topic: string | null;
  grammarFocus: string | null;
  text: string | null;
  feedback: string | null;
  wordCount: number;
  createdAt: string;
}

export function WritingList({ writings }: { writings: WritingData[] }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      const cards = Array.from(ref.current.children) as HTMLElement[];
      staggerFadeIn(cards, 0.06);
    }
  }, [writings]);

  if (writings.length === 0) {
    return <p className="text-center text-text-muted py-8">No writings yet</p>;
  }

  return (
    <div ref={ref} className="space-y-2">
      {writings.map((w) => (
        <WritingCard key={w.id} writing={w} />
      ))}
    </div>
  );
}
