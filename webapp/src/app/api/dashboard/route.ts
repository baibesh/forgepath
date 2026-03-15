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

  const now = new Date();
  const userOffset = user.tz_offset ?? 0;
  const userNow = new Date(now.getTime() + userOffset * 3600_000);
  const today = userNow.toISOString().slice(0, 10);

  const thirtyDaysAgo = new Date(userNow);
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

  const streaks = await prisma.streak.findMany({
    where: {
      user_id: userId,
      date: { gte: thirtyDaysAgo },
    },
    orderBy: { date: "asc" },
  });

  const streakMap = streaks.map((s) => ({
    date: s.date.toISOString().slice(0, 10),
    wordDone: s.word_done ?? false,
    writingDone: s.writing_done ?? false,
    reviewDone: s.review_done ?? false,
  }));

  // Current streak calculation
  let currentStreak = 0;
  const todayDate = new Date(today);
  for (let i = 0; i < 365; i++) {
    const d = new Date(todayDate);
    d.setDate(d.getDate() - i);
    const dateStr = d.toISOString().slice(0, 10);
    const s = streakMap.find((x) => x.date === dateStr);
    if (s && (s.wordDone || s.writingDone || s.reviewDone)) {
      currentStreak++;
    } else {
      break;
    }
  }

  const wordCount = await prisma.userWord.count({ where: { user_id: userId } });
  const writingCount = await prisma.writing.count({ where: { user_id: userId } });

  const grammarWeek = await prisma.grammarWeek.findFirst({
    where: {
      week_num: user.current_grammar_week,
      language: user.language ?? "en",
    },
  });

  return NextResponse.json({
    firstName: user.first_name,
    streakDays: streakMap,
    currentStreak,
    wordCount,
    writingCount,
    grammarWeek: grammarWeek
      ? {
          weekNum: grammarWeek.week_num,
          tenseName: grammarWeek.tense_name,
          family: grammarWeek.family,
          focus: grammarWeek.focus,
        }
      : null,
  });
}
