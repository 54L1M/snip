/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/54L1M/snip/internal/snippet"
	"github.com/54L1M/snip/internal/store"
	"github.com/54L1M/snip/internal/ui"
)

var editCmdDesc string

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a snippet's command template.",
	Long: `Opens the snippet's command template in your editor ($EDITOR). Save and
close the editor to update it. Use -d to also update the description.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSnippetName,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Load()
		if err != nil {
			return err
		}
		s, ok := st.Snippets[args[0]]
		if !ok {
			return fmt.Errorf("snippet %q not found", args[0])
		}

		edited, err := openEditor(s.Template)
		if err != nil {
			return err
		}
		if strings.TrimSpace(edited) == "" {
			return fmt.Errorf("empty template; snippet not changed")
		}
		s.Template = edited
		if cmd.Flags().Changed("desc") {
			s.Description = editCmdDesc
		}

		if err := store.Save(st); err != nil {
			return err
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("Updated snippet %q", s.Name)))
		if vars := snippet.Vars(s.Template); len(vars) > 0 {
			fmt.Printf("Variables: %s\n", strings.Join(vars, ", "))
		}
		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&editCmdDesc, "desc", "d", "", "update the description too")
}
