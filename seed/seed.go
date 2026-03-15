package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run seeds all initial data (idempotent — ON CONFLICT DO NOTHING).
func Run(pool *pgxpool.Pool) {
	ctx := context.Background()
	seedGrammarWeeksEN(pool, ctx)
	seedGrammarWeeksDE(pool, ctx)
	seedWordsEN(pool, ctx)
	seedWordsDE(pool, ctx)
	seedMediaEN(pool, ctx)
	seedMediaDE(pool, ctx)

	log.Println("Seeding completed")
}

// ==================== ENGLISH ====================

func seedGrammarWeeksEN(pool *pgxpool.Pool, ctx context.Context) {
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
		_, err := pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'en')
			 ON CONFLICT (week_num, language) DO NOTHING`,
			w.num, w.family, w.focus, w.tenseName, w.anchor, w.markers, w.formula, w.ex)
		if err != nil {
			log.Printf("Seed EN grammar week %d warning: %v", w.num, err)
		}
	}
}

func seedGrammarWeeksDE(pool *pgxpool.Pool, ctx context.Context) {
	weeks := []struct {
		num                                                    int
		family, focus, tenseName, anchor, markers, formula, ex string
	}{
		{1, "Einfach", "Präsens", "Präsens",
			"\U0001F504 Karussell — wiederholt sich immer wieder",
			"immer, oft, jeden Tag, manchmal, normalerweise",
			"S + Verb (ich -e, du -st, er/sie -t, wir -en)",
			"Ich lerne jeden Tag Deutsch."},
		{2, "Einfach", "Präteritum", "Präteritum",
			"\U0001F6AA Geschlossene Tür — abgeschlossen und vorbei",
			"gestern, letztes Jahr, damals, vor einer Woche, früher",
			"S + Verb (Präteritum-Stamm + Endung)",
			"Ich ging gestern ins Kino."},
		{3, "Zusammengesetzt", "Perfekt", "Perfekt",
			"\U0001F309 Brücke — Vergangenheit mit Bezug zur Gegenwart",
			"schon, bereits, gerade, noch nicht, heute",
			"S + haben/sein + ... + Partizip II",
			"Ich habe das Buch gelesen."},
		{4, "Einfach", "Futur I", "Futur I",
			"\U0001F52E Kristallkugel — Pläne und Vermutungen",
			"morgen, nächste Woche, bald, in Zukunft, wahrscheinlich",
			"S + werden + ... + Infinitiv",
			"Ich werde morgen nach Berlin fahren."},
		{5, "Konjunktiv", "Konjunktiv II", "Konjunktiv II",
			"\U0001F4AD Gedankenblase — Wünsche und irreale Situationen",
			"wenn, hätte, würde, könnte, gern",
			"S + würde + ... + Infinitiv / hätte / wäre",
			"Wenn ich reich wäre, würde ich reisen."},
		{6, "Zusammengesetzt", "Plusquamperfekt", "Plusquamperfekt",
			"\u23EA Rückspulen — vor einer anderen Vergangenheit",
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
		_, err := pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'de')
			 ON CONFLICT (week_num, language) DO NOTHING`,
			w.num, w.family, w.focus, w.tenseName, w.anchor, w.markers, w.formula, w.ex)
		if err != nil {
			log.Printf("Seed DE grammar week %d warning: %v", w.num, err)
		}
	}
}

func seedWordsEN(pool *pgxpool.Pool, ctx context.Context) {
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
		_, err := pool.Exec(ctx,
			`INSERT INTO words (word, definition, example, level, collocations, construction, language)
			 VALUES ($1, $2, $3, $4, $5, $6, 'en')
			 ON CONFLICT (word, language) DO NOTHING`,
			w.word, w.definition, w.example, w.level, w.collocations, w.construction)
		if err != nil {
			log.Printf("Seed EN word '%s' warning: %v", w.word, err)
		}
	}
}

func seedWordsDE(pool *pgxpool.Pool, ctx context.Context) {
	words := []struct {
		word, definition, example, level, collocations, construction string
	}{
		{"verstehen", "понять, понимать", "Ich verstehe die Frage nicht.", "A2",
			"gut verstehen, richtig verstehen, falsch verstehen", "verstehen + Akk"},
		{"brauchen", "нуждаться", "Ich brauche deine Hilfe.", "A2",
			"dringend brauchen, Zeit brauchen, Hilfe brauchen", "brauchen + Akk"},
		{"versuchen", "пытаться", "Ich versuche, pünktlich zu kommen.", "A2",
			"es versuchen, versuchen zu verstehen", "versuchen + zu + Inf"},
		{"erklären", "объяснять", "Können Sie das bitte erklären?", "A2",
			"genau erklären, einfach erklären", "erklären + Akk + Dat"},
		{"empfehlen", "рекомендовать", "Ich empfehle dir dieses Restaurant.", "A2",
			"sehr empfehlen, weiterempfehlen", "empfehlen + Dat + Akk"},
		{"sich freuen", "радоваться", "Ich freue mich auf das Wochenende.", "A2",
			"sich freuen auf, sich freuen über", "sich freuen auf + Akk / über + Akk"},
		{"sich entscheiden", "решиться", "Er hat sich für das rote Auto entschieden.", "A2",
			"sich schnell entscheiden, sich entscheiden für", "sich entscheiden für + Akk"},
		{"anfangen", "начинать", "Der Film fängt um 8 Uhr an.", "A2",
			"anfangen mit, anfangen zu arbeiten", "anfangen mit + Dat / zu + Inf"},
		{"aufhören", "прекращать", "Hör bitte auf zu reden.", "A2",
			"aufhören mit, aufhören zu rauchen", "aufhören mit + Dat / zu + Inf"},
		{"sich gewöhnen", "привыкнуть", "Ich gewöhne mich an das kalte Wetter.", "A2",
			"sich gewöhnen an, sich daran gewöhnen", "sich gewöhnen an + Akk"},
		{"vorschlagen", "предлагать", "Ich schlage vor, ins Kino zu gehen.", "A2",
			"einen Vorschlag machen, vorschlagen zu", "vorschlagen + zu + Inf / Dat + Akk"},
		{"sich unterhalten", "беседовать", "Wir haben uns über Politik unterhalten.", "A2",
			"sich unterhalten über, sich gut unterhalten", "sich unterhalten über + Akk / mit + Dat"},
		{"sich beschweren", "жаловаться", "Er beschwert sich über den Lärm.", "A2",
			"sich beschweren über, sich bei jemandem beschweren", "sich beschweren über + Akk / bei + Dat"},
		{"sich kümmern", "заботиться", "Sie kümmert sich um ihre Eltern.", "A2",
			"sich kümmern um, darum kümmern", "sich kümmern um + Akk"},
		{"abhängen", "зависеть", "Das hängt vom Wetter ab.", "A2",
			"abhängen von, davon abhängen", "abhängen von + Dat"},
		{"sich vorbereiten", "готовиться", "Ich bereite mich auf die Prüfung vor.", "A2",
			"sich vorbereiten auf, gut vorbereiten", "sich vorbereiten auf + Akk"},
		{"bestehen", "сдать (экзамен)", "Sie hat die Prüfung bestanden.", "A2",
			"eine Prüfung bestehen, bestehen aus", "bestehen + Akk / bestehen aus + Dat"},
		{"verbessern", "улучшить", "Ich möchte mein Deutsch verbessern.", "A2",
			"sich verbessern, Leistung verbessern", "verbessern + Akk / sich verbessern"},
		{"bemerken", "заметить", "Ich habe den Fehler nicht bemerkt.", "A2",
			"sofort bemerken, kaum bemerken", "bemerken + Akk / dass"},
		{"sich verändern", "измениться", "Die Stadt hat sich sehr verändert.", "A2",
			"sich stark verändern, sich kaum verändern", "sich verändern"},
		{"trotzdem", "тем не менее", "Es regnete, trotzdem ging er spazieren.", "A2",
			"trotzdem machen, trotzdem kommen", "trotzdem + Hauptsatz"},
		{"eigentlich", "на самом деле", "Eigentlich wollte ich zu Hause bleiben.", "A2",
			"eigentlich nicht, eigentlich schon", "eigentlich + Verb"},
		{"wahrscheinlich", "вероятно", "Er kommt wahrscheinlich morgen.", "A2",
			"sehr wahrscheinlich, wahrscheinlich nicht", "wahrscheinlich + Verb"},
		{"unbedingt", "обязательно", "Du musst unbedingt diesen Film sehen.", "A2",
			"unbedingt brauchen, nicht unbedingt", "unbedingt + Verb"},
		{"außerdem", "кроме того", "Außerdem habe ich keine Zeit.", "A2",
			"außerdem noch, außerdem auch", "außerdem + Hauptsatz"},
		{"deshalb", "поэтому", "Ich war müde, deshalb bin ich früh ins Bett gegangen.", "A2",
			"genau deshalb, deshalb auch", "deshalb + Verb (Position 1)"},
		{"obwohl", "хотя", "Obwohl es kalt war, ging sie ohne Jacke.", "A2",
			"obwohl es schwer ist, obwohl ich weiß", "obwohl + Nebensatz (Verb am Ende)"},
		{"falls", "в случае если", "Falls du Fragen hast, ruf mich an.", "A2",
			"falls nötig, falls möglich", "falls + Nebensatz (Verb am Ende)"},
		{"stattdessen", "вместо этого", "Ich blieb zu Hause. Stattdessen habe ich gelesen.", "A2",
			"stattdessen lieber, stattdessen machen", "stattdessen + Hauptsatz"},
		{"inzwischen", "тем временем", "Ich koche. Inzwischen kannst du den Tisch decken.", "A2",
			"inzwischen schon, inzwischen fertig", "inzwischen + Hauptsatz"},
		{"sich lohnen", "стоить того", "Es lohnt sich, Deutsch zu lernen.", "A2",
			"sich lohnen zu, es lohnt sich", "sich lohnen + zu + Inf"},
		{"schaffen", "справиться", "Ich habe es geschafft!", "A2",
			"es schaffen, rechtzeitig schaffen", "schaffen + Akk / es schaffen"},
		{"sich vorstellen", "представить себе", "Ich kann mir das nicht vorstellen.", "A2",
			"sich vorstellen können, sich etwas vorstellen", "sich (Dat) vorstellen + Akk"},
		{"vermissen", "скучать", "Ich vermisse meine Familie.", "A2",
			"jemanden vermissen, sehr vermissen", "vermissen + Akk"},
		{"sich entschuldigen", "извиниться", "Ich entschuldige mich für den Fehler.", "A2",
			"sich entschuldigen für, sich bei jemandem entschuldigen", "sich entschuldigen für + Akk / bei + Dat"},
		{"sich erinnern", "вспоминать", "Ich erinnere mich an den Urlaub.", "A2",
			"sich erinnern an, sich gut erinnern", "sich erinnern an + Akk"},
		{"sich interessieren", "интересоваться", "Ich interessiere mich für Kunst.", "A2",
			"sich interessieren für, sich sehr interessieren", "sich interessieren für + Akk"},
		{"sich beschäftigen", "заниматься", "Ich beschäftige mich mit dem Projekt.", "A2",
			"sich beschäftigen mit, sich intensiv beschäftigen", "sich beschäftigen mit + Dat"},
		{"gemütlich", "уютный", "Das Café ist sehr gemütlich.", "A2",
			"gemütliche Atmosphäre, gemütlicher Abend", "gemütlich + Nomen"},
		{"selbstverständlich", "само собой разумеется", "Selbstverständlich helfe ich dir.", "A2",
			"selbstverständlich machen, selbstverständlich sein", "selbstverständlich + Verb"},
		{"überraschend", "удивительно", "Das Ergebnis war überraschend.", "A2",
			"überraschend gut, überraschend schnell", "überraschend + Adj / Nomen"},
		{"geduldig", "терпеливый", "Sie ist sehr geduldig mit Kindern.", "A2",
			"geduldig sein, geduldig warten", "geduldig + mit + Dat"},
		{"zuverlässig", "надёжный", "Er ist ein zuverlässiger Freund.", "A2",
			"zuverlässig sein, zuverlässige Quelle", "zuverlässig + Nomen"},
		{"offensichtlich", "очевидно", "Offensichtlich hat er recht.", "A2",
			"offensichtlich falsch, offensichtlich richtig", "offensichtlich + Verb"},
		{"gründlich", "тщательно", "Du musst gründlich putzen.", "A2",
			"gründlich prüfen, gründlich lesen", "gründlich + Verb"},
		{"allmählich", "постепенно", "Allmählich wird es besser.", "A2",
			"allmählich verstehen, allmählich lernen", "allmählich + Verb"},
		{"ausgezeichnet", "отлично", "Das Essen war ausgezeichnet.", "A2",
			"ausgezeichnete Qualität, ausgezeichnet schmecken", "ausgezeichnet + Nomen"},
		{"angenehm", "приятный", "Es war ein angenehmer Abend.", "A2",
			"angenehme Atmosphäre, angenehm überrascht", "angenehm + Nomen"},
		{"gleichzeitig", "одновременно", "Er kann gleichzeitig lesen und essen.", "A2",
			"gleichzeitig machen, gleichzeitig passieren", "gleichzeitig + Verb"},
	}

	for _, w := range words {
		_, err := pool.Exec(ctx,
			`INSERT INTO words (word, definition, example, level, collocations, construction, language)
			 VALUES ($1, $2, $3, $4, $5, $6, 'de')
			 ON CONFLICT (word, language) DO NOTHING`,
			w.word, w.definition, w.example, w.level, w.collocations, w.construction)
		if err != nil {
			log.Printf("Seed DE word '%s' warning: %v", w.word, err)
		}
	}
}

func seedMediaEN(pool *pgxpool.Pool, ctx context.Context) {
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
		_, err := pool.Exec(ctx,
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
func seedMediaDE(pool *pgxpool.Pool, ctx context.Context) {}
