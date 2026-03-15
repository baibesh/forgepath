# Seed Scripts — заполнение базы данных

## Требования

В `.env` должны быть:

```
DATABASE_URL=postgres://user:pass@host:5432/forgepath
OPENAI_API_KEY=sk-...
YOUTUBE_API_KEY=AIza...
```

YouTube API ключ: https://console.cloud.google.com/apis/credentials
(включить YouTube Data API v3)

---

## 1. seed-media — YouTube видео

### Что делает

Ищет YouTube-видео через YouTube Data API v3, фильтрует и сохраняет в `media_resources`.

### Как работает

1. Берёт список поисковых запросов из `searchQueries` (`cmd/seed-media/main.go`)
2. Для каждого запроса вызывает YouTube Search API (`/v3/search`)
3. Получает детали видео (`/v3/videos` — duration, views, captions)
4. Фильтрует:
   - **>50,000 просмотров**
   - **Длительность 2-15 минут**
   - Проверяет наличие субтитров
5. Извлекает теги из заголовка и описания
6. Вставляет в `media_resources` с `ON CONFLICT (url) DO UPDATE`

### Параметры YouTube API

| Параметр | Значение |
|---|---|
| API endpoint (search) | `youtube.googleapis.com/youtube/v3/search` |
| API endpoint (details) | `youtube.googleapis.com/youtube/v3/videos` |
| part | `snippet` (search), `contentDetails,statistics` (details) |
| type | `video` |
| videoDuration | `medium` (4-20 min) |
| relevanceLanguage | `en` для английского, `de` для немецкого |
| maxResults | `10` на запрос |

### YouTube API квоты

- Лимит: **10,000 units/день** (бесплатный tier)
- Search = 100 units, Video details = 1 unit
- ~95 запросов = ~9,600 units
- Если получаешь `403` — квота исчерпана, подожди до 00:00 PST (10:00 UTC+3)

### Поисковые запросы

Определены в `var searchQueries` в `cmd/seed-media/main.go`:

| Язык | Уровень | Кол-во запросов | Темы |
|---|---|---|---|
| EN | A1 | 10 | basics, greetings, alphabet, listening |
| EN | A2 | 30 | grammar (tenses), phrasal verbs, daily life, listening, practical |
| EN | B1 | 15 | conditionals, reported speech, passive, idioms, business |
| EN | B2 | 12 | advanced grammar, academic vocab, TED talks, presentations |
| DE | A1 | 10 | Grundwortschatz, Begrüßung, Artikel, Präsens |
| DE | A2 | 21 | Grammatik, Alltag, Wortschatz, Hörverstehen, praktisch |
| DE | B1 | 12 | Konjunktiv, Nebensätze, Diskussion, Nachrichten, Beruf |

### Что сохраняется в БД

```sql
INSERT INTO media_resources (title, url, media_type, level, topic, duration, tags,
                             view_count, has_subtitles, description, active, language)
```

### Запуск

```bash
go run cmd/seed-media/main.go
```

Идемпотентный — `ON CONFLICT (url) DO UPDATE`, можно запускать повторно.
При повторном запуске обновит view_count, tags, описание.

### Добавление новых запросов

Добавь строку в `searchQueries`:

```go
{"your search query here", "topic,tags", "A2", "en"},
```

Поля:
- `query` — поисковый запрос YouTube
- `topic` — теги через запятую (для поиска в боте)
- `level` — CEFR уровень (A1/A2/B1/B2)
- `language` — `en` или `de`

---

## 2. seed-words — слова через OpenAI

### Что делает

Генерирует слова через OpenAI GPT-4o mini и сохраняет в `words`.

### Как работает

1. Берёт список категорий из `categories` (`cmd/seed-words/main.go`)
2. Для каждой категории отправляет промпт в OpenAI
3. Получает JSON-массив слов
4. Вставляет в `words` с `ON CONFLICT (word, language) DO NOTHING`

### Формат генерации

Каждое слово содержит:

```json
{
  "word": "figure out",
  "definition": "понять, разобраться",
  "example": "I finally figured out how to use this app.",
  "collocations": "figure out a problem, figure out the answer",
  "construction": "figure out + how/what/why",
  "level": "A2"
}
```

### Категории

| Язык | Уровень | Категории | Кол-во слов |
|---|---|---|---|
| EN | A1 | everyday words, basic phrases | ~50 |
| EN | A2 | phrasal verbs, connectors, verb constructions, adjectives, verbs, prepositions, fillers, questions | ~200 |
| EN | B1 | intermediate phrasal verbs, academic vocab, opinion phrases, advanced connectors | ~110 |
| EN | B2 | upper-intermediate vocab, idioms, business english | ~80 |
| DE | A1 | Alltagswörter, Grundphrasen | ~50 |
| DE | A2 | Verben mit Präpositionen, Konnektoren, trennbare Verben, Adjektive, reflexive Verben, Redewendungen | ~170 |
| DE | B1 | fortgeschrittene Verben, Konnektoren, Berufssprache, Redewendungen | ~100 |

### Что сохраняется в БД

```sql
INSERT INTO words (word, definition, example, level, collocations, construction, language)
```

### Запуск

```bash
go run cmd/seed-words/main.go
```

Идемпотентный — `ON CONFLICT (word, language) DO NOTHING`, дубликаты пропускаются.

### Добавление новых категорий

Добавь блок в `categories`:

```go
{`Generate 30 NEW_CATEGORY for LEVEL Language learners (Russian-speaking).
Description of what to generate.
Examples: example1, example2, example3.
DO NOT include these examples, generate NEW ones.`, 30, "en"},
```

Поля:
- `prompt` — промпт для OpenAI (описание + примеры)
- `count` — сколько слов генерировать
- `language` — `en` или `de`

### Стоимость

GPT-4o mini: ~$0.01 за батч. Все 29 батчей: ~$0.30.

---

## Порядок запуска

```bash
# 1. Убедись что .env заполнен
cat .env

# 2. Сначала media (YouTube API — есть дневной лимит)
go run cmd/seed-media/main.go

# 3. Потом words (OpenAI — без жёстких лимитов)
go run cmd/seed-words/main.go
```

### Проверка результатов

```bash
# Подключиться к БД и проверить
psql $DATABASE_URL -c "SELECT language, level, COUNT(*) FROM words GROUP BY language, level ORDER BY language, level;"
psql $DATABASE_URL -c "SELECT language, level, COUNT(*) FROM media_resources GROUP BY language, level ORDER BY language, level;"
```

---

## Миграции

Миграции запускаются автоматически при старте seed-скриптов и при старте бота.
Используется **goose** — SQL-файлы в `db/migrations/`.
Ручной запуск не нужен.
