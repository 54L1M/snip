/*
Copyright © 2026 54L1M
*/
package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	checkedStyle = lipgloss.NewStyle().Foreground(Green)
	cursorStyle  = lipgloss.NewStyle().Foreground(Blue)
)

// multiDelegate renders each row as a checkbox followed by the item name.
type multiDelegate struct {
	selected map[string]bool
}

func (d multiDelegate) Height() int                         { return 1 }
func (d multiDelegate) Spacing() int                        { return 0 }
func (d multiDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (d multiDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	name := string(i)

	box := "[ ]"
	if d.selected[name] {
		box = checkedStyle.Render("[x]")
	}

	// Styles are applied per segment (not to the whole line) so the checkbox
	// color doesn't reset the highlight of the rest of the row.
	cursor := "  "
	if index == m.Index() {
		cursor = cursorStyle.Render("> ")
		name = cursorStyle.Render(name)
	}
	fmt.Fprint(w, "  "+cursor+box+" "+name)
}

type multiSelectModel struct {
	list     list.Model
	order    []string
	selected map[string]bool
	done     bool
	canceled bool
}

func newMultiSelectModel(choices []string, title string) multiSelectModel {
	items := make([]list.Item, len(choices))
	for i, c := range choices {
		items[i] = item(c)
	}
	selected := map[string]bool{}

	l := list.New(items, multiDelegate{selected: selected}, 20, 14)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = selTitleStyle
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle all")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		}
	}

	return multiSelectModel{list: l, order: choices, selected: selected}
}

func (m multiSelectModel) Init() tea.Cmd { return nil }

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		// While the user is typing a filter, every key belongs to the filter input.
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m.canceled = true
			return m, tea.Quit
		case "esc":
			// With a filter applied, esc clears it (handled by the list); only
			// cancel when there is nothing to clear.
			if m.list.FilterState() == list.Unfiltered {
				m.canceled = true
				return m, tea.Quit
			}
		case " ", "x":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selected[string(i)] = !m.selected[string(i)]
			}
			return m, nil
		case "a":
			m.toggleVisible()
			return m, nil
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// toggleVisible selects every currently visible (filtered) item, or clears them
// all if they are already all selected.
func (m *multiSelectModel) toggleVisible() {
	visible := m.list.VisibleItems()
	all := len(visible) > 0
	for _, it := range visible {
		if !m.selected[string(it.(item))] {
			all = false
			break
		}
	}
	for _, it := range visible {
		m.selected[string(it.(item))] = !all
	}
}

func (m multiSelectModel) View() string {
	if m.done || m.canceled {
		return ""
	}
	return "\n" + m.list.View()
}

// RunMultiSelector shows a filterable checklist and returns the chosen values in
// the order they were given. The bool is false if the user quit without
// confirming (q/esc/ctrl+c).
func RunMultiSelector(choices []string, title string) ([]string, bool, error) {
	p := tea.NewProgram(newMultiSelectModel(choices, title), tea.WithOutput(os.Stderr))
	final, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	m := final.(multiSelectModel)
	if m.canceled {
		return nil, false, nil
	}
	var chosen []string
	for _, c := range m.order {
		if m.selected[c] {
			chosen = append(chosen, c)
		}
	}
	return chosen, true, nil
}
