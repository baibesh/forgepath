"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  type ReactNode,
} from "react";

interface TelegramContext {
  token: string | null;
  userId: number | null;
  loading: boolean;
}

const TgCtx = createContext<TelegramContext>({
  token: null,
  userId: null,
  loading: true,
});

export function useTelegramAuth() {
  return useContext(TgCtx);
}

export function TelegramProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<TelegramContext>({
    token: null,
    userId: null,
    loading: true,
  });

  const authenticate = useCallback(async () => {
    try {
      const tg = window.Telegram?.WebApp;
      if (!tg) {
        setState({ token: null, userId: null, loading: false });
        return;
      }

      tg.ready();
      tg.expand();

      const initDataRaw = tg.initData;
      if (!initDataRaw) {
        setState({ token: null, userId: null, loading: false });
        return;
      }

      const res = await fetch("/api/auth", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ initData: initDataRaw }),
      });

      if (!res.ok) {
        setState({ token: null, userId: null, loading: false });
        return;
      }

      const data = await res.json();
      setState({
        token: data.token,
        userId: data.userId,
        loading: false,
      });
    } catch {
      setState({ token: null, userId: null, loading: false });
    }
  }, []);

  useEffect(() => {
    authenticate();
  }, [authenticate]);

  return <TgCtx.Provider value={state}>{children}</TgCtx.Provider>;
}
