package db

import (
	"context"
	"database/sql"
	"embed"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Migrate runs versioned SQL migrations via goose, then seeds initial data.
func (d *DB) Migrate(databaseURL string) {
	// goose needs *sql.DB — open a separate connection for migrations only
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Migration: cannot open DB: %v", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Migration: cannot set dialect: %v", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatalf("Migration: %v", err)
	}

	log.Println("Goose migrations applied")

	// Seed initial data (idempotent — ON CONFLICT DO NOTHING)
	ctx := context.Background()
	d.seedGrammarWeeksEN(ctx)
	d.seedGrammarWeeksDE(ctx)
	d.seedWordsEN(ctx)
	d.seedWordsDE(ctx)
	d.seedMediaEN(ctx)
	d.seedMediaDE(ctx)

	log.Println("Database migration and seeding completed")
}

// ==================== ENGLISH ====================

func (d *DB) seedGrammarWeeksEN(ctx context.Context) {
	weeks := []struct {
		num                                                    int
		family, focus, tenseName, anchor, markers, formula, ex string
	}{
		{1, "Simple", "Past Simple", "Past Simple",
			"\U0001F6AA Closed door — done and finished",
			"yesterday, last week, ago, in 2020",
			"S + V2 (ed / irregular)",
			"I watched a movie yesterday."},
		{2, "Simple", "Present Simple", "Present Simple",
			"\U0001F504 Carousel — repeats again and again",
			"always, usually, every day, sometimes",
			"S + V1 (he/she +s)",
			"I usually wake up at 7."},
		{3, "Simple", "Future Simple", "Future Simple",
			"\U0001F52E Crystal ball — decision right now",
			"tomorrow, next week, I think, probably",
			"S + will + V1",
			"I will call you tomorrow."},
		{4, "Continuous", "Present Continuous", "Present Continuous",
			"\U0001F4F8 Photo — happening right now",
			"now, right now, at the moment, look!",
			"S + am/is/are + Ving",
			"I am reading a book right now."},
		{5, "Continuous", "Past Continuous", "Past Continuous",
			"\U0001F3AC Movie scene — background action in the past",
			"while, when, at that moment, all day yesterday",
			"S + was/were + Ving",
			"I was cooking when you called."},
		{6, "Perfect", "Present Perfect", "Present Perfect",
			"\U0001F309 Bridge — from past to present, result matters",
			"already, yet, just, ever, never, since, for",
			"S + have/has + V3",
			"I have already finished my homework."},
		{7, "Perfect", "Past Perfect", "Past Perfect",
			"\u23EA Rewind — action BEFORE another past action",
			"before, after, by the time, already (past context)",
			"S + had + V3",
			"I had eaten before she arrived."},
	}

	for _, w := range weeks {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'en')
			 ON CONFLICT (week_num, language) DO NOTHING`,
			w.num, w.family, w.focus, w.tenseName, w.anchor, w.markers, w.formula, w.ex)
		if err != nil {
			log.Printf("Seed EN grammar week %d warning: %v", w.num, err)
		}
	}
}

func (d *DB) seedGrammarWeeksDE(ctx context.Context) {
	weeks := []struct {
		num                                                    int
		family, focus, tenseName, anchor, markers, formula, ex string
	}{
		{1, "Einfach", "Pr\u00e4sens", "Pr\u00e4sens",
			"\U0001F504 Karussell — wiederholt sich immer wieder",
			"immer, oft, jeden Tag, manchmal, normalerweise",
			"S + Verb (ich -e, du -st, er/sie -t, wir -en)",
			"Ich lerne jeden Tag Deutsch."},
		{2, "Einfach", "Pr\u00e4teritum", "Pr\u00e4teritum",
			"\U0001F6AA Geschlossene T\u00fcr — abgeschlossen und vorbei",
			"gestern, letztes Jahr, damals, vor einer Woche, fr\u00fcher",
			"S + Verb (Pr\u00e4teritum-Stamm + Endung)",
			"Ich ging gestern ins Kino."},
		{3, "Zusammengesetzt", "Perfekt", "Perfekt",
			"\U0001F309 Br\u00fccke — Vergangenheit mit Bezug zur Gegenwart",
			"schon, bereits, gerade, noch nicht, heute",
			"S + haben/sein + ... + Partizip II",
			"Ich habe das Buch gelesen."},
		{4, "Einfach", "Futur I", "Futur I",
			"\U0001F52E Kristallkugel — Pl\u00e4ne und Vermutungen",
			"morgen, n\u00e4chste Woche, bald, in Zukunft, wahrscheinlich",
			"S + werden + ... + Infinitiv",
			"Ich werde morgen nach Berlin fahren."},
		{5, "Konjunktiv", "Konjunktiv II", "Konjunktiv II",
			"\U0001F4AD Gedankenblase — W\u00fcnsche und irreale Situationen",
			"wenn, h\u00e4tte, w\u00fcrde, k\u00f6nnte, gern",
			"S + w\u00fcrde + ... + Infinitiv / h\u00e4tte / w\u00e4re",
			"Wenn ich reich w\u00e4re, w\u00fcrde ich reisen."},
		{6, "Zusammengesetzt", "Plusquamperfekt", "Plusquamperfekt",
			"\u23EA R\u00fcckspulen — vor einer anderen Vergangenheit",
			"bevor, nachdem, bereits, schon (Vergangenheit)",
			"S + hatte/war + ... + Partizip II",
			"Ich hatte gegessen, bevor sie kam."},
		{7, "Passiv", "Passiv", "Passiv",
			"\U0001F3AD Maske — wer es tut ist unwichtig",
			"von, durch, es wird, wurde",
			"S + werden + ... + Partizip II",
			"Das Haus wird renoviert."},
	}

	for _, w := range weeks {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'de')
			 ON CONFLICT (week_num, language) DO NOTHING`,
			w.num, w.family, w.focus, w.tenseName, w.anchor, w.markers, w.formula, w.ex)
		if err != nil {
			log.Printf("Seed DE grammar week %d warning: %v", w.num, err)
		}
	}
}

func (d *DB) seedWordsEN(ctx context.Context) {
	words := []struct {
		word, definition, example, level, collocations, construction string
	}{
		{"figure out", "to understand, to solve", "I finally figured out how to use this app.", "A2",
			"figure out a problem, figure out the answer, figure out how", "figure out + how/what/why"},
		{"give up", "to stop trying", "Don't give up! You can do it.", "A2",
			"give up hope, give up trying, give up smoking", "give up + Ving / noun"},
		{"look forward to", "to wait with excitement", "I'm looking forward to the weekend.", "A2",
			"look forward to meeting, look forward to seeing", "look forward to + Ving / noun"},
		{"turn out", "to end up being", "It turned out to be a great movie.", "A2",
			"turn out to be, turn out well, turn out that", "turn out + to be / that"},
		{"come up with", "to think of, to create", "She came up with a brilliant idea.", "A2",
			"come up with an idea, come up with a plan, come up with a solution", "come up with + noun"},
		{"run out of", "to have none left", "We ran out of milk.", "A2",
			"run out of time, run out of money, run out of patience", "run out of + noun"},
		{"get along with", "to have a good relationship", "I get along with my coworkers.", "A2",
			"get along with people, get along well, get along with someone", "get along with + person"},
		{"pick up", "to learn, to collect", "I picked up some new words from the movie.", "A2",
			"pick up a language, pick up the phone, pick up a skill", "pick up + noun"},
		{"put off", "to delay, to postpone", "Stop putting off your homework!", "A2",
			"put off a meeting, put off doing something", "put off + Ving / noun"},
		{"carry on", "to continue", "Carry on with your work.", "A2",
			"carry on working, carry on with something", "carry on + Ving / with + noun"},
		{"break down", "to stop working, to cry", "My car broke down on the highway.", "A2",
			"break down in tears, break down a problem", "break down + (no object) / break down + noun"},
		{"set up", "to organize, to configure", "Let's set up a meeting for Monday.", "A2",
			"set up a business, set up a meeting, set up an account", "set up + noun"},
		{"find out", "to discover, to learn", "I found out the truth yesterday.", "A2",
			"find out the truth, find out about, find out that", "find out + about / that / noun"},
		{"go through", "to experience", "She went through a difficult time.", "A2",
			"go through a phase, go through changes, go through a process", "go through + noun"},
		{"bring up", "to mention, to raise", "Don't bring up that topic again.", "A2",
			"bring up a topic, bring up children, bring up an issue", "bring up + noun"},
		{"deal with", "to handle, to cope with", "I need to deal with this problem.", "A2",
			"deal with a problem, deal with stress, deal with people", "deal with + noun"},
		{"point out", "to indicate, to show", "She pointed out my mistake.", "A2",
			"point out a mistake, point out that, point out a problem", "point out + noun / that"},
		{"end up", "to finally be in a situation", "We ended up staying home.", "A2",
			"end up doing, end up in a place, end up being", "end up + Ving / in + place"},
		{"take off", "to leave the ground, to remove", "The plane took off on time.", "A2",
			"take off clothes, take off from work, plane takes off", "take off + noun / (no object)"},
		{"show up", "to appear, to arrive", "He didn't show up to the meeting.", "A2",
			"show up late, show up on time, show up unexpectedly", "show up + (no object) / at + place"},
		{"meanwhile", "at the same time", "I cooked dinner. Meanwhile, she set the table.", "A2",
			"meanwhile in, meanwhile back", "meanwhile + clause"},
		{"although", "even though", "Although it was raining, we went for a walk.", "A2",
			"although it seems, although I know", "although + clause"},
		{"actually", "in fact, really", "I actually enjoyed the movie.", "A2",
			"actually quite, actually think, actually happened", "actually + verb / adjective"},
		{"definitely", "certainly, for sure", "I will definitely come to your party.", "A2",
			"definitely agree, definitely need, definitely want", "definitely + verb"},
		{"probably", "most likely", "She will probably be late.", "A2",
			"probably not, probably the best, probably should", "probably + verb / adjective"},
		{"recently", "not long ago", "I recently started learning English.", "A2",
			"recently discovered, recently started, recently moved", "recently + Past Simple / Present Perfect"},
		{"especially", "particularly", "I love fruits, especially mangoes.", "A2",
			"especially when, especially important, especially for", "especially + noun / when / adjective"},
		{"instead", "as a replacement", "I didn't go out. Instead, I stayed home.", "A2",
			"instead of doing, instead of that", "instead + clause / instead of + Ving"},
		{"however", "but, nevertheless", "The test was hard. However, I passed it.", "A2",
			"however much, however difficult", "however + clause"},
		{"manage to", "to succeed in doing", "I managed to finish the project on time.", "A2",
			"manage to do, manage to find, manage to get", "manage to + V1"},
		{"be supposed to", "to be expected to", "You are supposed to be here at 9.", "A2",
			"supposed to do, supposed to be, supposed to know", "be supposed to + V1"},
		{"used to", "past habit", "I used to play football every day.", "A2",
			"used to live, used to be, used to think", "used to + V1"},
		{"be about to", "to be going to very soon", "The movie is about to start.", "A2",
			"about to leave, about to start, about to happen", "be about to + V1"},
		{"afford", "to be able to pay for", "I can't afford a new phone.", "A2",
			"afford to buy, can't afford, afford the time", "can/can't afford + to V1 / noun"},
		{"improve", "to make better", "I want to improve my English.", "A2",
			"improve skills, improve performance, improve quality", "improve + noun"},
		{"appreciate", "to be grateful for", "I really appreciate your help.", "A2",
			"appreciate help, appreciate the effort, appreciate it", "appreciate + noun / Ving"},
		{"avoid", "to stay away from", "Try to avoid making the same mistake.", "A2",
			"avoid doing, avoid mistakes, avoid problems", "avoid + Ving / noun"},
		{"recommend", "to suggest", "I recommend watching this movie.", "A2",
			"recommend doing, recommend a book, highly recommend", "recommend + Ving / noun"},
		{"require", "to need, to demand", "This job requires experience.", "A2",
			"require experience, require attention, require effort", "require + noun / Ving"},
		{"consider", "to think about", "I consider him a good friend.", "A2",
			"consider doing, consider the options, consider important", "consider + Ving / noun + adjective"},
		{"suggest", "to propose", "I suggest taking a break.", "A2",
			"suggest doing, suggest an idea, suggest that", "suggest + Ving / that + clause"},
		{"depend on", "to be determined by", "It depends on the weather.", "A2",
			"depend on someone, depend on the situation", "depend on + noun"},
		{"belong to", "to be owned by", "This book belongs to me.", "A2",
			"belong to someone, belong to a group", "belong to + noun"},
		{"consist of", "to be made up of", "The team consists of five people.", "A2",
			"consist of parts, consist of members", "consist of + noun"},
		{"respond to", "to answer, to reply", "She didn't respond to my message.", "A2",
			"respond to a question, respond to a message, respond quickly", "respond to + noun"},
		{"ordinary", "normal, usual", "It was just an ordinary day.", "A2",
			"ordinary people, ordinary life, ordinary day", "ordinary + noun"},
		{"essential", "very important, necessary", "Sleep is essential for health.", "A2",
			"essential for, essential part, essential information", "essential + for / noun"},
		{"obvious", "easy to see or understand", "The answer was obvious.", "A2",
			"obvious reason, obvious choice, obviously wrong", "obvious + noun / that"},
		{"entire", "whole, complete", "I spent the entire day reading.", "A2",
			"entire day, entire life, entire team", "entire + noun"},
		{"convenient", "easy to use, suitable", "This time is convenient for me.", "A2",
			"convenient time, convenient location, convenient for", "convenient + for + noun"},
	}

	for _, w := range words {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO words (word, definition, example, level, collocations, construction, language)
			 VALUES ($1, $2, $3, $4, $5, $6, 'en')
			 ON CONFLICT (word, language) DO NOTHING`,
			w.word, w.definition, w.example, w.level, w.collocations, w.construction)
		if err != nil {
			log.Printf("Seed EN word '%s' warning: %v", w.word, err)
		}
	}
}

func (d *DB) seedWordsDE(ctx context.Context) {
	words := []struct {
		word, definition, example, level, collocations, construction string
	}{
		{"verstehen", "\u043F\u043E\u043D\u044F\u0442\u044C, \u043F\u043E\u043D\u0438\u043C\u0430\u0442\u044C", "Ich verstehe die Frage nicht.", "A2",
			"gut verstehen, richtig verstehen, falsch verstehen", "verstehen + Akk"},
		{"brauchen", "\u043D\u0443\u0436\u0434\u0430\u0442\u044C\u0441\u044F", "Ich brauche deine Hilfe.", "A2",
			"dringend brauchen, Zeit brauchen, Hilfe brauchen", "brauchen + Akk"},
		{"versuchen", "\u043F\u044B\u0442\u0430\u0442\u044C\u0441\u044F", "Ich versuche, p\u00fcnktlich zu kommen.", "A2",
			"es versuchen, versuchen zu verstehen", "versuchen + zu + Inf"},
		{"erkl\u00e4ren", "\u043E\u0431\u044A\u044F\u0441\u043D\u044F\u0442\u044C", "K\u00f6nnen Sie das bitte erkl\u00e4ren?", "A2",
			"genau erkl\u00e4ren, einfach erkl\u00e4ren", "erkl\u00e4ren + Akk + Dat"},
		{"empfehlen", "\u0440\u0435\u043A\u043E\u043C\u0435\u043D\u0434\u043E\u0432\u0430\u0442\u044C", "Ich empfehle dir dieses Restaurant.", "A2",
			"sehr empfehlen, weiterempfehlen", "empfehlen + Dat + Akk"},
		{"sich freuen", "\u0440\u0430\u0434\u043E\u0432\u0430\u0442\u044C\u0441\u044F", "Ich freue mich auf das Wochenende.", "A2",
			"sich freuen auf, sich freuen \u00fcber", "sich freuen auf + Akk / \u00fcber + Akk"},
		{"sich entscheiden", "\u0440\u0435\u0448\u0438\u0442\u044C\u0441\u044F", "Er hat sich f\u00fcr das rote Auto entschieden.", "A2",
			"sich schnell entscheiden, sich entscheiden f\u00fcr", "sich entscheiden f\u00fcr + Akk"},
		{"anfangen", "\u043D\u0430\u0447\u0438\u043D\u0430\u0442\u044C", "Der Film f\u00e4ngt um 8 Uhr an.", "A2",
			"anfangen mit, anfangen zu arbeiten", "anfangen mit + Dat / zu + Inf"},
		{"aufh\u00f6ren", "\u043F\u0440\u0435\u043A\u0440\u0430\u0449\u0430\u0442\u044C", "H\u00f6r bitte auf zu reden.", "A2",
			"aufh\u00f6ren mit, aufh\u00f6ren zu rauchen", "aufh\u00f6ren mit + Dat / zu + Inf"},
		{"sich gew\u00f6hnen", "\u043F\u0440\u0438\u0432\u044B\u043A\u043D\u0443\u0442\u044C", "Ich gew\u00f6hne mich an das kalte Wetter.", "A2",
			"sich gew\u00f6hnen an, sich daran gew\u00f6hnen", "sich gew\u00f6hnen an + Akk"},
		{"vorschlagen", "\u043F\u0440\u0435\u0434\u043B\u0430\u0433\u0430\u0442\u044C", "Ich schlage vor, ins Kino zu gehen.", "A2",
			"einen Vorschlag machen, vorschlagen zu", "vorschlagen + zu + Inf / Dat + Akk"},
		{"sich unterhalten", "\u0431\u0435\u0441\u0435\u0434\u043E\u0432\u0430\u0442\u044C", "Wir haben uns \u00fcber Politik unterhalten.", "A2",
			"sich unterhalten \u00fcber, sich gut unterhalten", "sich unterhalten \u00fcber + Akk / mit + Dat"},
		{"sich beschweren", "\u0436\u0430\u043B\u043E\u0432\u0430\u0442\u044C\u0441\u044F", "Er beschwert sich \u00fcber den L\u00e4rm.", "A2",
			"sich beschweren \u00fcber, sich bei jemandem beschweren", "sich beschweren \u00fcber + Akk / bei + Dat"},
		{"sich k\u00fcmmern", "\u0437\u0430\u0431\u043E\u0442\u0438\u0442\u044C\u0441\u044F", "Sie k\u00fcmmert sich um ihre Eltern.", "A2",
			"sich k\u00fcmmern um, darum k\u00fcmmern", "sich k\u00fcmmern um + Akk"},
		{"abh\u00e4ngen", "\u0437\u0430\u0432\u0438\u0441\u0435\u0442\u044C", "Das h\u00e4ngt vom Wetter ab.", "A2",
			"abh\u00e4ngen von, davon abh\u00e4ngen", "abh\u00e4ngen von + Dat"},
		{"sich vorbereiten", "\u0433\u043E\u0442\u043E\u0432\u0438\u0442\u044C\u0441\u044F", "Ich bereite mich auf die Pr\u00fcfung vor.", "A2",
			"sich vorbereiten auf, gut vorbereiten", "sich vorbereiten auf + Akk"},
		{"bestehen", "\u0441\u0434\u0430\u0442\u044C (\u044D\u043A\u0437\u0430\u043C\u0435\u043D)", "Sie hat die Pr\u00fcfung bestanden.", "A2",
			"eine Pr\u00fcfung bestehen, bestehen aus", "bestehen + Akk / bestehen aus + Dat"},
		{"verbessern", "\u0443\u043B\u0443\u0447\u0448\u0438\u0442\u044C", "Ich m\u00f6chte mein Deutsch verbessern.", "A2",
			"sich verbessern, Leistung verbessern", "verbessern + Akk / sich verbessern"},
		{"bemerken", "\u0437\u0430\u043C\u0435\u0442\u0438\u0442\u044C", "Ich habe den Fehler nicht bemerkt.", "A2",
			"sofort bemerken, kaum bemerken", "bemerken + Akk / dass"},
		{"sich ver\u00e4ndern", "\u0438\u0437\u043C\u0435\u043D\u0438\u0442\u044C\u0441\u044F", "Die Stadt hat sich sehr ver\u00e4ndert.", "A2",
			"sich stark ver\u00e4ndern, sich kaum ver\u00e4ndern", "sich ver\u00e4ndern"},
		{"trotzdem", "\u0442\u0435\u043C \u043D\u0435 \u043C\u0435\u043D\u0435\u0435", "Es regnete, trotzdem ging er spazieren.", "A2",
			"trotzdem machen, trotzdem kommen", "trotzdem + Hauptsatz"},
		{"eigentlich", "\u043D\u0430 \u0441\u0430\u043C\u043E\u043C \u0434\u0435\u043B\u0435", "Eigentlich wollte ich zu Hause bleiben.", "A2",
			"eigentlich nicht, eigentlich schon", "eigentlich + Verb"},
		{"wahrscheinlich", "\u0432\u0435\u0440\u043E\u044F\u0442\u043D\u043E", "Er kommt wahrscheinlich morgen.", "A2",
			"sehr wahrscheinlich, wahrscheinlich nicht", "wahrscheinlich + Verb"},
		{"unbedingt", "\u043E\u0431\u044F\u0437\u0430\u0442\u0435\u043B\u044C\u043D\u043E", "Du musst unbedingt diesen Film sehen.", "A2",
			"unbedingt brauchen, nicht unbedingt", "unbedingt + Verb"},
		{"au\u00dferdem", "\u043A\u0440\u043E\u043C\u0435 \u0442\u043E\u0433\u043E", "Au\u00dferdem habe ich keine Zeit.", "A2",
			"au\u00dferdem noch, au\u00dferdem auch", "au\u00dferdem + Hauptsatz"},
		{"deshalb", "\u043F\u043E\u044D\u0442\u043E\u043C\u0443", "Ich war m\u00fcde, deshalb bin ich fr\u00fch ins Bett gegangen.", "A2",
			"genau deshalb, deshalb auch", "deshalb + Verb (Position 1)"},
		{"obwohl", "\u0445\u043E\u0442\u044F", "Obwohl es kalt war, ging sie ohne Jacke.", "A2",
			"obwohl es schwer ist, obwohl ich wei\u00df", "obwohl + Nebensatz (Verb am Ende)"},
		{"falls", "\u0432 \u0441\u043B\u0443\u0447\u0430\u0435 \u0435\u0441\u043B\u0438", "Falls du Fragen hast, ruf mich an.", "A2",
			"falls n\u00f6tig, falls m\u00f6glich", "falls + Nebensatz (Verb am Ende)"},
		{"stattdessen", "\u0432\u043C\u0435\u0441\u0442\u043E \u044D\u0442\u043E\u0433\u043E", "Ich blieb zu Hause. Stattdessen habe ich gelesen.", "A2",
			"stattdessen lieber, stattdessen machen", "stattdessen + Hauptsatz"},
		{"inzwischen", "\u0442\u0435\u043C \u0432\u0440\u0435\u043C\u0435\u043D\u0435\u043C", "Ich koche. Inzwischen kannst du den Tisch decken.", "A2",
			"inzwischen schon, inzwischen fertig", "inzwischen + Hauptsatz"},
		{"sich lohnen", "\u0441\u0442\u043E\u0438\u0442\u044C \u0442\u043E\u0433\u043E", "Es lohnt sich, Deutsch zu lernen.", "A2",
			"sich lohnen zu, es lohnt sich", "sich lohnen + zu + Inf"},
		{"schaffen", "\u0441\u043F\u0440\u0430\u0432\u0438\u0442\u044C\u0441\u044F", "Ich habe es geschafft!", "A2",
			"es schaffen, rechtzeitig schaffen", "schaffen + Akk / es schaffen"},
		{"sich vorstellen", "\u043F\u0440\u0435\u0434\u0441\u0442\u0430\u0432\u0438\u0442\u044C \u0441\u0435\u0431\u0435", "Ich kann mir das nicht vorstellen.", "A2",
			"sich vorstellen k\u00f6nnen, sich etwas vorstellen", "sich (Dat) vorstellen + Akk"},
		{"vermissen", "\u0441\u043A\u0443\u0447\u0430\u0442\u044C", "Ich vermisse meine Familie.", "A2",
			"jemanden vermissen, sehr vermissen", "vermissen + Akk"},
		{"sich entschuldigen", "\u0438\u0437\u0432\u0438\u043D\u0438\u0442\u044C\u0441\u044F", "Ich entschuldige mich f\u00fcr den Fehler.", "A2",
			"sich entschuldigen f\u00fcr, sich bei jemandem entschuldigen", "sich entschuldigen f\u00fcr + Akk / bei + Dat"},
		{"sich erinnern", "\u0432\u0441\u043F\u043E\u043C\u0438\u043D\u0430\u0442\u044C", "Ich erinnere mich an den Urlaub.", "A2",
			"sich erinnern an, sich gut erinnern", "sich erinnern an + Akk"},
		{"sich interessieren", "\u0438\u043D\u0442\u0435\u0440\u0435\u0441\u043E\u0432\u0430\u0442\u044C\u0441\u044F", "Ich interessiere mich f\u00fcr Kunst.", "A2",
			"sich interessieren f\u00fcr, sich sehr interessieren", "sich interessieren f\u00fcr + Akk"},
		{"sich besch\u00e4ftigen", "\u0437\u0430\u043D\u0438\u043C\u0430\u0442\u044C\u0441\u044F", "Ich besch\u00e4ftige mich mit dem Projekt.", "A2",
			"sich besch\u00e4ftigen mit, sich intensiv besch\u00e4ftigen", "sich besch\u00e4ftigen mit + Dat"},
		{"gem\u00fctlich", "\u0443\u044E\u0442\u043D\u044B\u0439", "Das Caf\u00e9 ist sehr gem\u00fctlich.", "A2",
			"gem\u00fctliche Atmosph\u00e4re, gem\u00fctlicher Abend", "gem\u00fctlich + Nomen"},
		{"selbstverst\u00e4ndlich", "\u0441\u0430\u043C\u043E \u0441\u043E\u0431\u043E\u0439 \u0440\u0430\u0437\u0443\u043C\u0435\u0435\u0442\u0441\u044F", "Selbstverst\u00e4ndlich helfe ich dir.", "A2",
			"selbstverst\u00e4ndlich machen, selbstverst\u00e4ndlich sein", "selbstverst\u00e4ndlich + Verb"},
		{"\u00fcberraschend", "\u0443\u0434\u0438\u0432\u0438\u0442\u0435\u043B\u044C\u043D\u043E", "Das Ergebnis war \u00fcberraschend.", "A2",
			"\u00fcberraschend gut, \u00fcberraschend schnell", "\u00fcberraschend + Adj / Nomen"},
		{"geduldig", "\u0442\u0435\u0440\u043F\u0435\u043B\u0438\u0432\u044B\u0439", "Sie ist sehr geduldig mit Kindern.", "A2",
			"geduldig sein, geduldig warten", "geduldig + mit + Dat"},
		{"zuverl\u00e4ssig", "\u043D\u0430\u0434\u0451\u0436\u043D\u044B\u0439", "Er ist ein zuverl\u00e4ssiger Freund.", "A2",
			"zuverl\u00e4ssig sein, zuverl\u00e4ssige Quelle", "zuverl\u00e4ssig + Nomen"},
		{"offensichtlich", "\u043E\u0447\u0435\u0432\u0438\u0434\u043D\u043E", "Offensichtlich hat er recht.", "A2",
			"offensichtlich falsch, offensichtlich richtig", "offensichtlich + Verb"},
		{"gr\u00fcndlich", "\u0442\u0449\u0430\u0442\u0435\u043B\u044C\u043D\u043E", "Du musst gr\u00fcndlich putzen.", "A2",
			"gr\u00fcndlich pr\u00fcfen, gr\u00fcndlich lesen", "gr\u00fcndlich + Verb"},
		{"allm\u00e4hlich", "\u043F\u043E\u0441\u0442\u0435\u043F\u0435\u043D\u043D\u043E", "Allm\u00e4hlich wird es besser.", "A2",
			"allm\u00e4hlich verstehen, allm\u00e4hlich lernen", "allm\u00e4hlich + Verb"},
		{"ausgezeichnet", "\u043E\u0442\u043B\u0438\u0447\u043D\u043E", "Das Essen war ausgezeichnet.", "A2",
			"ausgezeichnete Qualit\u00e4t, ausgezeichnet schmecken", "ausgezeichnet + Nomen"},
		{"angenehm", "\u043F\u0440\u0438\u044F\u0442\u043D\u044B\u0439", "Es war ein angenehmer Abend.", "A2",
			"angenehme Atmosph\u00e4re, angenehm \u00fcberrascht", "angenehm + Nomen"},
		{"gleichzeitig", "\u043E\u0434\u043D\u043E\u0432\u0440\u0435\u043C\u0435\u043D\u043D\u043E", "Er kann gleichzeitig lesen und essen.", "A2",
			"gleichzeitig machen, gleichzeitig passieren", "gleichzeitig + Verb"},
	}

	for _, w := range words {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO words (word, definition, example, level, collocations, construction, language)
			 VALUES ($1, $2, $3, $4, $5, $6, 'de')
			 ON CONFLICT (word, language) DO NOTHING`,
			w.word, w.definition, w.example, w.level, w.collocations, w.construction)
		if err != nil {
			log.Printf("Seed DE word '%s' warning: %v", w.word, err)
		}
	}
}

func (d *DB) seedMediaEN(ctx context.Context) {
	media := []struct {
		title, url, mediaType, level, topic, duration string
	}{
		{"Morning Routine — Easy English", "https://www.youtube.com/watch?v=GGp25fn25Cs", "video", "A2", "daily life", "5 min"},
		{"At the Restaurant — Easy English", "https://www.youtube.com/watch?v=BGHxLfRGk3I", "video", "A2", "food", "6 min"},
		{"My Daily Routine — Bob the Canadian", "https://www.youtube.com/watch?v=MIuoBGFMEAo", "video", "A2", "daily life", "8 min"},
		{"Shopping Vocabulary — English with Lucy", "https://www.youtube.com/watch?v=h4X-Oyl91sE", "video", "A2", "shopping", "10 min"},
		{"Travel English — Easy Conversations", "https://www.youtube.com/watch?v=tfJRwNo2SJI", "video", "A2", "travel", "7 min"},
		{"English Listening Practice — Slow Easy", "https://www.youtube.com/watch?v=MqR0GbVfIqk", "video", "A2", "listening", "10 min"},
		{"Past Simple Stories — Easy English", "https://www.youtube.com/watch?v=aBq4MJuxI2c", "video", "A2", "grammar", "6 min"},
		{"Present Perfect Explained — BBC Learning", "https://www.youtube.com/watch?v=WjpCNe_JwBs", "video", "A2", "grammar", "5 min"},
		{"Everyday Phrasal Verbs — Rachel's English", "https://www.youtube.com/watch?v=wLgS3t_EXak", "video", "A2", "vocabulary", "8 min"},
		{"How to Talk About Your Weekend", "https://www.youtube.com/watch?v=wBHLJGHxCgQ", "video", "A2", "speaking", "6 min"},
		{"English Weather Vocabulary", "https://www.youtube.com/watch?v=N4TBw9Y1hS0", "video", "A2", "vocabulary", "5 min"},
		{"Telling Time in English", "https://www.youtube.com/watch?v=IBBQXBhSNUs", "video", "A2", "basics", "7 min"},
		{"Describing People in English", "https://www.youtube.com/watch?v=A5fNZnpXBzQ", "video", "A2", "speaking", "6 min"},
		{"Job Interview English — Easy Level", "https://www.youtube.com/watch?v=naIkpQ_cIt0", "video", "A2", "work", "8 min"},
		{"English at the Doctor's Office", "https://www.youtube.com/watch?v=xdDbp6RnUfU", "video", "A2", "health", "6 min"},
		{"Cooking Vocabulary in English", "https://www.youtube.com/watch?v=ZjXwOIFbyoo", "video", "A2", "food", "7 min"},
		{"Giving Directions in English", "https://www.youtube.com/watch?v=bBMmEL5Fzno", "video", "A2", "travel", "5 min"},
		{"English Phone Conversations", "https://www.youtube.com/watch?v=VGMWQeEQPKQ", "video", "A2", "speaking", "6 min"},
		{"Feelings and Emotions Vocabulary", "https://www.youtube.com/watch?v=dEL3xPNpJJE", "video", "A2", "vocabulary", "5 min"},
		{"Weekend Plans — Easy Listening", "https://www.youtube.com/watch?v=CrHMjC99jas", "video", "A2", "listening", "8 min"},
	}

	for _, m := range media {
		_, err := d.Pool.Exec(ctx,
			`INSERT INTO media_resources (title, url, media_type, level, topic, duration, language)
			 VALUES ($1, $2, $3, $4, $5, $6, 'en')
			 ON CONFLICT (url) DO NOTHING`,
			m.title, m.url, m.mediaType, m.level, m.topic, m.duration)
		if err != nil {
			log.Printf("Seed EN media '%s' warning: %v", m.title, err)
		}
	}
}

// seedMediaDE is a no-op — German media is populated by running:
//
//	go run cmd/seed-media/main.go
//
// The seed-media script uses YouTube Data API to find real videos
// with >50K views, subtitles, and proper duration filtering.
func (d *DB) seedMediaDE(ctx context.Context) {}
