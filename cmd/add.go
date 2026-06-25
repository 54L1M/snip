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

const editorSeed = `# Write the command on the lines below. Use {{name}} for parts that change,
# e.g. --namespace={{namespace}} or :{{hash}}. Lines starting with # are ignored.
`

var (
	addCmdTemplate string
	addCmdDesc     string
	addCmdForce    bool
	addCmdLast     bool
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Save a new command snippet.",
	Long: `Save a command as a named snippet. Mark the parts that change with
{{placeholders}} so they can be filled in at run time.

Pass the command inline with --cmd, or omit it to compose the command in your
editor ($EDITOR). With --last, the command you most recently ran in your shell
is loaded into the editor so you can turn its changing parts into placeholders.`,
	Example: `  # Inline
  snip add deploy --cmd 'docker run -d --name {{app}} -p {{port}}:8080 -e ENV={{env}} registry.example.com/{{app}}:{{tag}}' -d "Run a service container"

  # Compose in your editor
  snip add deploy

  # Start from the last command you ran, then edit it into a template
  snip add deploy --last`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		st, err := store.Load()
		if err != nil {
			return err
		}
		if _, exists := st.Snippets[name]; exists && !addCmdForce {
			return fmt.Errorf("snippet %q already exists (use --force to overwrite)", name)
		}

		template := addCmdTemplate
		switch {
		case addCmdLast:
			last, err := lastShellCommand()
			if err != nil {
				return err
			}
			edited, err := openEditor(editorSeed + last + "\n")
			if err != nil {
				return err
			}
			template = stripComments(edited)
		case strings.TrimSpace(template) == "":
			edited, err := openEditor(editorSeed)
			if err != nil {
				return err
			}
			template = stripComments(edited)
		}
		if strings.TrimSpace(template) == "" {
			return fmt.Errorf("no command provided; snippet not saved")
		}

		st.Snippets[name] = &snippet.Snippet{
			Name:        name,
			Description: addCmdDesc,
			Template:    template,
		}
		if err := store.Save(st); err != nil {
			return err
		}

		vars := snippet.Vars(template)
		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("Saved snippet %q", name)))
		if len(vars) > 0 {
			fmt.Printf("Variables: %s\n", strings.Join(vars, ", "))
		}
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addCmdTemplate, "cmd", "", "the command template (omit to use $EDITOR)")
	addCmd.Flags().StringVarP(&addCmdDesc, "desc", "d", "", "a short description")
	addCmd.Flags().BoolVar(&addCmdForce, "force", false, "overwrite an existing snippet with the same name")
	addCmd.Flags().BoolVar(&addCmdLast, "last", false, "seed the editor with the last command run in your shell")
	addCmd.MarkFlagsMutuallyExclusive("cmd", "last")
}

// stripComments removes blank-leading '#' comment lines used in the editor seed.
func stripComments(s string) string {
	var kept []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}
