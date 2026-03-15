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
	Language  string
}

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
	err := d.Pool.QueryRow(context.Background(),
		`SELECT gw.week_num, gw.family, gw.focus, gw.tense_name, gw.anchor, gw.markers, gw.formula, gw.example,
		        COALESCE(gw.language,'en')
		 FROM users u
		 JOIN grammar_weeks gw ON gw.week_num = COALESCE(u.current_grammar_week, 1)
		                       AND COALESCE(gw.language,'en') = COALESCE(u.language,'en')
		 WHERE u.id = $1`, userID,
	).Scan(&g.WeekNum, &g.Family, &g.Focus, &g.TenseName, &g.Anchor, &g.Markers, &g.Formula, &g.Example, &g.Language)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func DefaultGrammar(language string) *GrammarWeek {
	switch language {
	case "de":
		return &GrammarWeek{
			TenseName: "Präsens",
			Family:    "Simple",
			Anchor:    "\U0001F504 Karussell — wiederholt sich immer wieder",
			Formula:   "S + Verb (konjugiert: ich -e, du -st, er -t)",
			Markers:   "immer, oft, jeden Tag, manchmal",
			Language:  "de",
		}
	default:
		return &GrammarWeek{
			TenseName: "Past Simple",
			Family:    "Simple",
			Anchor:    "\U0001F6AA Closed door — done and finished",
			Formula:   "S + V2 (ed / irregular)",
			Markers:   "yesterday, last week, ago",
			Language:  "en",
		}
	}
}
