/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/54L1M/snip/internal/snippet"
	"github.com/54L1M/snip/internal/store"
	"github.com/54L1M/snip/internal/ui"
)

var (
	runYes   bool
	runPrint bool
)

var runCmd = &cobra.Command{
	Use:   "run [name] [key=value ...]",
	Short: "Resolve a snippet's variables and run it.",
	Long: `Resolves a snippet's {{placeholders}} and runs the result.

Supply values as key=value arguments; any not given are prompted for, pre-filled
with the value you used last time. With no name, a picker lets you choose a
snippet. The fully-resolved command is shown for confirmation before it runs.`,
	Example: `  # Provide some values, get prompted for the rest
  snip run deploy app=web tag=1.4.2

  # Pick interactively and fill everything in a form
  snip run

  # Just print the resolved command, don't run it
  snip run deploy app=web tag=1.4.2 --print`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dr, _ := cmd.Flags().GetBool("dry-run"); dr {
			runPrint = true
		}

		st, err := store.Load()
		if err != nil {
			return err
		}
		if len(st.Snippets) == 0 {
			return fmt.Errorf("no snippets yet; add one with `snip add`")
		}

		name, provided := parseRunArgs(args)

		if name == "" {
			name, err = pickSnippet(st)
			if err != nil {
				return err
			}
			if name == "" {
				fmt.Println("No snippet selected.")
				return nil
			}
		}

		s, ok := st.Snippets[name]
		if !ok {
			return fmt.Errorf("snippet %q not found", name)
		}

		values, ok, err := resolveValues(s, provided)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Canceled.")
			return nil
		}

		resolved, err := snippet.Render(s.Template, values)
		if err != nil {
			return err
		}

		// Persist the values used as next time's defaults.
		if s.LastValues == nil {
			s.LastValues = map[string]string{}
		}
		for k, v := range values {
			s.LastValues[k] = v
		}
		if err := store.Save(st); err != nil {
			return err
		}

		fmt.Println(ui.RenderCommandBox(resolved))

		if runPrint {
			return nil
		}

		if !runYes && !viper.GetBool("AUTO_CONFIRM") {
			if !ui.Confirm("Run this?") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		code, runErr := runCommand(resolved)
		if runErr != nil {
			// The command itself failed; surface its exit code without an extra
			// cobra usage dump.
			os.Exit(code)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVarP(&runYes, "yes", "y", false, "skip the confirmation prompt")
	runCmd.Flags().BoolVarP(&runPrint, "print", "p", false, "print the resolved command without running it")
	runCmd.Flags().Bool("dry-run", false, "alias for --print")
	_ = runCmd.Flags().MarkHidden("dry-run")
}

// parseRunArgs splits args into an optional snippet name and the key=value pairs.
// The first argument without an '=' is treated as the name.
func parseRunArgs(args []string) (name string, provided map[string]string) {
	provided = map[string]string{}
	for _, a := range args {
		if k, v, ok := strings.Cut(a, "="); ok {
			provided[k] = v
			continue
		}
		if name == "" {
			name = a
		}
	}
	return name, provided
}

func pickSnippet(st *snippet.Store) (string, error) {
	names := make([]string, 0, len(st.Snippets))
	for n := range st.Snippets {
		names = append(names, n)
	}
	sort.Strings(names)
	return ui.RunSelector(names, "Select a snippet:")
}

// resolveValues fills every variable of s: from provided args first, otherwise
// via an interactive form pre-filled with last-used values. In a non-interactive
// context, any unprovided variable is an error.
func resolveValues(s *snippet.Snippet, provided map[string]string) (map[string]string, bool, error) {
	vars := s.Vars()
	values := map[string]string{}
	var missing []string
	for _, v := range vars {
		if val, ok := provided[v]; ok {
			values[v] = val
		} else {
			missing = append(missing, v)
		}
	}
	if len(missing) == 0 {
		return values, true, nil
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return nil, false, fmt.Errorf("missing value(s) for %s and no terminal to prompt", strings.Join(missing, ", "))
	}

	defaults := make(map[string]string, len(missing))
	for _, v := range missing {
		defaults[v] = s.LastValues[v]
	}
	filled, ok, err := ui.RunForm(missing, defaults)
	if err != nil || !ok {
		return nil, false, err
	}
	for k, v := range filled {
		values[k] = v
	}
	return values, true, nil
}
