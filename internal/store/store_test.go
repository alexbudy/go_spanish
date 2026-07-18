package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStoreFlow(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if err := s.EnsureSchema(ctx); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	if err := s.SeedNouns(ctx); err != nil {
		t.Fatalf("SeedNouns: %v", err)
	}

	words, err := s.GetAllWords(ctx, "en")
	if err != nil {
		t.Fatalf("GetAllWords: %v", err)
	}
	if len(words) == 0 {
		t.Fatal("expected seeded words, got none")
	}

	if _, err := s.InitProfile(ctx, "tester"); err != nil {
		t.Fatalf("InitProfile: %v", err)
	}

	profiles, err := s.GetProfiles(ctx)
	if err != nil {
		t.Fatalf("GetProfiles: %v", err)
	}
	if len(profiles) != 1 || profiles[0] != "tester" {
		t.Fatalf("expected [tester], got %v", profiles)
	}

	qWords, err := s.GetWordsForQuestion(ctx, "tester", EsToEn, nil, 10)
	if err != nil {
		t.Fatalf("GetWordsForQuestion: %v", err)
	}
	if len(qWords) != 10 {
		t.Fatalf("expected 10 words, got %d", len(qWords))
	}

	target := qWords[0]
	if err := s.UpdateRankingForWord(ctx, target.ID, EsToEn, "tester", true, 0.6); err != nil {
		t.Fatalf("UpdateRankingForWord: %v", err)
	}

	excluded, err := s.GetWordsForQuestion(ctx, "tester", EsToEn, []int64{target.ID}, 10)
	if err != nil {
		t.Fatalf("GetWordsForQuestion (excluded): %v", err)
	}
	for _, w := range excluded {
		if w.ID == target.ID {
			t.Fatalf("expected word %d to be excluded", target.ID)
		}
	}
}
