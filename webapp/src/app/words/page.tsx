"use client";

import { useEffect, useState, useCallback } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { WordList } from "@/components/words/WordList";
import { Search, Plus, X } from "lucide-react";
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
  const { token, loading: authLoading } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [words, setWords] = useState<WordItem[]>([]);
  const [filter, setFilter] = useState<Filter>("all");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(false);

  // Add word state
  const [showAdd, setShowAdd] = useState(false);
  const [addInput, setAddInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [addResult, setAddResult] = useState<{ type: "success" | "error" | "exists"; message: string } | null>(null);

  const fetchWords = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const params = new URLSearchParams({
        filter,
        search,
        page: String(page),
      });
      const res = await authFetch(`/api/words?${params}`);
      const data = await res.json();
      setWords(data.words ?? []);
      setTotalPages(data.totalPages ?? 1);
    } catch {
      setWords([]);
    }
    setLoading(false);
  }, [token, filter, search, page, authFetch]);

  useEffect(() => {
    fetchWords();
  }, [fetchWords]);

  useEffect(() => {
    setPage(1);
  }, [filter, search]);

  const handleAddWord = async () => {
    const word = addInput.trim().toLowerCase();
    if (!word || adding) return;

    setAdding(true);
    setAddResult(null);

    try {
      const res = await authFetch("/api/words", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ word }),
      });

      const data = await res.json();

      if (data.exists) {
        setAddResult({ type: "exists", message: `"${data.word.word}" is already in your list` });
      } else if (data.added) {
        setAddResult({
          type: "success",
          message: `Added: ${data.word.word} — ${data.word.definition}`,
        });
        setAddInput("");
        fetchWords();
      } else {
        setAddResult({ type: "error", message: data.error || "Word not found" });
      }
    } catch {
      setAddResult({ type: "error", message: "Something went wrong" });
    }

    setAdding(false);
  };

  const filters: { label: string; value: Filter }[] = [
    { label: "All", value: "all" },
    { label: "Due", value: "due" },
    { label: "Learned", value: "learned" },
  ];

  return (
    <PageTransition>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold">My Words</h1>
          <button
            onClick={() => {
              setShowAdd(!showAdd);
              setAddResult(null);
              setAddInput("");
            }}
            className={cn(
              "flex items-center gap-1 px-3 py-1.5 rounded-lg text-sm transition-colors",
              showAdd ? "bg-red/10 text-red" : "bg-accent text-white"
            )}
          >
            {showAdd ? <X size={14} /> : <Plus size={14} />}
            {showAdd ? "Close" : "Add word"}
          </button>
        </div>

        {showAdd && (
          <div className="bg-surface rounded-xl p-4 space-y-3">
            <div className="flex gap-2">
              <input
                type="text"
                placeholder="Type an English word..."
                value={addInput}
                onChange={(e) => {
                  setAddInput(e.target.value);
                  setAddResult(null);
                }}
                onKeyDown={(e) => e.key === "Enter" && handleAddWord()}
                className="flex-1 bg-surface-2 rounded-lg px-3 py-2.5 text-sm text-text placeholder:text-text-muted outline-none focus:ring-1 focus:ring-accent"
                autoFocus
              />
              <button
                onClick={handleAddWord}
                disabled={adding || !addInput.trim()}
                className="px-4 py-2.5 bg-accent text-white rounded-lg text-sm font-medium disabled:opacity-50 transition-opacity"
              >
                {adding ? "..." : "Add"}
              </button>
            </div>
            {addResult && (
              <p
                className={cn(
                  "text-sm",
                  addResult.type === "success" && "text-green",
                  addResult.type === "error" && "text-red",
                  addResult.type === "exists" && "text-yellow"
                )}
              >
                {addResult.message}
              </p>
            )}
          </div>
        )}

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

        {authLoading || loading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : !token ? (
          <p className="text-center text-text-muted py-8">Open this app from Telegram</p>
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
