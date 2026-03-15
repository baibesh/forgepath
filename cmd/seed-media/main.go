package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/baibesh/forgepath/db"
	"github.com/joho/godotenv"
)

type SearchResponse struct {
	Items         []SearchItem `json:"items"`
	NextPageToken string       `json:"nextPageToken"`
}

type SearchItem struct {
	ID      SearchID      `json:"id"`
	Snippet SearchSnippet `json:"snippet"`
}

type SearchID struct {
	VideoID string `json:"videoId"`
}

type SearchSnippet struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ChannelTitle string `json:"channelTitle"`
}

type VideoListResponse struct {
	Items []VideoItem `json:"items"`
}

type VideoItem struct {
	ID               string           `json:"id"`
	ContentDetails   ContentDetails   `json:"contentDetails"`
	Statistics       Statistics       `json:"statistics"`
}

type ContentDetails struct {
	Duration string `json:"duration"`
	Caption  string `json:"caption"`
}

type Statistics struct {
	ViewCount string `json:"viewCount"`
}

var searchQueries = []struct {
	query    string
	topic    string
	level    string
	language string
}{
	// ==================== ENGLISH A1 ====================
	{"very basic english words for beginners A1", "vocabulary,basics", "A1", "en"},
	{"english alphabet and numbers for beginners", "vocabulary,basics", "A1", "en"},
	{"simple english greetings and introductions", "speaking,greetings", "A1", "en"},
	{"english for absolute beginners conversation", "speaking,conversation", "A1", "en"},
	{"learn english colors shapes basic vocabulary", "vocabulary,basics", "A1", "en"},
	{"simple present tense english A1 beginner", "grammar,present simple", "A1", "en"},
	{"basic english listening very easy slow", "listening,practice", "A1", "en"},
	{"english daily words beginner A1 lesson", "vocabulary,daily life", "A1", "en"},
	{"how to introduce yourself in english beginner", "speaking,introductions", "A1", "en"},
	{"basic english questions and answers beginner", "speaking,questions", "A1", "en"},

	// ==================== ENGLISH A2 ====================
	{"easy english past simple for beginners", "grammar,past simple", "A2", "en"},
	{"present simple english lesson beginner", "grammar,present simple", "A2", "en"},
	{"future simple will english beginner", "grammar,future simple", "A2", "en"},
	{"present continuous english lesson easy", "grammar,present continuous", "A2", "en"},
	{"past continuous english beginner", "grammar,past continuous", "A2", "en"},
	{"present perfect english easy explanation", "grammar,present perfect", "A2", "en"},
	{"english tenses explained simply", "grammar,tenses", "A2", "en"},
	{"easy english conversation practice beginner", "speaking,conversation", "A2", "en"},
	{"daily routine english beginner", "daily life,routine", "A2", "en"},
	{"english speaking practice slow", "speaking,practice", "A2", "en"},
	{"how to describe your day in english", "speaking,daily life", "A2", "en"},
	{"english for everyday situations", "speaking,daily life", "A2", "en"},
	{"talking about weekend english", "speaking,weekend", "A2", "en"},
	{"how to tell a story in english beginner", "speaking,storytelling", "A2", "en"},
	{"most common phrasal verbs english", "vocabulary,phrasal verbs", "A2", "en"},
	{"easy phrasal verbs for beginners", "vocabulary,phrasal verbs", "A2", "en"},
	{"english vocabulary for daily life", "vocabulary,daily life", "A2", "en"},
	{"english connectors and linking words", "vocabulary,connectors", "A2", "en"},
	{"useful english phrases for conversation", "vocabulary,phrases", "A2", "en"},
	{"english collocations for beginners", "vocabulary,collocations", "A2", "en"},
	{"english listening practice easy slow", "listening,practice", "A2", "en"},
	{"english short stories for beginners", "listening,stories", "A2", "en"},
	{"english listening comprehension A2", "listening,comprehension", "A2", "en"},
	{"learn english through stories easy", "listening,stories", "A2", "en"},
	{"english at the restaurant beginner", "practical,restaurant", "A2", "en"},
	{"english for travel beginner", "practical,travel", "A2", "en"},
	{"english for shopping beginner", "practical,shopping", "A2", "en"},
	{"english job interview beginner", "practical,work", "A2", "en"},
	{"english for making friends", "practical,social", "A2", "en"},
	{"english for phone calls beginner", "practical,phone", "A2", "en"},

	// ==================== ENGLISH B1 ====================
	{"english B1 intermediate grammar lesson", "grammar,mixed tenses", "B1", "en"},
	{"conditional sentences english if clauses", "grammar,conditionals", "B1", "en"},
	{"reported speech english intermediate", "grammar,reported speech", "B1", "en"},
	{"passive voice english B1 intermediate", "grammar,passive", "B1", "en"},
	{"relative clauses english intermediate", "grammar,relative clauses", "B1", "en"},
	{"english intermediate conversation practice", "speaking,conversation", "B1", "en"},
	{"how to express opinions in english B1", "speaking,opinions", "B1", "en"},
	{"english debate and discussion phrases", "speaking,discussion", "B1", "en"},
	{"english intermediate vocabulary topics", "vocabulary,advanced", "B1", "en"},
	{"english idioms and expressions intermediate", "vocabulary,idioms", "B1", "en"},
	{"advanced phrasal verbs english B1", "vocabulary,phrasal verbs", "B1", "en"},
	{"english listening intermediate level", "listening,practice", "B1", "en"},
	{"english news for intermediate learners", "listening,news", "B1", "en"},
	{"english podcast for intermediate B1", "listening,podcast", "B1", "en"},
	{"business english intermediate level", "practical,business", "B1", "en"},

	// ==================== ENGLISH B2 ====================
	{"english B2 upper intermediate grammar", "grammar,advanced", "B2", "en"},
	{"mixed conditionals english advanced", "grammar,conditionals", "B2", "en"},
	{"english subjunctive mood upper intermediate", "grammar,subjunctive", "B2", "en"},
	{"inversion in english advanced grammar", "grammar,inversion", "B2", "en"},
	{"english B2 speaking practice fluency", "speaking,fluency", "B2", "en"},
	{"how to argue and persuade in english", "speaking,argumentation", "B2", "en"},
	{"english academic vocabulary upper intermediate", "vocabulary,academic", "B2", "en"},
	{"advanced english collocations B2", "vocabulary,collocations", "B2", "en"},
	{"english listening upper intermediate B2", "listening,practice", "B2", "en"},
	{"TED talks for english learners B2", "listening,ted talks", "B2", "en"},
	{"english for presentations upper intermediate", "practical,presentations", "B2", "en"},
	{"english email writing business B2", "practical,writing", "B2", "en"},

	// ==================== DEUTSCH A1 ====================
	{"deutsch lernen A1 Anfänger erste Wörter", "Wortschatz,Grundlagen", "A1", "de"},
	{"deutsch für Anfänger Begrüßung vorstellen", "Sprechen,Begrüßung", "A1", "de"},
	{"deutsche Zahlen und Alphabet lernen", "Wortschatz,Grundlagen", "A1", "de"},
	{"einfache deutsche Sätze für Anfänger A1", "Sprechen,Grundlagen", "A1", "de"},
	{"deutsch A1 Präsens einfach erklärt", "Grammatik,Präsens", "A1", "de"},
	{"deutsch hören A1 langsam und deutlich", "Hören,Übung", "A1", "de"},
	{"Alltag auf deutsch A1 erste Schritte", "Alltag,Grundlagen", "A1", "de"},
	{"deutsche Artikel der die das A1", "Grammatik,Artikel", "A1", "de"},
	{"sich vorstellen auf deutsch A1 Anfänger", "Sprechen,Vorstellung", "A1", "de"},
	{"einfache Fragen auf deutsch stellen A1", "Sprechen,Fragen", "A1", "de"},

	// ==================== DEUTSCH A2 ====================
	{"deutsch lernen A2 Präteritum einfach", "Grammatik,Präteritum", "A2", "de"},
	{"deutsch Perfekt haben sein einfach", "Grammatik,Perfekt", "A2", "de"},
	{"deutsch Präsens Konjugation Anfänger", "Grammatik,Präsens", "A2", "de"},
	{"trennbare Verben deutsch lernen", "Grammatik,trennbare Verben", "A2", "de"},
	{"Konjunktiv II deutsch einfach erklärt", "Grammatik,Konjunktiv", "A2", "de"},
	{"Passiv deutsch A2 Anfänger", "Grammatik,Passiv", "A2", "de"},
	{"deutsche Grammatik einfach erklärt A2", "Grammatik,Übersicht", "A2", "de"},
	{"deutsch sprechen üben Anfänger langsam", "Sprechen,Übung", "A2", "de"},
	{"Alltag auf deutsch beschreiben A2", "Alltag,Routine", "A2", "de"},
	{"deutsch Konversation einfach", "Sprechen,Konversation", "A2", "de"},
	{"deutsch für den Alltag lernen", "Alltag,Situationen", "A2", "de"},
	{"deutsch Wortschatz A2 wichtige Wörter", "Wortschatz,Grundwortschatz", "A2", "de"},
	{"deutsche Verben mit Präpositionen", "Wortschatz,Verben", "A2", "de"},
	{"deutsche Redewendungen Anfänger", "Wortschatz,Redewendungen", "A2", "de"},
	{"deutsch Hörverstehen A2 langsam", "Hören,Übung", "A2", "de"},
	{"deutsche Geschichten für Anfänger", "Hören,Geschichten", "A2", "de"},
	{"deutsch hören und verstehen A2", "Hören,Verstehen", "A2", "de"},
	{"im Restaurant bestellen deutsch lernen", "Praktisch,Restaurant", "A2", "de"},
	{"beim Arzt deutsch Vokabeln", "Praktisch,Gesundheit", "A2", "de"},
	{"einkaufen auf deutsch lernen", "Praktisch,Einkaufen", "A2", "de"},
	{"telefonieren auf deutsch A2", "Praktisch,Telefon", "A2", "de"},

	// ==================== DEUTSCH B1 ====================
	{"deutsch B1 Grammatik Konjunktiv II Übungen", "Grammatik,Konjunktiv", "B1", "de"},
	{"Nebensätze deutsch B1 weil dass obwohl", "Grammatik,Nebensätze", "B1", "de"},
	{"Passiv deutsch B1 Übungen erklärt", "Grammatik,Passiv", "B1", "de"},
	{"Relativsätze deutsch B1 einfach erklärt", "Grammatik,Relativsätze", "B1", "de"},
	{"deutsch B1 Konversation Meinung äußern", "Sprechen,Meinung", "B1", "de"},
	{"deutsch B1 Diskussion und Argumentation", "Sprechen,Diskussion", "B1", "de"},
	{"deutsch Wortschatz B1 erweitern", "Wortschatz,Fortgeschritten", "B1", "de"},
	{"deutsche Redewendungen B1 Mittelstufe", "Wortschatz,Redewendungen", "B1", "de"},
	{"deutsch Hörverstehen B1 Mittelstufe", "Hören,Übung", "B1", "de"},
	{"deutsche Nachrichten langsam B1", "Hören,Nachrichten", "B1", "de"},
	{"Bewerbungsgespräch deutsch B1", "Praktisch,Arbeit", "B1", "de"},
	{"deutsch für den Beruf B1 Mittelstufe", "Praktisch,Beruf", "B1", "de"},
}

func main() {
	godotenv.Load()

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY is required. Get one at https://console.cloud.google.com/apis/credentials")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	database := db.Connect(databaseURL)
	defer database.Close()

	database.Migrate(databaseURL)

	totalInserted := 0

	for _, sq := range searchQueries {
		log.Printf("Searching [%s]: %s", sq.language, sq.query)

		relevanceLang := "en"
		if sq.language == "de" {
			relevanceLang = "de"
		}
		videos, err := searchYouTube(apiKey, sq.query, 10, relevanceLang)
		if err != nil {
			log.Printf("  Search error: %v", err)
			continue
		}

		if len(videos) == 0 {
			log.Printf("  No results")
			continue
		}

		var videoIDs []string
		for _, v := range videos {
			videoIDs = append(videoIDs, v.ID.VideoID)
		}

		details, err := getVideoDetails(apiKey, videoIDs)
		if err != nil {
			log.Printf("  Details error: %v", err)
			continue
		}

		detailMap := make(map[string]VideoItem)
		for _, d := range details {
			detailMap[d.ID] = d
		}

		for _, v := range videos {
			detail, ok := detailMap[v.ID.VideoID]
			if !ok {
				continue
			}

			var viewCount int
			fmt.Sscanf(detail.Statistics.ViewCount, "%d", &viewCount)
			if viewCount < 50000 {
				continue
			}

			duration := parseDuration(detail.ContentDetails.Duration)
			if duration < 2*time.Minute || duration > 15*time.Minute {
				continue
			}

			hasSubtitles := detail.ContentDetails.Caption == "true"
			durationStr := formatDuration(duration)
			videoURL := "https://www.youtube.com/watch?v=" + v.ID.VideoID

			tags := extractTags(v.Snippet.Title, v.Snippet.Description, sq.topic)

			_, err := database.Pool.Exec(context.Background(),
				`INSERT INTO media_resources (title, url, media_type, level, topic, duration, tags, view_count, has_subtitles, description, active, language)
				 VALUES ($1, $2, 'video', $3, $4, $5, $6, $7, $8, $9, true, $10)
				 ON CONFLICT (url) DO UPDATE SET
				   view_count = $7, has_subtitles = $8, tags = $6, description = $9, language = $10`,
				v.Snippet.Title, videoURL, sq.level, sq.topic,
				durationStr, tags, viewCount, hasSubtitles,
				truncate(v.Snippet.Description, 500), sq.language)
			if err != nil {
				log.Printf("  Insert error: %v", err)
				continue
			}
			totalInserted++
			log.Printf("  ✅ %s (%s, %d views)", v.Snippet.Title, durationStr, viewCount)
		}

		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("\nDone! Inserted/updated %d videos", totalInserted)

	var count int
	database.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM media_resources").Scan(&count)
	log.Printf("Total media in database: %d", count)
}

func searchYouTube(apiKey, query string, maxResults int, relevanceLang string) ([]SearchItem, error) {
	u := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&q=%s&maxResults=%d&relevanceLanguage=%s&videoDuration=medium&key=%s",
		url.QueryEscape(query), maxResults, relevanceLang, apiKey)

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("YouTube API returned %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func getVideoDetails(apiKey string, videoIDs []string) ([]VideoItem, error) {
	ids := strings.Join(videoIDs, ",")
	u := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=contentDetails,statistics&id=%s&key=%s",
		ids, apiKey)

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VideoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func parseDuration(iso string) time.Duration {
	iso = strings.TrimPrefix(iso, "PT")
	var d time.Duration
	var num int

	for _, ch := range iso {
		switch {
		case ch >= '0' && ch <= '9':
			num = num*10 + int(ch-'0')
		case ch == 'H':
			d += time.Duration(num) * time.Hour
			num = 0
		case ch == 'M':
			d += time.Duration(num) * time.Minute
			num = 0
		case ch == 'S':
			d += time.Duration(num) * time.Second
			num = 0
		}
	}
	return d
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	return fmt.Sprintf("%d min", m)
}

func extractTags(title, description, baseTags string) string {
	tags := strings.Split(baseTags, ",")

	keywords := []string{
		"past simple", "present simple", "future simple",
		"present continuous", "past continuous",
		"present perfect", "past perfect",
		"phrasal verb", "vocabulary", "grammar",
		"conversation", "listening", "speaking",
		"beginner", "easy", "daily", "routine",
		"travel", "food", "work", "shopping",
	}

	combined := strings.ToLower(title + " " + description)
	for _, kw := range keywords {
		if strings.Contains(combined, kw) {
			tags = append(tags, kw)
		}
	}

	seen := make(map[string]bool)
	var unique []string
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			seen[t] = true
			unique = append(unique, t)
		}
	}

	return strings.Join(unique, ",")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
