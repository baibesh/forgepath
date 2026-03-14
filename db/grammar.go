package db

import "context"

type GrammarWeek struct {
	WeekNum   int
	Family    string
	Focus     string
	TenseName string
	Anchor    string
	Markers   string
	Formula   string
	Example   string
}

func (d *DB) GetGrammarWeek(weekNum int) (*GrammarWeek, error) {
	var g GrammarWeek
	err := d.Pool.QueryRow(context.Background(),
		`SELECT week_num, family, focus, tense_name, anchor, markers, formula, example
		 FROM grammar_weeks WHERE week_num = $1`, weekNum,
	).Scan(&g.WeekNum, &g.Family, &g.Focus, &g.TenseName, &g.Anchor, &g.Markers, &g.Formula, &g.Example)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (d *DB) GetCurrentGrammarFocus(userID int64) (*GrammarWeek, error) {
	var g GrammarWeek
	err := d.Pool.QueryRow(context.Background(),
		`SELECT gw.week_num, gw.family, gw.focus, gw.tense_name, gw.anchor, gw.markers, gw.formula, gw.example
		 FROM users u
		 JOIN grammar_weeks gw ON gw.week_num = COALESCE(u.current_grammar_week, 1)
		 WHERE u.id = $1`, userID,
	).Scan(&g.WeekNum, &g.Family, &g.Focus, &g.TenseName, &g.Anchor, &g.Markers, &g.Formula, &g.Example)
	if err != nil {
		return nil, err
	}
	return &g, nil
}
