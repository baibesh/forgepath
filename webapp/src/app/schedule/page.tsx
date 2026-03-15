"use client";

import { useEffect, useState } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { ScheduleForm } from "@/components/schedule/ScheduleForm";

interface TimeSlot {
  hour: number;
  min: number;
}

interface ScheduleData {
  word: TimeSlot;
  writing: TimeSlot;
  media: TimeSlot;
  review: TimeSlot;
}

export default function SchedulePage() {
  const { token, loading: authLoading } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [schedule, setSchedule] = useState<ScheduleData | null>(null);

  useEffect(() => {
    if (!token) return;
    authFetch("/api/schedule")
      .then((r) => r.json())
      .then(setSchedule)
      .catch(() => {});
  }, [token, authFetch]);

  return (
    <PageTransition>
      <div className="space-y-4">
        <h1 className="text-xl font-semibold">Daily Schedule</h1>
        <p className="text-sm text-text-muted">
          Choose when you want to receive your daily tasks.
        </p>
        {authLoading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : !token ? (
          <p className="text-center text-text-muted py-8">
            Open this app from Telegram
          </p>
        ) : schedule ? (
          <ScheduleForm initial={schedule} />
        ) : (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        )}
      </div>
    </PageTransition>
  );
}
