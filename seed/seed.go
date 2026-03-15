package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(pool *pgxpool.Pool) {
	ctx := context.Background()
	seedGrammarWeeksEN(pool, ctx)
	seedGrammarWeeksDE(pool, ctx)
	seedWordsEN(pool, ctx)
	seedWordsDE(pool, ctx)
	seedMediaEN(pool, ctx)
	log.Println("Seeding completed")
}

func seedGrammarWeeksEN(pool *pgxpool.Pool, ctx context.Context) {
	weeks := []struct {
		num                                                    int
		family, focus, tenseName, anchor, markers, formula, ex string
	}{
		{1, "Simple", "Past Simple", "Past Simple",
			"\U0001F6AA Closed door \u2014 done and finished",
			"yesterday, last week, ago, in 2020",
			"S + V2 (ed / irregular)",
			"I watched a movie yesterday."},
		{2, "Simple", "Present Simple", "Present Simple",
			"\U0001F504 Carousel \u2014 repeats again and again",
			"always, usually, every day, sometimes",
			"S + V1 (he/she +s)",
			"I usually wake up at 7."},
		{3, "Simple", "Future Simple", "Future Simple",
			"\U0001F52E Crystal ball \u2014 decision right now",
			"tomorrow, next week, I think, probably",
			"S + will + V1",
			"I will call you tomorrow."},
		{4, "Continuous", "Present Continuous", "Present Continuous",
			"\U0001F4F8 Photo \u2014 happening right now",
			"now, right now, at the moment, look!",
			"S + am/is/are + Ving",
			"I am reading a book right now."},
		{5, "Continuous", "Past Continuous", "Past Continuous",
			"\U0001F3AC Movie scene \u2014 background action in the past",
			"while, when, at that moment, all day yesterday",
			"S + was/were + Ving",
			"I was cooking when you called."},
		{6, "Perfect", "Present Perfect", "Present Perfect",
			"\U0001F309 Bridge \u2014 from past to present, result matters",
			"already, yet, just, ever, never, since, for",
			"S + have/has + V3",
			"I have already finished my homework."},
		{7, "Perfect", "Past Perfect", "Past Perfect",
			"\u23EA Rewind \u2014 action BEFORE another past action",
			"before, after, by the time, already (past context)",
			"S + had + V3",
			"I had eaten before she arrived."},
	}
	for _, w := range weeks {
		pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'en')
			 ON CONFLICT (week_num, language) DO NOTHING`,
			w.num, w.family, w.focus, w.tenseName, w.anchor, w.markers, w.formula, w.ex)
	}
}

func seedGrammarWeeksDE(pool *pgxpool.Pool, ctx context.Context) {
	weeks := []struct {
		num                                                    int
		family, focus, tenseName, anchor, markers, formula, ex string
	}{
		{1, "Einfach", "Pr\u00e4sens", "Pr\u00e4sens",
			"\U0001F504 Karussell \u2014 wiederholt sich immer wieder",
			"immer, oft, jeden Tag, manchmal, normalerweise",
			"S + Verb (ich -e, du -st, er/sie -t, wir -en)",
			"Ich lerne jeden Tag Deutsch."},
		{2, "Einfach", "Pr\u00e4teritum", "Pr\u00e4teritum",
			"\U0001F6AA Geschlossene T\u00fcr \u2014 abgeschlossen und vorbei",
			"gestern, letztes Jahr, damals, vor einer Woche, fr\u00fcher",
			"S + Verb (Pr\u00e4teritum-Stamm + Endung)",
			"Ich ging gestern ins Kino."},
		{3, "Zusammengesetzt", "Perfekt", "Perfekt",
			"\U0001F309 Br\u00fccke \u2014 Vergangenheit mit Bezug zur Gegenwart",
			"schon, bereits, gerade, noch nicht, heute",
			"S + haben/sein + ... + Partizip II",
			"Ich habe das Buch gelesen."},
		{4, "Einfach", "Futur I", "Futur I",
			"\U0001F52E Kristallkugel \u2014 Pl\u00e4ne und Vermutungen",
			"morgen, n\u00e4chste Woche, bald, in Zukunft, wahrscheinlich",
			"S + werden + ... + Infinitiv",
			"Ich werde morgen nach Berlin fahren."},
		{5, "Konjunktiv", "Konjunktiv II", "Konjunktiv II",
			"\U0001F4AD Gedankenblase \u2014 W\u00fcnsche und irreale Situationen",
			"wenn, h\u00e4tte, w\u00fcrde, k\u00f6nnte, gern",
			"S + w\u00fcrde + ... + Infinitiv / h\u00e4tte / w\u00e4re",
			"Wenn ich reich w\u00e4re, w\u00fcrde ich reisen."},
		{6, "Zusammengesetzt", "Plusquamperfekt", "Plusquamperfekt",
			"\u23EA R\u00fcckspulen \u2014 vor einer anderen Vergangenheit",
			"bevor, nachdem, bereits, schon (Vergangenheit)",
			"S + hatte/war + ... + Partizip II",
			"Ich hatte gegessen, bevor sie kam."},
		{7, "Passiv", "Passiv", "Passiv",
			"\U0001F3AD Maske \u2014 wer es tut ist unwichtig",
			"von, durch, es wird, wurde",
			"S + werden + ... + Partizip II",
			"Das Haus wird renoviert."},
	}
	for _, w := range weeks {
		pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'de')
			 ON CONFLICT (week_num, language) DO NOTHING`,
			w.num, w.family, w.focus, w.tenseName, w.anchor, w.markers, w.formula, w.ex)
	}
}

type wordSeed struct {
	word, def, example, level, collocations, construction string
}

func insertWords(pool *pgxpool.Pool, ctx context.Context, lang string, words []wordSeed) {
	for _, w := range words {
		_, err := pool.Exec(ctx,
			`INSERT INTO words (word, definition, example, level, collocations, construction, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (word, language) DO NOTHING`,
			w.word, w.def, w.example, w.level, w.collocations, w.construction, lang)
		if err != nil {
			log.Printf("Seed %s word '%s': %v", lang, w.word, err)
		}
	}
}

func seedWordsEN(pool *pgxpool.Pool, ctx context.Context) {
	// ==================== A1 ====================
	insertWords(pool, ctx, "en", []wordSeed{
		{"hello", "привет", "Hello! How are you?", "A1", "say hello, hello there", "hello + name"},
		{"thank you", "спасибо", "Thank you for your help.", "A1", "thank you very much, thank you so much", "thank you + for"},
		{"please", "пожалуйста", "Can I have some water, please?", "A1", "yes please, please help", "please + verb"},
		{"sorry", "извините", "Sorry, I don't understand.", "A1", "I'm sorry, sorry about that", "sorry + for"},
		{"good", "хороший", "This is a good book.", "A1", "good morning, good idea, good job", "good + noun"},
		{"bad", "плохой", "The weather is bad today.", "A1", "bad weather, bad idea, not bad", "bad + noun"},
		{"big", "большой", "This is a big house.", "A1", "big city, big problem, big family", "big + noun"},
		{"small", "маленький", "I have a small dog.", "A1", "small town, small size, small room", "small + noun"},
		{"go", "идти", "I go to work every day.", "A1", "go home, go to school, go shopping", "go + to + place"},
		{"come", "приходить", "Come here, please!", "A1", "come home, come back, come in", "come + direction"},
		{"eat", "есть", "I eat breakfast at 8.", "A1", "eat lunch, eat dinner, eat food", "eat + meal/food"},
		{"drink", "пить", "I drink coffee in the morning.", "A1", "drink water, drink tea, drink juice", "drink + liquid"},
		{"sleep", "спать", "I sleep 8 hours every night.", "A1", "go to sleep, sleep well, can't sleep", "sleep + time"},
		{"work", "работать", "I work in an office.", "A1", "work hard, work from home, go to work", "work + in/at + place"},
		{"like", "нравиться", "I like pizza.", "A1", "I like it, like very much, would like", "like + noun / Ving"},
		{"want", "хотеть", "I want to learn English.", "A1", "want to go, want to eat, I want", "want + to + V1"},
		{"have", "иметь", "I have two brothers.", "A1", "have a car, have time, have fun", "have + noun"},
		{"need", "нуждаться", "I need your help.", "A1", "need help, need time, need to go", "need + noun / to + V1"},
		{"can", "мочь", "I can swim.", "A1", "can do, can speak, can help", "can + V1"},
		{"know", "знать", "I know this word.", "A1", "I know, know how to, you know", "know + noun / how to"},
		{"understand", "понимать", "I don't understand this question.", "A1", "understand English, easy to understand", "understand + noun"},
		{"help", "помогать", "Can you help me?", "A1", "help someone, need help, help with", "help + person + with"},
		{"learn", "учить", "I learn new words every day.", "A1", "learn English, learn to swim, learn from", "learn + noun / to + V1"},
		{"read", "читать", "I read books before sleep.", "A1", "read a book, read news, read aloud", "read + noun"},
		{"write", "писать", "Please write your name.", "A1", "write a letter, write down, write to", "write + noun"},
		{"speak", "говорить", "Do you speak English?", "A1", "speak English, speak slowly, speak to", "speak + language / to"},
		{"listen", "слушать", "Listen to this song.", "A1", "listen to music, listen carefully", "listen + to + noun"},
		{"see", "видеть", "I can see the mountains.", "A1", "see you, nice to see, let me see", "see + noun"},
		{"house", "дом", "I live in a small house.", "A1", "my house, at home, house number", "in/at + house"},
		{"family", "семья", "My family is big.", "A1", "my family, family name, family member", "family + noun"},
		{"friend", "друг", "She is my best friend.", "A1", "best friend, old friend, make friends", "friend + of"},
		{"time", "время", "What time is it?", "A1", "free time, a long time, on time", "time + for"},
		{"day", "день", "Today is a beautiful day.", "A1", "every day, good day, bad day", "day + of the week"},
		{"water", "вода", "I drink water every morning.", "A1", "cold water, hot water, glass of water", "water + adjective"},
		{"food", "еда", "The food is delicious.", "A1", "fast food, food and drink, favourite food", "food + adjective"},
		{"money", "деньги", "I don't have much money.", "A1", "save money, spend money, a lot of money", "money + for"},
		{"happy", "счастливый", "I am happy today.", "A1", "feel happy, be happy, happy birthday", "happy + about/with"},
		{"tired", "усталый", "I am very tired after work.", "A1", "feel tired, get tired, tired of", "tired + of"},
	})

	// ==================== A2 ====================
	insertWords(pool, ctx, "en", []wordSeed{
		{"figure out", "понять, разобраться", "I finally figured out how to use this app.", "A2", "figure out a problem, figure out how", "figure out + how/what/why"},
		{"give up", "сдаться, бросить", "Don't give up! You can do it.", "A2", "give up hope, give up trying, give up smoking", "give up + Ving / noun"},
		{"look forward to", "ждать с нетерпением", "I'm looking forward to the weekend.", "A2", "look forward to meeting, look forward to seeing", "look forward to + Ving / noun"},
		{"turn out", "оказаться", "It turned out to be a great movie.", "A2", "turn out to be, turn out well", "turn out + to be / that"},
		{"come up with", "придумать", "She came up with a brilliant idea.", "A2", "come up with an idea, come up with a plan", "come up with + noun"},
		{"run out of", "закончиться", "We ran out of milk.", "A2", "run out of time, run out of money", "run out of + noun"},
		{"get along with", "ладить с", "I get along with my coworkers.", "A2", "get along with people, get along well", "get along with + person"},
		{"pick up", "подобрать, выучить", "I picked up some new words from the movie.", "A2", "pick up a language, pick up the phone", "pick up + noun"},
		{"put off", "откладывать", "Stop putting off your homework!", "A2", "put off a meeting, put off doing something", "put off + Ving / noun"},
		{"carry on", "продолжать", "Carry on with your work.", "A2", "carry on working, carry on with something", "carry on + Ving / with"},
		{"find out", "узнать", "I found out the truth yesterday.", "A2", "find out the truth, find out about", "find out + about / that"},
		{"set up", "организовать, настроить", "Let's set up a meeting for Monday.", "A2", "set up a business, set up a meeting", "set up + noun"},
		{"deal with", "справляться с", "I need to deal with this problem.", "A2", "deal with a problem, deal with stress", "deal with + noun"},
		{"end up", "в итоге оказаться", "We ended up staying home.", "A2", "end up doing, end up in a place", "end up + Ving / in"},
		{"show up", "появиться", "He didn't show up to the meeting.", "A2", "show up late, show up on time", "show up + at/to"},
		{"although", "хотя", "Although it was raining, we went for a walk.", "A2", "although it seems, although I know", "although + clause"},
		{"actually", "на самом деле", "I actually enjoyed the movie.", "A2", "actually quite, actually think", "actually + verb"},
		{"definitely", "определённо", "I will definitely come to your party.", "A2", "definitely agree, definitely need", "definitely + verb"},
		{"probably", "вероятно", "She will probably be late.", "A2", "probably not, probably should", "probably + verb"},
		{"recently", "недавно", "I recently started learning English.", "A2", "recently discovered, recently started", "recently + Past Simple / Present Perfect"},
		{"especially", "особенно", "I love fruits, especially mangoes.", "A2", "especially when, especially important", "especially + noun / when"},
		{"instead", "вместо этого", "I didn't go out. Instead, I stayed home.", "A2", "instead of doing, instead of that", "instead + clause / instead of + Ving"},
		{"however", "однако", "The test was hard. However, I passed it.", "A2", "however much, however difficult", "however + clause"},
		{"manage to", "суметь", "I managed to finish the project on time.", "A2", "manage to do, manage to find", "manage to + V1"},
		{"used to", "раньше (привычка в прошлом)", "I used to play football every day.", "A2", "used to live, used to be, used to think", "used to + V1"},
		{"improve", "улучшить", "I want to improve my English.", "A2", "improve skills, improve performance", "improve + noun"},
		{"appreciate", "ценить", "I really appreciate your help.", "A2", "appreciate help, appreciate the effort", "appreciate + noun / Ving"},
		{"avoid", "избегать", "Try to avoid making the same mistake.", "A2", "avoid doing, avoid mistakes", "avoid + Ving / noun"},
		{"recommend", "рекомендовать", "I recommend watching this movie.", "A2", "recommend doing, highly recommend", "recommend + Ving / noun"},
		{"suggest", "предложить", "I suggest taking a break.", "A2", "suggest doing, suggest that", "suggest + Ving / that"},
		{"depend on", "зависеть от", "It depends on the weather.", "A2", "depend on someone, depend on the situation", "depend on + noun"},
		{"essential", "необходимый", "Sleep is essential for health.", "A2", "essential for, essential part", "essential + for"},
		{"convenient", "удобный", "This time is convenient for me.", "A2", "convenient time, convenient location", "convenient + for"},
		{"ordinary", "обычный", "It was just an ordinary day.", "A2", "ordinary people, ordinary life", "ordinary + noun"},
		{"entire", "весь, целый", "I spent the entire day reading.", "A2", "entire day, entire life, entire team", "entire + noun"},
		{"afford", "позволить себе", "I can't afford a new phone.", "A2", "afford to buy, can't afford", "can/can't afford + to V1"},
		{"obvious", "очевидный", "The answer was obvious.", "A2", "obvious reason, obvious choice", "obvious + that / noun"},
		{"meanwhile", "тем временем", "I cooked dinner. Meanwhile, she set the table.", "A2", "meanwhile in, meanwhile back", "meanwhile + clause"},
		{"be about to", "вот-вот, собираться", "The movie is about to start.", "A2", "about to leave, about to start", "be about to + V1"},
	})

	// ==================== B1 ====================
	insertWords(pool, ctx, "en", []wordSeed{
		{"come across", "случайно найти", "I came across an interesting article.", "B1", "come across a problem, come across as", "come across + noun"},
		{"get over", "пережить, преодолеть", "It took her a long time to get over the breakup.", "B1", "get over it, get over a cold", "get over + noun"},
		{"put up with", "терпеть, мириться", "I can't put up with this noise.", "B1", "put up with someone, put up with behaviour", "put up with + noun/Ving"},
		{"stand out", "выделяться", "Her essay really stood out from the rest.", "B1", "stand out from the crowd, stand out as", "stand out + from/as"},
		{"bring about", "вызвать, привести к", "Technology has brought about many changes.", "B1", "bring about change, bring about results", "bring about + noun"},
		{"take for granted", "принимать как должное", "Don't take your health for granted.", "B1", "take it for granted, take someone for granted", "take + noun + for granted"},
		{"nevertheless", "тем не менее", "It was raining; nevertheless, we went hiking.", "B1", "but nevertheless, nevertheless important", "nevertheless + clause"},
		{"furthermore", "более того", "The hotel was cheap. Furthermore, it was clean.", "B1", "furthermore it is, furthermore we", "furthermore + clause"},
		{"consequently", "следовательно", "He didn't study. Consequently, he failed the exam.", "B1", "and consequently, consequently the", "consequently + clause"},
		{"whereas", "тогда как", "I like tea, whereas my sister prefers coffee.", "B1", "whereas in fact, whereas the other", "clause + whereas + clause"},
		{"provided that", "при условии что", "You can go, provided that you finish your work.", "B1", "provided that you, provided that the", "provided that + clause"},
		{"contribute", "вносить вклад", "Everyone should contribute to the project.", "B1", "contribute to, contribute ideas", "contribute + to + noun"},
		{"influence", "влиять", "Music can influence your mood.", "B1", "influence on, influence decision", "influence + noun / on"},
		{"significant", "значительный", "There was a significant increase in sales.", "B1", "significant change, significant difference", "significant + noun"},
		{"establish", "установить, основать", "The company was established in 1990.", "B1", "establish a business, establish rules", "establish + noun"},
		{"on the other hand", "с другой стороны", "It's expensive. On the other hand, it's high quality.", "B1", "on the other hand however", "on the other hand + clause"},
		{"as far as I know", "насколько я знаю", "As far as I know, the meeting is at 3.", "B1", "as far as I know, as far as I can tell", "as far as I know + clause"},
		{"it seems to me", "мне кажется", "It seems to me that this plan won't work.", "B1", "it seems to me that, it seems like", "it seems to me + that"},
		{"in my experience", "по моему опыту", "In my experience, practice is more important than theory.", "B1", "in my experience with, based on my experience", "in my experience + clause"},
		{"turn down", "отклонить, отказать", "She turned down the job offer.", "B1", "turn down an offer, turn down a request", "turn down + noun"},
	})

	// ==================== B2 ====================
	insertWords(pool, ctx, "en", []wordSeed{
		{"acknowledge", "признавать", "He acknowledged his mistake publicly.", "B2", "acknowledge a problem, acknowledge receipt", "acknowledge + noun / that"},
		{"comprehensive", "всеобъемлющий", "We need a comprehensive review of the project.", "B2", "comprehensive guide, comprehensive analysis", "comprehensive + noun"},
		{"pursue", "преследовать, стремиться к", "She decided to pursue a career in medicine.", "B2", "pursue a goal, pursue a career", "pursue + noun"},
		{"reluctant", "неохотный", "He was reluctant to share his opinion.", "B2", "reluctant to do, somewhat reluctant", "reluctant + to + V1"},
		{"inevitable", "неизбежный", "Change is inevitable in any organization.", "B2", "inevitable consequence, inevitable result", "inevitable + noun / that"},
		{"undermine", "подрывать", "His comments undermined her confidence.", "B2", "undermine trust, undermine authority", "undermine + noun"},
		{"hit the nail on the head", "попасть в точку", "You hit the nail on the head with that comment.", "B2", "really hit the nail, exactly hit the nail", "hit the nail on the head"},
		{"a blessing in disguise", "нет худа без добра", "Losing that job was a blessing in disguise.", "B2", "turned out to be a blessing", "noun + is a blessing in disguise"},
		{"cut corners", "срезать углы, экономить", "We shouldn't cut corners on safety.", "B2", "cut corners on, try to cut corners", "cut corners + on + noun"},
		{"benchmark", "эталон, ориентир", "This product sets a new benchmark in quality.", "B2", "set a benchmark, industry benchmark", "benchmark + for/in"},
		{"leverage", "использовать в своих интересах", "We need to leverage our existing resources.", "B2", "leverage technology, leverage experience", "leverage + noun"},
		{"stakeholder", "заинтересованная сторона", "All stakeholders should be involved in the decision.", "B2", "key stakeholder, stakeholder meeting", "stakeholder + in"},
		{"streamline", "упростить, оптимизировать", "We need to streamline our processes.", "B2", "streamline operations, streamline workflow", "streamline + noun"},
		{"elaborate", "подробно объяснить", "Could you elaborate on that point?", "B2", "elaborate on, elaborate plan", "elaborate + on + noun"},
		{"in a nutshell", "в двух словах", "In a nutshell, the project was a success.", "B2", "put it in a nutshell", "in a nutshell + clause"},
	})
}

func seedWordsDE(pool *pgxpool.Pool, ctx context.Context) {
	// ==================== A1 ====================
	insertWords(pool, ctx, "de", []wordSeed{
		{"hallo", "привет", "Hallo! Wie geht es dir?", "A1", "Hallo sagen, Hallo zusammen", "Hallo + Name"},
		{"danke", "спасибо", "Danke f\u00fcr deine Hilfe.", "A1", "vielen Dank, danke sch\u00f6n", "danke + f\u00fcr"},
		{"bitte", "пожалуйста", "Kann ich bitte etwas Wasser haben?", "A1", "ja bitte, bitte sch\u00f6n", "bitte + Verb"},
		{"gut", "хороший", "Das Essen ist sehr gut.", "A1", "guten Morgen, guten Tag, sehr gut", "gut + Nomen"},
		{"schlecht", "плохой", "Das Wetter ist schlecht.", "A1", "schlechtes Wetter, schlecht gelaunt", "schlecht + Nomen"},
		{"gro\u00df", "большой", "Berlin ist eine gro\u00dfe Stadt.", "A1", "gro\u00dfe Stadt, gro\u00dfe Familie", "gro\u00df + Nomen"},
		{"klein", "маленький", "Ich habe einen kleinen Hund.", "A1", "kleine Stadt, kleines Kind", "klein + Nomen"},
		{"gehen", "идти", "Ich gehe jeden Tag zur Arbeit.", "A1", "nach Hause gehen, spazieren gehen", "gehen + Richtung"},
		{"kommen", "приходить", "Komm bitte her!", "A1", "nach Hause kommen, aus ... kommen", "kommen + Richtung"},
		{"essen", "есть", "Ich esse Fr\u00fchst\u00fcck um 8 Uhr.", "A1", "zu Mittag essen, zu Abend essen", "essen + Akk"},
		{"trinken", "пить", "Ich trinke morgens Kaffee.", "A1", "Wasser trinken, Tee trinken", "trinken + Akk"},
		{"schlafen", "спать", "Ich schlafe 8 Stunden pro Nacht.", "A1", "gut schlafen, schlafen gehen", "schlafen + Zeitangabe"},
		{"arbeiten", "работать", "Ich arbeite in einem B\u00fcro.", "A1", "hart arbeiten, von zu Hause arbeiten", "arbeiten + in/bei"},
		{"lernen", "учить", "Ich lerne jeden Tag Deutsch.", "A1", "Deutsch lernen, f\u00fcr die Pr\u00fcfung lernen", "lernen + Akk"},
		{"sprechen", "говорить", "Sprechen Sie Deutsch?", "A1", "Deutsch sprechen, langsam sprechen", "sprechen + Sprache / mit"},
		{"lesen", "читать", "Ich lese gern B\u00fccher.", "A1", "ein Buch lesen, Zeitung lesen", "lesen + Akk"},
		{"schreiben", "писать", "Bitte schreiben Sie Ihren Namen.", "A1", "einen Brief schreiben, aufschreiben", "schreiben + Akk"},
		{"m\u00f6gen", "любить, нравиться", "Ich mag Pizza.", "A1", "gern m\u00f6gen, ich mag", "m\u00f6gen + Akk"},
		{"wollen", "хотеть", "Ich will Deutsch lernen.", "A1", "ich will, wollen wir", "wollen + Inf"},
		{"k\u00f6nnen", "мочь", "Ich kann schwimmen.", "A1", "kann ich, k\u00f6nnen Sie", "k\u00f6nnen + Inf"},
		{"wissen", "знать", "Ich wei\u00df das nicht.", "A1", "ich wei\u00df, Bescheid wissen", "wissen + Akk / dass"},
		{"verstehen", "понимать", "Ich verstehe die Frage nicht.", "A1", "gut verstehen, richtig verstehen", "verstehen + Akk"},
		{"helfen", "помогать", "K\u00f6nnen Sie mir helfen?", "A1", "jemandem helfen, gern helfen", "helfen + Dat"},
		{"brauchen", "нуждаться", "Ich brauche deine Hilfe.", "A1", "dringend brauchen, Hilfe brauchen", "brauchen + Akk"},
		{"Haus", "дом", "Ich wohne in einem kleinen Haus.", "A1", "zu Hause, nach Hause", "in/zu + Haus"},
		{"Familie", "семья", "Meine Familie ist gro\u00df.", "A1", "meine Familie, Familienmitglied", "Familie + Nomen"},
		{"Freund", "друг", "Er ist mein bester Freund.", "A1", "bester Freund, guter Freund", "Freund + von"},
		{"Zeit", "время", "Ich habe keine Zeit.", "A1", "keine Zeit, freie Zeit, viel Zeit", "Zeit + f\u00fcr"},
		{"Tag", "день", "Heute ist ein sch\u00f6ner Tag.", "A1", "guten Tag, jeden Tag, sch\u00f6ner Tag", "Tag + Wochentag"},
		{"Wasser", "вода", "Ich trinke jeden Morgen Wasser.", "A1", "kaltes Wasser, hei\u00dfes Wasser", "Wasser + Adjektiv"},
		{"Geld", "деньги", "Ich habe nicht viel Geld.", "A1", "Geld sparen, Geld ausgeben", "Geld + f\u00fcr"},
		{"gl\u00fccklich", "счастливый", "Ich bin heute sehr gl\u00fccklich.", "A1", "gl\u00fccklich sein, gl\u00fccklich machen", "gl\u00fccklich + \u00fcber"},
		{"m\u00fcde", "усталый", "Ich bin sehr m\u00fcde nach der Arbeit.", "A1", "m\u00fcde sein, m\u00fcde werden", "m\u00fcde + von"},
	})

	// ==================== A2 ====================
	insertWords(pool, ctx, "de", []wordSeed{
		{"sich freuen", "радоваться", "Ich freue mich auf das Wochenende.", "A2", "sich freuen auf, sich freuen \u00fcber", "sich freuen auf + Akk / \u00fcber + Akk"},
		{"sich entscheiden", "решиться", "Er hat sich f\u00fcr das rote Auto entschieden.", "A2", "sich schnell entscheiden, sich entscheiden f\u00fcr", "sich entscheiden f\u00fcr + Akk"},
		{"anfangen", "начинать", "Der Film f\u00e4ngt um 8 Uhr an.", "A2", "anfangen mit, anfangen zu arbeiten", "anfangen mit + Dat / zu + Inf"},
		{"aufh\u00f6ren", "прекращать", "H\u00f6r bitte auf zu reden.", "A2", "aufh\u00f6ren mit, aufh\u00f6ren zu rauchen", "aufh\u00f6ren mit + Dat / zu + Inf"},
		{"sich gew\u00f6hnen", "привыкнуть", "Ich gew\u00f6hne mich an das kalte Wetter.", "A2", "sich gew\u00f6hnen an, sich daran gew\u00f6hnen", "sich gew\u00f6hnen an + Akk"},
		{"vorschlagen", "предлагать", "Ich schlage vor, ins Kino zu gehen.", "A2", "einen Vorschlag machen, vorschlagen zu", "vorschlagen + zu + Inf"},
		{"sich k\u00fcmmern", "заботиться", "Sie k\u00fcmmert sich um ihre Eltern.", "A2", "sich k\u00fcmmern um, darum k\u00fcmmern", "sich k\u00fcmmern um + Akk"},
		{"abh\u00e4ngen", "зависеть", "Das h\u00e4ngt vom Wetter ab.", "A2", "abh\u00e4ngen von, davon abh\u00e4ngen", "abh\u00e4ngen von + Dat"},
		{"sich vorbereiten", "готовиться", "Ich bereite mich auf die Pr\u00fcfung vor.", "A2", "sich vorbereiten auf, gut vorbereiten", "sich vorbereiten auf + Akk"},
		{"verbessern", "улучшить", "Ich m\u00f6chte mein Deutsch verbessern.", "A2", "sich verbessern, Leistung verbessern", "verbessern + Akk"},
		{"bemerken", "заметить", "Ich habe den Fehler nicht bemerkt.", "A2", "sofort bemerken, kaum bemerken", "bemerken + Akk / dass"},
		{"empfehlen", "рекомендовать", "Ich empfehle dir dieses Restaurant.", "A2", "sehr empfehlen, weiterempfehlen", "empfehlen + Dat + Akk"},
		{"sich erinnern", "вспоминать", "Ich erinnere mich an den Urlaub.", "A2", "sich erinnern an, sich gut erinnern", "sich erinnern an + Akk"},
		{"sich interessieren", "интересоваться", "Ich interessiere mich f\u00fcr Kunst.", "A2", "sich interessieren f\u00fcr, sich sehr interessieren", "sich interessieren f\u00fcr + Akk"},
		{"trotzdem", "тем не менее", "Es regnete, trotzdem ging er spazieren.", "A2", "trotzdem machen, trotzdem kommen", "trotzdem + Hauptsatz"},
		{"eigentlich", "на самом деле", "Eigentlich wollte ich zu Hause bleiben.", "A2", "eigentlich nicht, eigentlich schon", "eigentlich + Verb"},
		{"wahrscheinlich", "вероятно", "Er kommt wahrscheinlich morgen.", "A2", "sehr wahrscheinlich, wahrscheinlich nicht", "wahrscheinlich + Verb"},
		{"deshalb", "поэтому", "Ich war m\u00fcde, deshalb bin ich fr\u00fch ins Bett gegangen.", "A2", "genau deshalb, deshalb auch", "deshalb + Verb (Position 1)"},
		{"obwohl", "хотя", "Obwohl es kalt war, ging sie ohne Jacke.", "A2", "obwohl es schwer ist, obwohl ich wei\u00df", "obwohl + Nebensatz"},
		{"gem\u00fctlich", "уютный", "Das Caf\u00e9 ist sehr gem\u00fctlich.", "A2", "gem\u00fctliche Atmosph\u00e4re, gem\u00fctlicher Abend", "gem\u00fctlich + Nomen"},
		{"schaffen", "справиться", "Ich habe es geschafft!", "A2", "es schaffen, rechtzeitig schaffen", "schaffen + Akk"},
		{"vermissen", "скучать", "Ich vermisse meine Familie.", "A2", "jemanden vermissen, sehr vermissen", "vermissen + Akk"},
		{"sich entschuldigen", "извиниться", "Ich entschuldige mich f\u00fcr den Fehler.", "A2", "sich entschuldigen f\u00fcr, bei jemandem", "sich entschuldigen f\u00fcr + Akk"},
		{"sich lohnen", "стоить того", "Es lohnt sich, Deutsch zu lernen.", "A2", "sich lohnen zu, es lohnt sich", "sich lohnen + zu + Inf"},
		{"au\u00dferdem", "кроме того", "Au\u00dferdem habe ich keine Zeit.", "A2", "au\u00dferdem noch, au\u00dferdem auch", "au\u00dferdem + Hauptsatz"},
	})

	// ==================== B1 ====================
	insertWords(pool, ctx, "de", []wordSeed{
		{"sich auseinandersetzen", "разбираться, анализировать", "Man muss sich mit dem Problem auseinandersetzen.", "B1", "sich auseinandersetzen mit, kritisch", "sich auseinandersetzen mit + Dat"},
		{"beitragen", "вносить вклад", "Jeder kann zum Erfolg beitragen.", "B1", "beitragen zu, viel beitragen", "beitragen zu + Dat"},
		{"verzichten", "отказаться", "Er verzichtet auf S\u00fc\u00dfigkeiten.", "B1", "verzichten auf, freiwillig verzichten", "verzichten auf + Akk"},
		{"beeinflussen", "влиять", "Die Medien beeinflussen die Meinung.", "B1", "stark beeinflussen, positiv beeinflussen", "beeinflussen + Akk"},
		{"erheblich", "значительный", "Es gab erhebliche Ver\u00e4nderungen.", "B1", "erheblich steigen, erhebliche Unterschiede", "erheblich + Nomen / Verb"},
		{"Zusammenhang", "связь, контекст", "In diesem Zusammenhang ist es wichtig.", "B1", "im Zusammenhang mit, in diesem Zusammenhang", "Zusammenhang + mit/zwischen"},
		{"dennoch", "тем не менее", "Es war schwer, dennoch hat er es geschafft.", "B1", "aber dennoch, und dennoch", "dennoch + Hauptsatz"},
		{"sowohl...als auch", "как...так и", "Sowohl Kinder als auch Erwachsene waren begeistert.", "B1", "sowohl A als auch B", "sowohl + A + als auch + B"},
		{"weder...noch", "ни...ни", "Weder er noch sie waren zu Hause.", "B1", "weder A noch B", "weder + A + noch + B"},
		{"auf dem Laufenden bleiben", "быть в курсе", "Ich versuche, auf dem Laufenden zu bleiben.", "B1", "auf dem Laufenden halten, immer auf dem Laufenden", "auf dem Laufenden + bleiben/halten"},
		{"den Nagel auf den Kopf treffen", "попасть в точку", "Mit deinem Kommentar hast du den Nagel auf den Kopf getroffen.", "B1", "genau den Nagel, damit den Nagel", "den Nagel auf den Kopf treffen"},
		{"unter vier Augen", "наедине", "K\u00f6nnen wir unter vier Augen sprechen?", "B1", "unter vier Augen reden, besprechen", "unter vier Augen + Verb"},
		{"in Betracht ziehen", "принять во внимание", "Man sollte alle M\u00f6glichkeiten in Betracht ziehen.", "B1", "in Betracht ziehen ob, ernsthaft", "in Betracht ziehen + Akk / ob"},
		{"sich herausstellen", "оказаться", "Es hat sich herausgestellt, dass er recht hatte.", "B1", "es stellte sich heraus, sich als ... herausstellen", "sich herausstellen + dass"},
	})
}

func seedMediaEN(pool *pgxpool.Pool, ctx context.Context) {
	media := []struct {
		title, url, mediaType, level, topic, duration string
	}{
		{"Morning Routine \u2014 Easy English", "https://www.youtube.com/watch?v=GGp25fn25Cs", "video", "A2", "daily life", "5 min"},
		{"At the Restaurant \u2014 Easy English", "https://www.youtube.com/watch?v=BGHxLfRGk3I", "video", "A2", "food", "6 min"},
		{"My Daily Routine \u2014 Bob the Canadian", "https://www.youtube.com/watch?v=MIuoBGFMEAo", "video", "A2", "daily life", "8 min"},
		{"Shopping Vocabulary \u2014 English with Lucy", "https://www.youtube.com/watch?v=h4X-Oyl91sE", "video", "A2", "shopping", "10 min"},
		{"Travel English \u2014 Easy Conversations", "https://www.youtube.com/watch?v=tfJRwNo2SJI", "video", "A2", "travel", "7 min"},
	}
	for _, m := range media {
		pool.Exec(ctx,
			`INSERT INTO media_resources (title, url, media_type, level, topic, duration, language)
			 VALUES ($1, $2, $3, $4, $5, $6, 'en')
			 ON CONFLICT (url) DO NOTHING`,
			m.title, m.url, m.mediaType, m.level, m.topic, m.duration)
	}
}
