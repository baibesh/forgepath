# ForgePath Bot

Study through daily habits. A system, not a course.

---

## Философия

Наука и опыт полиглотов сходятся: **маленькие ежедневные действия** работают лучше любого курса.

Ключевые принципы:
- **Регулярность > интенсивность** — 15-30 мин/день лучше 3 часов в субботу
- **Spaced Repetition** — интервальное повторение (x1.5-2 лучше зубрёжки)
- **Active Recall** — вспоминать, а не перечитывать (до x4 улучшение)
- **Comprehensible Input (i+1)** — контент чуть выше текущего уровня
- **Output (writing/speaking)** — производство языка заставляет мозг замечать пробелы
- **Погружение** — интегрировать английский в повседневную жизнь

Привычка формируется ~66 дней. Streak — инструмент, а не самоцель.

---

## Stack

| Layer | Tech |
|---|---|
| Bot | Go + `telebot.v3` |
| Cron | `robfig/cron` |
| DB | PostgreSQL + `pgx` |
| AI | OpenAI `gpt-4o-mini` |
| Languages | English, Deutsch |
| Config | `.env` + `godotenv` |
| Deploy | GitHub Actions → VPS (systemd) |
| Mini App (v2) | Next.js |

---

## Структура проекта

```
forgepath/
├── main.go
├── .env / .env.example
├── go.mod
│
├── config/
│   └── config.go         # чтение .env
│
├── db/
│   ├── postgres.go       # подключение + миграции
│   ├── users.go          # CRUD пользователей
│   ├── words.go          # слова + user_words
│   ├── streak.go         # логика streak
│   ├── writings.go       # free writing тексты
│   └── state.go          # FSM состояние диалога
│
├── ai/
│   └── openai.go         # GPT-4o mini — фидбек, генерация слов, медиа, квизы
│
├── content/
│   └── content.go        # общие темы для writing/медиа по языкам
│
├── bot/
│   ├── handlers.go       # команды: /start /today /skip /stats /settings
│   ├── callbacks.go      # inline callback handlers (кнопки квиза, действия)
│   ├── keyboards.go      # inline + reply кнопки
│   └── middleware.go      # логирование, проверка регистрации
│
├── cron/
│   ├── scheduler.go      # регистрация всех задач
│   └── jobs.go           # логика каждой рассылки
│
├── srs/
│   └── algorithm.go      # SM-2 алгоритм интервального повторения
│
└── .github/
    └── workflows/
        └── deploy.yml    # CI/CD
```

---

## Дневной цикл обучения

```
07:30  📖 Слово дня (новое) + повторение 2-3 старых слов (quiz)
12:00  ✍️ Free writing (5 мин) — тема + AI фидбек
18:00  🎬 Медиа рекомендация (подкаст/видео/статья по уровню)
21:30  📊 Daily review — вспомни слово дня + итог дня
Вс 09:00  📈 Недельный отчёт — прогресс, слова, streak
```

Все времена — по локальному времени юзера (tz_offset).

---

## Схема БД

```sql
-- Пользователи
CREATE TABLE users (
    id              BIGINT PRIMARY KEY,  -- telegram user id
    username        TEXT,
    first_name      TEXT,
    language        TEXT DEFAULT 'en',   -- en / de
    tz_offset       INT DEFAULT 5,       -- часовой пояс (UTC+N)
    level           TEXT DEFAULT 'A2',   -- A2 / B1 / B2 / C1
    active          BOOL DEFAULT TRUE,
    skip_count      INT DEFAULT 0,       -- использовано skip-ов в текущей неделе
    created_at      TIMESTAMP DEFAULT NOW()
);

-- FSM — состояние диалога
CREATE TABLE user_state (
    user_id         BIGINT PRIMARY KEY REFERENCES users(id),
    state           TEXT DEFAULT 'idle',  -- idle / waiting_writing / waiting_quiz / waiting_feedback
    context         JSONB DEFAULT '{}',   -- доп данные (topic, word_id, quiz_options и т.д.)
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- Streak (дневная отметка)
CREATE TABLE streaks (
    user_id         BIGINT REFERENCES users(id),
    date            DATE,
    word_done       BOOL DEFAULT FALSE,
    writing_done    BOOL DEFAULT FALSE,
    review_done     BOOL DEFAULT FALSE,
    PRIMARY KEY (user_id, date)
);

-- Словарь (глобальный)
CREATE TABLE words (
    id              SERIAL PRIMARY KEY,
    word            TEXT NOT NULL UNIQUE,
    definition      TEXT,
    example         TEXT,
    level           TEXT,  -- A2 / B1 / B2 / C1
    language        TEXT DEFAULT 'en',   -- en / de
    created_at      TIMESTAMP DEFAULT NOW()
);

-- Связь юзер-слово + SRS
CREATE TABLE user_words (
    user_id         BIGINT REFERENCES users(id),
    word_id         INT REFERENCES words(id),
    seen_at         TIMESTAMP DEFAULT NOW(),
    next_review     TIMESTAMP,           -- когда повторить (SRS)
    interval_days   INT DEFAULT 1,       -- текущий интервал SM-2
    ease_factor     REAL DEFAULT 2.5,    -- фактор лёгкости SM-2
    repetitions     INT DEFAULT 0,       -- кол-во успешных повторений
    score           INT DEFAULT 0,       -- последняя оценка (0-5)
    PRIMARY KEY (user_id, word_id)
);

-- Free writing тексты
CREATE TABLE writings (
    id              SERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id),
    topic           TEXT,
    text            TEXT,
    feedback        TEXT,                -- AI фидбек
    word_count      INT,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- Медиа рекомендации (кэш, чтобы не повторять)
CREATE TABLE media_recommendations (
    id              SERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id),
    title           TEXT,
    url             TEXT,
    media_type      TEXT,  -- podcast / video / article
    level           TEXT,
    language        TEXT DEFAULT 'en',   -- en / de
    created_at      TIMESTAMP DEFAULT NOW()
);
```

---

## Spaced Repetition (SM-2)

Алгоритм SuperMemo 2 — проверенный, простой, эффективный.

### Как работает

При каждом повторении юзер оценивает (через кнопки):
- **Забыл** (score 0-1) → сброс, повторить завтра
- **Трудно** (score 2-3) → короткий интервал
- **Легко** (score 4-5) → увеличить интервал

```
if score < 3:
    repetitions = 0
    interval = 1 day
else:
    if repetitions == 0: interval = 1 day
    if repetitions == 1: interval = 3 days
    if repetitions >= 2: interval = prev_interval * ease_factor

ease_factor = max(1.3, ease_factor + (0.1 - (5-score) * (0.08 + (5-score) * 0.02)))
next_review = now + interval
```

### Интеграция с cron

Утренний cron (07:30):
1. Генерирует **новое слово** через OpenAI → сохраняет в `words` + `user_words`
2. Находит **2-3 слова для повторения** (`WHERE next_review <= NOW()`)
3. Отправляет всё одним сообщением: новое слово + quiz по старым

### Quiz формат

```
🔄 Повтори слово:

"persistent" — это значит:
  A) быстрый
  B) настойчивый ✅
  C) ленивый
  D) случайный
```

Варианты генерируются через OpenAI (3 неправильных + 1 правильный).
Ответ через inline кнопки → обновление SRS.

---

## Cron Jobs — детали

### 07:30 — Слово дня + повторение

```
1. Для каждого активного юзера (с учётом tz_offset):
2. OpenAI генерирует новое слово по уровню юзера
3. Проверяем что слово не дубликат (ищем в user_words)
4. Сохраняем в words + user_words (next_review = now + 1 day)
5. Находим слова для повторения (next_review <= now, LIMIT 3)
6. Отправляем: новое слово + quiz по старым
7. Ставим state = 'waiting_quiz' с контекстом (word_ids, answers)
```

### 12:00 — Free Writing

```
1. Для каждого активного юзера:
2. Выбираем тему (пул тем, ротация по дню, учёт уровня)
3. Отправляем тему
4. Ставим state = 'waiting_writing'
5. Когда юзер присылает текст:
   - Сохраняем в writings
   - Отправляем в OpenAI GPT-4o mini для фидбека
   - Сохраняем фидбек
   - Отвечаем юзеру
   - Обновляем streak (writing_done = true)
   - Ставим state = 'idle'
```

### 18:00 — Медиа рекомендация

```
1. Через день (чётные дни) для каждого юзера:
2. OpenAI генерирует рекомендацию по уровню + интересам
3. Проверяем что не дубликат (media_recommendations)
4. Отправляем с кнопками: [Посмотрел ✅] [Пропустить]
5. "Посмотрел" → обновляет streak
```

### 21:30 — Daily Review

```
1. Для каждого юзера:
2. Собираем итоги дня:
   - Слово дня (вспомни!) — ещё один active recall
   - Writing: сделано/нет
   - Streak: текущий
3. Если не всё сделано — мягкое напоминание
4. Обновляем streak (review_done = true)
```

### Вс 09:00 — Недельный отчёт

```
1. OpenAI анализирует данные за неделю:
   - Слов выучено / повторено
   - Writing-ов написано
   - Средняя длина текстов (растёт ли?)
   - Streak дней
   - Слабые слова (низкий score)
2. Формирует персональный отчёт
3. Сбрасываем skip_count
```

---

## FSM (Finite State Machine) — управление диалогом

Юзер может быть в одном из состояний:

| State | Описание | Ожидаемый ввод |
|---|---|---|
| `idle` | Ничего не ждём | Команды |
| `waiting_writing` | Ждём текст free writing | Любой текст → AI фидбек |
| `waiting_quiz` | Ждём ответ на quiz | Inline кнопка A/B/C/D |
| `waiting_feedback` | AI обрабатывает | Ничего (показываем typing) |

Хранится в `user_state`. При каждом входящем сообщении проверяем state и роутим.

---

## Cron + часовые пояса

Cron работает в UTC. Для каждого задания:

```go
// Запускаем каждые 30 минут, проверяем кому пора
c.AddFunc("*/30 * * * *", func() {
    now := time.Now().UTC()
    users := database.GetActiveUsers()
    for _, user := range users {
        localHour := (now.Hour() + user.TzOffset) % 24
        localMinute := now.Minute()

        switch {
        case localHour == 7 && localMinute == 30:
            SendWordOfDay(b, database, user)
        case localHour == 12 && localMinute == 0:
            SendFreeWriting(b, database, user)
        case localHour == 18 && localMinute == 0:
            SendMediaRec(b, database, user)
        case localHour == 21 && localMinute == 30:
            SendDailyReview(b, database, user)
        }
    }
})

// Недельный — раз в час проверяем воскресенье
c.AddFunc("0 * * * 0", func() {
    // аналогично, проверяем localHour == 9
})
```

---

## AI интеграция

### Распределение моделей

| Задача | Модель | Вызовов/юзер/день | Почему |
|---|---|---|---|
| Фидбек по writing | GPT-4o mini | 1 | Лучше понимает нюансы языка |
| Генерация слова дня | GPT-4o mini | 1 | Единая модель, простой стек |
| Quiz варианты | GPT-4o mini | 1 | Простая генерация |
| Медиа рекомендация | GPT-4o mini | 0.5 (через день) | Не критично |
| Недельный отчёт | GPT-4o mini | 0.14 (раз/нед) | Структурированный вывод |
| Daily review подсказка | — | 0 | Берём из БД, без AI |

### Стоимость

| Юзеров | OpenAI/мес |
|---|---|
| 50 | ~$2 |
| 200 | ~$8 |
| 1000 | ~$35 |

### Промпты

**Генерация слова:**
```
Give one English word for level {level} that the student hasn't seen before.
Avoid these words: {seen_words_list}
Format:
Word: [word]
Definition: [clear, simple definition]
Example: [natural sentence a native would say]
Tip: [one memory tip, collocation, or common usage note]
```

**Quiz генерация:**
```
The correct answer is "{word}" meaning "{definition}".
Generate 3 wrong but plausible options for a multiple choice quiz.
Return as JSON: ["wrong1", "wrong2", "wrong3"]
```

**Free writing фидбек:**
```
You are an English tutor. Student level: {level}.
Review this free writing:
- 2-3 grammar/style corrections (with examples)
- 1 thing done well
- 1 actionable tip
Keep under 100 words. Be direct but encouraging.

Text: {text}
```

**Медиа рекомендация:**
```
Recommend one {media_type} for English level {level}.
Topic preference: {interests or "general"}.
Avoid: {previously_recommended}
Format:
Title: [name]
Type: [podcast/video/article]
Link: [real URL if known, otherwise "search for: query"]
Why: [1 sentence why it's good for this level]
Duration: [estimated time]
```

---

## Команды бота

| Команда | Действие |
|---|---|
| `/start` | Регистрация + онбординг (выбор языка, уровня, часовой пояс) |
| `/today` | Показать текущее/следующее задание |
| `/word` | Получить слово вне расписания |
| `/write` | Начать free writing вне расписания |
| `/quiz` | Получить quiz по словам для повторения |
| `/skip` | Пропустить сегодня (макс 2/нед без потери streak) |
| `/stats` | Streak + слова + прогресс + график |
| `/words` | Список выученных слов с уровнем запоминания |
| `/settings` | Язык, часовой пояс, уровень, время рассылок |
| `/help` | Описание системы |

---

## Онбординг (/start)

```
1. Приветствие
2. "Какой язык учим?" → кнопки [English] [Deutsch]
3. "Какой у тебя уровень?" → кнопки [A2] [B1] [B2] [C1]
4. "Твой часовой пояс?" → кнопки популярных или ввод UTC+N
5. "Отлично! Вот как это работает:" → описание цикла
6. "Начнём? Вот твоё первое слово:" → сразу слово дня
```

---

## Skip логика

- Максимум **2 skip-а в неделю** без потери streak
- 3-й skip = streak обнуляется
- Счётчик `skip_count` сбрасывается в воскресенье (в недельном отчёте)
- Это не "пропустить навсегда", а "отложить на завтра"

---

## Deploy

**Сервер:** 95.164.18.138 (VPS)
**Сервис:** systemd (`forgepath.service`)
**CI/CD:** GitHub Actions → build → scp → restart

```
push to main → build linux/amd64 → scp to /opt/forgepath/ → systemctl restart forgepath
```

---

## Roadmap

### Phase 1 — MVP (сейчас)
- [x] Название и концепция
- [x] Go проект + структура
- [x] PostgreSQL на сервере + таблицы
- [x] GitHub repo + CI/CD
- [x] `/start` + приветствие
- [x] Онбординг (выбор языка, уровня, timezone)
- [x] SRS алгоритм + quiz
- [x] FSM для состояний диалога
- [x] Free writing + OpenAI фидбек
- [x] Медиа рекомендации
- [x] Daily review (21:30)
- [x] /stats, /words, /today
- [x] Skip логика (2/нед)
- [x] /settings — время рассылок, уровень, язык
- [x] /word, /write, /quiz — ручной запуск
- [ ] Недельный отчёт

### Phase 2 — Polish
- [ ] Темы для writing по уровням
- [ ] Улучшенные промпты (на основе реального использования)
- [ ] Обработка edge cases (нет интернета, AI timeout)

### Phase 3 — Scale
- [ ] Mini App dashboard (Next.js)
- [ ] Voice messages → speech-to-text → фидбек по произношению
- [ ] Групповые челленджи
- [ ] Адаптивный уровень (автоповышение по прогрессу)
