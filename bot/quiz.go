package bot

import (
	"sync"
	"time"
)

type quizEntry struct {
	UserID     int64
	WordID     int
	CorrectIdx int
	CreatedAt  time.Time
}

var (
	quizPolls   = make(map[string]quizEntry)
	quizPollsMu sync.Mutex
)

func RegisterQuizPoll(pollID string, userID int64, wordID int, correctIdx int) {
	quizPollsMu.Lock()
	defer quizPollsMu.Unlock()
	quizPolls[pollID] = quizEntry{
		UserID:     userID,
		WordID:     wordID,
		CorrectIdx: correctIdx,
		CreatedAt:  time.Now(),
	}
	if len(quizPolls) > 500 {
		cutoff := time.Now().Add(-2 * time.Hour)
		for k, e := range quizPolls {
			if e.CreatedAt.Before(cutoff) {
				delete(quizPolls, k)
			}
		}
	}
}

func GetQuizPoll(pollID string) (quizEntry, bool) {
	quizPollsMu.Lock()
	defer quizPollsMu.Unlock()
	e, ok := quizPolls[pollID]
	if ok {
		delete(quizPolls, pollID)
	}
	return e, ok
}
