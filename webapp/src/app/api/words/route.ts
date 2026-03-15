import { NextResponse } from "next/server";
import { getUserIdFromRequest } from "@/lib/auth";
import { prisma } from "@/lib/prisma";

export async function GET(req: Request) {
  const userId = await getUserIdFromRequest(req);
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const url = new URL(req.url);
  const filter = url.searchParams.get("filter") || "all";
  const search = url.searchParams.get("search") || "";
  const page = parseInt(url.searchParams.get("page") || "1");
  const limit = 20;
  const offset = (page - 1) * limit;

  const now = new Date();

  let words;
  let total;

  if (filter === "due") {
    const where = {
      user_id: userId,
      next_review: { lte: now },
      ...(search
        ? { word: { word: { contains: search, mode: "insensitive" as const } } }
        : {}),
    };
    total = await prisma.userWord.count({ where });
    words = await prisma.userWord.findMany({
      where,
      include: { word: true },
      orderBy: { next_review: "asc" },
      skip: offset,
      take: limit,
    });
  } else if (filter === "learned") {
    const where = {
      user_id: userId,
      repetitions: { gte: 3 },
      ...(search
        ? { word: { word: { contains: search, mode: "insensitive" as const } } }
        : {}),
    };
    total = await prisma.userWord.count({ where });
    words = await prisma.userWord.findMany({
      where,
      include: { word: true },
      orderBy: { seen_at: "desc" },
      skip: offset,
      take: limit,
    });
  } else {
    const where = {
      user_id: userId,
      ...(search
        ? { word: { word: { contains: search, mode: "insensitive" as const } } }
        : {}),
    };
    total = await prisma.userWord.count({ where });
    words = await prisma.userWord.findMany({
      where,
      include: { word: true },
      orderBy: { seen_at: "desc" },
      skip: offset,
      take: limit,
    });
  }

  const items = words.map((uw) => ({
    id: uw.word.id,
    word: uw.word.word,
    definition: uw.word.definition,
    example: uw.word.example,
    level: uw.word.level,
    intervalDays: uw.interval_days,
    repetitions: uw.repetitions,
    easeFactor: uw.ease_factor,
    nextReview: uw.next_review?.toISOString() ?? null,
    seenAt: uw.seen_at.toISOString(),
    srsStatus:
      uw.repetitions >= 3 ? "learned" : uw.next_review && uw.next_review <= now ? "due" : "learning",
  }));

  return NextResponse.json({
    words: items,
    total,
    page,
    totalPages: Math.ceil(total / limit),
  });
}
