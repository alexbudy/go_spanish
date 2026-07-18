// Command go_spanish is a terminal vocabulary trainer for Spanish/English
// nouns, backed by a local SQLite database.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"go_spanish/internal/store"
	"go_spanish/internal/tui"
)

const defaultDBPath = "my_db.db"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := runInit(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "path to the SQLite database file")
	dropDB := fs.Bool("drop-db", false, "delete the existing database file before initializing")
	addProfile := fs.String("add-profile", "", "create a profile with this name after initializing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *dropDB {
		if err := os.Remove(*dbPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing existing database: %w", err)
		}
	}

	s, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer s.Close()

	ctx := context.Background()
	if err := s.EnsureSchema(ctx); err != nil {
		return err
	}
	if err := s.SeedNouns(ctx); err != nil {
		return err
	}

	if *addProfile != "" {
		if _, err := s.InitProfile(ctx, *addProfile); err != nil {
			return fmt.Errorf("adding profile %q: %w", *addProfile, err)
		}
	}

	fmt.Printf("Database ready at %s\n", *dbPath)
	return nil
}

func run(args []string) error {
	fs := flag.NewFlagSet("go_spanish", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "path to the SQLite database file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer s.Close()

	ctx := context.Background()
	if err := s.EnsureSchema(ctx); err != nil {
		return err
	}
	if err := s.SeedNouns(ctx); err != nil {
		return err
	}

	p := tea.NewProgram(tui.New(s))
	_, err = p.Run()
	return err
}
