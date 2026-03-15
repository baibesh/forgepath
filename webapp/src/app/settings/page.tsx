"use client";

import { useEffect, useState } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { SettingsForm } from "@/components/settings/SettingsForm";

interface SettingsData {
  language: string;
  level: string;
  tzOffset: number;
}

export default function SettingsPage() {
  const { token, loading: authLoading } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [settings, setSettings] = useState<SettingsData | null>(null);

  useEffect(() => {
    if (!token) return;
    authFetch("/api/settings")
      .then((r) => r.json())
      .then(setSettings)
      .catch(() => {});
  }, [token, authFetch]);

  return (
    <PageTransition>
      <div className="space-y-4">
        <h1 className="text-xl font-semibold">Settings</h1>
        {authLoading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : !token ? (
          <p className="text-center text-text-muted py-8">Open this app from Telegram</p>
        ) : settings ? (
          <SettingsForm initial={settings} />
        ) : (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        )}
      </div>
    </PageTransition>
  );
}
