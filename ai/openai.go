package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	if apiKey == "" {
		return nil
	}
	return &OpenAIClient{client: openai.NewClient(apiKey)}
}

func langFullName(code string) string {
	switch code {
	case "ru":
		return "English"
	case "kk":
		return "English"
	default:
		return "English"
	}
}

func uiLangName(code string) string {
	switch code {
	case "ru":
		return "Russian"
	case "kk":
		return "Kazakh"
	default:
		return "Russian"
	}
}

func (o *OpenAIClient) CheckWriting(text, grammarFocus, level, language string) (string, error) {
	if o == nil {
		return "AI feedback is not available right now. Keep writing!", nil
	}

	targetLang := langFullName(language)
	uiLang := uiLangName(language)

	prompt := fmt.Sprintf(`You are an %s language tutor. The student speaks %s.
Their level: %s. Grammar focus this week: %s

Review the following text and respond in %s in this EXACT format:
✅ Good: (1-2 things done well)
❌ Errors: (each error → correction, explain briefly)
💡 Better version: (rewrite 1-2 sentences more naturally)
🎯 Grammar tip: (one tip about %s)
🧠 Hack: (one memory trick for the main error)

Under 150 words. Direct but encouraging.

Student's text:
%s`, targetLang, uiLang, level, grammarFocus, uiLang, grammarFocus, text)

	return o.complete(prompt)
}

func (o *OpenAIClient) CheckSentences(sentences, mediaTitle, level, language string) (string, error) {
	if o == nil {
		return "AI feedback is not available right now. Good job writing!", nil
	}

	targetLang := langFullName(language)
	uiLang := uiLangName(language)

	prompt := fmt.Sprintf(`You are an %s language tutor. The student speaks %s. Level: %s.
They watched: "%s"
They wrote these sentences as a post-watching task.

Check grammar, suggest improvements. Be brief and encouraging.
Respond in %s. Format:
✅ Good: (what's correct)
❌ Fix: (corrections with brief explanations)
💡 Better: (improved versions)

Under 100 words.

Student wrote:
%s`, targetLang, uiLang, level, mediaTitle, uiLang, sentences)

	return o.complete(prompt)
}

func (o *OpenAIClient) GenerateQuizOptions(word, definition, language string, count int) ([]string, error) {
	if o == nil {
		return defaultQuizOptions(definition, language), nil
	}

	defLang := uiLangName(language)

	prompt := fmt.Sprintf(`Generate %d wrong answer options for a vocabulary quiz.
The word is: "%s" and the correct definition is: "%s"
The definitions are written in %s. Generate %d WRONG definitions in %s that are plausible but incorrect.
Keep them short (1-3 words), similar style to the correct definition.
Return ONLY the wrong options, one per line, no numbering, no quotes.`, count, word, definition, defLang, count, defLang)

	text, err := o.complete(prompt)
	if err != nil {
		return defaultQuizOptions(definition, language), nil
	}

	lines := strings.Split(strings.TrimSpace(text), "\n")
	var options []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			options = append(options, line)
		}
	}
	if len(options) < count {
		return defaultQuizOptions(definition, language), nil
	}
	return options[:count], nil
}

func (o *OpenAIClient) GenerateWeeklyReport(wordsLearned, writingsDone, streakDays int, grammarFocus, level, language string) (string, error) {
	if o == nil {
		return fmt.Sprintf("Great week! %d words learned, %d writings done, %d day streak!", wordsLearned, writingsDone, streakDays), nil
	}

	uiLang := uiLangName(language)

	prompt := fmt.Sprintf(`Write a brief, encouraging weekly report for an English language learner.
Their native language is %s, level: %s.
Stats: %d words learned, %d writings completed, %d day streak, grammar focus: %s.
2-3 sentences. Encouraging but specific. Include one tip for next week.
Write your response in %s.`, uiLang, level, wordsLearned, writingsDone, streakDays, grammarFocus, uiLang)

	return o.complete(prompt)
}

func (o *OpenAIClient) SuggestMediaKeywords(grammarFocus, todayWord, level string) ([]string, error) {
	if o == nil {
		return []string{grammarFocus}, nil
	}

	prompt := fmt.Sprintf(`You help pick YouTube videos for an %s English language learner.
Their grammar focus: %s
Today's word: %s

Return 3-5 search keywords (comma-separated, lowercase) to find a relevant YouTube video.
Focus on practical topics connected to the grammar and word.
Example output: past simple,daily routine,telling stories
Return ONLY the keywords, nothing else.`, level, grammarFocus, todayWord)

	text, err := o.complete(prompt)
	if err != nil {
		return []string{grammarFocus}, nil
	}

	var keywords []string
	for _, kw := range strings.Split(text, ",") {
		kw = strings.TrimSpace(strings.ToLower(kw))
		if kw != "" {
			keywords = append(keywords, kw)
		}
	}
	if len(keywords) == 0 {
		return []string{grammarFocus}, nil
	}
	return keywords, nil
}

func (o *OpenAIClient) TextToSpeech(text string) (string, error) {
	if o == nil {
		return "", fmt.Errorf("OpenAI client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := o.client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Voice: openai.VoiceAlloy,
		Input: text,
	})
	if err != nil {
		return "", fmt.Errorf("TTS error: %w", err)
	}
	defer resp.Close()

	tmpFile, err := os.CreateTemp("", "forgepath-tts-*.mp3")
	if err != nil {
		return "", fmt.Errorf("temp file error: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("write error: %w", err)
	}
	tmpFile.Close()

	return tmpFile.Name(), nil
}

func (o *OpenAIClient) SpeechToText(filePath string) (string, error) {
	if o == nil {
		return "", fmt.Errorf("OpenAI client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := o.client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filePath,
	})
	if err != nil {
		return "", fmt.Errorf("STT error: %w", err)
	}

	return strings.TrimSpace(resp.Text), nil
}

// WordInfo holds AI-generated info for a custom word.
type WordInfo struct {
	Definition   string
	Example      string
	Collocations string
	Construction string
	Synonyms     string
	Antonyms     string
	Examples     string
}

func (o *OpenAIClient) LookupWord(word, userLanguage string) (*WordInfo, error) {
	if o == nil {
		return nil, fmt.Errorf("OpenAI client not configured")
	}

	uiLang := uiLangName(userLanguage)

	prompt := fmt.Sprintf(`You are an English language dictionary. The user speaks %s.
For the English word/phrase: "%s"

Return EXACTLY this JSON format, no extra text:
{
  "definition": "short translation/definition in %s (1-3 words)",
  "example": "one natural example sentence in English using this word",
  "examples": "two more example sentences in different contexts, separated by |",
  "collocations": "3-4 common collocations, comma separated",
  "construction": "grammar pattern, e.g. 'verb + noun' or 'adjective + about'",
  "synonyms": "2-3 synonyms, comma separated",
  "antonyms": "1-2 antonyms, comma separated (empty string if none)"
}

If the word doesn't exist or is not English, return:
{"error": "not found"}`, uiLang, word, uiLang)

	text, err := o.complete(prompt)
	if err != nil {
		return nil, err
	}

	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var parsed struct {
		Definition   string `json:"definition"`
		Example      string `json:"example"`
		Examples     string `json:"examples"`
		Collocations string `json:"collocations"`
		Construction string `json:"construction"`
		Synonyms     string `json:"synonyms"`
		Antonyms     string `json:"antonyms"`
		Error        string `json:"error"`
	}

	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		log.Printf("LookupWord JSON parse error: %v, raw: %s", err, text)
		return nil, fmt.Errorf("could not parse word info")
	}

	if parsed.Error != "" {
		return nil, fmt.Errorf("word not found")
	}

	if parsed.Definition == "" {
		return nil, fmt.Errorf("could not parse word info")
	}

	return &WordInfo{
		Definition:   parsed.Definition,
		Example:      parsed.Example,
		Collocations: parsed.Collocations,
		Construction: parsed.Construction,
		Synonyms:     parsed.Synonyms,
		Antonyms:     parsed.Antonyms,
		Examples:     parsed.Examples,
	}, nil
}

// EnrichWord generates extra context for an existing word missing synonyms/examples.
func (o *OpenAIClient) EnrichWord(word, definition, userLanguage string) (synonyms, antonyms, examples string, err error) {
	if o == nil {
		return "", "", "", fmt.Errorf("OpenAI client not configured")
	}

	uiLang := uiLangName(userLanguage)

	prompt := fmt.Sprintf(`For the English word "%s" (meaning: %s), the user speaks %s.
Return EXACTLY this JSON, no extra text:
{
  "synonyms": "2-3 synonyms, comma separated",
  "antonyms": "1-2 antonyms, comma separated (empty string if none)",
  "examples": "two example sentences in different contexts, separated by |"
}`, word, definition, uiLang)

	text, err := o.complete(prompt)
	if err != nil {
		return "", "", "", err
	}

	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var parsed struct {
		Synonyms string `json:"synonyms"`
		Antonyms string `json:"antonyms"`
		Examples string `json:"examples"`
	}

	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return "", "", "", fmt.Errorf("could not parse enrichment")
	}

	return parsed.Synonyms, parsed.Antonyms, parsed.Examples, nil
}

// GenerateClozeOptions generates a cloze sentence and wrong word options for a quiz.
func (o *OpenAIClient) GenerateClozeOptions(word, definition, language string) (sentence string, wrongWords []string, err error) {
	if o == nil {
		return "", nil, fmt.Errorf("OpenAI client not configured")
	}

	prompt := fmt.Sprintf(`Create a fill-in-the-blank quiz for the English word "%s" (meaning: %s).
Return EXACTLY this JSON, no extra text:
{
  "sentence": "A natural sentence with ___ where the word should go",
  "wrong_words": ["wrong1", "wrong2", "wrong3"]
}
The wrong words should be plausible but incorrect alternatives at a similar level.
Keep the sentence simple (A2-B1 level).`, word, definition)

	text, err := o.complete(prompt)
	if err != nil {
		return "", nil, err
	}

	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var parsed struct {
		Sentence   string   `json:"sentence"`
		WrongWords []string `json:"wrong_words"`
	}

	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return "", nil, fmt.Errorf("could not parse cloze")
	}

	if parsed.Sentence == "" || len(parsed.WrongWords) < 3 {
		return "", nil, fmt.Errorf("incomplete cloze data")
	}

	return parsed.Sentence, parsed.WrongWords[:3], nil
}

// GenerateCollocationQuiz generates a collocation quiz for a word.
func (o *OpenAIClient) GenerateCollocationQuiz(word, collocations, language string) (question string, options []string, correctIdx int, err error) {
	if o == nil {
		return "", nil, 0, fmt.Errorf("OpenAI client not configured")
	}

	uiLang := uiLangName(language)

	prompt := fmt.Sprintf(`Create a collocation quiz for the English word "%s".
Known collocations: %s
The user speaks %s.

Return EXACTLY this JSON, no extra text:
{
  "question": "Which phrase is correct?",
  "correct": "one correct collocation with the word",
  "wrong": ["wrong collocation 1", "wrong collocation 2", "wrong collocation 3"]
}
Wrong options should use the word but in unnatural combinations.`, word, collocations, uiLang)

	text, err := o.complete(prompt)
	if err != nil {
		return "", nil, 0, err
	}

	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var parsed struct {
		Question string   `json:"question"`
		Correct  string   `json:"correct"`
		Wrong    []string `json:"wrong"`
	}

	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return "", nil, 0, fmt.Errorf("could not parse collocation quiz")
	}

	if parsed.Correct == "" || len(parsed.Wrong) < 3 {
		return "", nil, 0, fmt.Errorf("incomplete collocation data")
	}

	options = []string{parsed.Correct}
	options = append(options, parsed.Wrong[:3]...)

	return parsed.Question, options, 0, nil
}

func defaultQuizOptions(definition, language string) []string {
	defaults := []string{
		"увеличить", "радоваться", "бежать быстро",
		"запомнить", "согласиться", "потерять надежду",
		"пробовать", "исчезнуть", "собирать вместе",
	}
	var result []string
	for _, d := range defaults {
		if d != definition && len(result) < 3 {
			result = append(result, d)
		}
	}
	return result
}

func (o *OpenAIClient) complete(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			MaxTokens:   500,
			Temperature: 0.7,
		})
		if err != nil {
			lastErr = err
			log.Printf("OpenAI attempt %d failed: %v", attempt+1, err)
			time.Sleep(1 * time.Second)
			continue
		}
		if len(resp.Choices) > 0 {
			return resp.Choices[0].Message.Content, nil
		}
		return "", fmt.Errorf("empty response from OpenAI")
	}
	return "", fmt.Errorf("OpenAI failed after retries: %w", lastErr)
}
