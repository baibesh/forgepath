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

// YouTube Data API v3 response structs
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
	query string
	topic string
	level string
}{
	// Grammar focused
	{"easy english past simple for beginners", "grammar,past simple", "A2"},
	{"present simple english lesson beginner", "grammar,present simple", "A2"},
	{"future simple will english beginner", "grammar,future simple", "A2"},
	{"present continuous english lesson easy", "grammar,present continuous", "A2"},
	{"past continuous english beginner", "grammar,past continuous", "A2"},
	{"present perfect english easy explanation", "grammar,present perfect", "A2"},
	{"english tenses explained simply", "grammar,tenses", "A2"},

	// Daily life & speaking
	{"easy english conversation practice beginner", "speaking,conversation", "A2"},
	{"daily routine english beginner", "daily life,routine", "A2"},
	{"english speaking practice slow", "speaking,practice", "A2"},
	{"how to describe your day in english", "speaking,daily life", "A2"},
	{"english for everyday situations", "speaking,daily life", "A2"},
	{"talking about weekend english", "speaking,weekend", "A2"},
	{"how to tell a story in english beginner", "speaking,storytelling", "A2"},

	// Vocabulary & phrasal verbs
	{"most common phrasal verbs english", "vocabulary,phrasal verbs", "A2"},
	{"easy phrasal verbs for beginners", "vocabulary,phrasal verbs", "A2"},
	{"english vocabulary for daily life", "vocabulary,daily life", "A2"},
	{"english connectors and linking words", "vocabulary,connectors", "A2"},
	{"useful english phrases for conversation", "vocabulary,phrases", "A2"},
	{"english collocations for beginners", "vocabulary,collocations", "A2"},

	// Listening
	{"english listening practice easy slow", "listening,practice", "A2"},
	{"english short stories for beginners", "listening,stories", "A2"},
	{"english listening comprehension A2", "listening,comprehension", "A2"},
	{"learn english through stories easy", "listening,stories", "A2"},

	// Practical situations
	{"english at the restaurant beginner", "practical,restaurant", "A2"},
	{"english for travel beginner", "practical,travel", "A2"},
	{"english for shopping beginner", "practical,shopping", "A2"},
	{"english job interview beginner", "practical,work", "A2"},
	{"english for making friends", "practical,social", "A2"},
	{"english for phone calls beginner", "practical,phone", "A2"},
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

	database.Migrate()

	totalInserted := 0

	for _, sq := range searchQueries {
		log.Printf("Searching: %s", sq.query)

		videos, err := searchYouTube(apiKey, sq.query, 10)
		if err != nil {
			log.Printf("  Search error: %v", err)
			continue
		}

		if len(videos) == 0 {
			log.Printf("  No results")
			continue
		}

		// Get video details (duration, views, captions)
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

			// Filter: >50K views
			var viewCount int
			fmt.Sscanf(detail.Statistics.ViewCount, "%d", &viewCount)
			if viewCount < 50000 {
				continue
			}

			// Filter: 2-15 min duration
			duration := parseDuration(detail.ContentDetails.Duration)
			if duration < 2*time.Minute || duration > 15*time.Minute {
				continue
			}

			hasSubtitles := detail.ContentDetails.Caption == "true"
			durationStr := formatDuration(duration)
			videoURL := "https://www.youtube.com/watch?v=" + v.ID.VideoID

			// Extract tags from title + description
			tags := extractTags(v.Snippet.Title, v.Snippet.Description, sq.topic)

			_, err := database.Pool.Exec(context.Background(),
				`INSERT INTO media_resources (title, url, media_type, level, topic, duration, tags, view_count, has_subtitles, description, active)
				 VALUES ($1, $2, 'video', $3, $4, $5, $6, $7, $8, $9, true)
				 ON CONFLICT (url) DO UPDATE SET
				   view_count = $7, has_subtitles = $8, tags = $6, description = $9`,
				v.Snippet.Title, videoURL, sq.level, sq.topic,
				durationStr, tags, viewCount, hasSubtitles,
				truncate(v.Snippet.Description, 500))
			if err != nil {
				log.Printf("  Insert error: %v", err)
				continue
			}
			totalInserted++
			log.Printf("  ✅ %s (%s, %d views)", v.Snippet.Title, durationStr, viewCount)
		}

		// Rate limit: YouTube API quota
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("\nDone! Inserted/updated %d videos", totalInserted)

	// Show total count
	var count int
	database.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM media_resources").Scan(&count)
	log.Printf("Total media in database: %d", count)
}

func searchYouTube(apiKey, query string, maxResults int) ([]SearchItem, error) {
	u := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&q=%s&maxResults=%d&relevanceLanguage=en&videoDuration=medium&key=%s",
		url.QueryEscape(query), maxResults, apiKey)

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

// parseDuration parses ISO 8601 duration (PT5M30S)
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

	// Extract keywords from title
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

	// Deduplicate
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
