package db

import (
	"context"
	"fmt"
	"math/rand/v2"
)

func (d *DB) GetWordByID(id int) (*Word, error) {
	var w Word
	err := d.Pool.QueryRow(context.Background(),
		`SELECT id, word, COALESCE(definition,''), COALESCE(example,''), COALESCE(collocations,''),
		        COALESCE(construction,''), COALESCE(synonyms,''), COALESCE(antonyms,''),
		        COALESCE(examples,''), COALESCE(level,'A2'), COALESCE(language,'en')
		 FROM words WHERE id = $1`, id,
	).Scan(&w.ID, &w.Word, &w.Definition, &w.Example, &w.Collocations, &w.Construction,
		&w.Synonyms, &w.Antonyms, &w.Examples, &w.Level, &w.Language)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (d *DB) GetRandomUnseen(userID int64, level, language string) (*Word, error) {
	var count int
	err := d.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM words w
		 WHERE w.level = $2 AND COALESCE(w.language,'en') = $3
		   AND w.id NOT IN (SELECT word_id FROM user_words WHERE user_id = $1)`,
		userID, level, language).Scan(&count)
	if err != nil || count == 0 {
		return nil, fmt.Errorf("no unseen words")
	}

	offset := rand.IntN(count)

	var w Word
	err = d.Pool.QueryRow(context.Background(),
		`SELECT w.id, w.word, COALESCE(w.definition,''), COALESCE(w.example,''),
		        COALESCE(w.collocations,''), COALESCE(w.construction,''),
		        COALESCE(w.synonyms,''), COALESCE(w.antonyms,''), COALESCE(w.examples,''),
		        COALESCE(w.level,'A2'), COALESCE(w.language,'en')
		 FROM words w
		 WHERE w.level = $2 AND COALESCE(w.language,'en') = $3
		   AND w.id NOT IN (SELECT word_id FROM user_words WHERE user_id = $1)
		 LIMIT 1 OFFSET $4`, userID, level, language, offset,
	).Scan(&w.ID, &w.Word, &w.Definition, &w.Example, &w.Collocations, &w.Construction,
		&w.Synonyms, &w.Antonyms, &w.Examples, &w.Level, &w.Language)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (d *DB) GetWordsForReview(userID int64, limit int) ([]Word, error) {
	rows, err := d.Pool.Query(context.Background(),
		`SELECT w.id, w.word, COALESCE(w.definition,''), COALESCE(w.example,''),
		        COALESCE(w.collocations,''), COALESCE(w.construction,''),
		        COALESCE(w.synonyms,''), COALESCE(w.antonyms,''), COALESCE(w.examples,''),
		        COALESCE(w.level,'A2'), COALESCE(w.language,'en')
		 FROM user_words uw
		 JOIN words w ON w.id = uw.word_id
		 WHERE uw.user_id = $1 AND uw.next_review <= NOW()
		 ORDER BY uw.next_review ASC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var w Word
		if err := rows.Scan(&w.ID, &w.Word, &w.Definition, &w.Example, &w.Collocations, &w.Construction,
			&w.Synonyms, &w.Antonyms, &w.Examples, &w.Level, &w.Language); err != nil {
			return nil, err
		}
		words = append(words, w)
	}
	return words, nil
}

func (d *DB) MarkWordSeen(userID int64, wordID int) error {
	_, err := d.Pool.Exec(context.Background(),
		`INSERT INTO user_words (user_id, word_id, seen_at, next_review, interval_days, ease_factor, repetitions, score)
		 VALUES ($1, $2, NOW(), NOW() + INTERVAL '1 day', 1, 2.5, 0, 0)
		 ON CONFLICT (user_id, word_id) DO NOTHING`, userID, wordID)
	return err
}

func (d *DB) UpdateWordReview(userID int64, wordID int, intervalDays int, easeFactor float64, repetitions int) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE user_words
		 SET interval_days = $3, ease_factor = $4, repetitions = $5,
		     next_review = NOW() + ($3 || ' days')::INTERVAL
		 WHERE user_id = $1 AND word_id = $2`,
		userID, wordID, intervalDays, easeFactor, repetitions)
	return err
}

func (d *DB) GetUserWordCount(userID int64) (int, error) {
	var count int
	err := d.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM user_words WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (d *DB) GetUserWords(userID int64, offset, limit int) ([]Word, error) {
	rows, err := d.Pool.Query(context.Background(),
		`SELECT w.id, w.word, COALESCE(w.definition,''), COALESCE(w.example,''),
		        COALESCE(w.collocations,''), COALESCE(w.construction,''),
		        COALESCE(w.synonyms,''), COALESCE(w.antonyms,''), COALESCE(w.examples,''),
		        COALESCE(w.level,'A2'), COALESCE(w.language,'en')
		 FROM user_words uw
		 JOIN words w ON w.id = uw.word_id
		 WHERE uw.user_id = $1
		 ORDER BY uw.seen_at DESC
		 LIMIT $3 OFFSET $2`, userID, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var w Word
		if err := rows.Scan(&w.ID, &w.Word, &w.Definition, &w.Example, &w.Collocations, &w.Construction,
			&w.Synonyms, &w.Antonyms, &w.Examples, &w.Level, &w.Language); err != nil {
			return nil, err
		}
		words = append(words, w)
	}
	return words, nil
}

func (d *DB) GetUserWordRepetitions(userID int64, wordID int) (int, error) {
	var reps int
	err := d.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(repetitions, 0) FROM user_words WHERE user_id = $1 AND word_id = $2`,
		userID, wordID).Scan(&reps)
	if err != nil {
		return 0, nil
	}
	return reps, nil
}

func (d *DB) GetUserWordSRS(userID int64, wordID int) (repetitions int, intervalDays int, easeFactor float64, err error) {
	err = d.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(repetitions, 0), COALESCE(interval_days, 1), COALESCE(ease_factor, 2.5)
		 FROM user_words WHERE user_id = $1 AND word_id = $2`,
		userID, wordID).Scan(&repetitions, &intervalDays, &easeFactor)
	if err != nil {
		return 0, 1, 2.5, nil
	}
	return
}

func (d *DB) InsertCustomWord(word, definition, example, collocations, construction, level, language string) (int, error) {
	var id int
	err := d.Pool.QueryRow(context.Background(),
		`INSERT INTO words (word, definition, example, collocations, construction, level, language)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (word, language) DO UPDATE SET
		   definition = EXCLUDED.definition, example = EXCLUDED.example,
		   collocations = EXCLUDED.collocations, construction = EXCLUDED.construction
		 RETURNING id`,
		word, definition, example, collocations, construction, level, language,
	).Scan(&id)
	return id, err
}

func (d *DB) UpdateWordEnrichment(wordID int, synonyms, antonyms, examples string) error {
	_, err := d.Pool.Exec(context.Background(),
		`UPDATE words SET synonyms = $2, antonyms = $3, examples = $4 WHERE id = $1`,
		wordID, synonyms, antonyms, examples)
	return err
}

func (d *DB) GetWordByText(word, language string) (*Word, error) {
	var w Word
	err := d.Pool.QueryRow(context.Background(),
		`SELECT id, word, COALESCE(definition,''), COALESCE(example,''), COALESCE(collocations,''),
		        COALESCE(construction,''), COALESCE(synonyms,''), COALESCE(antonyms,''),
		        COALESCE(examples,''), COALESCE(level,'A2'), COALESCE(language,'en')
		 FROM words WHERE LOWER(word) = LOWER($1) AND COALESCE(language,'en') = $2`, word, language,
	).Scan(&w.ID, &w.Word, &w.Definition, &w.Example, &w.Collocations, &w.Construction,
		&w.Synonyms, &w.Antonyms, &w.Examples, &w.Level, &w.Language)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (d *DB) GetRandomWordsExcluding(wordID int, level string, limit int) ([]Word, error) {
	rows, err := d.Pool.Query(context.Background(),
		`SELECT id, word, COALESCE(definition,''), COALESCE(example,''),
		        COALESCE(collocations,''), COALESCE(construction,''),
		        COALESCE(synonyms,''), COALESCE(antonyms,''), COALESCE(examples,''),
		        COALESCE(level,'A2'), COALESCE(language,'en')
		 FROM words WHERE id != $1 AND level = $2
		 ORDER BY RANDOM() LIMIT $3`, wordID, level, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var w Word
		if err := rows.Scan(&w.ID, &w.Word, &w.Definition, &w.Example, &w.Collocations, &w.Construction,
			&w.Synonyms, &w.Antonyms, &w.Examples, &w.Level, &w.Language); err != nil {
			return nil, err
		}
		words = append(words, w)
	}
	return words, nil
}
