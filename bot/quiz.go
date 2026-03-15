package bot

import "sync"

type quizEntry struct {
	UserID     int64
	WordID     int
	CorrectIdx int
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
	}
	// Keep map small: remove old entries if > 1000
	if len(quizPolls) > 1000 {
		i := 0
		for k := range quizPolls {
			if i > 500 {
				break
			}
			delete(quizPolls, k)
			i++
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
