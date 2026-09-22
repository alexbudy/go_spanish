// Package tui implements the Spanish Buddy bubbletea interface: choosing or
// creating a profile, picking a training direction, and running a
// multiple-choice vocabulary quiz against the store package.
package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"go_spanish/internal/store"
)

type screen int

const (
	screenProfileSelect screen = iota
	screenNewProfile
	screenDeleteProfileConfirm
	screenDeleteProfileConfirmFinal // final confirmation, require typing 'delete'
	screenDirectionSelect
	screenQuizModeSelect
	screenManageWords
	screenWordDetails // show details of a selected word
	screenNumQuestions
	screenNumOptions
	screenQuestion
	screenFeedback
	screenResults
	screenGoodbye
	screenError
)

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

// New builds the initial Model, loading whatever profiles already exist.
func New(s *store.Store) Model {
	m := Model{
		store: s,
		ctx:   context.Background(),
	}
	m.newProfileInput = textinput.New()
	m.newProfileInput.Placeholder = "profile name"
	m.newProfileInput.CharLimit = 32

	defaultNumQuestions := "10"
	m.numQuestionsInput = textinput.New()
	m.numQuestionsInput.Placeholder = defaultNumQuestions
	m.numQuestionsInput.SetValue(defaultNumQuestions)
	m.numQuestionsInput.Width = 5 // Ensure full placeholder shown
	m.numQuestionsInput.CharLimit = 3

	defaultNumAnswers := "4"
	m.numOptionsInput = textinput.New()
	m.numOptionsInput.Placeholder = defaultNumAnswers
	m.numOptionsInput.SetValue(defaultNumAnswers)
	m.numOptionsInput.CharLimit = 1

	m.specialDeletePhraseInput = textinput.New()
	m.specialDeletePhraseInput.Placeholder = "type deletion phrase"
	m.specialDeletePhraseInput.Width = 50

	m.loadProfiles()
	return m
}

func (m *Model) loadProfiles() {
	profiles, err := m.store.GetProfiles(m.ctx)
	if err != nil {
		m.fail(err)
		return
	}
	m.existingProfiles = profiles
	m.buildProfileMenu()
	m.screen = screenProfileSelect
}

func (m *Model) buildProfileMenu() {
	var items []choiceItem
	for _, name := range m.existingProfiles {
		items = append(items, choiceItem{label: name, value: name})
	}

	// add a visual line break to the last element
	items[len(items)-1].label += "\n    ----------"

	// add one new profile entry for dynamic profile creation
	if len(m.existingProfiles) < maxProfileSlots {
		items = append(items, choiceItem{label: "New Profile", value: newProfileValue})
	}

	items = append(items, choiceItem{label: "Exit :(", value: exitValue})
	m.profileMenu = newChoiceList("Select a profile (or exit)", items)
}

func (m *Model) fail(err error) {
	m.err = err
	m.screen = screenError
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	switch m.screen {
	case screenProfileSelect:
		return m.updateProfileSelect(msg)
	case screenNewProfile:
		return m.updateNewProfile(msg)
	case screenDeleteProfileConfirm:
		return m.updateDeleteProfileConfirm(msg)
	case screenDeleteProfileConfirmFinal:
		return m.updateDeleteProfileConfirmFinal(msg)
	case screenDirectionSelect:
		return m.updateDirectionSelect(msg)
	case screenQuizModeSelect:
		return m.updateQuizModeSelect(msg)
	case screenManageWords:
		return m.updateManageWords(msg)
	case screenWordDetails:
		return m.updateWordDetails(msg)
	case screenNumQuestions:
		return m.updateNumQuestions(msg)
	case screenNumOptions:
		return m.updateNumOptions(msg)
	case screenQuestion:
		return m.updateQuestion(msg)
	case screenFeedback:
		return m.updateFeedback(msg)
	case screenResults:
		return m.updateResults(msg)
	case screenGoodbye, screenError:
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenProfileSelect:
		return m.viewProfileSelect()
	case screenNewProfile:
		return m.viewNewProfile()
	case screenDeleteProfileConfirm:
		return m.viewDeleteProfileConfirm()
	case screenDeleteProfileConfirmFinal:
		return m.viewDeleteProfileConfirmFinal()
	case screenDirectionSelect:
		return m.viewDirectionSelect()
	case screenQuizModeSelect:
		return m.viewQuizModeSelect()
	case screenManageWords:
		return m.viewManageWords()
	case screenWordDetails:
		return m.viewWordDetails()
	case screenNumQuestions:
		return m.viewNumQuestions()
	case screenNumOptions:
		return m.viewNumOptions()
	case screenQuestion:
		return m.viewQuestion()
	case screenFeedback:
		return m.viewFeedback()
	case screenResults:
		return m.viewResults()
	case screenGoodbye:
		return titleStyle.Render("Exiting the program, thanks for training!") + "\n"
	case screenError:
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + helpStyle.Render("\n\npress any key to exit")
	}
	return ""
}
