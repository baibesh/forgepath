"use client";

import { useEffect, useState, useCallback } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { WritingList } from "@/components/writings/WritingList";

interface WritingData {
  id: number;
  topic: string | null;
  grammarFocus: string | null;
  text: string | null;
  feedback: string | null;
  wordCount: number;
  createdAt: string;
}

export default function WritingsPage() {
  const { token } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [writings, setWritings] = useState<WritingData[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);

  const fetchWritings = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    const res = await authFetch(`/api/writings?page=${page}`);
    const data = await res.json();
    setWritings(data.writings);
    setTotalPages(data.totalPages);
    setLoading(false);
  }, [token, page, authFetch]);

  useEffect(() => {
    fetchWritings();
  }, [fetchWritings]);

  return (
    <PageTransition>
      <div className="space-y-4">
        <h1 className="text-xl font-semibold">My Writings</h1>

        {loading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : (
          <WritingList writings={writings} />
        )}

        {totalPages > 1 && (
          <div className="flex justify-center gap-2 pt-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-3 py-1.5 rounded-lg text-sm bg-surface text-text-muted disabled:opacity-30"
            >
              Prev
            </button>
            <span className="px-3 py-1.5 text-sm text-text-muted">
              {page}/{totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="px-3 py-1.5 rounded-lg text-sm bg-surface text-text-muted disabled:opacity-30"
            >
              Next
            </button>
          </div>
        )}
      </div>
    </PageTransition>
  );
}
