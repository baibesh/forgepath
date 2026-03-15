package db

import "time"

type User struct {
	ID                 int64
	Username           string
	FirstName          string
	TzOffset           int
	Level              string
	Language           string
	Active             bool
	Onboarded          bool
	SkipCount          int
	CurrentGrammarWeek int
	CreatedAt          time.Time
}

type Word struct {
	ID           int
	Word         string
	Definition   string
	Example      string
	Collocations string
	Construction string
	Level        string
	Language     string
}

type UserWord struct {
	UserID       int64
	WordID       int
	SeenAt       time.Time
	NextReview   time.Time
	IntervalDays int
	EaseFactor   float64
	Repetitions  int
	Score        int
}

type TodayStreak struct {
	WordDone    bool
	WritingDone bool
	ReviewDone  bool
}

type WeeklyStats struct {
	Days         int
	WordsDone    int
	WritingsDone int
	ReviewsDone  int
}

type UserState struct {
	UserID  int64
	State   string
	Context map[string]string
}

type GrammarWeek struct {
	WeekNum   int
	Family    string
	Focus     string
	TenseName string
	Anchor    string
	Markers   string
	Formula   string
	Example   string
	Language  string
}

type MediaResource struct {
	ID           int
	Title        string
	URL          string
	MediaType    string
	Level        string
	Topic        string
	Duration     string
	Active       bool
	Tags         string
	ViewCount    int
	HasSubtitles bool
	Description  string
	Language     string
}

type UserMedia struct {
	UserID       int64
	MediaID      int
	SentAt       time.Time
	TaskSent     bool
	TaskResponse string
	Completed    bool
}

type Writing struct {
	ID           int
	UserID       int64
	Topic        string
	GrammarFocus string
	Text         string
	Feedback     string
	WordCount    int
	WritingType  string
	CreatedAt    time.Time
}
