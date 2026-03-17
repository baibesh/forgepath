package ai

import (
	"context"
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
