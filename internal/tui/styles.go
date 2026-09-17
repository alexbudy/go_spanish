package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle             = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtleStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cursorStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	selectedStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	selectedStyleSecondary = lipgloss.NewStyle() // Style for non-primary text in selected item
	errorStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	helpStyle              = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	promptStyle            = lipgloss.NewStyle().Bold(true)
)
