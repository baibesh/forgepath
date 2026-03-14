# ForgePath Bot

Study through daily habits. A system, not a course.

---

## Stack

| Layer | Tech |
|---|---|
| Bot | Go + `telebot.v3` |
| Cron | `robfig/cron` |
| DB | PostgreSQL + `pgx` |
| Config | `.env` + `godotenv` |
| Mini App (позже) | Next.js |

---

## Структура проекта

```
forgepath/
├── main.go
├── .env
├── .env.example
├── go.mod
│
├── bot/
│   ├── handlers.go       # /start /today /skip /stats /settings
│   └── keyboards.go      # inline кнопки
│
├── cron/
│   ├── scheduler.go      # регистрация всех задач
│   └── jobs.go           # логика каждой рассылки
│
├── db/
│   ├── postgres.go       # подключение
│   ├── users.go          # CRUD пользователей
│   └── streak.go         # логика streak
│
└── config/
    └── config.go         # чтение .env
```

---

## Быстрый старт локально

### 1. Установи Go

```bash
# macOS
brew install go

# проверь версию (нужно 1.21+)
go version
```

### 2. Клонируй и инициализируй проект

```bash
mkdir forgepath && cd forgepath
go mod init github.com/yourname/forgepath
```

### 3. Установи зависимости

```bash
go get gopkg.in/telebot.v3
go get github.com/robfig/cron/v3
go get github.com/jackc/pgx/v5
go get github.com/joho/godotenv
```

### 4. Настрой .env

```env
BOT_TOKEN=your_token_from_botfather
DATABASE_URL=postgres://user:password@localhost:5432/forgepath
```

### 5. Запусти PostgreSQL локально

```bash
# через Docker — проще всего
docker run --name forgepath-db \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=forgepath \
  -p 5432:5432 \
  -d postgres:16
```

### 6. Запусти бота

```bash
go run main.go
```

---

## main.go — точка входа

```go
package main

import (
    "log"
    "os"
    "time"

    "github.com/joho/godotenv"
    tele "gopkg.in/telebot.v3"

    "github.com/yourname/forgepath/bot"
    "github.com/yourname/forgepath/cron"
    "github.com/yourname/forgepath/db"
)

func main() {
    godotenv.Load()

    database := db.Connect(os.Getenv("DATABASE_URL"))

    b, err := tele.NewBot(tele.Settings{
        Token:  os.Getenv("BOT_TOKEN"),
        Poller: &tele.LongPoller{Timeout: 10 * time.Second},
    })
    if err != nil {
        log.Fatal(err)
    }

    bot.RegisterHandlers(b, database)
    cron.StartScheduler(b, database)

    log.Println("ForgePath bot started")
    b.Start()
}
```

---

## Cron расписание

```go
// cron/scheduler.go

package cron

import (
    "github.com/robfig/cron/v3"
    tele "gopkg.in/telebot.v3"
    "github.com/yourname/forgepath/db"
)

func StartScheduler(b *tele.Bot, database *db.DB) {
    c := cron.New(cron.WithLocation(time.UTC))

    c.AddFunc("30 7 * * *", func() { SendWordOfDay(b, database) })   // 07:30 — слово дня
    c.AddFunc("0 12 * * *", func() { SendFreeWriting(b, database) }) // 12:00 — free writing
    c.AddFunc("0 18 * * *", func() { SendMediaRec(b, database) })    // 18:00 — медиа (через день)
    c.AddFunc("30 21 * * *", func() { SendDailyReview(b, database) }) // 21:30 — ревизия
    c.AddFunc("0 9 * * 0", func() { SendWeeklyReport(b, database) }) // 09:00 вс — недельный отчёт

    c.Start()
}
```

---

## Схема БД

```sql
CREATE TABLE users (
    id          BIGINT PRIMARY KEY,  -- telegram user id
    username    TEXT,
    tz_offset   INT DEFAULT 0,       -- часовой пояс в часах
    level       TEXT DEFAULT 'A2',   -- текущий уровень
    active      BOOL DEFAULT TRUE,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE streaks (
    user_id     BIGINT REFERENCES users(id),
    date        DATE,
    completed   BOOL DEFAULT FALSE,
    PRIMARY KEY (user_id, date)
);

CREATE TABLE words (
    id          SERIAL PRIMARY KEY,
    word        TEXT NOT NULL,
    definition  TEXT,
    example     TEXT,
    level       TEXT  -- A2 / B1 / B2 / C1
);

CREATE TABLE user_words (
    user_id     BIGINT REFERENCES users(id),
    word_id     INT REFERENCES words(id),
    seen_at     TIMESTAMP DEFAULT NOW(),
    next_review TIMESTAMP,  -- интервальное повторение
    score       INT DEFAULT 0,
    PRIMARY KEY (user_id, word_id)
);
```

---

## AI интеграция

Используем **GPT-4o mini** для критичного фидбека и **Gemini Flash** для простых задач.

### Зависимость

```bash
go get github.com/sashabaranov/go-openai
go get github.com/google/generative-ai-go/genai
```

### .env

```env
OPENAI_API_KEY=your_openai_key
GEMINI_API_KEY=your_gemini_key
```

### Где что используется

| Задача | Модель | Почему |
|---|---|---|
| Фидбек по free writing | GPT-4o mini | Лучше понимает нюансы English writing |
| Слово дня + пример | Gemini Flash | Простая генерация, free tier |
| Медиа-рекомендация | Gemini Flash | Не критично, экономим |
| Анализ прогресса | Gemini Flash | Структурированный вывод |

### Пример — фидбек по writing

```go
// bot/ai.go

package bot

import (
    "context"
    "fmt"
    "os"

    openai "github.com/sashabaranov/go-openai"
)

func GetWritingFeedback(text string) (string, error) {
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

    prompt := fmt.Sprintf(`You are a concise English writing coach.
Analyze this student's free writing and give brief feedback:
- 2-3 grammar/style mistakes with corrections
- One thing they did well
- One tip for next time

Keep it under 100 words. Be encouraging but direct.

Text: %s`, text)

    resp, err := client.CreateChatCompletion(context.Background(),
        openai.ChatCompletionRequest{
            Model: openai.GPT4oMini,
            Messages: []openai.ChatCompletionMessage{
                {Role: openai.ChatMessageRoleUser, Content: prompt},
            },
            MaxTokens: 200,
        },
    )
    if err != nil {
        return "", err
    }

    return resp.Choices[0].Message.Content, nil
}
```

### Пример — слово дня через Gemini

```go
// bot/word.go

package bot

import (
    "context"
    "fmt"
    "os"

    "google.golang.org/api/option"
    "github.com/google/generative-ai-go/genai"
)

func GenerateWordOfDay(level string) (string, error) {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
    if err != nil {
        return "", err
    }
    defer client.Close()

    model := client.GenerativeModel("gemini-1.5-flash")
    prompt := fmt.Sprintf(`Give one useful English word for level %s.
Format exactly like this:
Word: [word]
Definition: [short definition]
Example: [natural example sentence]
Tip: [one memory tip]`, level)

    resp, err := model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}
```

---

## Команды бота

| Команда | Действие |
|---|---|
| `/start` | Регистрация + онбординг |
| `/today` | Текущее задание |
| `/skip` | Пропустить сегодня без потери streak |
| `/stats` | Streak + слова + прогресс |
| `/settings` | Изменить время рассылок |

---

## AI интеграция

### Модели

| Задача | Модель | Почему |
|---|---|---|
| Free writing фидбек | `gpt-4o-mini` | Лучший фидбек по английскому письму |
| Слово дня + пример | `gemini-1.5-flash` | Бесплатный tier, достаточно качества |
| Медиа-рекомендация | `gemini-1.5-flash` | Не критично к качеству |
| Анализ прогресса | `gpt-4o-mini` | Нужна точность |

### .env

```env
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=AI...
```

### Зависимости

```bash
go get github.com/sashabaranov/go-openai
go get github.com/google/generative-ai-go/genai
```

### Free writing фидбек — OpenAI

```go
// ai/openai.go

package ai

import (
    "context"
    "fmt"
    "os"

    openai "github.com/sashabaranov/go-openai"
)

func CheckFreeWriting(text string, level string) (string, error) {
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

    prompt := fmt.Sprintf(`
You are an English tutor. The student's level is %s.
Review this free writing and give brief feedback:
- 2-3 grammar mistakes (if any), with corrections
- 1 style suggestion
- 1 positive note

Keep feedback under 100 words. Be direct, not overly encouraging.

Student's text:
%s
`, level, text)

    resp, err := client.CreateChatCompletion(context.Background(),
        openai.ChatCompletionRequest{
            Model: openai.GPT4oMini,
            Messages: []openai.ChatCompletionMessage{
                {Role: openai.ChatMessageRoleUser, Content: prompt},
            },
            MaxTokens: 200,
        },
    )
    if err != nil {
        return "", err
    }

    return resp.Choices[0].Message.Content, nil
}
```

### Слово дня — Gemini

```go
// ai/gemini.go

package ai

import (
    "context"
    "fmt"
    "os"

    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

func GenerateWordOfDay(level string) (string, error) {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
    if err != nil {
        return "", err
    }
    defer client.Close()

    model := client.GenerativeModel("gemini-1.5-flash")

    prompt := fmt.Sprintf(`
Give one useful English word for level %s.
Format exactly like this:
Word: [word]
Definition: [one short sentence]
Example: [natural example sentence]
Tip: [one memory tip or collocation]

No extra text.
`, level)

    resp, err := model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}
```

### Использование в cron jobs

```go
// cron/jobs.go

func SendWordOfDay(b *tele.Bot, database *db.DB) {
    users := database.GetActiveUsers()
    for _, user := range users {
        word, err := ai.GenerateWordOfDay(user.Level)
        if err != nil {
            log.Println("gemini error:", err)
            continue
        }
        b.Send(&tele.User{ID: user.ID}, word, wordKeyboard())
    }
}

func SendFreeWriting(b *tele.Bot, database *db.DB) {
    topics := []string{
        "Describe your morning routine",
        "What would you change about your city?",
        "A skill you want to learn and why",
    }
    topic := topics[time.Now().Day() % len(topics)]

    users := database.GetActiveUsers()
    for _, user := range users {
        msg := fmt.Sprintf("✍️ *Free writing — 5 min*\n\nTopic: _%s_\n\nWrite without stopping. Send when done.", topic)
        b.Send(&tele.User{ID: user.ID}, msg, tele.ModeMarkdown)
    }
}
```

### Обработка ответа на free writing

```go
// bot/handlers.go

b.Handle(tele.OnText, func(c tele.Context) error {
    user := database.GetUser(c.Sender().ID)
    if !user.WaitingForWriting {
        return nil
    }

    // сохраняем текст
    database.SaveWriting(user.ID, c.Text())

    // получаем фидбек
    feedback, err := ai.CheckFreeWriting(c.Text(), user.Level)
    if err != nil {
        return c.Send("✅ Saved! Feedback coming soon...")
    }

    database.UpdateStreak(user.ID)

    return c.Send(fmt.Sprintf("✅ *Done!*\n\n%s", feedback), tele.ModeMarkdown)
})
```

### Примерная стоимость

| Пользователей/день | OpenAI (writing) | Gemini (word+media) | Итого/мес |
|---|---|---|---|
| 50 | ~$1.5 | $0 (free tier) | ~$1.5 |
| 200 | ~$6 | ~$0.5 | ~$6.5 |
| 1000 | ~$30 | ~$2 | ~$32 |

---

## Деплой (позже)

Минимальный вариант — VPS на **Hetzner** (€4/мес):

```bash
# собери бинарник
GOOS=linux GOARCH=amd64 go build -o forgepath .

# отправь на сервер
scp forgepath user@server:/opt/forgepath/

# запусти через systemd или screen
./forgepath
```

---

## Roadmap

- [x] Название и концепция
- [ ] Базовый бот — `/start` + приветствие
- [ ] Cron — слово дня
- [ ] Free writing + сохранение ответа
- [ ] Streak логика
- [ ] /stats команда
- [ ] Настройка часового пояса
- [ ] Mini App — dashboard (v2)
