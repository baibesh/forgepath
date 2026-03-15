import { NextResponse } from "next/server";
import { getUserIdFromRequest } from "@/lib/auth";
import { prisma } from "@/lib/prisma";

export async function GET(req: Request) {
  const userId = await getUserIdFromRequest(req);
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const url = new URL(req.url);
  const page = parseInt(url.searchParams.get("page") || "1");
  const limit = 10;
  const offset = (page - 1) * limit;

  const total = await prisma.writing.count({ where: { user_id: userId } });
  const writings = await prisma.writing.findMany({
    where: { user_id: userId },
    orderBy: { created_at: "desc" },
    skip: offset,
    take: limit,
  });

  return NextResponse.json({
    writings: writings.map((w) => ({
      id: w.id,
      topic: w.topic,
      grammarFocus: w.grammar_focus,
      text: w.text,
      feedback: w.feedback,
      wordCount: w.word_count,
      createdAt: w.created_at.toISOString(),
    })),
    total,
    page,
    totalPages: Math.ceil(total / limit),
  });
}
