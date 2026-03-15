import { NextResponse } from "next/server";
import { validateInitData, parseInitData } from "@/lib/telegram";
import { createToken } from "@/lib/auth";
import { prisma } from "@/lib/prisma";

export async function POST(req: Request) {
  const { initData } = await req.json();
  if (!initData) {
    return NextResponse.json({ error: "Missing initData" }, { status: 400 });
  }

  const botToken = process.env.BOT_TOKEN;
  if (!botToken) {
    return NextResponse.json({ error: "Server misconfigured" }, { status: 500 });
  }

  if (!validateInitData(initData, botToken)) {
    return NextResponse.json({ error: "Invalid initData" }, { status: 401 });
  }

  const user = parseInitData(initData);
  if (!user) {
    return NextResponse.json({ error: "Cannot parse user" }, { status: 400 });
  }

  const dbUser = await prisma.user.findUnique({
    where: { id: BigInt(user.id) },
  });

  if (!dbUser) {
    return NextResponse.json({ error: "User not found" }, { status: 404 });
  }

  const token = await createToken(user.id);
  return NextResponse.json({
    token,
    userId: user.id,
    firstName: dbUser.first_name,
  });
}
