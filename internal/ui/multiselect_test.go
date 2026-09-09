/*
Copyright © 2026 54L1M
*/
package ui

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func press(m multiSelectModel, keys ...string) multiSelectModel {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case " ":
			msg = tea.KeyMsg{Type: tea.KeySpace}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		next, _ := m.Update(msg)
		m = next.(multiSelectModel)
	}
	return m
}

func chosen(m multiSelectModel) []string {
	var out []string
	for _, c := range m.order {
		if m.selected[c] {
			out = append(out, c)
		}
	}
	return out
}

func TestMultiSelectToggleAndConfirm(t *testing.T) {
	m := newMultiSelectModel([]string{"a", "b", "c"}, "pick")
	m = press(m, " ", "down", "down", "x", "enter")
	if !m.done || m.canceled {
		t.Fatalf("expected done, got done=%v canceled=%v", m.done, m.canceled)
	}
	if got := chosen(m); !reflect.DeepEqual(got, []string{"a", "c"}) {
		t.Errorf("chosen = %v, want [a c]", got)
	}
}

func TestMultiSelectToggleTwiceUnselects(t *testing.T) {
	m := press(newMultiSelectModel([]string{"a"}, ""), " ", " ")
	if len(chosen(m)) != 0 {
		t.Errorf("expected nothing selected, got %v", chosen(m))
	}
}

func TestMultiSelectAll(t *testing.T) {
	m := press(newMultiSelectModel([]string{"a", "b", "c"}, ""), "a")
	if got := chosen(m); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("after 'a': %v", got)
	}
	m = press(m, "a")
	if got := chosen(m); len(got) != 0 {
		t.Errorf("second 'a' should clear all, got %v", got)
	}
}

func TestMultiSelectCancel(t *testing.T) {
	for _, k := range []string{"q", "esc"} {
		m := press(newMultiSelectModel([]string{"a"}, ""), " ", k)
		if !m.canceled || m.done {
			t.Errorf("%q: expected canceled, got done=%v canceled=%v", k, m.done, m.canceled)
		}
	}
}

func TestMultiSelectFilterTypingDoesNotToggle(t *testing.T) {
	// "/" starts filtering; keys typed afterwards go to the filter, not the
	// toggle/cancel bindings, until enter applies the filter.
	m := press(newMultiSelectModel([]string{"alpha", "beta"}, ""), "/", "a", "q", " ")
	if m.canceled || m.done || len(chosen(m)) != 0 {
		t.Errorf("filter typing leaked: done=%v canceled=%v chosen=%v", m.done, m.canceled, chosen(m))
	}
}
