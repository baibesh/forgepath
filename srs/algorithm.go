package srs

import "math"

type ReviewResult struct {
	IntervalDays int
	EaseFactor   float64
	Repetitions  int
}

func Calculate(repetitions int, intervalDays int, easeFactor float64, score int) ReviewResult {
	if score < 3 {
		return ReviewResult{
			IntervalDays: 1,
			EaseFactor:   math.Max(1.3, easeFactor-0.2),
			Repetitions:  0,
		}
	}

	var newInterval int
	switch repetitions {
	case 0:
		newInterval = 1
	case 1:
		newInterval = 3
	default:
		newInterval = int(math.Round(float64(intervalDays) * easeFactor))
	}

	newEF := easeFactor + (0.1 - float64(5-score)*(0.08+float64(5-score)*0.02))
	if newEF < 1.3 {
		newEF = 1.3
	}

	return ReviewResult{
		IntervalDays: newInterval,
		EaseFactor:   newEF,
		Repetitions:  repetitions + 1,
	}
}
