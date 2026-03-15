"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { ChevronDown, MessageSquare } from "lucide-react";

interface WritingData {
  id: number;
  topic: string | null;
  grammarFocus: string | null;
  text: string | null;
  feedback: string | null;
  wordCount: number;
  createdAt: string;
}

export function WritingCard({ writing }: { writing: WritingData }) {
  const [open, setOpen] = useState(false);

  const date = new Date(writing.createdAt).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });

  return (
    <div className="bg-surface rounded-xl p-3" onClick={() => setOpen(!open)}>
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <div className="font-medium text-sm">{writing.topic || "Untitled"}</div>
          <div className="text-xs text-text-muted flex items-center gap-2">
            <span>{date}</span>
            <span>{writing.wordCount} words</span>
            {writing.grammarFocus && <span>{writing.grammarFocus}</span>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {writing.feedback && <MessageSquare size={14} className="text-accent" />}
          <ChevronDown
            size={16}
            className={cn("text-text-muted transition-transform", open && "rotate-180")}
          />
        </div>
      </div>
      {open && (
        <div className="mt-3 pt-3 border-t border-border text-sm space-y-3">
          {writing.text && (
            <div>
              <div className="text-xs text-text-muted mb-1">Your text</div>
              <p className="text-text whitespace-pre-wrap">{writing.text}</p>
            </div>
          )}
          {writing.feedback && (
            <div>
              <div className="text-xs text-text-muted mb-1">AI Feedback</div>
              <p className="text-text-muted whitespace-pre-wrap">{writing.feedback}</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
