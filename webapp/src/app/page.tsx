"use client";

import { useEffect, useState } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { StreakCalendar } from "@/components/dashboard/StreakCalendar";
import { StatsCards } from "@/components/dashboard/StatsCards";
import { GrammarBadge } from "@/components/dashboard/GrammarBadge";

interface DashboardData {
  firstName: string;
  streakDays: { date: string; wordDone: boolean; writingDone: boolean; reviewDone: boolean }[];
  currentStreak: number;
  wordCount: number;
  writingCount: number;
  grammarWeek: { weekNum: number; tenseName: string; family: string; focus: string } | null;
}

export default function DashboardPage() {
  const { token, loading } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [data, setData] = useState<DashboardData | null>(null);

  useEffect(() => {
    if (!token) return;
    authFetch("/api/dashboard")
      .then((r) => r.json())
      .then(setData);
  }, [token, authFetch]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!token) {
    return (
      <div className="text-center pt-20 text-text-muted">
        <p>Open this app from Telegram</p>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <PageTransition>
      <div className="space-y-5">
        <h1 className="text-xl font-semibold">
          Hey, {data.firstName || "there"}
        </h1>
        <StreakCalendar days={data.streakDays} />
        <StatsCards
          streak={data.currentStreak}
          words={data.wordCount}
          writings={data.writingCount}
        />
        {data.grammarWeek && (
          <GrammarBadge
            weekNum={data.grammarWeek.weekNum}
            tenseName={data.grammarWeek.tenseName}
            family={data.grammarWeek.family}
            focus={data.grammarWeek.focus}
          />
        )}
      </div>
    </PageTransition>
  );
}
