"use client";

import { useTelegramAuth } from "@/providers/TelegramProvider";

export function Header() {
  const { loading } = useTelegramAuth();

  if (loading) return null;

  return (
    <header className="px-4 pt-4 pb-2">
      <div className="flex items-center gap-2">
        <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center text-white font-bold text-sm">
          F
        </div>
        <span className="text-lg font-semibold text-text">ForgePath</span>
      </div>
    </header>
  );
}
