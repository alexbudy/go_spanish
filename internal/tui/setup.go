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
		newChoiceItem("Provide spanish words, select english words", string(store.EsToEn)),
		newChoiceItem("Provide english words, select spanish words", string(store.EnToEs)),
		newSeparatorItem(),
		newChoiceItem("Profile Settings", profileSettings),
	})
}

func (m *Model) buildDeleteProfileConfirmMenu() {
	m.deleteProfileConfirmMenu = newChoiceList("Are you sure you want to delete this profile?", []choiceItem{
		newChoiceItem("Yes, delete this profile (cannot be undone)", "yes"),
		newChoiceItem("No, go back", "no"),
	})
}

func (m *Model) buildQuizModeMenu() {
	b := strings.Builder{}
	b.WriteString("Select a quiz mode for ")

	if m.locale == store.EsToEn {
		b.WriteString("Spanish -> English translations")
	} else {
		b.WriteString("English -> Spanish translations")
	}

	m.quizModeMenu = newChoiceList(b.String(), []choiceItem{
		newChoiceItem("Profile's well-known words", string(store.WellKnown)),
		newChoiceItem("Any words", string(store.Any)),
		newChoiceItem("Profile's least-known words", string(store.LeastKnown)),
		newSeparatorItem(),
		newChoiceItem("Manage words", string(store.ManageWords)),
	})
}

func (m Model) viewDeleteProfileConfirm() string {
	return m.deleteProfileConfirmMenu.view() + helpStyle.Render("\n↑/↓ to navigate • enter to select • esc to go back")
}

func (m Model) updateDeleteProfileConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Allow deleting by number (1-based)
	n, err := strconv.Atoi(keyMsg.String())
	if err == nil && (n == 1 || n == 2) { // Add one for the Exit option
		m.deleteProfileConfirmMenu.cursor = n - 1

		keyMsg = tea.KeyMsg{Type: tea.KeyEnter} // continue as if "enter" was pressed
	}

	switch keyMsg.String() {
	case "up", "k":
		m.deleteProfileConfirmMenu.up()
	case "down", "j":
		m.deleteProfileConfirmMenu.down()
	case "enter":
		if m.deleteProfileConfirmMenu.selected().value == "yes" {
			m.specialDeletePhraseInput.CharLimit = 50 // allow for typing mistakes
			m.specialDeletePhraseInput.Width = 50
			m.specialDeletePhraseInput.Focus()

			m.screen = screenDeleteProfileConfirmFinal
		} else {
			m.screen = screenProfileSelect
		}
		return m, nil
	case "esc":
		m.screen = screenProfileSelect
	}
	return m, nil
}

// Final confirmation menu, requiring user to type deletion phrase
func (m Model) viewDeleteProfileConfirmFinal() string {
	deletionPhrase := "delete " + m.profile // phrase to type to delete profile

	var b strings.Builder
	b.WriteString(deleteConfirmStyle.Render("Type '"))
	b.WriteString(deleteConfirmSpecialWordStyle.Render(deletionPhrase))
	b.WriteString(deleteConfirmStyle.Render("'to confirm deletion"))
	b.WriteString("\n\n")
	b.WriteString(m.specialDeletePhraseInput.View())

	if m.invalidDeletePhraseErr != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.invalidDeletePhraseErr))
	}

	b.WriteString(helpStyle.Render("\n\nenter to confirm • esc to go back"))
	return b.String()
}

func (m Model) updateDeleteProfileConfirmFinal(msg tea.Msg) (tea.Model, tea.Cmd) {
	deletionPhrase := "delete " + m.profile

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "enter":
		if m.specialDeletePhraseInput.Value() != deletionPhrase {
			m.invalidDeletePhraseErr = "Invalid deletion phrase"
			return m, nil
		}

		if _, err := m.store.DeleteProfile(m.ctx, m.profile); err != nil {
			m.fail(err)
			return m, nil
		}

		// remove the deleted profile from existing profiles
		for i, existingProfile := range m.existingProfiles {
			if existingProfile == m.profile {
				m.existingProfiles = append(m.existingProfiles[:i], m.existingProfiles[i+1:]...)
				break
			}
		}
		m.buildProfileMenu() // rebuild profile menu in case using new profile
		m.screen = screenProfileSelect

		return m, nil
	case "esc":
		m.buildProfileMenu() // rebuild profile menu in case using new profile
		m.screen = screenProfileSelect

		return m, nil
	}

	var cmd tea.Cmd
	m.specialDeletePhraseInput, cmd = m.specialDeletePhraseInput.Update(msg)

	return m, cmd
}

func (m Model) updateDirectionSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	// Allow deleting by number (1-based)
	n, err := strconv.Atoi(keyMsg.String())
	if err == nil && (n >= 1 && n <= 3) { // Add one for the Exit option

		m.directionMenu.cursor = n - 1

		keyMsg = tea.KeyMsg{Type: tea.KeyEnter} // continue as if "enter" was pressed
	}

	switch keyMsg.String() {
	case "up", "k":
		m.directionMenu.up()
	case "down", "j":
		m.directionMenu.down()
	case "enter":
		selectableNumber := 0

		for i, item := range m.directionMenu.items {
			if !item.selectable {
				continue
			}

			selectableNumber++

			if selectableNumber == n {
				m.directionMenu.cursor = i
			}
		}

		if m.directionMenu.selected().value == profileSettings {
			m.buildProfileSettingsMenu()
			m.screen = screenProfileSettings
			return m, nil
		}

		m.locale = store.Locale(m.directionMenu.selected().value)

		m.buildQuizModeMenu()
		m.screen = screenQuizModeSelect
		return m, textinput.Blink
	case "esc":
		m.buildProfileMenu() // rebuild profile menu in case using new profile
		m.screen = screenProfileSelect
	}
	return m, nil
}

func (m Model) viewDirectionSelect() string {
	return m.directionMenu.view(m.profile) + helpStyle.Render("\n↑/↓ to navigate • enter to select • esc to go back")
}

func (m Model) viewProfileSettings() string {
	return m.profileSettings.view() + helpStyle.Render("\n↑/↓/⇆ to navigate • enter to select • esc to go back")
}

func (m Model) updateQuizModeSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// load all words into model
	if len(m.allWords) == 0 {
		wordPool, err := m.store.GetAllWords(m.ctx, m.profile)
		if err != nil {
			return m, nil
		}
		m.allWords = wordPool
	}

	// Allow navigating by number (1-based) - TODO can refactor to avoid duplication with updateDirectionSelect
	n, err := strconv.Atoi(keyMsg.String())
	if err == nil && n >= 1 && n <= 4 { // three modes + manage words
		m.quizModeMenu.cursor = n - 1

		keyMsg = tea.KeyMsg{Type: tea.KeyEnter} // continue as if "enter" was pressed
	}

	switch keyMsg.String() {
	case "up", "k":
		m.quizModeMenu.up()
	case "down", "j":
		m.quizModeMenu.down()
	case "enter":
		selectableNumber := 0

		for i, item := range m.quizModeMenu.items {
			if !item.selectable {
				continue
			}

			selectableNumber++

			if selectableNumber == n {
				m.quizModeMenu.cursor = i
			}
		}

		m.quizMode = store.QuizMode(m.quizModeMenu.selected().value)

		if m.quizModeMenu.selected().value == string(store.ManageWords) {
			m.buildManageWordsMenu()
			m.screen = screenManageWords
			return m, nil
		}

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
	b.WriteString(promptStyle.Render("How many words would you like to try? (1-20)"))
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
