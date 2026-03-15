"use client";

import { useEffect, useState, useCallback } from "react";
import { useAuthFetch } from "@/hooks/useAuth";
import { useTelegramAuth } from "@/providers/TelegramProvider";
import { PageTransition } from "@/components/layout/PageTransition";
import { MediaList } from "@/components/media/MediaList";

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

export default function MediaPage() {
  const { token, loading: authLoading } = useTelegramAuth();
  const authFetch = useAuthFetch();
  const [media, setMedia] = useState<MediaData[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchMedia = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await authFetch("/api/media");
      const data = await res.json();
      setMedia(data.media ?? []);
    } catch {
      setMedia([]);
    }
    setLoading(false);
  }, [token, authFetch]);

  useEffect(() => {
    fetchMedia();
  }, [fetchMedia]);

  return (
    <PageTransition>
      <div className="space-y-4">
        <h1 className="text-xl font-semibold">My Videos</h1>

        {authLoading || loading ? (
          <div className="flex justify-center py-8">
            <div className="w-6 h-6 border-2 border-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : !token ? (
          <p className="text-center text-text-muted py-8">Open this app from Telegram</p>
        ) : media.length === 0 ? (
          <p className="text-center text-text-muted py-8">No videos yet. They come at 18:00!</p>
        ) : (
          <MediaList media={media} />
        )}
      </div>
    </PageTransition>
  );
}
