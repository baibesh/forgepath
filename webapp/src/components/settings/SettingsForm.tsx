"use client";

import { useState } from "react";
import { useAuthFetch } from "@/hooks/useAuth";

interface SettingsData {
  language: string;
  level: string;
  tzOffset: number;
}

export function SettingsForm({ initial }: { initial: SettingsData }) {
  const authFetch = useAuthFetch();
  const [language, setLanguage] = useState(initial.language);
  const [level, setLevel] = useState(initial.level);
  const [tzOffset, setTzOffset] = useState(initial.tzOffset);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    setSaved(false);
    await authFetch("/api/settings", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ language, level, tzOffset }),
    });
    setSaving(false);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const selectClass =
    "w-full bg-surface-2 rounded-lg px-3 py-2.5 text-sm text-text outline-none focus:ring-1 focus:ring-accent appearance-none";

  return (
    <div className="space-y-5">
      <div>
        <label className="text-sm text-text-muted block mb-1.5">Language</label>
        <select value={language} onChange={(e) => setLanguage(e.target.value)} className={selectClass}>
          <option value="en">English</option>
          <option value="de">Deutsch</option>
        </select>
      </div>

      <div>
        <label className="text-sm text-text-muted block mb-1.5">Level</label>
        <select value={level} onChange={(e) => setLevel(e.target.value)} className={selectClass}>
          {["A1", "A2", "B1", "B2", "C1"].map((l) => (
            <option key={l} value={l}>{l}</option>
          ))}
        </select>
      </div>

      <div>
        <label className="text-sm text-text-muted block mb-1.5">Timezone (UTC offset)</label>
        <select
          value={tzOffset}
          onChange={(e) => setTzOffset(Number(e.target.value))}
          className={selectClass}
        >
          {Array.from({ length: 27 }, (_, i) => i - 12).map((tz) => (
            <option key={tz} value={tz}>
              UTC{tz >= 0 ? `+${tz}` : tz}
            </option>
          ))}
        </select>
      </div>

      <button
        onClick={handleSave}
        disabled={saving}
        className="w-full bg-accent text-white rounded-xl py-3 text-sm font-medium disabled:opacity-50 transition-opacity"
      >
        {saving ? "Saving..." : saved ? "Saved!" : "Save Settings"}
      </button>
    </div>
  );
}
