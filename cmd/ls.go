/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/54L1M/snip/internal/store"
	"github.com/54L1M/snip/internal/ui"
)

var lsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List saved snippets.",
	Long:    `Lists every saved snippet with its description.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Load()
		if err != nil {
			return err
		}
		if len(st.Snippets) == 0 {
			fmt.Println(ui.WarningStyle.Render("No snippets yet. Add one with `snip add <name> --cmd \"...\"`."))
			return nil
		}

		names := make([]string, 0, len(st.Snippets))
		for name := range st.Snippets {
			names = append(names, name)
		}
		sort.Strings(names)

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ui.Gray)).
			Headers("NAME", "DESCRIPTION").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return ui.HeaderStyle.Padding(0, 1)
				}
				if row%2 == 0 {
					return ui.EvenRowStyle
				}
				return ui.OddRowStyle
			})

		for _, name := range names {
			t.Row(name, st.Snippets[name].Description)
		}

		fmt.Println(t)
		return nil
	},
}
