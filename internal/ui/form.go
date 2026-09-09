/*
Copyright © 2026 54L1M
*/
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

type formModel struct {
	vars     []string
	inputs   []textinput.Model
	focus    int
	done     bool
	canceled bool
}

func newFormModel(vars []string, defaults map[string]string) formModel {
	inputs := make([]textinput.Model, len(vars))
	for i, v := range vars {
		ti := textinput.New()
		ti.Prompt = ""
		ti.SetValue(defaults[v])
		ti.CursorEnd()
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return formModel{vars: vars, inputs: inputs}
}

func (m formModel) Init() tea.Cmd { return textinput.Blink }

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "enter", "tab", "down":
			if m.focus == len(m.inputs)-1 && km.String() == "enter" {
				m.done = true
				return m, tea.Quit
			}
			m.advance(1)
			return m, nil
		case "shift+tab", "up":
			m.advance(-1)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

func (m *formModel) advance(delta int) {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + delta + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	m.inputs[m.focus].CursorEnd()
}

func (m formModel) View() string {
	if m.done || m.canceled {
		return ""
	}
	var b strings.Builder
	b.WriteString(formHintStyle.Render("Fill in the values (enter/tab to move, enter on the last field to submit, esc to cancel)"))
	b.WriteString("\n\n")
	for i, in := range m.inputs {
		cursor := "  "
		if i == m.focus {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, formLabelStyle.Render(m.vars[i]), in.View()))
	}
	return b.String()
}

// RunForm prompts for each variable with its default pre-filled and returns the
// collected values. The bool is false if the user canceled (esc/ctrl+c).
func RunForm(vars []string, defaults map[string]string) (map[string]string, bool, error) {
	p := tea.NewProgram(newFormModel(vars, defaults), tea.WithOutput(os.Stderr))
	final, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	fm := final.(formModel)
	if fm.canceled {
		return nil, false, nil
	}
	values := make(map[string]string, len(vars))
	for i, v := range vars {
		values[v] = fm.inputs[i].Value()
	}
	return values, true, nil
}

// RenderCommandBox returns the resolved command framed in a styled box. When
// the command is wider than the terminal it is wrapped inside the box, so the
// border isn't broken up by the terminal's own line wrapping.
func RenderCommandBox(command string) string {
	width := 0
	if w, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
		width = w
	}
	return renderCommandBox(command, width)
}

// renderCommandBox frames command, wrapping it to fit a terminal termWidth
// columns wide. A termWidth of 0 (unknown, e.g. not a TTY) disables wrapping.
func renderCommandBox(command string, termWidth int) string {
	style := CommandBoxStyle
	// Border (2) + padding (2) around the content; leave one spare column so a
	// line filling the terminal exactly doesn't trigger an extra wrap.
	const frame = 2 + 2 + 1
	if termWidth > frame && lipgloss.Width(command)+frame > termWidth {
		// Width covers content plus padding; the border is added on top.
		style = style.Width(termWidth - 3)
	}
	return style.Render(command)
}

// Confirm prints a [y/N] prompt and returns true only on an affirmative answer.
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
