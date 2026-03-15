"use client";

import { useCallback } from "react";
import { useTelegramAuth } from "@/providers/TelegramProvider";

export function useAuthFetch() {
  const { token } = useTelegramAuth();

  return useCallback(
    async (url: string, options?: RequestInit) => {
      const headers = new Headers(options?.headers);
      if (token) headers.set("Authorization", `Bearer ${token}`);
      return fetch(url, { ...options, headers });
    },
    [token]
  );
}
