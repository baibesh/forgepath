package db

import "context"

func (d *DB) GetGrammarWeek(weekNum int, language string) (*GrammarWeek, error) {
	var g GrammarWeek
	err := d.Pool.QueryRow(context.Background(),
		`SELECT week_num, family, focus, tense_name, anchor, markers, formula, example, COALESCE(language,'en')
		 FROM grammar_weeks WHERE week_num = $1 AND COALESCE(language,'en') = $2`, weekNum, language,
	).Scan(&g.WeekNum, &g.Family, &g.Focus, &g.TenseName, &g.Anchor, &g.Markers, &g.Formula, &g.Example, &g.Language)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (d *DB) GetCurrentGrammarFocus(userID int64) (*GrammarWeek, error) {
	var g GrammarWeek
	// Use target_language (what user is learning), not language (UI language)
	err := d.Pool.QueryRow(context.Background(),
		`SELECT gw.week_num, gw.family, gw.focus, gw.tense_name, gw.anchor, gw.markers, gw.formula, gw.example,
		        COALESCE(gw.language,'en')
		 FROM users u
		 JOIN grammar_weeks gw ON gw.week_num = COALESCE(u.current_grammar_week, 1)
		                       AND COALESCE(gw.language,'en') = COALESCE(u.target_language,'en')
		 WHERE u.id = $1`, userID,
	).Scan(&g.WeekNum, &g.Family, &g.Focus, &g.TenseName, &g.Anchor, &g.Markers, &g.Formula, &g.Example, &g.Language)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func DefaultGrammar(language string) *GrammarWeek {
	// Default grammar is always English (the target learning language)
	return &GrammarWeek{
		TenseName: "Past Simple",
		Family:    "Simple",
		Anchor:    "\U0001F6AA Закрытая дверь — действие завершено, дверь захлопнулась. Представь дверь, которая закрылась — назад не вернёшься.",
		Formula:   "S + V2 (ed / irregular)",
		Markers:   "yesterday, last week, ago",
		Language:  "en",
	}
}
