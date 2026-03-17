package content

import "fmt"

type Messages struct {
	// Onboarding
	Welcome          func(name string) string
	LevelSet         func(lang string) string
	LevelPrompt      string
	TimezonePrompt   string
	AllSet           string
	TzCustomPrompt   string
	TzInvalid        string

	// Start (returning user)
	StartReturning   func(name, flag, langName, level, schedule string) string
	ChooseAction     string

	// Commands
	TodayAllDone     string
	TodayLeft        string
	TodayWord        string
	TodayWriting     string
	TodayQuiz        string
	AllWordsLearned  string
	NoWordsYet       string
	WordsYouKnow     string
	AndMore          func(n int) string
	NothingToReview  string
	SkipMaxReached   string
	SkipConfirm      func(left int) string
	SkipDone         func(left int) string
	SkipCancelled    string
	CancelNothing    string
	CancelDone       string
	PrevTaskCancelled string
	SettingsTitle    string

	// Help
	Help             string

	// Quiz
	QuizCorrect      string
	QuizWrong        func(word, def string) string
	QuizWrongSimple  string
	QuizTrySentence  string

	// Writing
	WritingTooShort  string
	WritingSaved     func(count int) string
	WritingSaveError string
	MediaTooShort    string
	MediaGotIt       string
	MediaGoodJob     string

	// Voice
	VoiceNotAvailable string
	VoiceError        string
	VoiceCantHear     string
	VoiceIdleHint     string

	// General
	SomethingWrong   string
	NotStarted       string
	FinishSetup      string
	ActiveTask       string
	AudioNotAvail    string
	AudioFailed      string
	AudioGenerating  string

	// Buttons (main keyboard)
	BtnNewWord       string
	BtnWrite         string
	BtnQuiz          string
	BtnToday         string
	BtnProgress      string
	BtnSettings      string

	// Inline buttons
	BtnYesSkip       string
	BtnNoSkip        string
	BtnListen        string
	BtnDoneWatching  string
	BtnCustomize     string
	BtnTimezone      string
	BtnLevel         string
	BtnLanguage      string
	BtnSchedule      string

	// Format labels
	LabelNewWord     string
	LabelWriting     string
	LabelQuiz        string
	LabelWord        string
	LabelWritings    string
	LabelStreak      string
	LabelThisWeek    string
	LabelLast7Days   string
	LabelWords       string
	LabelQuizzes     string
	LabelEndOfDay    string
	LabelSeeYouTmrw string

	// Writing / quiz format
	LabelTimeToWrite string
	LabelTopic       string
	LabelTryToUse    string
	LabelExample     string
	LabelMarkers     string
	LabelNewWordDay  string
	LabelHowToUse   string
	LabelGoesWith    string
	LabelGrammar     string
	LabelQuickQuiz   string
	LabelRemember    string
	LabelWhatsWord   string
	LabelTypeAnswer  string
	LabelSentence    string
	LabelMakeSentence string
	LabelTypeSentence string
	LabelWatchThis   string
	LabelAfterWatch  string
	LabelWhatThink   string
	LabelWriteAbout  string
	LabelWhatAbout   string
	LabelNewWordHeard string
	LabelWhatDoYouThink string
	LabelTryUseGrammar string
	LabelTakeQuiz    string
	LabelYourWeek    string
	LabelNewGrammar  string
	LabelDuration    string
	LabelScheduleWord    string
	LabelScheduleWriting string
	LabelScheduleMedia   string
	LabelScheduleReview  string

	// Quiz poll
	QuizPollQuestion func(word string) string
	TryUseWord       func(word string) string
}

var MessagesEN = Messages{
	Welcome: func(name string) string {
		return fmt.Sprintf(
			"Hey, %s! \U0001F44B\n\n"+
				"I'm *ForgePath* — I'll help you learn English every day.\n\n"+
				"Here's how it works:\n"+
				"\U0001F31F Morning — a new word + quiz\n"+
				"\u270D\uFE0F Afternoon — write a few sentences\n"+
				"\U0001F3AC Evening — watch something fun\n"+
				"\U0001F31B Night — see how your day went\n\n"+
				"15-30 minutes a day is all you need!\n\n"+
				"Now select your level:", name)
	},
	LevelSet: func(lang string) string {
		return fmt.Sprintf("\u2705 Language: *%s*\n\nNow select your level:", lang)
	},
	LevelPrompt:    "Now select your level:",
	TimezonePrompt: "Now select your timezone:",
	AllSet:         "\u2705 All set! Your first word is coming! \U0001F680",
	TzCustomPrompt: "Type your UTC offset (e.g. 5 for UTC+5, -3 for UTC-3):",
	TzInvalid:      "Please enter a number between -12 and 14:",

	StartReturning: func(name, flag, langName, level, schedule string) string {
		return fmt.Sprintf(
			"Hey, %s! %s\n\n"+
				"You're learning %s, level *%s*\n\n"+
				"*Your daily schedule:*\n%s\n\n"+
				"Pick what you want to do!",
			name, flag, langName, level, schedule)
	},
	ChooseAction: "Choose an action:",

	TodayAllDone:    "\u2705 *All done for today!* Great job! See you tomorrow \U0001F4AA",
	TodayLeft:       "*What's left today:*\n\n",
	TodayWord:       "\U0001F31F New word — /word",
	TodayWriting:    "\u270D\uFE0F Writing — /write",
	TodayQuiz:       "\U0001F9E9 Quiz — /quiz",
	AllWordsLearned: "You've learned all available words! Amazing! \U0001F389",
	NoWordsYet:      "No words yet! Start with /word to learn your first one.",
	WordsYouKnow:    "\U0001F4DA *Words you know:*\n\n",
	AndMore:         func(n int) string { return fmt.Sprintf("\n_...and %d more_", n) },
	NothingToReview: "Nothing to review yet! Learn some words first with /word",
	SkipMaxReached:  "You've already taken 2 days off this week. You got this! \U0001F4AA",
	SkipConfirm:     func(left int) string { return fmt.Sprintf("*Take a day off?*\n\nYou have *%d* day(s) off left this week.", left) },
	SkipDone:        func(left int) string { return fmt.Sprintf("\U0001F634 Rest day! You have %d day(s) off left this week.", left) },
	SkipCancelled:   "\u2705 Good choice! Let's keep going!",
	CancelNothing:   "Nothing to cancel right now.",
	CancelDone:      "\u2705 Done! You can start something new anytime.",
	PrevTaskCancelled: "Previous task cancelled.",
	SettingsTitle:   "\u2699\uFE0F *Settings*\n\nWhat do you want to change?",

	Help: "\U0001F4DA *How ForgePath works*\n\n" +
		"Every day you get:\n" +
		"\U0001F31F *New word* — learn it and take a quiz\n" +
		"\u270D\uFE0F *Writing* — write a few sentences on a topic\n" +
		"\U0001F3AC *Video* — watch something and write about it\n" +
		"\U0001F31B *Review* — see how your day went\n\n" +
		"*Main commands:*\n" +
		"/word — learn a new word\n" +
		"/write — write something\n" +
		"/quiz — practice your words\n" +
		"/today — what's left for today\n" +
		"/stats — your progress\n" +
		"/skip — take a day off\n\n" +
		"Each week focuses on one grammar topic.\n" +
		"Don't worry about mistakes — that's how you learn! \U0001F4AA",

	QuizCorrect:     "\u2705 Yes! You got it! \U0001F389",
	QuizWrong:       func(word, def string) string { return fmt.Sprintf("\u274C Close! The answer was: *%s*\n(%s)\n\nNo worries, you'll see it again!", word, def) },
	QuizWrongSimple: "\u274C Not this time. You'll see it again soon!",
	QuizTrySentence: "Try writing a full sentence!",

	WritingTooShort:  "That's a bit short! Try to write at least a few sentences.",
	WritingSaved:     func(count int) string { return fmt.Sprintf("\u2705 Saved! (%d words)\n\nAnalyzing...", count) },
	WritingSaveError: "Error saving your writing. Try again.",
	MediaTooShort:    "Try to write a bit more!",
	MediaGotIt:       "\u2705 Got it! Let me check...",
	MediaGoodJob:     "Good job! \U0001F4AA",

	VoiceNotAvailable: "Voice recognition is not available right now.",
	VoiceError:        "Could not process your voice message. Try again.",
	VoiceCantHear:     "Could not hear anything. Try again or send text.",
	VoiceIdleHint:     "I can hear you during writing or quiz tasks! Start one with /write or /quiz",

	SomethingWrong: "Something went wrong. Try again later.",
	NotStarted:     "Hi! Type /start to get started.",
	FinishSetup:    "Let's finish setting up first! Type /start",
	ActiveTask:     "You're in the middle of something. Finish it or type /cancel first.",
	AudioNotAvail:  "Audio not available",
	AudioFailed:    "Audio generation failed. Try again later.",
	AudioGenerating: "Generating audio...",

	BtnNewWord:  "\U0001F31F New word",
	BtnWrite:    "\u270D\uFE0F Write",
	BtnQuiz:     "\U0001F9E9 Quiz",
	BtnToday:    "\U0001F4CB Today",
	BtnProgress: "\U0001F4CA Progress",
	BtnSettings: "\u2699\uFE0F Settings",

	BtnYesSkip:      "\u2705 Yes, skip",
	BtnNoSkip:       "\u274C No, I'll do it",
	BtnListen:       "\U0001F50A Listen",
	BtnDoneWatching: "\u2705 Done watching!",
	BtnCustomize:    "\u2699\uFE0F Customize schedule",
	BtnTimezone:     "\U0001F550 Timezone",
	BtnLevel:        "\U0001F4DA Level",
	BtnLanguage:     "\U0001F310 Language",
	BtnSchedule:     "\U0001F514 Schedule",

	LabelNewWord:     "\U0001F31F *New word for you!*",
	LabelWriting:     "\u270D\uFE0F Writing",
	LabelQuiz:        "\U0001F9E9 Quiz",
	LabelWord:        "Word",
	LabelWritings:    "Writings",
	LabelStreak:      "Streak",
	LabelThisWeek:    "This week",
	LabelLast7Days:   "*Last 7 days:*",
	LabelWords:       "Words",
	LabelQuizzes:     "Quizzes",
	LabelEndOfDay:    "\U0001F31B *End of day!*",
	LabelSeeYouTmrw:  "See you tomorrow!",

	LabelTimeToWrite:    "\u270D\uFE0F *Time to write!*",
	LabelTopic:          "*Topic:*",
	LabelTryToUse:       "Try to use",
	LabelExample:        "Example",
	LabelMarkers:        "Markers",
	LabelNewWordDay:     "\U0001F31F *New word for you!*",
	LabelHowToUse:       "\U0001F527 How to use",
	LabelGoesWith:       "\U0001F517 Goes with",
	LabelGrammar:        "\U0001F4A1 *This week's grammar:*",
	LabelQuickQuiz:      "\U0001F9E9 *Quick quiz!*",
	LabelRemember:       "\U0001F9E9 *Can you remember?*",
	LabelWhatsWord:      "What's the word for:",
	LabelTypeAnswer:     "Type your answer:",
	LabelSentence:       "\U0001F9E9 *Use it in a sentence!*",
	LabelMakeSentence:   "Make a sentence with:",
	LabelTypeSentence:   "Type your sentence:",
	LabelWatchThis:      "\U0001F3AC *Watch this!*",
	LabelAfterWatch:     "After watching, press the button and I'll give you a small task!",
	LabelWhatThink:      "\U0001F4DD *What did you think?*",
	LabelWriteAbout:     "Write a few sentences about what you watched.",
	LabelWhatAbout:      "What was it about?",
	LabelNewWordHeard:   "What new word did you hear?",
	LabelWhatDoYouThink: "What do you think about it?",
	LabelTryUseGrammar:  "Try to use",
	LabelTakeQuiz:       "Take a /quiz to complete today!",
	LabelYourWeek:       "\U0001F389 *Your week!*",
	LabelNewGrammar:     "New grammar topic starts now!",
	LabelDuration:       "\u23F1",
	LabelScheduleWord:    "\U0001F31F Word",
	LabelScheduleWriting: "\u270D\uFE0F Writing",
	LabelScheduleMedia:   "\U0001F3AC Media",
	LabelScheduleReview:  "\U0001F31B Review",

	QuizPollQuestion: func(word string) string { return fmt.Sprintf("What does \"%s\" mean?", word) },
	TryUseWord:       func(word string) string { return fmt.Sprintf("Try using the word *%s* in your sentence!", word) },
}

var MessagesRU = Messages{
	Welcome: func(name string) string {
		return fmt.Sprintf(
			"Привет, %s! \U0001F44B\n\n"+
				"Я *ForgePath* — помогу тебе учить английский каждый день.\n\n"+
				"Как это работает:\n"+
				"\U0001F31F Утро — новое слово + квиз\n"+
				"\u270D\uFE0F День — напиши несколько предложений\n"+
				"\U0001F3AC Вечер — посмотри что-то интересное\n"+
				"\U0001F31B Ночь — итоги дня\n\n"+
				"15-30 минут в день — это всё что нужно!\n\n"+
				"Выбери свой уровень:", name)
	},
	LevelSet: func(lang string) string {
		return fmt.Sprintf("\u2705 Язык: *%s*\n\nТеперь выбери уровень:", lang)
	},
	LevelPrompt:    "Теперь выбери уровень:",
	TimezonePrompt: "Выбери часовой пояс:",
	AllSet:         "\u2705 Всё готово! Первое слово уже летит! \U0001F680",
	TzCustomPrompt: "Введи смещение UTC (например 5 для UTC+5, -3 для UTC-3):",
	TzInvalid:      "Введи число от -12 до 14:",

	StartReturning: func(name, flag, langName, level, schedule string) string {
		return fmt.Sprintf(
			"Привет, %s! %s\n\n"+
				"Ты учишь %s, уровень *%s*\n\n"+
				"*Твоё расписание:*\n%s\n\n"+
				"Выбери что хочешь делать!",
			name, flag, langName, level, schedule)
	},
	ChooseAction: "Выбери действие:",

	TodayAllDone:    "\u2705 *Всё сделано на сегодня!* Молодец! До завтра \U0001F4AA",
	TodayLeft:       "*Что осталось на сегодня:*\n\n",
	TodayWord:       "\U0001F31F Новое слово — /word",
	TodayWriting:    "\u270D\uFE0F Письмо — /write",
	TodayQuiz:       "\U0001F9E9 Квиз — /quiz",
	AllWordsLearned: "Ты выучил все доступные слова! Круто! \U0001F389",
	NoWordsYet:      "Слов пока нет! Начни с /word чтобы выучить первое.",
	WordsYouKnow:    "\U0001F4DA *Слова которые ты знаешь:*\n\n",
	AndMore:         func(n int) string { return fmt.Sprintf("\n_...и ещё %d_", n) },
	NothingToReview: "Пока нечего повторять! Сначала выучи слова через /word",
	SkipMaxReached:  "Ты уже взял 2 выходных на этой неделе. Ты справишься! \U0001F4AA",
	SkipConfirm:     func(left int) string { return fmt.Sprintf("*Взять выходной?*\n\nОсталось *%d* выходных на этой неделе.", left) },
	SkipDone:        func(left int) string { return fmt.Sprintf("\U0001F634 День отдыха! Осталось %d выходных на этой неделе.", left) },
	SkipCancelled:   "\u2705 Правильный выбор! Продолжаем!",
	CancelNothing:   "Нечего отменять.",
	CancelDone:      "\u2705 Готово! Можешь начать что-то новое.",
	PrevTaskCancelled: "Предыдущее задание отменено.",
	SettingsTitle:   "\u2699\uFE0F *Настройки*\n\nЧто хочешь изменить?",

	Help: "\U0001F4DA *Как работает ForgePath*\n\n" +
		"Каждый день ты получаешь:\n" +
		"\U0001F31F *Новое слово* — выучи и пройди квиз\n" +
		"\u270D\uFE0F *Письмо* — напиши несколько предложений на тему\n" +
		"\U0001F3AC *Видео* — посмотри и напиши об этом\n" +
		"\U0001F31B *Обзор* — итоги дня\n\n" +
		"*Основные команды:*\n" +
		"/word — выучить новое слово\n" +
		"/write — написать текст\n" +
		"/quiz — тренировать слова\n" +
		"/today — что осталось на сегодня\n" +
		"/stats — твой прогресс\n" +
		"/skip — взять выходной\n\n" +
		"Каждую неделю — новая тема грамматики.\n" +
		"Не бойся ошибок — так учатся! \U0001F4AA",

	QuizCorrect:     "\u2705 Да! Правильно! \U0001F389",
	QuizWrong:       func(word, def string) string { return fmt.Sprintf("\u274C Почти! Ответ: *%s*\n(%s)\n\nНе переживай, увидишь снова!", word, def) },
	QuizWrongSimple: "\u274C Не в этот раз. Скоро увидишь снова!",
	QuizTrySentence: "Попробуй написать целое предложение!",

	WritingTooShort:  "Маловато! Попробуй написать хотя бы несколько предложений.",
	WritingSaved:     func(count int) string { return fmt.Sprintf("\u2705 Сохранено! (%d слов)\n\nАнализирую...", count) },
	WritingSaveError: "Ошибка сохранения. Попробуй ещё раз.",
	MediaTooShort:    "Попробуй написать чуть больше!",
	MediaGotIt:       "\u2705 Принято! Проверяю...",
	MediaGoodJob:     "Отлично! \U0001F4AA",

	VoiceNotAvailable: "Распознавание голоса сейчас недоступно.",
	VoiceError:        "Не удалось обработать голосовое сообщение. Попробуй ещё раз.",
	VoiceCantHear:     "Ничего не слышно. Попробуй ещё раз или отправь текст.",
	VoiceIdleHint:     "Я слышу тебя во время письма или квиза! Начни через /write или /quiz",

	SomethingWrong: "Что-то пошло не так. Попробуй позже.",
	NotStarted:     "Привет! Напиши /start чтобы начать.",
	FinishSetup:    "Давай сначала закончим настройку! Напиши /start",
	ActiveTask:     "Ты сейчас в процессе. Заверши или напиши /cancel.",
	AudioNotAvail:  "Аудио недоступно",
	AudioFailed:    "Ошибка генерации аудио. Попробуй позже.",
	AudioGenerating: "Генерирую аудио...",

	BtnNewWord:  "\U0001F31F Новое слово",
	BtnWrite:    "\u270D\uFE0F Писать",
	BtnQuiz:     "\U0001F9E9 Квиз",
	BtnToday:    "\U0001F4CB Сегодня",
	BtnProgress: "\U0001F4CA Прогресс",
	BtnSettings: "\u2699\uFE0F Настройки",

	BtnYesSkip:      "\u2705 Да, пропустить",
	BtnNoSkip:       "\u274C Нет, я сделаю",
	BtnListen:       "\U0001F50A Слушать",
	BtnDoneWatching: "\u2705 Просмотрено!",
	BtnCustomize:    "\u2699\uFE0F Настроить расписание",
	BtnTimezone:     "\U0001F550 Часовой пояс",
	BtnLevel:        "\U0001F4DA Уровень",
	BtnLanguage:     "\U0001F310 Язык",
	BtnSchedule:     "\U0001F514 Расписание",

	LabelNewWord:     "\U0001F31F *Новое слово для тебя!*",
	LabelWriting:     "\u270D\uFE0F Письмо",
	LabelQuiz:        "\U0001F9E9 Квиз",
	LabelWord:        "Слово",
	LabelWritings:    "Тексты",
	LabelStreak:      "Серия",
	LabelThisWeek:    "На этой неделе",
	LabelLast7Days:   "*Последние 7 дней:*",
	LabelWords:       "Слова",
	LabelQuizzes:     "Квизы",
	LabelEndOfDay:    "\U0001F31B *Итоги дня!*",
	LabelSeeYouTmrw:  "До завтра!",

	LabelTimeToWrite:    "\u270D\uFE0F *Время писать!*",
	LabelTopic:          "*Тема:*",
	LabelTryToUse:       "Попробуй использовать",
	LabelExample:        "Пример",
	LabelMarkers:        "Маркеры",
	LabelNewWordDay:     "\U0001F31F *Новое слово для тебя!*",
	LabelHowToUse:       "\U0001F527 Как использовать",
	LabelGoesWith:       "\U0001F517 Сочетается с",
	LabelGrammar:        "\U0001F4A1 *Грамматика на этой неделе:*",
	LabelQuickQuiz:      "\U0001F9E9 *Быстрый квиз!*",
	LabelRemember:       "\U0001F9E9 *Помнишь?*",
	LabelWhatsWord:      "Какое слово означает:",
	LabelTypeAnswer:     "Напиши ответ:",
	LabelSentence:       "\U0001F9E9 *Составь предложение!*",
	LabelMakeSentence:   "Составь предложение со словом:",
	LabelTypeSentence:   "Напиши предложение:",
	LabelWatchThis:      "\U0001F3AC *Посмотри это!*",
	LabelAfterWatch:     "После просмотра нажми кнопку — я дам тебе задание!",
	LabelWhatThink:      "\U0001F4DD *Что думаешь?*",
	LabelWriteAbout:     "Напиши несколько предложений о том что посмотрел.",
	LabelWhatAbout:      "О чём это было?",
	LabelNewWordHeard:   "Какое новое слово услышал?",
	LabelWhatDoYouThink: "Что ты об этом думаешь?",
	LabelTryUseGrammar:  "Попробуй использовать",
	LabelTakeQuiz:       "Пройди /quiz чтобы завершить сегодня!",
	LabelYourWeek:       "\U0001F389 *Твоя неделя!*",
	LabelNewGrammar:     "Новая тема грамматики начинается!",
	LabelDuration:       "\u23F1",
	LabelScheduleWord:    "\U0001F31F Слово",
	LabelScheduleWriting: "\u270D\uFE0F Письмо",
	LabelScheduleMedia:   "\U0001F3AC Медиа",
	LabelScheduleReview:  "\U0001F31B Обзор",

	QuizPollQuestion: func(word string) string { return fmt.Sprintf("Что означает \"%s\"?", word) },
	TryUseWord:       func(word string) string { return fmt.Sprintf("Попробуй использовать слово *%s* в предложении!", word) },
}

var MessagesKK = Messages{
	Welcome: func(name string) string {
		return fmt.Sprintf(
			"Сәлем, %s! \U0001F44B\n\n"+
				"Мен *ForgePath* — күнделікті ағылшын тілін үйренуге көмектесемін.\n\n"+
				"Қалай жұмыс істейді:\n"+
				"\U0001F31F Таңертең — жаңа сөз + квиз\n"+
				"\u270D\uFE0F Күндіз — бірнеше сөйлем жаз\n"+
				"\U0001F3AC Кешке — қызықты нәрсе көр\n"+
				"\U0001F31B Түнде — күн қорытындысы\n\n"+
				"Күніне 15-30 минут жеткілікті!\n\n"+
				"Деңгейіңді таңда:", name)
	},
	LevelSet: func(lang string) string {
		return fmt.Sprintf("\u2705 Тіл: *%s*\n\nЕнді деңгейіңді таңда:", lang)
	},
	LevelPrompt:    "Деңгейіңді таңда:",
	TimezonePrompt: "Уақыт белдеуін таңда:",
	AllSet:         "\u2705 Бәрі дайын! Бірінші сөз келе жатыр! \U0001F680",
	TzCustomPrompt: "UTC ығысуын жаз (мысалы 5 — UTC+5, -3 — UTC-3):",
	TzInvalid:      "-12 мен 14 арасындағы сан жаз:",

	StartReturning: func(name, flag, langName, level, schedule string) string {
		return fmt.Sprintf(
			"Сәлем, %s! %s\n\n"+
				"Сен %s оқып жатырсың, деңгей *%s*\n\n"+
				"*Күн тәртібің:*\n%s\n\n"+
				"Не істегің келеді?",
			name, flag, langName, level, schedule)
	},
	ChooseAction: "Әрекет таңда:",

	TodayAllDone:    "\u2705 *Бүгінге бәрі жасалды!* Жарайсың! Ертеңге дейін \U0001F4AA",
	TodayLeft:       "*Бүгінге не қалды:*\n\n",
	TodayWord:       "\U0001F31F Жаңа сөз — /word",
	TodayWriting:    "\u270D\uFE0F Жазу — /write",
	TodayQuiz:       "\U0001F9E9 Квиз — /quiz",
	AllWordsLearned: "Барлық сөздерді үйрендің! Керемет! \U0001F389",
	NoWordsYet:      "Сөздер әлі жоқ! /word арқылы бірінші сөзді үйрен.",
	WordsYouKnow:    "\U0001F4DA *Сен білетін сөздер:*\n\n",
	AndMore:         func(n int) string { return fmt.Sprintf("\n_...және тағы %d_", n) },
	NothingToReview: "Қайталайтын ештеңе жоқ! Алдымен /word арқылы сөз үйрен",
	SkipMaxReached:  "Бұл аптада 2 демалыс алдың. Сен қолыңнан келеді! \U0001F4AA",
	SkipConfirm:     func(left int) string { return fmt.Sprintf("*Демалыс алу керек пе?*\n\nОсы аптада *%d* демалыс қалды.", left) },
	SkipDone:        func(left int) string { return fmt.Sprintf("\U0001F634 Демалыс күні! Осы аптада %d демалыс қалды.", left) },
	SkipCancelled:   "\u2705 Дұрыс таңдау! Жалғастырамыз!",
	CancelNothing:   "Тоқтатуға ештеңе жоқ.",
	CancelDone:      "\u2705 Дайын! Жаңа нәрсе бастай аласың.",
	PrevTaskCancelled: "Алдыңғы тапсырма тоқтатылды.",
	SettingsTitle:   "\u2699\uFE0F *Баптаулар*\n\nНені өзгерткің келеді?",

	Help: "\U0001F4DA *ForgePath қалай жұмыс істейді*\n\n" +
		"Күн сайын сен аласың:\n" +
		"\U0001F31F *Жаңа сөз* — үйрен және квиз өт\n" +
		"\u270D\uFE0F *Жазу* — тақырыпқа бірнеше сөйлем жаз\n" +
		"\U0001F3AC *Видео* — қара және жаз\n" +
		"\U0001F31B *Қорытынды* — күн қорытындысы\n\n" +
		"*Негізгі командалар:*\n" +
		"/word — жаңа сөз үйрену\n" +
		"/write — мәтін жазу\n" +
		"/quiz — сөздерді жаттығу\n" +
		"/today — бүгінге не қалды\n" +
		"/stats — прогресс\n" +
		"/skip — демалыс алу\n\n" +
		"Әр апта — жаңа грамматика тақырыбы.\n" +
		"Қателіктен қорықпа — солай үйренеді! \U0001F4AA",

	QuizCorrect:     "\u2705 Иә! Дұрыс! \U0001F389",
	QuizWrong:       func(word, def string) string { return fmt.Sprintf("\u274C Жақын! Жауабы: *%s*\n(%s)\n\nАлаңдама, тағы көресің!", word, def) },
	QuizWrongSimple: "\u274C Бұл жолы емес. Жақында тағы көресің!",
	QuizTrySentence: "Толық сөйлем жазып көр!",

	WritingTooShort:  "Аз! Кем дегенде бірнеше сөйлем жазып көр.",
	WritingSaved:     func(count int) string { return fmt.Sprintf("\u2705 Сақталды! (%d сөз)\n\nТалдап жатырмын...", count) },
	WritingSaveError: "Сақтау қатесі. Қайта байқап көр.",
	MediaTooShort:    "Тағы аздап жазып көр!",
	MediaGotIt:       "\u2705 Қабылданды! Тексеремін...",
	MediaGoodJob:     "Жарайсың! \U0001F4AA",

	VoiceNotAvailable: "Дауысты тану қазір қол жетімді емес.",
	VoiceError:        "Дауыстық хабарды өңдеу мүмкін болмады. Қайта байқап көр.",
	VoiceCantHear:     "Ештеңе естілмейді. Қайта байқа немесе мәтін жібер.",
	VoiceIdleHint:     "Жазу немесе квиз кезінде естимін! /write немесе /quiz арқылы баста",

	SomethingWrong: "Бірдеңе дұрыс болмады. Кейінірек байқап көр.",
	NotStarted:     "Сәлем! Бастау үшін /start жаз.",
	FinishSetup:    "Алдымен баптауды аяқтайық! /start жаз",
	ActiveTask:     "Сен қазір тапсырмадасың. Аяқта немесе /cancel жаз.",
	AudioNotAvail:  "Аудио қол жетімді емес",
	AudioFailed:    "Аудио жасау қатесі. Кейінірек байқап көр.",
	AudioGenerating: "Аудио жасалуда...",

	BtnNewWord:  "\U0001F31F Жаңа сөз",
	BtnWrite:    "\u270D\uFE0F Жазу",
	BtnQuiz:     "\U0001F9E9 Квиз",
	BtnToday:    "\U0001F4CB Бүгін",
	BtnProgress: "\U0001F4CA Прогресс",
	BtnSettings: "\u2699\uFE0F Баптаулар",

	BtnYesSkip:      "\u2705 Иә, өткіз",
	BtnNoSkip:       "\u274C Жоқ, жасаймын",
	BtnListen:       "\U0001F50A Тыңдау",
	BtnDoneWatching: "\u2705 Көрдім!",
	BtnCustomize:    "\u2699\uFE0F Кестені баптау",
	BtnTimezone:     "\U0001F550 Уақыт белдеуі",
	BtnLevel:        "\U0001F4DA Деңгей",
	BtnLanguage:     "\U0001F310 Тіл",
	BtnSchedule:     "\U0001F514 Кесте",

	LabelNewWord:     "\U0001F31F *Сен үшін жаңа сөз!*",
	LabelWriting:     "\u270D\uFE0F Жазу",
	LabelQuiz:        "\U0001F9E9 Квиз",
	LabelWord:        "Сөз",
	LabelWritings:    "Мәтіндер",
	LabelStreak:      "Серия",
	LabelThisWeek:    "Осы аптада",
	LabelLast7Days:   "*Соңғы 7 күн:*",
	LabelWords:       "Сөздер",
	LabelQuizzes:     "Квиздер",
	LabelEndOfDay:    "\U0001F31B *Күн қорытындысы!*",
	LabelSeeYouTmrw:  "Ертеңге дейін!",

	LabelTimeToWrite:    "\u270D\uFE0F *Жазу уақыты!*",
	LabelTopic:          "*Тақырып:*",
	LabelTryToUse:       "Қолдануға тырыс",
	LabelExample:        "Мысал",
	LabelMarkers:        "Маркерлер",
	LabelNewWordDay:     "\U0001F31F *Сен үшін жаңа сөз!*",
	LabelHowToUse:       "\U0001F527 Қалай қолданылады",
	LabelGoesWith:       "\U0001F517 Бірге қолданылады",
	LabelGrammar:        "\U0001F4A1 *Осы аптаның грамматикасы:*",
	LabelQuickQuiz:      "\U0001F9E9 *Жылдам квиз!*",
	LabelRemember:       "\U0001F9E9 *Есіңде ме?*",
	LabelWhatsWord:      "Мынаны қалай айтады:",
	LabelTypeAnswer:     "Жауабыңды жаз:",
	LabelSentence:       "\U0001F9E9 *Сөйлем құра!*",
	LabelMakeSentence:   "Мына сөзбен сөйлем құра:",
	LabelTypeSentence:   "Сөйлемді жаз:",
	LabelWatchThis:      "\U0001F3AC *Мынаны көр!*",
	LabelAfterWatch:     "Көргеннен кейін батырманы бас — тапсырма беремін!",
	LabelWhatThink:      "\U0001F4DD *Не ойлайсың?*",
	LabelWriteAbout:     "Көргенің туралы бірнеше сөйлем жаз.",
	LabelWhatAbout:      "Ол не туралы еді?",
	LabelNewWordHeard:   "Қандай жаңа сөз естідің?",
	LabelWhatDoYouThink: "Бұл туралы не ойлайсың?",
	LabelTryUseGrammar:  "Қолдануға тырыс",
	LabelTakeQuiz:       "Бүгінді аяқтау үшін /quiz өт!",
	LabelYourWeek:       "\U0001F389 *Сенің аптаң!*",
	LabelNewGrammar:     "Жаңа грамматика тақырыбы басталды!",
	LabelDuration:       "\u23F1",
	LabelScheduleWord:    "\U0001F31F Сөз",
	LabelScheduleWriting: "\u270D\uFE0F Жазу",
	LabelScheduleMedia:   "\U0001F3AC Медиа",
	LabelScheduleReview:  "\U0001F31B Қорытынды",

	QuizPollQuestion: func(word string) string { return fmt.Sprintf("\"%s\" не дегенді білдіреді?", word) },
	TryUseWord:       func(word string) string { return fmt.Sprintf("*%s* сөзін сөйлемде қолдануға тырыс!", word) },
}

func GetMessages(language string) *Messages {
	switch language {
	case "ru":
		return &MessagesRU
	case "kk":
		return &MessagesKK
	default:
		return &MessagesEN
	}
}
