"use client";

import { useState } from "react";
import { useAuthFetch } from "@/hooks/useAuth";

interface TimeSlot {
  hour: number;
  min: number;
}

interface ScheduleData {
  word: TimeSlot;
  writing: TimeSlot;
  media: TimeSlot;
  review: TimeSlot;
}

const TASKS = [
  { key: "word" as const, label: "New word", icon: "\u2B50" },
  { key: "writing" as const, label: "Writing", icon: "\u270D\uFE0F" },
  { key: "media" as const, label: "Media", icon: "\uD83C\uDFAC" },
  { key: "review" as const, label: "Review", icon: "\uD83C\uDF1B" },
];

function formatTime(hour: number, min: number): string {
  return `${String(hour).padStart(2, "0")}:${String(min).padStart(2, "0")}`;
}

export function ScheduleForm({ initial }: { initial: ScheduleData }) {
  const authFetch = useAuthFetch();
  const [schedule, setSchedule] = useState<ScheduleData>(initial);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const updateSlot = (key: keyof ScheduleData, value: string) => {
    const [h, m] = value.split(":").map(Number);
    if (isNaN(h) || isNaN(m)) return;
    setSchedule((prev) => ({ ...prev, [key]: { hour: h, min: m } }));
  };

  const handleSave = async () => {
    setSaving(true);
    setSaved(false);
    await authFetch("/api/schedule", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(schedule),
    });
    setSaving(false);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const inputClass =
    "w-full bg-surface-2 rounded-lg px-3 py-2.5 text-sm text-text outline-none focus:ring-1 focus:ring-accent";

  return (
    <div className="space-y-4">
      {TASKS.map(({ key, label, icon }) => (
        <div key={key}>
          <label className="text-sm text-text-muted block mb-1.5">
            {icon} {label}
          </label>
          <input
            type="time"
            value={formatTime(schedule[key].hour, schedule[key].min)}
            onChange={(e) => updateSlot(key, e.target.value)}
            className={inputClass}
          />
        </div>
      ))}

      <button
        onClick={handleSave}
        disabled={saving}
        className="w-full bg-accent text-white rounded-xl py-3 text-sm font-medium disabled:opacity-50 transition-opacity"
      >
        {saving ? "Saving..." : saved ? "Saved!" : "Save Schedule"}
      </button>
    </div>
  );
}
