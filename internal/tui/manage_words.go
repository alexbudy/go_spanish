package tui

import (
	"strconv"
	"strings"
	"time"

	"go_spanish/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

/* components needed for word manage screen */

const entriesPerPage = 25

type manageWordsList struct {
	title     string       // the title for managing words
	items     []store.Word // all words
	pageStart int          // 0 base start of words to show
	cursor    int          // between pageStart and pageEnd, inclusive
	pageEnd   int          // 0 base end of words to show (TODO - can do without)
	direction store.Locale // which direction the words are shown in - es ( goes to english) or en ( goes to spanish)
}

type numberInputTimeoutMsg struct {
	generation int
}

func numberInputTimeout(generation int) tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg {
		return numberInputTimeoutMsg{
			generation: generation,
		}
	})
}

func newManageWordsList(title string, items []store.Word, direction store.Locale) manageWordsList {
	return manageWordsList{
		title: title, items: items, pageStart: 0, cursor: 0,
		pageEnd: entriesPerPage, direction: direction,
	}
}

func (mwl *manageWordsList) up() {
	if mwl.cursor == 0 {
		return // top of list - do nothing
	}
	mwl.cursor--

	if mwl.cursor < mwl.pageStart {
		mwl.pageStart--
		mwl.pageEnd--
	}
}

func (mwl *manageWordsList) down() {
	if mwl.cursor >= mwl.pageEnd || mwl.cursor == len(mwl.items)-1 {
		return // bottom of list - do nothing
	}
	mwl.cursor++

	if mwl.cursor >= mwl.pageEnd {
		mwl.pageStart++
		mwl.pageEnd++
	}
}

// PgUp pressed
func (mwl *manageWordsList) prevPage() {
	if mwl.cursor-entriesPerPage < 0 {
		// top of list - don't move cursor
		mwl.pageStart = 0
		mwl.pageEnd = entriesPerPage
	} else {
		mwl.pageStart = max(0, mwl.pageStart-entriesPerPage)
		mwl.cursor -= entriesPerPage
		mwl.pageEnd -= entriesPerPage
	}
}

// PgdDown pressed
func (mwl *manageWordsList) nextPage() {
	if mwl.cursor+entriesPerPage > len(mwl.items) {
		// bottom of list - no scroll
		return
	} else {
		mwl.pageStart += entriesPerPage
		mwl.cursor += entriesPerPage
		mwl.pageEnd = mwl.pageStart + entriesPerPage
	}
}

func (mwl *manageWordsList) first() {
	mwl.pageStart = 0
	mwl.cursor = 0
	mwl.pageEnd = entriesPerPage
}

func (mwl *manageWordsList) last() {
	mwl.pageStart = len(mwl.items) - entriesPerPage
	mwl.cursor = len(mwl.items) - 1
	mwl.pageEnd = len(mwl.items)
}

func (mwl *manageWordsList) swapDirection() {
	if mwl.direction == store.EnToEs {
		mwl.direction = store.EsToEn
	} else {
		mwl.direction = store.EnToEs
	}
}

func (mwl manageWordsList) view() string {
	var b strings.Builder
	if mwl.title != "" {
		b.WriteString(promptStyle.Render(mwl.title) + "\n\n")
	}

	for i := mwl.pageStart; i <= mwl.pageEnd && i < len(mwl.items); i++ {
		item := mwl.items[i]

		if i == mwl.cursor {
			b.WriteString(cursorStyle.Render("> "))
		} else {
			b.WriteString("  ")
		}
		wordLine := strconv.FormatInt(item.ID, 10) + ". "
		var score float64 // how well you know word in given direction

		if mwl.direction == store.EnToEs {
			wordLine += item.English + " -> " + item.Spanish
			score = item.EnToEsScore
		} else {
			wordLine += item.Spanish + " -> " + item.English
			score = item.EsToEnScore
		}

		b.WriteString(scoreToKnowledgeLevelStyle(score).Render(wordLine))

		b.WriteString("\n")

	}

	// Pages subtext - // offset by 1 since page numbers start at 1, round up when dividing
	pageCounterSubText := ("Page " + strconv.Itoa((mwl.pageStart/entriesPerPage)+1) +
		" of " + strconv.Itoa((len(mwl.items)+entriesPerPage-1)/entriesPerPage))
	b.WriteString(pageCounterStyle.Render(pageCounterSubText))

	return b.String()
}

// index is start of words to show
func (m *Model) buildManageWordsMenu() {
	m.manageWordsMenu = newManageWordsList("Select a word", m.allWords, m.locale)
}

// Manage words screen - show all words, allow for reset, removal (TODO)
func (m Model) viewManageWords() string {
	return m.manageWordsMenu.view() + helpStyle.Render("\n↑/↓ to navigate • enter to select • esc to go back • [PGUP/PGDN/HOME/END] to page by "+strconv.Itoa(entriesPerPage)+" words • [TAB] to swap direction • p to pronounce")
}

func (m Model) updateManageWords(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.manageWordsMenu.up()
		case "down", "j":
			m.manageWordsMenu.down()
		case "pgup":
			m.manageWordsMenu.prevPage()
		case "pgdown":
			m.manageWordsMenu.nextPage()
		case "home":
			m.manageWordsMenu.first()
		case "end":
			m.manageWordsMenu.last()
		case "tab":
			m.manageWordsMenu.swapDirection()
		case "p":
			word := m.manageWordsMenu.items[m.manageWordsMenu.cursor]
			// pronounce both words - TODO - cleaner way to do this?
			if m.manageWordsMenu.direction == store.EnToEs {
				go func() {
					_ = speak(word.English, store.EnToEs)
					go func() {
						_ = speak(word.Spanish, store.EsToEn)
					}()
				}()
			} else {
				go func() {
					_ = speak(word.Spanish, store.EsToEn)
					go func() {
						_ = speak(word.English, store.EnToEs)
					}()
				}()
			}
		case "enter":
			m.buildWordDetailsMenu() // build the detail menu
			m.screen = screenWordDetails
		case "esc":
			m.screen = screenQuizModeSelect
		default:
			// Check if the user pressed a number
			if len(msg.Runes) == 1 &&
				msg.Runes[0] >= '0' &&
				msg.Runes[0] <= '9' {

				m.manageWordsNumberInput += string(msg.Runes)

				// Only allow 2 digits
				if len(m.manageWordsNumberInput) > 2 {
					m.manageWordsNumberInput = ""
					m.manageWordsNumberInputGeneration++
					return m, nil
				}

				// This is a new generation.
				m.manageWordsNumberInputGeneration++

				generation := m.manageWordsNumberInputGeneration

				// Start a timer for this generation.
				return m, numberInputTimeout(generation)
			}
		}
	case numberInputTimeoutMsg:
		// Ignore old timer
		if msg.generation != m.manageWordsNumberInputGeneration {
			return m, nil
		}

		if m.manageWordsNumberInput == "" {
			return m, nil
		}

		n, err := strconv.Atoi(m.manageWordsNumberInput)
		if err != nil {
			m.manageWordsNumberInput = ""
			return m, nil
		}

		if n >= 1 && n <= len(m.manageWordsMenu.items) {
			m.manageWordsMenu.cursor = n - 1

			// Keep selected item visible.
			if m.manageWordsMenu.cursor < m.manageWordsMenu.pageStart {
				m.manageWordsMenu.pageStart = m.manageWordsMenu.cursor
			}

			if m.manageWordsMenu.cursor > m.manageWordsMenu.pageEnd {
				m.manageWordsMenu.pageEnd = m.manageWordsMenu.cursor
			}
		}

		m.manageWordsNumberInput = ""
	}

	return m, nil
}
