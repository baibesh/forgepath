package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/baibesh/forgepath/db"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type WordEntry struct {
	Word         string `json:"word"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
	Collocations string `json:"collocations"`
	Construction string `json:"construction"`
	Level        string `json:"level"`
}

var categories = []struct {
	prompt string
	count  int
}{
	{`Generate 30 essential PHRASAL VERBS for A2 English learners (Russian-speaking).
These should be the most frequently used in daily conversation.
Focus on phrasal verbs that help express actions, decisions, and daily activities.
Examples of what I mean: figure out, give up, look forward to, turn out, come up with.
DO NOT include these examples, generate NEW ones.`, 30},

	{`Generate 30 essential CONNECTORS and LINKING WORDS for A2 English learners (Russian-speaking).
These help build longer, connected speech — the main goal.
Include: time connectors (then, after that), contrast (but, however), addition (also, besides),
cause/effect (because, so), opinion (I think, in my opinion).
Examples: meanwhile, although, actually, however, instead.
DO NOT include these examples, generate NEW ones.`, 30},

	{`Generate 30 essential VERB CONSTRUCTIONS for A2 English learners (Russian-speaking).
These are verb patterns that are critical for natural speech.
Examples: manage to, be supposed to, used to, be about to, afford to.
Include: want to, need to, try to, decide to, plan to, hope to, etc.
DO NOT include these examples, generate NEW ones.`, 30},

	{`Generate 25 practical ADJECTIVES and ADVERBS for A2 English learners (Russian-speaking).
These should help describe things, people, and situations in everyday conversation.
Examples: ordinary, essential, obvious, entire, convenient.
DO NOT include these examples, generate NEW ones.`, 25},

	{`Generate 25 essential EVERYDAY VERBS for A2 English learners (Russian-speaking).
NOT phrasal verbs — single verbs that are used constantly in conversation.
Examples: improve, appreciate, avoid, recommend, suggest, consider.
DO NOT include these examples, generate NEW ones.`, 25},

	{`Generate 20 useful PREPOSITIONAL PHRASES and FIXED EXPRESSIONS for A2 English learners (Russian-speaking).
These are chunks that native speakers use constantly.
Examples: depend on, belong to, consist of, respond to.
Include: in fact, on purpose, by the way, in case, as well, etc.
DO NOT include these examples, generate NEW ones.`, 20},

	{`Generate 20 conversational FILLER PHRASES and REACTION EXPRESSIONS for A2 English learners (Russian-speaking).
These make speech natural and give time to think.
Include: you know, I mean, by the way, to be honest, the thing is, no wonder, etc.
These are essential for sounding natural in conversation.`, 20},

	{`Generate 20 essential QUESTION PATTERNS and PHRASES for A2 English learners (Russian-speaking).
Being able to ask questions is critical for conversation.
Include patterns like: How come...?, What about...?, Do you mind if...?, Could you...?, etc.`, 20},
}

func main() {
	godotenv.Load()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	database := db.Connect(databaseURL)
	defer database.Close()

	database.Migrate()

	client := openai.NewClient(apiKey)
	totalInserted := 0

	for i, cat := range categories {
		log.Printf("\n[%d/%d] Generating batch...", i+1, len(categories))

		words, err := generateWords(client, cat.prompt, cat.count)
		if err != nil {
			log.Printf("  Error: %v", err)
			continue
		}

		for _, w := range words {
			_, err := database.Pool.Exec(context.Background(),
				`INSERT INTO words (word, definition, example, level, collocations, construction)
				 VALUES ($1, $2, $3, $4, $5, $6)
				 ON CONFLICT (word) DO NOTHING`,
				strings.ToLower(strings.TrimSpace(w.Word)),
				w.Definition, w.Example, w.Level,
				w.Collocations, w.Construction)
			if err != nil {
				log.Printf("  Insert '%s' error: %v", w.Word, err)
				continue
			}
			totalInserted++
			log.Printf("  ✅ %s — %s", w.Word, w.Definition)
		}

		// Rate limit
		time.Sleep(1 * time.Second)
	}

	log.Printf("\nDone! Inserted %d new words", totalInserted)

	var count int
	database.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM words").Scan(&count)
	log.Printf("Total words in database: %d", count)
}

func generateWords(client *openai.Client, prompt string, count int) ([]WordEntry, error) {
	fullPrompt := fmt.Sprintf(`%s

Return EXACTLY %d entries as a JSON array. Each entry must have:
- "word": the word/phrase in English (lowercase)
- "definition": translation to Russian (brief, 2-5 words)
- "example": one natural example sentence using this word
- "collocations": 2-3 common collocations separated by comma
- "construction": grammatical pattern (e.g., "verb + to V1", "adj + noun")
- "level": "A2"

Return ONLY the JSON array, no other text. Make sure it's valid JSON.`, prompt, count)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: fullPrompt},
		},
		MaxTokens:   4000,
		Temperature: 0.8,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	text := resp.Choices[0].Message.Content
	// Clean up markdown code fences if present
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var words []WordEntry
	if err := json.Unmarshal([]byte(text), &words); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w\nRaw: %s", err, text[:min(200, len(text))])
	}

	return words, nil
}
