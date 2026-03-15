import { NextResponse } from "next/server";
import { getUserIdFromRequest } from "@/lib/auth";
import { prisma } from "@/lib/prisma";

export async function GET(req: Request) {
  const userId = await getUserIdFromRequest(req);
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const userMedia = await prisma.userMedia.findMany({
    where: { user_id: userId },
    orderBy: { sent_at: "desc" },
    take: 20,
    include: { media: true },
  });

  return NextResponse.json({
    media: userMedia.map((um) => ({
      id: um.media_id,
      title: um.media.title,
      url: um.media.url,
      duration: um.media.duration,
      level: um.media.level,
      watched: um.task_sent,
      responded: um.completed,
      response: um.task_response || null,
      sentAt: um.sent_at.toISOString(),
    })),
  });
}
