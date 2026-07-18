package tui

import (
	"context"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"go_spanish/internal/store"
)

func enter() tea.Msg { return tea.KeyMsg{Type: tea.KeyEnter} }
func down() tea.Msg  { return tea.KeyMsg{Type: tea.KeyDown} }
func runes(s string) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func newTestModel(t *testing.T) Model {
	t.Helper()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	if err := s.EnsureSchema(ctx); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	if err := s.SeedNouns(ctx); err != nil {
		t.Fatalf("SeedNouns: %v", err)
	}

	return New(s)
}

func send(t *testing.T, m tea.Model, msgs ...tea.Msg) tea.Model {
	t.Helper()
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

// TestFullJourney drives the whole app: create a profile, pick a direction,
// answer every question, and reach the results screen, then decline to go
// again.
func TestFullJourney(t *testing.T) {
	mm := newTestModel(t)

	if mm.screen != screenProfileSelect {
		t.Fatalf("expected screenProfileSelect, got %v", mm.screen)
	}
	if got := mm.profileMenu.selected().value; got != newProfileValue {
		t.Fatalf("expected first item to be new-profile placeholder, got %q", got)
	}

	// Create a new profile named "tester".
	m := send(t, mm, enter())
	mm = m.(Model)
	if mm.screen != screenNewProfile {
		t.Fatalf("expected screenNewProfile, got %v", mm.screen)
	}
	m = send(t, mm, runes("tester"), enter())
	mm = m.(Model)
	if mm.screen != screenDirectionSelect {
		t.Fatalf("expected screenDirectionSelect, got %v", mm.screen)
	}
	if mm.profile != "tester" {
		t.Fatalf("expected profile 'tester', got %q", mm.profile)
	}

	// Pick "es_to_en" (first option).
	m = send(t, mm, enter())
	mm = m.(Model)
	if mm.screen != screenNumQuestions {
		t.Fatalf("expected screenNumQuestions, got %v", mm.screen)
	}
	if mm.locale != store.EsToEn {
		t.Fatalf("expected locale EsToEn, got %v", mm.locale)
	}

	// 3 questions, 3 options each.
	m = send(t, mm, runes("3"), enter())
	mm = m.(Model)
	if mm.screen != screenNumOptions {
		t.Fatalf("expected screenNumOptions, got %v", mm.screen)
	}
	m = send(t, mm, runes("3"), enter())
	mm = m.(Model)
	if mm.screen != screenQuestion {
		t.Fatalf("expected screenQuestion, got %v", mm.screen)
	}
	if mm.quiz.totalQuestions != 3 || mm.quiz.numOptions != 3 {
		t.Fatalf("expected 3 questions / 3 options, got %+v", mm.quiz)
	}
	if len(mm.answerMenu.items) != 3 {
		t.Fatalf("expected 3 answer choices, got %d", len(mm.answerMenu.items))
	}

	// Answer all 3 questions (always pick whatever is highlighted).
	for i := 1; i <= 3; i++ {
		if mm.screen != screenQuestion {
			t.Fatalf("question %d: expected screenQuestion, got %v", i, mm.screen)
		}
		m = send(t, mm, enter()) // submit answer -> feedback
		mm = m.(Model)
		if mm.screen != screenFeedback {
			t.Fatalf("question %d: expected screenFeedback, got %v", i, mm.screen)
		}
		m = send(t, mm, enter()) // continue
		mm = m.(Model)
	}

	if mm.screen != screenResults {
		t.Fatalf("expected screenResults, got %v", mm.screen)
	}
	if mm.quiz.questionsCorrect+len(mm.quiz.incorrectWords) != 3 {
		t.Fatalf("expected 3 total answered, got correct=%d incorrect=%d",
			mm.quiz.questionsCorrect, len(mm.quiz.incorrectWords))
	}

	// Decline to go again -> goodbye.
	m = send(t, mm, down(), enter())
	mm = m.(Model)
	if mm.screen != screenGoodbye {
		t.Fatalf("expected screenGoodbye, got %v", mm.screen)
	}
}

// TestGoAgainReturnsToNumQuestions verifies choosing "Yes" on the results
// screen loops back to the question-count prompt with the same profile and
// locale intact.
func TestGoAgainReturnsToNumQuestions(t *testing.T) {
	mm := newTestModel(t)
	m := send(t, mm, enter()) // new profile
	m = send(t, m, runes("looper"), enter())
	m = send(t, m, enter()) // pick direction
	m = send(t, m, runes("1"), enter())
	m = send(t, m, runes("2"), enter())

	mm = m.(Model)
	for i := 1; i <= 1; i++ {
		m = send(t, mm, enter(), enter())
		mm = m.(Model)
	}
	if mm.screen != screenResults {
		t.Fatalf("expected screenResults, got %v", mm.screen)
	}

	m = send(t, mm, enter()) // "Yes" is the default selection
	mm = m.(Model)
	if mm.screen != screenNumQuestions {
		t.Fatalf("expected screenNumQuestions after go-again, got %v", mm.screen)
	}
	if mm.profile != "looper" || mm.locale != store.EsToEn {
		t.Fatalf("expected profile/locale preserved, got profile=%q locale=%v", mm.profile, mm.locale)
	}
}

// TestNewProfileValidation checks empty and duplicate profile names are
// rejected without leaving the new-profile screen.
func TestNewProfileValidation(t *testing.T) {
	mm := newTestModel(t)
	m := send(t, mm, enter())
	mm = m.(Model)

	m = send(t, mm, enter()) // empty name
	mm = m.(Model)
	if mm.screen != screenNewProfile || mm.newProfileErr == "" {
		t.Fatalf("expected validation error on empty name, got screen=%v err=%q", mm.screen, mm.newProfileErr)
	}

	m = send(t, mm, runes("bad name!"), enter()) // non-alnum
	mm = m.(Model)
	if mm.screen != screenNewProfile || mm.newProfileErr == "" {
		t.Fatalf("expected validation error on invalid name, got screen=%v err=%q", mm.screen, mm.newProfileErr)
	}
}
