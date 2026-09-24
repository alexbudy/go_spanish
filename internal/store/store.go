// Package store provides SQLite-backed persistence for words, profiles, and
// per-profile rankings, mirroring the schema used by the original Python
// spanish_buddy project.
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

//go:embed nouns.csv
var nounsCSV string

// QuizMode identifies which words to select for a quiz: well-known, least-known, or any.
type QuizMode string

const (
	// WellKnown selects words the user knows well.
	WellKnown  QuizMode = "well_known"
	Any        QuizMode = "any" // Any selects words regardless of the words ranking for the profile.
	LeastKnown QuizMode = "least_known"
)

// Locale identifies a training direction: which language the question is
// shown in, and which language the answer must be given in.
type Locale string

const (
	EsToEn Locale = "es_to_en"
	EnToEs Locale = "en_to_es"
	En     Locale = "en"
	Es     Locale = "es"
)

const (
	// ManageWords is a special profile name used to manage words.
	ManageWords = "manage_words"
)

// QuestionLang returns the language ("es" or "en") the target word is shown in.
func (l Locale) QuestionLang() string { return string(l)[:2] }

// AnswerLang returns the language ("es" or "en") the correct answer is in.
func (l Locale) AnswerLang() string { return string(l)[len(l)-2:] }

func (l Locale) column() (string, error) {
	switch l {
	case EsToEn, EnToEs:
		return string(l), nil
	default:
		return "", fmt.Errorf("store: unknown locale %q", l)
	}
}

// Word is a single Spanish/English noun pair.
type Word struct {
	ID          int64   // 1-based ind
	Spanish     string  // Spanish Word
	English     string  // English Word
	Gender      string  // TODO use/implement/delete?
	EsToEnScore float64 // how well the user knows spanish word from english
	EnToEsScore float64 // how well the user knows english word from spanish
}

// Text returns the word's text in the given locale's question language.
func (w Word) Text(lang string) string {
	if lang == "en" {
		return w.English
	}
	return w.Spanish
}

// Store wraps a SQLite database connection.
type Store struct {
	db *sql.DB
}

// Open opens (and if necessary creates) the SQLite database at path with
// foreign key enforcement turned on.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// EnsureSchema creates the nouns/profiles/rankings tables if they don't exist.
func (s *Store) EnsureSchema(ctx context.Context) error {
	for _, stmt := range strings.Split(schemaSQL, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("store: ensure schema: %w", err)
		}
	}

	if err := s.migrate(ctx); err != nil {
	}
	return nil
}

func (s *Store) migrate(ctx context.Context) error {
	var exists int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM pragma_table_info('profiles')
		WHERE name = 'enable_speech'
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("store: check some_new_column: %w", err)
	}

	// migration #1
	if exists == 0 {
		// add enable_speech to profile table
		_, err = s.db.ExecContext(ctx, `
			ALTER TABLE profiles
			ADD COLUMN enable_speech BOOLEAN NOT NULL DEFAULT TRUE
		`)
		if err != nil {
			return fmt.Errorf("store: add some_new_column: %w", err)
		}
	}

	// migration #2

	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM pragma_table_info('profiles')
		WHERE name = 'default_num_questions' OR name='default_num_answers'
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("store: check some_new_column: %w", err)
	}
	if exists == 0 {
		// add number of answers and questions default
		_, err = s.db.ExecContext(ctx, `
			ALTER TABLE profiles
			ADD COLUMN default_num_questions INT NOT NULL DEFAULT 10

		`)
		if err != nil {
			return fmt.Errorf("store: add some_new_column: %w", err)
		}

		_, err = s.db.ExecContext(ctx, `
			ALTER TABLE profiles
			ADD COLUMN default_num_answers INT NOT NULL DEFAULT 4
		`)
		if err != nil {
			return fmt.Errorf("store: add some_new_column: %w", err)
		}
	}

	return nil
}

// SeedNouns loads the embedded word list into the nouns table, unless it has
// already been populated.
func (s *Store) SeedNouns(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM nouns").Scan(&count); err != nil {
		return fmt.Errorf("store: count nouns: %w", err)
	}
	if count > 0 {
		return nil
	}

	reader := csv.NewReader(strings.NewReader(nounsCSV))
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("store: read nouns header: %w", err)
	}
	colIdx := make(map[string]int, len(header))
	for i, name := range header {
		colIdx[name] = i
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: seed nouns: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO nouns (spanish, english, gender) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("store: seed nouns: %w", err)
	}
	defer stmt.Close()

	for {
		row, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("store: read nouns row: %w", err)
		}
		spanish := row[colIdx["spanish"]]
		english := row[colIdx["english"]]
		gender := row[colIdx["gender"]]
		if _, err := stmt.ExecContext(ctx, spanish, english, gender); err != nil {
			return fmt.Errorf("store: insert noun %q: %w", english, err)
		}
	}

	return tx.Commit()
}

// InitProfile creates a new profile and seeds a zeroed ranking row for every
// existing noun.
func (s *Store) InitProfile(ctx context.Context, name string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("store: init profile: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, "INSERT INTO profiles (name) VALUES (?)", name)
	if err != nil {
		return 0, fmt.Errorf("store: insert profile %q: %w", name, err)
	}
	profileID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("store: init profile: %w", err)
	}

	rows, err := tx.QueryContext(ctx, "SELECT id FROM nouns")
	if err != nil {
		return 0, fmt.Errorf("store: init profile: %w", err)
	}
	var nounIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, fmt.Errorf("store: init profile: %w", err)
		}
		nounIDs = append(nounIDs, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("store: init profile: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO rankings (profile_id, noun_id) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("store: init profile: %w", err)
	}
	defer stmt.Close()
	for _, nounID := range nounIDs {
		if _, err := stmt.ExecContext(ctx, profileID, nounID); err != nil {
			return 0, fmt.Errorf("store: init profile: %w", err)
		}
	}

	return profileID, tx.Commit()
}

// DeleteProfile hard deletes rankings and the profile itself, returning the number of deleted ranking rows.
func (s *Store) DeleteProfile(ctx context.Context, name string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("store: delete profile: %w", err)
	}
	defer tx.Rollback()

	// hard delete rankings for the profile
	res, err := tx.ExecContext(ctx, "DELETE FROM rankings WHERE profile_id = (SELECT id FROM profiles WHERE name = ?)", name)
	if err != nil {
		return 0, fmt.Errorf("store: delete profile %q: %w", name, err)
	}

	deletedRows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("store: delete profile %q: %w", name, err)
	}

	// hard delete profile
	_, err = tx.ExecContext(ctx, "DELETE FROM profiles WHERE name = ?", name)
	if err != nil {
		return 0, fmt.Errorf("store: delete profile %q: %w", name, err)
	}

	return deletedRows, tx.Commit()
}

// GetProfiles returns up to 8 non-deleted profile names, oldest first.
func (s *Store) GetProfiles(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT name FROM profiles WHERE deleted_at IS NULL ORDER BY created_at ASC LIMIT 8")
	if err != nil {
		return nil, fmt.Errorf("store: get profiles: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("store: get profiles: %w", err)
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

type Profile struct {
	ID                  int64
	Name                string
	EnableSpeech        bool
	DefaultNumQuestions int
	DefaultNumAnswers   int
}

// GetProfile returns the profile matching the provided name.
func (s *Store) GetProfile(ctx context.Context, profile string) (Profile, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, enable_speech, default_num_questions, default_num_answers FROM profiles WHERE name = ?", profile)
	if err != nil {
		return Profile{}, fmt.Errorf("store: get profile: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return Profile{}, fmt.Errorf("store: get profile %q: %w", profile, err)
		}
		return Profile{}, fmt.Errorf("store: get profile %q: %w", profile, sql.ErrNoRows)
	}

	var p Profile
	if err := rows.Scan(&p.ID, &p.Name, &p.EnableSpeech, &p.DefaultNumQuestions, &p.DefaultNumAnswers); err != nil {
		return Profile{}, fmt.Errorf("store: get profile %q: %w", profile, err)
	}
	if err := rows.Err(); err != nil {
		return Profile{}, fmt.Errorf("store: get profile %q: %w", profile, err)
	}
	return p, nil
}

// UpdateProfile updates a profile based on p.name.
func (s *Store) UpdateProfile(ctx context.Context, p Profile) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE profiles SET enable_speech = ?, default_num_questions = ?, default_num_answers = ? WHERE name = ?",
		p.EnableSpeech, p.DefaultNumQuestions, p.DefaultNumAnswers, p.Name)
	return err
}

// UpdateProfileName updates the profile name for an existing profile.
func (s *Store) UpdateProfileName(ctx context.Context, oldName, newName string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE profiles SET name = ? WHERE name = ?", newName, oldName)
	return err
}

// GetAllWords returns every noun's text in both languages, as well as its rankings into the other direction,
// used to build the pool of multiple-choice answers.
func (s *Store) GetAllWords(ctx context.Context, profile string) ([]Word, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT n.spanish, n.english, r.es_to_en, r.en_to_es
		FROM nouns n
		INNER JOIN rankings r ON n.id = r.noun_id
		INNER JOIN profiles p ON p.id = r.profile_id
		WHERE p.name = ?`, profile)
	if err != nil {
		return nil, fmt.Errorf("store: get all words: %w", err)
	}
	defer rows.Close()

	var words []Word

	for rows.Next() {
		var word Word

		if err := rows.Scan(&word.Spanish, &word.English, &word.EsToEnScore, &word.EnToEsScore); err != nil {
			return nil, fmt.Errorf("store: get all words: %w", err)
		}

		word.ID = int64(len(words) + 1)

		words = append(words, word)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: get all words: %w", err)
	}

	return words, nil
}

// GetWordsForQuestion returns the words for profile in the given
// locale, selected by 'mode' (random, least/most known) excluding excludeWordIDs,
// ordered by ranking ascending with random tie-breaking, limited to numWords.
func (s *Store) GetWordsForQuestion(ctx context.Context, profile string, quizMode QuizMode, locale Locale, excludeWordIDs []int64, numWords int) ([]Word, error) {
	column, err := locale.column()
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	sb.WriteString(`SELECT n.id, n.spanish, n.english, n.gender
FROM nouns n
INNER JOIN rankings r ON n.id = r.noun_id
INNER JOIN profiles p ON p.id = r.profile_id
WHERE p.name = ?`)
	args := []any{profile}

	if len(excludeWordIDs) > 0 {
		placeholders := make([]string, len(excludeWordIDs))
		for i, id := range excludeWordIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		sb.WriteString(" AND n.id NOT IN (" + strings.Join(placeholders, ",") + ") ")
	}

	switch quizMode {
	case Any:
		sb.WriteString(" ORDER BY RANDOM() LIMIT ?")
	case LeastKnown:
		sb.WriteString(fmt.Sprintf(" ORDER BY %s ASC, RANDOM() LIMIT ?", column))
	case WellKnown:
		sb.WriteString(fmt.Sprintf(" ORDER BY %s DESC, RANDOM() LIMIT ?", column))
	}

	args = append(args, numWords)

	rows, err := s.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("store: get words for question: %w", err)
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var w Word
		if err := rows.Scan(&w.ID, &w.Spanish, &w.English, &w.Gender); err != nil {
			return nil, fmt.Errorf("store: get words for question: %w", err)
		}
		words = append(words, w)
	}
	return words, rows.Err()
}

// UpdateRankingForWord nudges a word's ranking for profile in the given
// locale up (correct) or down (incorrect) by rankingAdjustment.
func (s *Store) UpdateRankingForWord(ctx context.Context, wordID int64, locale Locale, profile string, gotCorrect bool, rankingAdjustment float64) error {
	column, err := locale.column()
	if err != nil {
		return err
	}
	if !gotCorrect {
		rankingAdjustment = -rankingAdjustment
	}

	var profileID int64
	if err := s.db.QueryRowContext(ctx, "SELECT id FROM profiles WHERE name = ?", profile).Scan(&profileID); err != nil {
		return fmt.Errorf("store: update ranking: lookup profile %q: %w", profile, err)
	}

	query := fmt.Sprintf("UPDATE rankings SET %s = %s + ? WHERE profile_id = ? AND noun_id = ?", column, column)
	if _, err := s.db.ExecContext(ctx, query, rankingAdjustment, profileID, wordID); err != nil {
		return fmt.Errorf("store: update ranking: %w", err)
	}
	return nil
}

// ResetRankingForWord resets a word's ranking to zero for the given profile and locale.
// wordID is 1-based id of word
func (s *Store) ResetRankingForWord(ctx context.Context, wordID int64, locale Locale, profile string) error {
	column, err := locale.column()
	if err != nil {
		return err
	}

	var profileID int64
	if err := s.db.QueryRowContext(ctx, "SELECT id FROM profiles WHERE name = ?", profile).Scan(&profileID); err != nil {
		return fmt.Errorf("store: update ranking: lookup profile %q: %w", profile, err)
	}

	query := fmt.Sprintf("UPDATE rankings SET %s = 0.0 WHERE profile_id = ? AND noun_id = ?", column)
	if _, err := s.db.ExecContext(ctx, query, profileID, wordID); err != nil {
		return fmt.Errorf("store: update ranking: %w", err)
	}
	return nil
}
