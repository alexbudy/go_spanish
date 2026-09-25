package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateProfileSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Handle input while renaming a profile
	if m.profileRenaming {
		switch keyMsg.String() {
		case "esc":
			m.profileRenaming = false
			m.profileRenameInput.Blur()

			m.profileMenu.renameIndex = -1
			m.profileMenu.renameInput = nil

			return m, nil

		case "enter":
			// save the input
			newName := m.profileRenameInput.Value()

			m.store.UpdateProfileName(m.ctx, m.profileMenu.selected().value, newName)

			m.profileMenu.items[m.profileRenameIndex].label = newName

			m.profileRenaming = false
			m.profileRenameInput.Blur()

			m.profileMenu.renameIndex = -1
			m.profileMenu.renameInput = nil

			return m, nil
		}

		// Everything else goes to the text input
		var cmd tea.Cmd

		m.profileRenameInput, cmd = m.profileRenameInput.Update(msg)

		m.profileMenu.renameInput = &m.profileRenameInput

		return m, cmd
	}

	// Allow selecting a profile or option (new profile, exit) by number (1-based)
	n, err := strconv.Atoi(keyMsg.String())
	if err == nil { // 1-9 was pressed
		if n >= 1 && n <= len(m.profileMenu.items) {
			selectableNumber := 0

			for i, item := range m.profileMenu.items {
				if !item.selectable {
					continue
				}
				selectableNumber++
				if selectableNumber == n {
					m.profileMenu.cursor = i
				}
			}
		}
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
			m.newProfileErr = ""
			m.screen = screenDirectionSelect
		}
	case "r":
		selected := m.profileMenu.selected()

		if selected.value == exitValue || selected.value == newProfileValue {
			// can't rename 'exit' or 'new profile' options as they are not profiles
			m.profileRenameErr = "Invalid renaming option selected"
			break
		}

		m.profileRenaming = true
		m.profileRenameIndex = m.profileMenu.cursor

		m.profileRenameInput.SetValue(selected.label)
		m.profileRenameInput.CursorEnd()
		m.profileRenameInput.Focus()

		m.profileMenu.renameIndex = m.profileRenameIndex
		m.profileMenu.renameInput = &m.profileRenameInput

		return m, textinput.Blink
	case "delete", "d": // DEL was pressed
		selected := m.profileMenu.selected()
		if selected.value == exitValue || selected.value == newProfileValue {
			// can't delete 'exit' or 'new profile' options as they are not profiles
			m.delProfileErr = "Invalid deletion option selected"
			break
		}

		m.profile = selected.value        // profile to delete
		m.buildDeleteProfileConfirmMenu() // confirm Deletion
		m.screen = screenDeleteProfileConfirm
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

	if m.delProfileErr != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.delProfileErr))
	}

	if m.profileRenameErr != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.profileRenameErr))
	}

	if m.profileRenaming {
		b.WriteString(helpStyle.Render("\nenter to save name change • esc to cancel"))
	} else {
		b.WriteString(helpStyle.Render("\n↑/↓ to navigate • enter to select • r to rename a profile • [DEL]/'d' to delete a profile • q to quit"))
	}
	return b.String()
}

// has at least one char or int, spaces, _, - valid
func isValidProfileName(name string) bool {
	hasLetterOrNumber := false

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			hasLetterOrNumber = true
		case r >= 'A' && r <= 'Z':
			hasLetterOrNumber = true
		case r >= '0' && r <= '9':
			hasLetterOrNumber = true
		case r == ' ' || r == '_' || r == '-':
			// valid characters
		default:
			return false
		}
	}

	return hasLetterOrNumber
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
		case !isValidProfileName(name):
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
