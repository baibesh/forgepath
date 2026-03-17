package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(pool *pgxpool.Pool) {
	ctx := context.Background()
	seedGrammarWeeksEN(pool, ctx)
	seedWordsEN(pool, ctx)
	seedMediaEN(pool, ctx)
	log.Println("Seeding completed")
}

func seedGrammarWeeksEN(pool *pgxpool.Pool, ctx context.Context) {
	weeks := []struct {
		num                                                    int
		family, focus, tenseName, anchor, markers, formula, ex string
	}{
		{1, "Simple", "Past Simple", "Past Simple",
			"\U0001F6AA Закрытая дверь — действие завершено, дверь захлопнулась. Носитель думает: «это было и прошло». Представь дверь, которая закрылась — назад не вернёшься.",
			"yesterday, last week, ago, in 2020",
			"S + V2 (ed / irregular)",
			"I watched a movie yesterday."},
		{2, "Simple", "Present Simple", "Present Simple",
			"\U0001F504 Карусель — крутится снова и снова. Носитель описывает привычки и факты, которые повторяются. Представь карусель: каждый день одно и то же.",
			"always, usually, every day, sometimes",
			"S + V1 (he/she +s)",
			"I usually wake up at 7."},
		{3, "Simple", "Future Simple", "Future Simple",
			"\U0001F52E Хрустальный шар — решение прямо сейчас, спонтанное. Носитель не планировал заранее, а решил в момент разговора. Представь: «О, я позвоню!» — решение на ходу.",
			"tomorrow, next week, I think, probably",
			"S + will + V1",
			"I will call you tomorrow."},
		{4, "Continuous", "Present Continuous", "Present Continuous",
			"\U0001F4F8 Фотография — ты ловишь момент прямо сейчас. Носитель видит действие в процессе, как будто фоткает. Также: временные ситуации и планы на ближайшее будущее.",
			"now, right now, at the moment, look!",
			"S + am/is/are + Ving",
			"I am reading a book right now."},
		{5, "Continuous", "Past Continuous", "Past Continuous",
			"\U0001F3AC Кинокадр — фоновое действие в прошлом. Представь сцену из фильма: камера показывает, что происходило, когда вдруг что-то случилось. Одно — фон, другое — событие.",
			"while, when, at that moment, all day yesterday",
			"S + was/were + Ving",
			"I was cooking when you called."},
		{6, "Continuous", "Future Continuous", "Future Continuous",
			"\U0001F3A5 Видеозвонок из будущего — ты «подключаешься» к моменту в будущем и видишь процесс. Носитель говорит: «В 8 вечера я буду заниматься этим». Действие будет в процессе.",
			"at 5 pm tomorrow, this time next week, when you arrive",
			"S + will be + Ving",
			"I will be working at 6 pm tomorrow."},
		{7, "Perfect", "Present Perfect", "Present Perfect",
			"\U0001F309 Мост — соединяет прошлое с настоящим. Носителю важен не момент, а результат сейчас: «Я уже сделал» = готово. Секрет: если важно КОГДА — Past Simple, если важно ЧТО — Present Perfect.",
			"already, yet, just, ever, never, since, for",
			"S + have/has + V3",
			"I have already finished my homework."},
		{8, "Perfect", "Past Perfect", "Past Perfect",
			"\u23EA Перемотка — действие ДО другого прошлого. Представь два события в прошлом: одно раньше другого. Past Perfect — то, что было первым. «Я уже поел, когда она пришла».",
			"before, after, by the time, already (past context)",
			"S + had + V3",
			"I had eaten before she arrived."},
		{9, "Perfect", "Future Perfect", "Future Perfect",
			"\U0001F3C1 Финишная черта — к определённому моменту в будущем будет готово. Носитель уверен: «К пятнице я закончу». Представь дедлайн — и результат к нему.",
			"by tomorrow, by next year, by the time",
			"S + will have + V3",
			"I will have finished the project by Friday."},
		{10, "Perfect Continuous", "Present Perfect Continuous", "Present Perfect Continuous",
			"\u23F3 Песочные часы — действие началось в прошлом и ВСЁ ЕЩЁ идёт. Или только что закончилось, и виден след. «Я учу английский 2 года» = начал и продолжаю.",
			"for, since, all day, how long, lately, recently",
			"S + have/has been + Ving",
			"I have been learning English for two years."},
		{11, "Perfect Continuous", "Past Perfect Continuous", "Past Perfect Continuous",
			"\U0001F4A8 Мокрый след — действие шло ДОЛГО до другого события в прошлом. «Я бежал 30 минут, поэтому устал». Акцент на длительности процесса до момента в прошлом.",
			"for, since, before, when, all day",
			"S + had been + Ving",
			"I had been waiting for an hour when the bus finally came."},
		{12, "Special", "Going to", "Going to",
			"\U0001F4CB План в блокноте — заранее решённое намерение. Разница с will: going to = уже решил, will = решаю сейчас. «I'm going to travel» = я уже купил билет в голове.",
			"tonight, next week, soon, I plan",
			"S + am/is/are going to + V1",
			"I am going to start a new course next month."},
	}
	for _, w := range weeks {
		pool.Exec(ctx,
			`INSERT INTO grammar_weeks (week_num, family, focus, tense_name, anchor, markers, formula, example, language)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'en')
			 ON CONFLICT (week_num, language) DO UPDATE SET
			   family = EXCLUDED.family, focus = EXCLUDED.focus, tense_name = EXCLUDED.tense_name,
			   anchor = EXCLUDED.anchor, markers = EXCLUDED.markers, formula = EXCLUDED.formula, example = EXCLUDED.example`,
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
