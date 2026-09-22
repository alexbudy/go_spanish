package tui

import (
	"context"

	"go_spanish/internal/store"

	"github.com/charmbracelet/bubbles/textinput"
)

// model is getting large so place it in its own file

const (
	// Limit to 8, so we can allow selecting by number (1-9) in the profile menu (allow 1 for exit).
	maxProfileSlots = 8
	newProfileValue = "__new__"
	exitValue       = "__exit__"
)

// Model is the root bubbletea model driving the whole application.
type Model struct {
	store *store.Store
	ctx   context.Context
	err   error

	screen screen

	existingProfiles []string
	profileMenu      choiceList

	newProfileInput textinput.Model
	newProfileErr   string
	delProfileErr   string

	deleteProfileConfirmMenu   choiceList
	deleteProfileSpecialPhrase string // user must type this to confirm deletion
	specialDeletePhraseInput   textinput.Model
	invalidDeletePhraseErr     string

	directionMenu   choiceList
	quizModeMenu    choiceList
	manageWordsMenu manageWordsList
	wordDetailsMenu wordDetails

	numQuestionsInput textinput.Model
	numQuestionsErr   string

	numOptionsInput textinput.Model
	numOptionsErr   string

	profile  string
	locale   store.Locale   // 'es' or 'en'
	quizMode store.QuizMode // aka difficulty, e.g. 'well_known', 'any', 'least_known'

	quiz       quizState // state of the current quiz
	answerMenu choiceList

	allWords []store.Word // slice of all words for current profile (both languages, both rankings)
}
