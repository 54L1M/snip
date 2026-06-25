/*
Copyright © 2026 54L1M
*/

// Package ui holds the lipgloss styling, cobra help template, and the bubbletea
// components (snippet picker and variable form) used across snip.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	Primary   = lipgloss.Color("#3b82f6")
	Blue      = lipgloss.Color("#00bfda")
	Gray      = lipgloss.Color("#8a8a8a")
	LightGray = lipgloss.Color("#c6c6c6")
	Orange    = lipgloss.Color("202")
	Yellow    = lipgloss.Color("226")
	Green     = lipgloss.Color("#22c55e")

	ErrorStyle   = lipgloss.NewStyle().Foreground(Orange)
	SuccessStyle = lipgloss.NewStyle().Foreground(Primary)
	WarningStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder(), true).
			BorderForeground(Yellow).Padding(0, 1)

	// CommandBoxStyle frames the resolved command shown before execution.
	CommandBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(Primary).
			Foreground(Green).
			Padding(0, 1)

	// Table styles, mirrored from i2c.
	HeaderStyle  = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	CellStyle    = lipgloss.NewStyle().Padding(0, 1)
	OddRowStyle  = CellStyle.Foreground(Gray)
	EvenRowStyle = CellStyle.Foreground(LightGray)

	// Help template styles.
	helpHeaderStyle  = lipgloss.NewStyle().Foreground(Blue).Bold(true)
	helpCmdDescStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true).MarginLeft(2)
	helpSubtleStyle  = lipgloss.NewStyle().Foreground(Gray)
	helpExampleStyle = lipgloss.NewStyle().MarginLeft(2)

	// Form styles.
	formLabelStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	formHintStyle  = lipgloss.NewStyle().Foreground(Gray)
)
