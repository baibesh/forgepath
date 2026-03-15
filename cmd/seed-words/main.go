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
	prompt   string
	count    int
	language string
}{
	// === ENGLISH ===

	{`Generate 30 essential PHRASAL VERBS for A2 English learners (Russian-speaking).
These should be the most frequently used in daily conversation.
Focus on phrasal verbs that help express actions, decisions, and daily activities.
Examples of what I mean: figure out, give up, look forward to, turn out, come up with.
DO NOT include these examples, generate NEW ones.`, 30, "en"},

	{`Generate 30 essential CONNECTORS and LINKING WORDS for A2 English learners (Russian-speaking).
These help build longer, connected speech — the main goal.
Include: time connectors (then, after that), contrast (but, however), addition (also, besides),
cause/effect (because, so), opinion (I think, in my opinion).
Examples: meanwhile, although, actually, however, instead.
DO NOT include these examples, generate NEW ones.`, 30, "en"},

	{`Generate 30 essential VERB CONSTRUCTIONS for A2 English learners (Russian-speaking).
These are verb patterns that are critical for natural speech.
Examples: manage to, be supposed to, used to, be about to, afford to.
Include: want to, need to, try to, decide to, plan to, hope to, etc.
DO NOT include these examples, generate NEW ones.`, 30, "en"},

	{`Generate 25 practical ADJECTIVES and ADVERBS for A2 English learners (Russian-speaking).
These should help describe things, people, and situations in everyday conversation.
Examples: ordinary, essential, obvious, entire, convenient.
DO NOT include these examples, generate NEW ones.`, 25, "en"},

	{`Generate 25 essential EVERYDAY VERBS for A2 English learners (Russian-speaking).
NOT phrasal verbs — single verbs that are used constantly in conversation.
Examples: improve, appreciate, avoid, recommend, suggest, consider.
DO NOT include these examples, generate NEW ones.`, 25, "en"},

	{`Generate 20 useful PREPOSITIONAL PHRASES and FIXED EXPRESSIONS for A2 English learners (Russian-speaking).
These are chunks that native speakers use constantly.
Examples: depend on, belong to, consist of, respond to.
Include: in fact, on purpose, by the way, in case, as well, etc.
DO NOT include these examples, generate NEW ones.`, 20, "en"},

	{`Generate 20 conversational FILLER PHRASES and REACTION EXPRESSIONS for A2 English learners (Russian-speaking).
These make speech natural and give time to think.
Include: you know, I mean, by the way, to be honest, the thing is, no wonder, etc.
These are essential for sounding natural in conversation.`, 20, "en"},

	{`Generate 20 essential QUESTION PATTERNS and PHRASES for A2 English learners (Russian-speaking).
Being able to ask questions is critical for conversation.
Include patterns like: How come...?, What about...?, Do you mind if...?, Could you...?, etc.`, 20, "en"},

	// === DEUTSCH ===

	{`Generate 30 essential VERBEN MIT PRÄPOSITIONEN for A2 German learners (Russian-speaking).
These are verbs with fixed prepositions critical for natural German speech.
Examples: sich freuen auf, sich interessieren für, abhängen von, sich kümmern um.
DO NOT include these examples, generate NEW ones.
Include the grammatical case (Akk/Dat) in the construction field.`, 30, "de"},

	{`Generate 30 essential KONNEKTOREN und BINDEWÖRTER for A2 German learners (Russian-speaking).
These help build connected speech in German.
Include: temporal (dann, danach), kausal (weil, deshalb), konzessiv (obwohl, trotzdem),
adversativ (aber, jedoch), additional (außerdem, auch).
Examples: trotzdem, eigentlich, wahrscheinlich, deshalb, obwohl.
DO NOT include these examples, generate NEW ones.`, 30, "de"},

	{`Generate 25 essential TRENNBARE und UNTRENNBARE VERBEN for A2 German learners (Russian-speaking).
These are separable and inseparable verbs critical for daily German.
Examples: anfangen, aufhören, vorschlagen, sich vorstellen, bestehen.
DO NOT include these examples, generate NEW ones.
Include whether the verb is trennbar or untrennbar.`, 25, "de"},

	{`Generate 25 practical ADJEKTIVE und ADVERBIEN for A2 German learners (Russian-speaking).
These should help describe things, people, and situations in everyday German conversation.
Examples: gemütlich, zuverlässig, geduldig, offensichtlich, ausgezeichnet.
DO NOT include these examples, generate NEW ones.`, 25, "de"},

	{`Generate 20 essential REFLEXIVE VERBEN for A2 German learners (Russian-speaking).
Reflexive verbs are very common in German and different from Russian/English.
Examples: sich entscheiden, sich beschweren, sich gewöhnen, sich erinnern.
DO NOT include these examples, generate NEW ones.
Include the preposition and case in construction.`, 20, "de"},

	{`Generate 20 useful DEUTSCHE REDEWENDUNGEN and FESTE AUSDRÜCKE for A2 German learners (Russian-speaking).
These are fixed expressions native German speakers use constantly.
Include: es lohnt sich, auf jeden Fall, im Grunde genommen, etc.
These are essential for sounding natural in German conversation.`, 20, "de"},

	// ==================== ENGLISH A1 ====================

	{`Generate 30 basic EVERYDAY WORDS for A1 English learners (Russian-speaking).
These are the first 30 words a beginner absolutely needs: family, food, home, body, time, etc.
Simple nouns, basic verbs (go, eat, sleep, like, want), basic adjectives (big, small, good, bad).
Level: absolute beginner. Definition in Russian, simple example.`, 30, "en"},

	{`Generate 20 essential BASIC PHRASES for A1 English learners (Russian-speaking).
These are survival phrases: hello, goodbye, thank you, excuse me, how much, where is, I want, I like,
can I, please, sorry, yes, no, my name is, I don't understand, etc.
Simple, practical, daily use.`, 20, "en"},

	// ==================== ENGLISH B1 ====================

	{`Generate 30 INTERMEDIATE PHRASAL VERBS for B1 English learners (Russian-speaking).
These are more advanced phrasal verbs beyond A2 level.
Examples: come across, get over, put up with, stand out, bring about.
DO NOT include these examples, generate NEW ones.`, 30, "en"},

	{`Generate 30 ACADEMIC and FORMAL VOCABULARY for B1 English learners (Russian-speaking).
Words used in news, articles, work emails. Not slang, not too simple.
Examples: contribute, influence, significant, frequently, approximately, establish.
DO NOT include these examples, generate NEW ones.`, 30, "en"},

	{`Generate 25 OPINION and ARGUMENTATION PHRASES for B1 English learners (Russian-speaking).
These help express complex opinions, agree/disagree, make arguments.
Examples: as far as I know, on the other hand, in my experience, it seems to me.
DO NOT include these examples, generate NEW ones.`, 25, "en"},

	{`Generate 25 ADVANCED CONNECTORS for B1 English learners (Russian-speaking).
Beyond A2 connectors. More sophisticated linking words.
Examples: nevertheless, furthermore, consequently, whereas, provided that.
DO NOT include these examples, generate NEW ones.`, 25, "en"},

	// ==================== ENGLISH B2 ====================

	{`Generate 30 UPPER-INTERMEDIATE VOCABULARY for B2 English learners (Russian-speaking).
Academic and professional words. Used in presentations, essays, discussions.
Examples: acknowledge, comprehensive, pursue, reluctant, inevitable.
DO NOT include these examples, generate NEW ones.`, 30, "en"},

	{`Generate 25 ADVANCED IDIOMS and EXPRESSIONS for B2 English learners (Russian-speaking).
Common idiomatic expressions that educated native speakers use.
Examples: hit the nail on the head, a blessing in disguise, cut corners, the last straw.
DO NOT include these examples, generate NEW ones.`, 25, "en"},

	{`Generate 25 BUSINESS and PROFESSIONAL ENGLISH for B2 English learners (Russian-speaking).
Vocabulary for meetings, emails, negotiations, presentations.
Examples: benchmark, leverage, stakeholder, streamline, scalable.
DO NOT include these examples, generate NEW ones.`, 25, "en"},

	// ==================== DEUTSCH A1 ====================

	{`Generate 30 grundlegende ALLTAGSWÖRTER for A1 German learners (Russian-speaking).
These are the first 30 German words a beginner needs: Familie, Essen, Haus, Körper, Zeit.
Simple nouns, basic verbs (gehen, essen, schlafen, mögen, wollen), basic adjectives (groß, klein, gut, schlecht).
Level: absolute beginner. Definition in Russian, simple example in German.`, 30, "de"},

	{`Generate 20 essential GRUNDPHRASEN for A1 German learners (Russian-speaking).
Survival phrases: Hallo, Tschüss, Danke, Entschuldigung, Wie viel kostet, Wo ist,
Ich möchte, Ich verstehe nicht, Wie heißen Sie, etc.
Simple, practical, daily use. Definition in Russian.`, 20, "de"},

	// ==================== DEUTSCH B1 ====================

	{`Generate 30 FORTGESCHRITTENE VERBEN MIT PRÄPOSITIONEN for B1 German learners (Russian-speaking).
More complex verbs with prepositions, beyond A2 level.
Examples: sich auseinandersetzen mit, beitragen zu, verzichten auf, bestehen auf.
DO NOT include these examples, generate NEW ones.
Include case (Akk/Dat) in construction field.`, 30, "de"},

	{`Generate 25 KONNEKTOREN und SATZVERBINDUNGEN for B1 German learners (Russian-speaking).
Advanced connectors for complex sentences: je...desto, sowohl...als auch, weder...noch,
indem, sodass, anstatt dass. More complex than A2 level.
DO NOT include the examples above, generate NEW ones.`, 25, "de"},

	{`Generate 25 BERUFS- und WISSENSCHAFTSSPRACHE for B1 German learners (Russian-speaking).
Vocabulary for work, news, academic texts in German.
Examples: beitragen, beeinflussen, erheblich, Untersuchung, Zusammenhang.
DO NOT include these examples, generate NEW ones.`, 25, "de"},

	{`Generate 20 DEUTSCHE REDEWENDUNGEN B1 for B1 German learners (Russian-speaking).
More advanced German idioms and fixed expressions beyond A2.
Examples: den Nagel auf den Kopf treffen, unter vier Augen, auf dem Laufenden bleiben.
DO NOT include these examples, generate NEW ones.`, 20, "de"},
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

	database.Migrate(databaseURL)

	client := openai.NewClient(apiKey)
	totalInserted := 0

	for i, cat := range categories {
		log.Printf("\n[%d/%d] Generating batch [%s]...", i+1, len(categories), cat.language)

		langName := "English"
		if cat.language == "de" {
			langName = "German"
		}

		words, err := generateWords(client, cat.prompt, cat.count, langName)
		if err != nil {
			log.Printf("  Error: %v", err)
			continue
		}

		for _, w := range words {
			_, err := database.Pool.Exec(context.Background(),
				`INSERT INTO words (word, definition, example, level, collocations, construction, language)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)
				 ON CONFLICT (word, language) DO NOTHING`,
				strings.ToLower(strings.TrimSpace(w.Word)),
				w.Definition, w.Example, w.Level,
				w.Collocations, w.Construction, cat.language)
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

func generateWords(client *openai.Client, prompt string, count int, langName string) ([]WordEntry, error) {
	wordLang := "English"
	if langName == "German" {
		wordLang = "German"
	}

	// Extract level from the prompt (A1/A2/B1/B2)
	level := "A2"
	for _, l := range []string{"A1", "B2", "B1"} {
		if strings.Contains(prompt, l) {
			level = l
			break
		}
	}

	fullPrompt := fmt.Sprintf(`%s

Return EXACTLY %d entries as a JSON array. Each entry must have:
- "word": the word/phrase in %s (lowercase)
- "definition": translation to Russian (brief, 2-5 words)
- "example": one natural example sentence using this word in %s
- "collocations": 2-3 common collocations separated by comma
- "construction": grammatical pattern (e.g., "verb + zu + Inf", "adj + Nomen")
- "level": "%s"

Return ONLY the JSON array, no other text. Make sure it's valid JSON.`, prompt, count, wordLang, wordLang, level)

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
