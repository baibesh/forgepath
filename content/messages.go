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
}

var MessagesDE = Messages{
	Welcome: func(name string) string {
		return fmt.Sprintf(
			"Hey, %s! \U0001F44B\n\n"+
				"Ich bin *ForgePath* — ich helfe dir jeden Tag Deutsch zu lernen.\n\n"+
				"So funktioniert's:\n"+
				"\U0001F31F Morgens — ein neues Wort + Quiz\n"+
				"\u270D\uFE0F Mittags — schreib ein paar Sätze\n"+
				"\U0001F3AC Abends — schau dir etwas Lustiges an\n"+
				"\U0001F31B Nachts — sieh wie dein Tag war\n\n"+
				"15-30 Minuten am Tag reichen!\n\n"+
				"Wähle jetzt dein Niveau:", name)
	},
	LevelSet: func(lang string) string {
		return fmt.Sprintf("\u2705 Sprache: *%s*\n\nWähle jetzt dein Niveau:", lang)
	},
	LevelPrompt:    "Wähle dein Niveau:",
	TimezonePrompt: "Wähle jetzt deine Zeitzone:",
	AllSet:         "\u2705 Alles klar! Dein erstes Wort kommt! \U0001F680",
	TzCustomPrompt: "Gib deine UTC-Abweichung ein (z.B. 5 für UTC+5, -3 für UTC-3):",
	TzInvalid:      "Bitte eine Zahl zwischen -12 und 14 eingeben:",

	StartReturning: func(name, flag, langName, level, schedule string) string {
		return fmt.Sprintf(
			"Hey, %s! %s\n\n"+
				"Du lernst %s, Niveau *%s*\n\n"+
				"*Dein Tagesplan:*\n%s\n\n"+
				"Was möchtest du machen?",
			name, flag, langName, level, schedule)
	},
	ChooseAction: "Wähle eine Aktion:",

	TodayAllDone:    "\u2705 *Alles erledigt für heute!* Super! Bis morgen \U0001F4AA",
	TodayLeft:       "*Was heute noch fehlt:*\n\n",
	TodayWord:       "\U0001F31F Neues Wort — /word",
	TodayWriting:    "\u270D\uFE0F Schreiben — /write",
	TodayQuiz:       "\U0001F9E9 Quiz — /quiz",
	AllWordsLearned: "Du hast alle Wörter gelernt! Toll! \U0001F389",
	NoWordsYet:      "Noch keine Wörter! Starte mit /word um dein erstes zu lernen.",
	WordsYouKnow:    "\U0001F4DA *Wörter die du kennst:*\n\n",
	AndMore:         func(n int) string { return fmt.Sprintf("\n_...und %d weitere_", n) },
	NothingToReview: "Noch nichts zum Wiederholen! Lerne zuerst Wörter mit /word",
	SkipMaxReached:  "Du hast diese Woche schon 2 Tage frei genommen. Du schaffst das! \U0001F4AA",
	SkipConfirm:     func(left int) string { return fmt.Sprintf("*Einen Tag frei nehmen?*\n\nDu hast noch *%d* freie(n) Tag(e) diese Woche.", left) },
	SkipDone:        func(left int) string { return fmt.Sprintf("\U0001F634 Ruhetag! Du hast noch %d freie(n) Tag(e) diese Woche.", left) },
	SkipCancelled:   "\u2705 Gute Wahl! Weiter geht's!",
	CancelNothing:   "Nichts zum Abbrechen gerade.",
	CancelDone:      "\u2705 Fertig! Du kannst jederzeit etwas Neues starten.",
	PrevTaskCancelled: "Vorherige Aufgabe abgebrochen.",
	SettingsTitle:   "\u2699\uFE0F *Einstellungen*\n\nWas möchtest du ändern?",

	Help: "\U0001F4DA *So funktioniert ForgePath*\n\n" +
		"Jeden Tag bekommst du:\n" +
		"\U0001F31F *Neues Wort* — lerne es und mach ein Quiz\n" +
		"\u270D\uFE0F *Schreiben* — schreib ein paar Sätze zu einem Thema\n" +
		"\U0001F3AC *Video* — schau dir etwas an und schreib darüber\n" +
		"\U0001F31B *Rückblick* — sieh wie dein Tag war\n\n" +
		"*Hauptbefehle:*\n" +
		"/word — ein neues Wort lernen\n" +
		"/write — etwas schreiben\n" +
		"/quiz — Wörter üben\n" +
		"/today — was heute noch fehlt\n" +
		"/stats — dein Fortschritt\n" +
		"/skip — einen Tag frei nehmen\n\n" +
		"Jede Woche ein Grammatik-Thema.\n" +
		"Keine Angst vor Fehlern — so lernt man! \U0001F4AA",

	QuizCorrect:     "\u2705 Ja! Richtig! \U0001F389",
	QuizWrong:       func(word, def string) string { return fmt.Sprintf("\u274C Knapp! Die Antwort war: *%s*\n(%s)\n\nKeine Sorge, du siehst es wieder!", word, def) },
	QuizWrongSimple: "\u274C Diesmal nicht. Du siehst es bald wieder!",
	QuizTrySentence: "Versuche einen ganzen Satz zu schreiben!",

	WritingTooShort:  "Das ist etwas kurz! Versuche ein paar Sätze zu schreiben.",
	WritingSaved:     func(count int) string { return fmt.Sprintf("\u2705 Gespeichert! (%d Wörter)\n\nAnalysiere...", count) },
	WritingSaveError: "Fehler beim Speichern. Versuche es nochmal.",
	MediaTooShort:    "Versuche etwas mehr zu schreiben!",
	MediaGotIt:       "\u2705 Hab's! Lass mich prüfen...",
	MediaGoodJob:     "Gut gemacht! \U0001F4AA",

	VoiceNotAvailable: "Spracherkennung ist gerade nicht verfügbar.",
	VoiceError:        "Konnte deine Sprachnachricht nicht verarbeiten. Versuche es nochmal.",
	VoiceCantHear:     "Konnte nichts hören. Versuche es nochmal oder sende Text.",
	VoiceIdleHint:     "Ich kann dich bei Schreib- oder Quiz-Aufgaben hören! Starte eine mit /write oder /quiz",

	SomethingWrong: "Etwas ist schiefgelaufen. Versuche es später nochmal.",
	NotStarted:     "Hi! Tippe /start um loszulegen.",
	FinishSetup:    "Lass uns zuerst die Einrichtung beenden! Tippe /start",
	ActiveTask:     "Du bist mitten in etwas. Beende es oder tippe /cancel.",
	AudioNotAvail:  "Audio nicht verfügbar",
	AudioFailed:    "Audioerzeugung fehlgeschlagen. Versuche es später.",
	AudioGenerating: "Audio wird erzeugt...",
}

func GetMessages(language string) *Messages {
	switch language {
	case "de":
		return &MessagesDE
	default:
		return &MessagesEN
	}
}
