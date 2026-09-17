package tui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateProfileSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Allow selecting a profile by number (1-based)
	n, err := strconv.Atoi(keyMsg.String())
	if err == nil && n >= 1 && n <= maxProfileSlots+1 { // Add one for the Exit option
		m.profileMenu.cursor = n - 1

		keyMsg = tea.KeyMsg{Type: tea.KeyEnter} // continue as if "enter" was pressed
	}

	switch keyMsg.String() {
	case "up", "k":
		m.profileMenu.up()
	case "down", "j":
		m.profileMenu.down()
	case "enter":
		switch selected := m.profileMenu.selected(); selected.value {
		case exitValue:
			m.screen = screenGoodbye
		case newProfileValue:
			m.newProfileInput.SetValue("")
			m.newProfileInput.Focus()
			m.newProfileErr = ""
			m.screen = screenNewProfile
			return m, textinput.Blink
		default:
			m.profile = selected.value
			m.buildDirectionMenu()
			m.screen = screenDirectionSelect
		}
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) viewProfileSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Welcome to Spanish Buddy!"))
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("Practice your Spanish <-> English vocabulary."))
	b.WriteString("\n\n")
	b.WriteString(m.profileMenu.view())
	b.WriteString(helpStyle.Render("\n↑/↓ to navigate • enter to select • q to quit"))
	return b.String()
}

func isAlnum(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (m *Model) profileNameTaken(name string) bool {
	lower := strings.ToLower(name)
	for _, p := range m.existingProfiles {
		if strings.ToLower(p) == lower {
			return true
		}
	}
	return false
}

func (m Model) updateNewProfile(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "esc":
		m.newProfileInput.Blur()
		m.screen = screenProfileSelect
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.newProfileInput.Value())
		switch {
		case name == "":
			m.newProfileErr = "Please enter a non-empty profile"
			return m, nil
		case m.profileNameTaken(name):
			m.newProfileErr = fmt.Sprintf("Profile %s already exists, please enter a unique one", name)
			return m, nil
		case !isAlnum(name):
			m.newProfileErr = "Invalid profile name, please try again"
			return m, nil
		}

		if _, err := m.store.InitProfile(m.ctx, name); err != nil {
			m.fail(err)
			return m, nil
		}
		m.profile = name
		m.existingProfiles = append(m.existingProfiles, name)
		m.newProfileInput.Blur()
		m.buildDirectionMenu()
		m.screen = screenDirectionSelect
		return m, nil
	}

	var cmd tea.Cmd
	m.newProfileInput, cmd = m.newProfileInput.Update(msg)
	return m, cmd
}

func (m Model) viewNewProfile() string {
	var b strings.Builder
	b.WriteString(promptStyle.Render("What would you like to call this profile?"))
	b.WriteString("\n\n")
	b.WriteString(m.newProfileInput.View())
	if m.newProfileErr != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.newProfileErr))
	}
	b.WriteString(helpStyle.Render("\n\nenter to confirm • esc to go back"))
	return b.String()
}
