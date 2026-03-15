"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { ChevronDown, ExternalLink, CheckCircle, Clock } from "lucide-react";

interface MediaData {
  id: number;
  title: string;
  url: string;
  duration: string;
  level: string;
  watched: boolean;
  responded: boolean;
  response: string | null;
  sentAt: string;
}

export function MediaCard({ media }: { media: MediaData }) {
  const [open, setOpen] = useState(false);

  const date = new Date(media.sentAt).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });

  return (
    <div className="bg-surface rounded-xl p-3">
      <div className="flex items-center justify-between" onClick={() => setOpen(!open)}>
        <div className="flex-1">
          <div className="font-medium text-sm">{media.title}</div>
          <div className="text-xs text-text-muted flex items-center gap-2 mt-0.5">
            <span>{date}</span>
            <span>{media.duration}</span>
            {media.responded ? (
              <span className="flex items-center gap-0.5 text-green-500">
                <CheckCircle size={10} /> Done
              </span>
            ) : media.watched ? (
              <span className="flex items-center gap-0.5 text-amber-500">
                <Clock size={10} /> Watched
              </span>
            ) : (
              <span className="text-text-muted">Sent</span>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <a
            href={media.url}
            target="_blank"
            rel="noopener noreferrer"
            onClick={(e) => e.stopPropagation()}
            className="text-accent"
          >
            <ExternalLink size={16} />
          </a>
          {media.response && (
            <ChevronDown
              size={16}
              className={cn("text-text-muted transition-transform", open && "rotate-180")}
            />
          )}
        </div>
      </div>
      {open && media.response && (
        <div className="mt-3 pt-3 border-t border-border text-sm">
          <div className="text-xs text-text-muted mb-1">Your response</div>
          <p className="text-text whitespace-pre-wrap">{media.response}</p>
        </div>
      )}
    </div>
  );
}
