/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/54L1M/snip/internal/store"
	"github.com/54L1M/snip/internal/ui"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show a snippet's template and variables.",
	Long:  `Prints the full template for a snippet, its description, and the variables it declares with their remembered defaults.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Load()
		if err != nil {
			return err
		}
		s, ok := st.Snippets[args[0]]
		if !ok {
			return fmt.Errorf("snippet %q not found", args[0])
		}

		fmt.Println(ui.HeaderStyle.Render(s.Name))
		if s.Description != "" {
			fmt.Println(s.Description)
		}
		fmt.Println()
		fmt.Println(ui.RenderCommandBox(s.Template))

		vars := s.Vars()
		if len(vars) > 0 {
			fmt.Println()
			fmt.Println(ui.HeaderStyle.Render("VARIABLES"))
			for _, v := range vars {
				if def := s.LastValues[v]; def != "" {
					fmt.Printf("  %s (last: %s)\n", v, def)
				} else {
					fmt.Printf("  %s\n", v)
				}
			}
		}
		return nil
	},
}
