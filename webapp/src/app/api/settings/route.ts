import { NextResponse } from "next/server";
import { getUserIdFromRequest } from "@/lib/auth";
import { prisma } from "@/lib/prisma";

export async function GET(req: Request) {
  const userId = await getUserIdFromRequest(req);
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const user = await prisma.user.findUnique({ where: { id: userId } });
  if (!user) {
    return NextResponse.json({ error: "User not found" }, { status: 404 });
  }

  return NextResponse.json({
    language: user.language,
    level: user.level,
    tzOffset: user.tz_offset,
  });
}

export async function PUT(req: Request) {
  const userId = await getUserIdFromRequest(req);
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const body = await req.json();
  const { language, level, tzOffset } = body;

  const data: Record<string, unknown> = {};
  if (language && ["en", "ru", "kk"].includes(language)) data.language = language;
  if (level && ["A1", "A2", "B1", "B2", "C1"].includes(level)) data.level = level;
  if (tzOffset !== undefined && typeof tzOffset === "number" && tzOffset >= -12 && tzOffset <= 14) {
    data.tz_offset = tzOffset;
  }

  if (Object.keys(data).length === 0) {
    return NextResponse.json({ error: "No valid fields" }, { status: 400 });
  }

  await prisma.user.update({ where: { id: userId }, data });

  return NextResponse.json({ ok: true });
}
