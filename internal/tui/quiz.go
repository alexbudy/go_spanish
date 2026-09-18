package tui

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"go_spanish/internal/store"
)

type quizState struct {
	quizMode store.QuizMode

	totalQuestions    int
	numOptions        int
	index             int // 1-based index of the question currently shown/answered
	questionsCorrect  int
	questionedWordIDs []int64
	incorrectWords    []store.Word
	answerPool        []string

	target         store.Word
	targetText     string
	correctAnswer  string
	selectedAnswer string
	wasCorrect     bool
}

func (m *Model) startQuiz() error {
	pool, err := m.store.GetAllWords(m.ctx, m.locale.AnswerLang())
	if err != nil {
		return err
	}
	m.quiz = quizState{
		quizMode:       m.quiz.quizMode,
		totalQuestions: m.quiz.totalQuestions,
		numOptions:     m.quiz.numOptions,
		index:          1,
		answerPool:     pool,
	}
	return m.loadNextQuestion()
}

func (m *Model) loadNextQuestion() error {
	words, err := m.store.GetWordsForQuestion(m.ctx, m.profile, m.quiz.quizMode, m.locale, m.quiz.questionedWordIDs, 10)
	if err != nil {
		return err
	}
	if len(words) == 0 {
		return fmt.Errorf("no more words available for %q", m.profile)
	}
	target := words[rand.Intn(len(words))]

	qLang := m.locale.QuestionLang()
	aLang := m.locale.AnswerLang()
	targetText := target.Text(qLang)
	correct := target.Text(aLang)

	options := sampleDistinct(m.quiz.answerPool, m.quiz.numOptions)
	options = removeFirst(options, correct)
	if len(options) > m.quiz.numOptions-1 {
		options = options[:m.quiz.numOptions-1]
	}
	answerSet := append(append([]string{}, options...), correct)
	rand.Shuffle(len(answerSet), func(i, j int) { answerSet[i], answerSet[j] = answerSet[j], answerSet[i] })

	m.quiz.target = target
	m.quiz.targetText = targetText
	m.quiz.correctAnswer = correct
	title := fmt.Sprintf("%d/%d Please select the translation for %q", m.quiz.index, m.quiz.totalQuestions, targetText)
	m.answerMenu = newChoiceList(title, labeledItems(answerSet...))
	m.screen = screenQuestion
	return nil
}

// sampleDistinct returns up to n distinct random elements from pool, without
// replacement, mirroring Python's random.sample.
func sampleDistinct(pool []string, n int) []string {
	if n > len(pool) {
		n = len(pool)
	}
	perm := rand.Perm(len(pool))[:n]
	out := make([]string, n)
	for i, idx := range perm {
		out[i] = pool[idx]
	}
	return out
}

// removeFirst returns a copy of s with the first occurrence of val removed.
func removeFirst(s []string, val string) []string {
	for i, v := range s {
		if v == val {
			out := make([]string, 0, len(s)-1)
			out = append(out, s[:i]...)
			out = append(out, s[i+1:]...)
			return out
		}
	}
	return s
}

func (m Model) updateQuestion(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Allow selecting an answer by number (1-based)
	n, err := strconv.Atoi(keyMsg.String())
	if err == nil && n >= 1 && n <= len(m.answerMenu.items) {
		m.answerMenu.cursor = n - 1
		keyMsg = tea.KeyMsg{Type: tea.KeyEnter} // continue as if "enter" was pressed
	}

	switch keyMsg.String() {
	case "up", "k":
		m.answerMenu.up()
	case "down", "j":
		m.answerMenu.down()
	case "enter":
		selected := m.answerMenu.selected().value
		correct := selected == m.quiz.correctAnswer

		adjustment := 0.15 * float64(m.quiz.numOptions)
		if err := m.store.UpdateRankingForWord(m.ctx, m.quiz.target.ID, m.locale, m.profile, correct, adjustment); err != nil {
			m.fail(err)
			return m, nil
		}

		if correct {
			m.quiz.questionsCorrect++
		} else {
			m.quiz.incorrectWords = append(m.quiz.incorrectWords, m.quiz.target)
		}
		m.quiz.questionedWordIDs = append(m.quiz.questionedWordIDs, m.quiz.target.ID)
		m.quiz.selectedAnswer = selected
		m.quiz.wasCorrect = correct
		m.screen = screenFeedback
	}
	return m, nil
}

func (m Model) viewQuestion() string {
	return m.answerMenu.view() + helpStyle.Render("\n↑/↓ to navigate • enter to select")
}

func (m Model) updateFeedback(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); !ok {
		return m, nil
	}
	if m.quiz.index >= m.quiz.totalQuestions {
		m.buildGoAgainMenu()
		m.screen = screenResults
		return m, nil
	}
	m.quiz.index++
	if err := m.loadNextQuestion(); err != nil {
		m.fail(err)
	}
	return m, nil
}

func (m Model) viewFeedback() string {
	var b strings.Builder
	if m.quiz.wasCorrect {
		b.WriteString(successStyle.Render("Correct!"))
	} else {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Wrong! You selected %q, the correct answer was %q.", m.quiz.selectedAnswer, m.quiz.correctAnswer)))
	}
	b.WriteString(helpStyle.Render("\n\npress enter to continue"))
	return b.String()
}

func (m *Model) buildGoAgainMenu() {
	m.answerMenu = newChoiceList("Would you like to go again?", labeledItems("Yes", "No"))
}

func (m Model) updateResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Allow selecting an answer by number (1-based)
	n, err := strconv.Atoi(keyMsg.String())
	if err == nil && (n == 1 || n == 2) {
		m.answerMenu.cursor = n - 1
		keyMsg = tea.KeyMsg{Type: tea.KeyEnter} // continue as if "enter" was pressed
	}

	switch keyMsg.String() {
	case "up", "k":
		m.answerMenu.up()
	case "down", "j":
		m.answerMenu.down()
	case "enter":
		if m.answerMenu.selected().value == "Yes" {
			m.numQuestionsInput.SetValue("10")
			m.numQuestionsInput.Focus()
			m.numQuestionsErr = ""
			m.screen = screenNumQuestions
			return m, textinput.Blink
		}
		m.screen = screenGoodbye
	}
	return m, nil
}

func (m Model) viewResults() string {
	var b strings.Builder
	pct := int(math.Round(float64(m.quiz.questionsCorrect) / float64(m.quiz.totalQuestions) * 100))
	b.WriteString(titleStyle.Render(fmt.Sprintf("Quiz complete! You got %d/%d = %d%% correct.", m.quiz.questionsCorrect, m.quiz.totalQuestions, pct)))
	b.WriteString("\n\n")

	if len(m.quiz.incorrectWords) == 0 {
		b.WriteString(successStyle.Render("Looks like you know all the words given!"))
	} else {
		b.WriteString(promptStyle.Render("Here are the words you got wrong for review:"))
		b.WriteString("\n")
		for _, w := range m.quiz.incorrectWords {
			if m.locale.QuestionLang() == "en" {
				b.WriteString(fmt.Sprintf("    %s: %s\n", w.English, w.Spanish))
			} else {
				b.WriteString(fmt.Sprintf("    %s: %s\n", w.Spanish, w.English))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(m.answerMenu.view())
	b.WriteString(helpStyle.Render("\n↑/↓ to navigate • enter to select"))
	return b.String()
}
