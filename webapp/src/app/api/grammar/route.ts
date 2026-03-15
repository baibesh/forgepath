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

  const weeks = await prisma.grammarWeek.findMany({
    where: { language: user.language ?? "en" },
    orderBy: { week_num: "asc" },
  });

  return NextResponse.json({
    weeks: weeks.map((w) => ({
      weekNum: w.week_num,
      family: w.family,
      focus: w.focus,
      tenseName: w.tense_name,
      anchor: w.anchor,
      markers: w.markers,
      formula: w.formula,
      example: w.example,
    })),
    currentWeek: user.current_grammar_week,
  });
}
