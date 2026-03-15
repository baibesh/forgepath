"use client";

import { useEffect, useState, useCallback } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { WordList } from "@/components/words/WordList";
import { Search } from "lucide-react";
import { cn } from "@/lib/utils";

type Filter = "all" | "due" | "learned";

interface WordItem {
  id: number;
  word: string;
  definition: string | null;
  example: string | null;
  level: string;
  intervalDays: number;
  repetitions: number;
  srsStatus: "learned" | "learning" | "due";
}

export default function WordsPage() {
  const { token } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [words, setWords] = useState<WordItem[]>([]);
  const [filter, setFilter] = useState<Filter>("all");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);

  const fetchWords = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    const params = new URLSearchParams({
      filter,
      search,
      page: String(page),
    });
    const res = await authFetch(`/api/words?${params}`);
    const data = await res.json();
    setWords(data.words);
    setTotalPages(data.totalPages);
    setLoading(false);
  }, [token, filter, search, page, authFetch]);

  useEffect(() => {
    fetchWords();
  }, [fetchWords]);

  useEffect(() => {
    setPage(1);
  }, [filter, search]);

  const filters: { label: string; value: Filter }[] = [
    { label: "All", value: "all" },
    { label: "Due", value: "due" },
    { label: "Learned", value: "learned" },
  ];

  return (
    <PageTransition>
      <div className="space-y-4">
        <h1 className="text-xl font-semibold">My Words</h1>

        <div className="relative">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            type="text"
            placeholder="Search words..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-surface rounded-xl pl-9 pr-4 py-2.5 text-sm text-text placeholder:text-text-muted outline-none focus:ring-1 focus:ring-accent"
          />
        </div>

        <div className="flex gap-2">
          {filters.map((f) => (
            <button
              key={f.value}
              onClick={() => setFilter(f.value)}
              className={cn(
                "px-3 py-1.5 rounded-lg text-sm transition-colors",
                filter === f.value
                  ? "bg-accent text-white"
                  : "bg-surface text-text-muted"
              )}
            >
              {f.label}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : (
          <WordList words={words} />
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
