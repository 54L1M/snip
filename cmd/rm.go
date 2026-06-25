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

var rmYes bool

var rmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove", "delete"},
	Short:   "Delete a snippet.",
	Long:              `Deletes a saved snippet. Asks for confirmation unless -y is given.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSnippetName,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		st, err := store.Load()
		if err != nil {
			return err
		}
		if _, ok := st.Snippets[name]; !ok {
			return fmt.Errorf("snippet %q not found", name)
		}

		if !rmYes && !ui.Confirm(fmt.Sprintf("Delete snippet %q?", name)) {
			fmt.Println("Aborted.")
			return nil
		}

		delete(st.Snippets, name)
		if err := store.Save(st); err != nil {
			return err
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("Deleted snippet %q", name)))
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmYes, "yes", "y", false, "delete without confirmation")
}
