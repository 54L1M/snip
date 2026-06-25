/*
Copyright © 2026 54L1M
*/
package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item string

func (i item) FilterValue() string { return string(i) }

var (
	selTitleStyle    = lipgloss.NewStyle().MarginLeft(2)
	itemStyle        = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(Blue)
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(tea.Msg, *list.Model) tea.Cmd     { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type selectorModel struct {
	list     list.Model
	Selected string
}

func newSelectorModel(choices []string, title string) selectorModel {
	items := make([]list.Item, len(choices))
	for i, choice := range choices {
		items[i] = item(choice)
	}

	l := list.New(items, itemDelegate{}, 20, 14)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = selTitleStyle
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	return selectorModel{list: l}
}

func (m selectorModel) Init() tea.Cmd { return nil }

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.Selected = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectorModel) View() string {
	if m.Selected != "" {
		return ""
	}
	return "\n" + m.list.View()
}

// RunSelector shows a filterable list and returns the chosen value. An empty
// string means the user quit without selecting.
func RunSelector(choices []string, title string) (string, error) {
	p := tea.NewProgram(newSelectorModel(choices, title))
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	return finalModel.(selectorModel).Selected, nil
}
