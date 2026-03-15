"use client";

import { useEffect, useState } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { WeekTimeline } from "@/components/grammar/WeekTimeline";

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

export default function GrammarPage() {
  const { token } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [weeks, setWeeks] = useState<GrammarWeekData[]>([]);
  const [currentWeek, setCurrentWeek] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    authFetch("/api/grammar")
      .then((r) => r.json())
      .then((data) => {
        setWeeks(data.weeks);
        setCurrentWeek(data.currentWeek);
        setLoading(false);
      });
  }, [token, authFetch]);

  return (
    <PageTransition>
      <div className="space-y-4">
        <h1 className="text-xl font-semibold">Grammar Timeline</h1>
        {loading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : weeks.length === 0 ? (
          <p className="text-center text-text-muted py-8">No grammar weeks available</p>
        ) : (
          <WeekTimeline weeks={weeks} currentWeek={currentWeek} />
        )}
      </div>
    </PageTransition>
  );
}
