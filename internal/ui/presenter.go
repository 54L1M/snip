/*
Copyright © 2026 54L1M
*/
package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// wordmark is the "snip" block-letter logo (figlet "standard").
const wordmark = `           _
 ___ _ __ (_)_ __
/ __| '_ \| | '_ \
\__ \ | | | | |_) |
|___/_| |_|_| .__/
            |_|    `

// scissors is an ASCII scissors snipping to the right of the wordmark. It uses a
// quoted string (not a raw literal) because the art contains backticks.
const scissors = "    _       ,/'\n" +
	"   (_).  ,/'\n" +
	"   __  ::\n" +
	"  (__)'  `\\.\n" +
	"            `\\."

// DisplayWelcomeScreen prints the snip banner: the wordmark with scissors beside
// it, a short description, and a hint, framed in a rounded box.
func DisplayWelcomeScreen() {
	logoStyle := lipgloss.NewStyle().Foreground(Primary).Bold(true)
	scissorsStyle := lipgloss.NewStyle().Foreground(Yellow)

	banner := lipgloss.JoinHorizontal(
		lipgloss.Center,
		scissorsStyle.Render(scissors),
		"  ",
		logoStyle.Render(wordmark),
	)

	descStyle := lipgloss.NewStyle().Foreground(Gray).Align(lipgloss.Center)
	description := descStyle.Render("Save long, parameterized commands as snippets and run them with ease.")
	hint := descStyle.Render("\nUse 'snip --help' to see all available commands.")

	content := lipgloss.JoinVertical(lipgloss.Center, banner, description, hint)

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(1, 4).
		Align(lipgloss.Center)

	fmt.Println(containerStyle.Render(content))
}
