package tui

import (
	"context"
	"strconv"
	"strings"

	"go_spanish/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

/* components for showing the details of a word */

type wordDetails struct {
	title         string
	word          store.Word
	selected      store.Locale // which word is selected (en or es)
	profile       string       // profile name for associated rankings
	wordUpdateMsg string       // if word was updated (score reset), show this msg
}

func newWordDetails(title string, word store.Word, selected store.Locale, profile string) wordDetails {
	return wordDetails{title: title, word: word, selected: selected, profile: profile, wordUpdateMsg: ""}
}

func (wd *wordDetails) toggleSelectedLocale() {
	if wd.selected == store.EnToEs { // simple toggle, probably better way to do it
		wd.selected = store.EsToEn
	} else {
		wd.selected = store.EnToEs
	}
}

// reset the score for the word, start locale is locale (ex locale = 'es' resets es_to_en score)
func (wd *wordDetails) resetScore(ctx context.Context, s *store.Store) store.Word {
	if wd.selected == store.EnToEs {
		wd.word.EnToEsScore = 0
		wd.wordUpdateMsg = "Reset score for '" + wd.word.English + "' <-> '" + wd.word.Spanish + "'"
	} else {
		wd.word.EsToEnScore = 0
		wd.wordUpdateMsg = "Reset score for '" + wd.word.Spanish + "' <-> '" + wd.word.English + "'"
	}
	s.ResetRankingForWord(ctx, wd.word.ID, wd.selected, wd.profile)

	return wd.word
}

func (wd wordDetails) view() string {
	var b strings.Builder
	if wd.selected == store.EnToEs {
		b.WriteString("Word Details for '" + wd.word.English + "' <-> '" + wd.word.Spanish + "'\n")
	} else {
		b.WriteString("Word Details for '" + wd.word.Spanish + "' <-> '" + wd.word.English + "'\n")
	}

	b.WriteString("ID: " + strconv.FormatInt(wd.word.ID, 10) + "\n")

	if wd.selected == store.EnToEs {
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
	m.wordDetailsMenu = newWordDetails("Word Details for ", m.manageWordsMenu.items[m.manageWordsMenu.cursor], m.manageWordsMenu.direction, m.profile) // default to English 'from' selection
}

// Manage words screen - show all words, allow for reset, removal (TODO?)
func (m Model) viewWordDetails() string {
	return m.wordDetailsMenu.view() + helpStyle.Render("\n ↑/↓ to navigate • esc to go back • tab to pronounce • ctrl+r to reset selected score")
}

func (m Model) updateWordDetails(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k", "down", "j":
		m.wordDetailsMenu.toggleSelectedLocale()
	case "tab":
		// pronounce both words - TODO - cleaner way to do this?
		if m.wordDetailsMenu.selected == store.EnToEs {
			go func() {
				_ = speak(m.wordDetailsMenu.word.English, store.EnToEs)
				go func() {
					_ = speak(m.wordDetailsMenu.word.Spanish, store.EsToEn)
				}()
			}()
		} else {
			go func() {
				_ = speak(m.wordDetailsMenu.word.Spanish, store.EsToEn)
				go func() {
					_ = speak(m.wordDetailsMenu.word.English, store.EnToEs)
				}()
			}()
		}
	case "ctrl+r":
		updatedWord := m.wordDetailsMenu.resetScore(m.ctx, m.store)

		m.manageWordsMenu.items[m.manageWordsMenu.cursor] = updatedWord
	case "esc":
		m.screen = screenManageWords
	}

	return m, nil
}
