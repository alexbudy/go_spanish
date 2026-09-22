package tui

import "github.com/charmbracelet/lipgloss"

func scoreToKnowledgeLevelStyle(score float64) lipgloss.Style {
	if score > 1.5 {
		return knowWordVeryWellStyle
	} else if score > 0.5 {
		return knowWordWellStyle
	} else if score > -0.5 {
		return knowWordNeutralStyle
	} else if score > -1.5 {
		return knowWordPoorlyStyle
	} else {
		return knowWordVeryPoorlyStyle
	}
}
