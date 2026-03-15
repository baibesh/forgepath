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
    word: { hour: user.word_hour, min: user.word_min },
    writing: { hour: user.writing_hour, min: user.writing_min },
    media: { hour: user.media_hour, min: user.media_min },
    review: { hour: user.review_hour, min: user.review_min },
  });
}

function isValidTime(hour: unknown, min: unknown): boolean {
  return (
    typeof hour === "number" &&
    typeof min === "number" &&
    Number.isInteger(hour) &&
    Number.isInteger(min) &&
    hour >= 0 &&
    hour <= 23 &&
    min >= 0 &&
    min <= 59
  );
}

export async function PUT(req: Request) {
  const userId = await getUserIdFromRequest(req);
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const body = await req.json();
  const { word, writing, media, review } = body;

  const data: Record<string, number> = {};

  if (word && isValidTime(word.hour, word.min)) {
    data.word_hour = word.hour;
    data.word_min = word.min;
  }
  if (writing && isValidTime(writing.hour, writing.min)) {
    data.writing_hour = writing.hour;
    data.writing_min = writing.min;
  }
  if (media && isValidTime(media.hour, media.min)) {
    data.media_hour = media.hour;
    data.media_min = media.min;
  }
  if (review && isValidTime(review.hour, review.min)) {
    data.review_hour = review.hour;
    data.review_min = review.min;
  }

  if (Object.keys(data).length === 0) {
    return NextResponse.json({ error: "No valid fields" }, { status: 400 });
  }

  await prisma.user.update({ where: { id: userId }, data });

  return NextResponse.json({ ok: true });
}
