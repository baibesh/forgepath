"use client";

import { useRef, useEffect, type ReactNode } from "react";
import { fadeInUp } from "@/lib/animations";

export function PageTransition({ children }: { children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (ref.current) fadeInUp(ref.current);
  }, []);

  return (
    <div ref={ref} style={{ opacity: 0 }}>
      {children}
    </div>
  );
}
