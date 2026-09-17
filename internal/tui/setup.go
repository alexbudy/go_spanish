package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"go_spanish/internal/store"
)

func (m *Model) buildDirectionMenu() {
	m.directionMenu = newChoiceList("Select a training direction", []choiceItem{
		{label: "Provide spanish words, select english words", value: string(store.EsToEn)},
		{label: "Provide english words, select spanish words", value: string(store.EnToEs)},
	})
}

func (m *Model) buildQuizModeMenu() {
	m.quizModeMenu = newChoiceList("Select a quiz mode", []choiceItem{
		{label: "Well-known words", value: string(store.WellKnown)},
		{label: "Any words", value: string(store.Any)},
		{label: "Least-known words\n     ----------", value: string(store.LeastKnown)},
		{label: "Manage words", value: string(store.ManageWords)},
	})
}

func (m Model) updateDirectionSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "up", "k":
		m.directionMenu.up()
	case "down", "j":
		m.directionMenu.down()
	case "enter":
		m.locale = store.Locale(m.directionMenu.selected().value)

		m.buildQuizModeMenu()
		m.screen = screenQuizModeSelect
		return m, textinput.Blink
	case "esc":
		m.screen = screenProfileSelect
	}
	return m, nil
}

func (m Model) viewDirectionSelect() string {
	return m.directionMenu.view(m.profile) + helpStyle.Render("\n↑/↓ to navigate • enter to select • esc to go back")
}

func (m Model) updateQuizModeSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "up", "k":
		m.quizModeMenu.up()
	case "down", "j":
		m.quizModeMenu.down()
	case "enter":
		m.locale = store.Locale(m.directionMenu.selected().value)

		m.numQuestionsInput.SetValue("")
		m.numQuestionsInput.Focus()
		m.numQuestionsErr = ""
		m.screen = screenNumQuestions

		return m, textinput.Blink
	case "esc":
		m.screen = screenProfileSelect
	}
	return m, nil
}

func (m Model) viewQuizModeSelect() string {
	return m.quizModeMenu.view() + helpStyle.Render("\n↑/↓ to navigate • enter to select • esc to go back")
}

func (m Model) updateNumQuestions(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "esc":
		m.numQuestionsInput.Blur()
		m.screen = screenDirectionSelect
		return m, nil
	case "enter":
		n, err := strconv.Atoi(strings.TrimSpace(m.numQuestionsInput.Value()))
		switch {
		case err != nil:
			m.numQuestionsErr = "Please enter a valid integer"
			return m, nil
		case n <= 0:
			m.numQuestionsErr = "Number of questions cannot be negative or 0."
			return m, nil
		case n > 20:
			m.numQuestionsErr = "Let's just stick to a reasonable number of questions for now (<= 20)"
			return m, nil
		}
		m.quiz.totalQuestions = n
		m.numQuestionsInput.Blur()
		m.numOptionsInput.SetValue("")
		m.numOptionsInput.Focus()
		m.numOptionsErr = ""
		m.screen = screenNumOptions
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	m.numQuestionsInput, cmd = m.numQuestionsInput.Update(msg)
	return m, cmd
}

func (m Model) viewNumQuestions() string {
	var b strings.Builder
	b.WriteString(promptStyle.Render("How many questions would you like to solve? (1-20)"))
	b.WriteString("\n\n")
	b.WriteString(m.numQuestionsInput.View())
	if m.numQuestionsErr != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.numQuestionsErr))
	}
	b.WriteString(helpStyle.Render("\n\nenter to confirm • esc to go back"))
	return b.String()
}

func (m Model) updateNumOptions(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "esc":
		m.numOptionsInput.Blur()
		m.screen = screenNumQuestions
		return m, nil
	case "enter":
		n, err := strconv.Atoi(strings.TrimSpace(m.numOptionsInput.Value()))
		switch {
		case err != nil:
			m.numOptionsErr = "Please enter a valid integer"
			return m, nil
		case n < 2:
			m.numOptionsErr = "Number of options must be >= 2."
			return m, nil
		case n > 6:
			m.numOptionsErr = "Number of options cannot exceed 6."
			return m, nil
		}
		m.quiz.numOptions = n
		m.numOptionsInput.Blur()
		if err := m.startQuiz(); err != nil {
			m.fail(err)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.numOptionsInput, cmd = m.numOptionsInput.Update(msg)
	return m, cmd
}

func (m Model) viewNumOptions() string {
	var b strings.Builder
	b.WriteString(promptStyle.Render("How many options would you like to see per word? (2-6)"))
	b.WriteString("\n\n")
	b.WriteString(m.numOptionsInput.View())
	if m.numOptionsErr != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.numOptionsErr))
	}
	b.WriteString(helpStyle.Render("\n\nenter to confirm • esc to go back"))
	return b.String()
}
