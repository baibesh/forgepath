"use client";

import { useRef, useEffect } from "react";
import { MediaCard } from "./MediaCard";
import { staggerFadeIn } from "@/lib/animations";

interface MediaData {
  id: number;
  title: string;
  url: string;
  duration: string;
  level: string;
  watched: boolean;
  responded: boolean;
  response: string | null;
  sentAt: string;
}

export function MediaList({ media }: { media: MediaData[] }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) {
      const cards = Array.from(ref.current.children) as HTMLElement[];
      staggerFadeIn(cards, 0.06);
    }
  }, [media]);

  return (
    <div ref={ref} className="space-y-2">
      {media.map((m) => (
        <MediaCard key={m.id} media={m} />
      ))}
    </div>
  );
}
