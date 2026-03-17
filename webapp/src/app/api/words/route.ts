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

interface WordLookup {
  definition: string;
  example: string;
  collocations: string;
  construction: string;
}

async function lookupWord(word: string, userLanguage: string): Promise<WordLookup | null> {
  const apiKey = process.env.OPENAI_API_KEY;
  if (!apiKey) return null;

  const uiLang = userLanguage === "kk" ? "Kazakh" : userLanguage === "en" ? "Russian" : "Russian";

  const prompt = `You are an English language dictionary. The user speaks ${uiLang}.
For the English word/phrase: "${word}"

Return EXACTLY this JSON format, no extra text:
{
  "definition": "short translation/definition in ${uiLang} (1-3 words)",
  "example": "one natural example sentence in English using this word",
  "collocations": "3-4 common collocations, comma separated",
  "construction": "grammar pattern, e.g. 'verb + noun' or 'adjective + about'"
}

If the word doesn't exist or is not English, return:
{"error": "not found"}`;

  try {
    const res = await fetch("https://api.openai.com/v1/chat/completions", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${apiKey}`,
      },
      body: JSON.stringify({
        model: "gpt-4o-mini",
        messages: [{ role: "user", content: prompt }],
        max_tokens: 300,
        temperature: 0.5,
      }),
    });

    if (!res.ok) return null;

    const data = await res.json();
    const text = data.choices?.[0]?.message?.content?.trim();
    if (!text) return null;

    // Strip markdown code fences
    const clean = text.replace(/```json\s*/g, "").replace(/```\s*/g, "").trim();
    const parsed = JSON.parse(clean);

    if (parsed.error) return null;
    if (!parsed.definition) return null;

    return {
      definition: parsed.definition,
      example: parsed.example || "",
      collocations: parsed.collocations || "",
      construction: parsed.construction || "",
    };
  } catch {
    return null;
  }
}

export async function POST(req: Request) {
  const userId = await getUserIdFromRequest(req);
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const body = await req.json();
  const word = (body.word || "").trim().toLowerCase();
  if (!word || word.length > 100) {
    return NextResponse.json({ error: "Invalid word" }, { status: 400 });
  }

  const user = await prisma.user.findUnique({ where: { id: userId } });
  if (!user) {
    return NextResponse.json({ error: "User not found" }, { status: 404 });
  }

  // Check if word exists in DB
  const existing = await prisma.word.findFirst({
    where: { word: { equals: word, mode: "insensitive" }, language: "en" },
  });

  if (existing) {
    // Check if user already has it
    const userWord = await prisma.userWord.findUnique({
      where: { user_id_word_id: { user_id: userId, word_id: existing.id } },
    });

    if (userWord) {
      return NextResponse.json({
        exists: true,
        word: {
          id: existing.id,
          word: existing.word,
          definition: existing.definition,
          example: existing.example,
        },
      });
    }

    // Add to user's words
    await prisma.userWord.create({
      data: {
        user_id: userId,
        word_id: existing.id,
        seen_at: new Date(),
        next_review: new Date(Date.now() + 86400000),
        interval_days: 1,
        ease_factor: 2.5,
        repetitions: 0,
        score: 0,
      },
    });

    return NextResponse.json({
      added: true,
      word: {
        id: existing.id,
        word: existing.word,
        definition: existing.definition,
        example: existing.example,
      },
    });
  }

  // Look up via OpenAI
  const info = await lookupWord(word, user.language);
  if (!info) {
    return NextResponse.json({ error: "Word not found" }, { status: 404 });
  }

  // Insert word
  const newWord = await prisma.word.create({
    data: {
      word,
      definition: info.definition,
      example: info.example,
      collocations: info.collocations,
      construction: info.construction,
      level: user.level,
      language: "en",
    },
  });

  // Link to user
  await prisma.userWord.create({
    data: {
      user_id: userId,
      word_id: newWord.id,
      seen_at: new Date(),
      next_review: new Date(Date.now() + 86400000),
      interval_days: 1,
      ease_factor: 2.5,
      repetitions: 0,
      score: 0,
    },
  });

  return NextResponse.json({
    added: true,
    word: {
      id: newWord.id,
      word: newWord.word,
      definition: newWord.definition,
      example: newWord.example,
    },
  });
}
