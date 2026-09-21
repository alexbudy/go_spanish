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
	promptStyleProfile     = lipgloss.NewStyle().Bold(true).Italic(true)

	// Deletion confirmation styles for emphasis
	deleteConfirmStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	deleteConfirmSpecialWordStyle = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("196"))

	// Page counter styles
	pageCounterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("247")).MarginTop(1)

	// Manage Words styles
	knowWordVeryWellStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("41"))
	knowWordWellStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	knowWordNeutralStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252505"))
	knowWordPoorlyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	knowWordVeryPoorlyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("124"))
)
