import { createHmac } from "crypto";

export function validateInitData(initDataRaw: string, botToken: string): boolean {
  const params = new URLSearchParams(initDataRaw);
  const hash = params.get("hash");
  if (!hash) return false;

  params.delete("hash");
  const entries = Array.from(params.entries());
  entries.sort(([a], [b]) => a.localeCompare(b));
  const dataCheckString = entries.map(([k, v]) => `${k}=${v}`).join("\n");

  const secretKey = createHmac("sha256", "WebAppData").update(botToken).digest();
  const computedHash = createHmac("sha256", secretKey)
    .update(dataCheckString)
    .digest("hex");

  return computedHash === hash;
}

export function parseInitData(initDataRaw: string) {
  const params = new URLSearchParams(initDataRaw);
  const userStr = params.get("user");
  if (!userStr) return null;
  try {
    return JSON.parse(userStr) as {
      id: number;
      first_name: string;
      last_name?: string;
      username?: string;
      language_code?: string;
    };
  } catch {
    return null;
  }
}
