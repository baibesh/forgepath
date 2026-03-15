"use client";

import { useRef, useEffect } from "react";
import { WordCard } from "./WordCard";
import { staggerFadeIn } from "@/lib/animations";

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

export function WordList({ words }: { words: WordItem[] }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      const cards = Array.from(ref.current.children) as HTMLElement[];
      staggerFadeIn(cards, 0.05);
    }
  }, [words]);

  if (words.length === 0) {
    return <p className="text-center text-text-muted py-8">No words found</p>;
  }

  return (
    <div ref={ref} className="space-y-2">
      {words.map((w) => (
        <WordCard key={w.id} item={w} />
      ))}
    </div>
  );
}
