/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/54L1M/snip/internal/snippet"
	"github.com/54L1M/snip/internal/store"
	"github.com/54L1M/snip/internal/ui"
)

var (
	importForce  bool
	importSelect bool
)

var importCmd = &cobra.Command{
	Use:   "import [file] [name ...]",
	Short: "Import snippets from a snip export file.",
	Long: `Imports snippets from JSON produced by "snip export" (or a snippets.json
backup). Reads the given file, or stdin when the file is "-" or omitted while
input is piped in.

Every snippet in the file is imported unless you list names after the file, or
pass --select to pick them from a checklist. Snippets whose names you already
have are skipped unless --force is given.`,
	Example: `  # Import everything from a file
  snip import deploy.json

  # Import only some of them
  snip import deploy.json deploy deploy-worker

  # Pick from a checklist
  snip import deploy.json --select

  # From stdin
  pbpaste | snip import
  curl -s https://example.com/snips.json | snip import -

  # Overwrite snippets you already have
  snip import deploy.json --force`,
	Args:              cobra.ArbitraryArgs,
	ValidArgsFunction: completeImport,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			src       string
			names     []string
			fromStdin bool
		)
		switch {
		case len(args) == 0:
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no file given; pass a path, or pipe JSON on stdin")
			}
			fromStdin = true
		case args[0] == "-":
			fromStdin = true
			names = args[1:]
		default:
			src = args[0]
			names = args[1:]
		}
		if importSelect && len(names) > 0 {
			return fmt.Errorf("cannot combine --select with snippet names")
		}
		if importSelect && fromStdin {
			return fmt.Errorf("cannot pick interactively while reading from stdin; save the JSON to a file first")
		}

		var (
			data  []byte
			err   error
			label = src
		)
		if fromStdin {
			label = "stdin"
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(src)
		}
		if err != nil {
			return fmt.Errorf("could not read %s: %w", label, err)
		}

		in, err := parseImport(data)
		if err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
		available := sortedNames(in)

		switch {
		case len(names) > 0:
			for _, n := range names {
				if _, ok := in.Snippets[n]; !ok {
					return fmt.Errorf("snippet %q not found in %s", n, label)
				}
			}
		case importSelect:
			var ok bool
			names, ok, err = ui.RunMultiSelector(available, "Select snippets to import:")
			if err != nil {
				return err
			}
			if !ok || len(names) == 0 {
				fmt.Println("No snippets selected.")
				return nil
			}
		default:
			names = available
		}

		st, err := store.Load()
		if err != nil {
			return err
		}
		imported, skipped := mergeImport(st, in, names, importForce)
		if len(imported) > 0 {
			if err := store.Save(st); err != nil {
				return err
			}
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("Imported %s: %s", plural(len(imported), "snippet"), strings.Join(imported, ", "))))
		}
		if len(skipped) > 0 {
			fmt.Println(ui.WarningStyle.Render(fmt.Sprintf("Skipped %s you already have: %s\nUse --force to overwrite.", plural(len(skipped), "snippet"), strings.Join(skipped, ", "))))
		}
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVarP(&importForce, "force", "f", false, "overwrite snippets that already exist")
	importCmd.Flags().BoolVarP(&importSelect, "select", "s", false, "pick which snippets to import from a checklist")
}

// parseImport decodes an export document and validates it: at least one
// snippet, each with a name and a template. Snippet names come from the map
// keys, matching how the store is laid out.
func parseImport(data []byte) (*snippet.Store, error) {
	var in snippet.Store
	if err := json.Unmarshal(data, &in); err != nil {
		return nil, fmt.Errorf("not a valid snip export: %w", err)
	}
	if len(in.Snippets) == 0 {
		return nil, fmt.Errorf("no snippets found")
	}
	for name, s := range in.Snippets {
		if strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("a snippet has an empty name")
		}
		if s == nil || strings.TrimSpace(s.Template) == "" {
			return nil, fmt.Errorf("snippet %q has no template", name)
		}
		s.Name = name
	}
	return &in, nil
}

// mergeImport copies the named snippets from src into dst. Existing names are
// skipped unless force is set; when a forced overwrite carries no last-used
// values of its own, the local ones are kept as defaults.
func mergeImport(dst, src *snippet.Store, names []string, force bool) (imported, skipped []string) {
	for _, n := range names {
		s := src.Snippets[n]
		existing, exists := dst.Snippets[n]
		if exists && !force {
			skipped = append(skipped, n)
			continue
		}
		ns := &snippet.Snippet{
			Name:        n,
			Description: s.Description,
			Template:    s.Template,
			LastValues:  s.LastValues,
		}
		if ns.LastValues == nil && exists {
			ns.LastValues = existing.LastValues
		}
		dst.Snippets[n] = ns
		imported = append(imported, n)
	}
	return imported, skipped
}

// completeImport completes `snip import`: a file path first, then the names of
// the snippets inside that file that haven't been listed yet.
func completeImport(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}
	if args[0] == "-" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	in, err := parseImport(data)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	given := map[string]bool{}
	for _, n := range args[1:] {
		given[n] = true
	}
	var comps []string
	for _, n := range sortedNames(in) {
		if given[n] {
			continue
		}
		if d := in.Snippets[n].Description; d != "" {
			comps = append(comps, n+"\t"+d)
		} else {
			comps = append(comps, n)
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}
