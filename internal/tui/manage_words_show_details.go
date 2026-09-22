package tui

import (
	"strconv"
	"strings"

	"go_spanish/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

/* components for showing the details of a word */

type wordDetails struct {
	title         string
	word          store.Word
	selected      store.Locale // TODO which 'from' direction is selected, can be toggled
	wordUpdateMsg string       // if word was updated (score reset), show this msg
}

func newWordDetails(title string, word store.Word, selected store.Locale) wordDetails {
	return wordDetails{title: title, word: word, selected: selected, wordUpdateMsg: ""}
}

func (wd *wordDetails) toggleSelectedLocale() {
	if wd.selected == store.En { // simple toggle, probably better way to do it
		wd.selected = store.Es
	} else {
		wd.selected = store.En
	}
}

// reset the score for the word, start locale is locale (ex locale = 'es' resets es_to_en score)
func (wd *wordDetails) resetScore() {
	if wd.selected == store.En {
		wd.word.EnToEsScore = 0
		wd.wordUpdateMsg = "Reset score for '" + wd.word.English + "' <-> '" + wd.word.Spanish + "'"
	} else {
		wd.word.EsToEnScore = 0
		wd.wordUpdateMsg = "Reset score for '" + wd.word.Spanish + "' <-> '" + wd.word.English + "'"
	}
}

func (wd wordDetails) view() string {
	var b strings.Builder

	b.WriteString("Word Details for '" + wd.word.English + "' <-> '" + wd.word.Spanish + "'\n")

	b.WriteString("ID: " + strconv.FormatInt(wd.word.ID, 10) + "\n")

	if wd.selected == store.En {
		b.WriteString(cursorStyle.Render("> "))
		b.WriteString(scoreToKnowledgeLevelStyle(wd.word.EnToEsScore).Render(wd.word.English + " -> " + wd.word.Spanish + ": " + strconv.FormatFloat(wd.word.EnToEsScore, 'f', 2, 64)))
		b.WriteString("\n")
		b.WriteString(scoreToKnowledgeLevelStyle(wd.word.EsToEnScore).Render("  " + wd.word.Spanish + " -> " + wd.word.English + ": " + strconv.FormatFloat(wd.word.EsToEnScore, 'f', 2, 64)))
		b.WriteString("\n")
	} else {
		b.WriteString(scoreToKnowledgeLevelStyle(wd.word.EnToEsScore).Render("  " + wd.word.English + " -> " + wd.word.Spanish + ": " + strconv.FormatFloat(wd.word.EnToEsScore, 'f', 2, 64)))
		b.WriteString("\n")
		b.WriteString(cursorStyle.Render("> "))
		b.WriteString(scoreToKnowledgeLevelStyle(wd.word.EsToEnScore).Render(wd.word.Spanish + " -> " + wd.word.English + ": " + strconv.FormatFloat(wd.word.EsToEnScore, 'f', 2, 64)))
		b.WriteString("\n")
	}

	if wd.wordUpdateMsg != "" {
		b.WriteString(successStyle.Render(wd.wordUpdateMsg))
		b.WriteString("\n")
	}

	return b.String()
}

// index is start of words to show
func (m *Model) buildWordDetailsMenu() {
	m.wordDetailsMenu = newWordDetails("Word Details for ", m.manageWordsMenu.items[m.manageWordsMenu.cursor], store.En) // default to English 'from' selection
}

// Manage words screen - show all words, allow for reset, removal (TODO?)
func (m Model) viewWordDetails() string {
	return m.wordDetailsMenu.view() + helpStyle.Render("\n ↑/↓ to navigate • esc to go back • ctrl+r to reset selected score")
}

func (m Model) updateWordDetails(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k", "down", "j":
		m.wordDetailsMenu.toggleSelectedLocale()
	case "ctrl+r":
		m.wordDetailsMenu.resetScore()
	case "esc":
		m.screen = screenManageWords
	}

	return m, nil
}
