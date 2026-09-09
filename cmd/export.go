/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/54L1M/snip/internal/snippet"
	"github.com/54L1M/snip/internal/store"
	"github.com/54L1M/snip/internal/ui"
)

var (
	exportAll     bool
	exportResolve bool
	exportOutput  string
	exportFormat  string
)

var exportCmd = &cobra.Command{
	Use:     "export [name ...] [key=value ...]",
	Aliases: []string{"share"},
	Short:   "Export snippets as JSON to share or back up.",
	Long: `Exports snippets as JSON that "snip import" can read back. Name the snippets
to export, pass --all, or give no names to pick them from a checklist.

By default each snippet is exported as its template, with {{placeholders}}
intact and without your last-used values. With --resolve the placeholders are
filled in instead, so what you share is the full, ready-to-run command:
key=value arguments set the values you choose, anything else takes the value
you used last, and you are prompted for whatever is left.

Output is JSON by default, which is what "snip import" reads. With --format sh
the commands are written as a plain shell script instead (one per snippet,
each under a comment with its name), handy for pasting into a chat. A script
can't be imported back.

Output goes to stdout unless -o names a file.`,
	Example: `  # Pick snippets from a checklist and print them
  snip export

  # Export named snippets to a file
  snip export deploy deploy-worker -o deploy.json

  # Everything, viewed with bat
  snip export --all | bat -l json

  # Full command with the values you used last
  snip export deploy --resolve

  # Full command with values you choose (implies --resolve)
  snip export deploy ns=staging hash=abc1234

  # Ready-to-paste shell script with the full commands
  snip export deploy deploy-worker --resolve --format sh`,
	Args:              cobra.ArbitraryArgs,
	ValidArgsFunction: completeExport,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Load()
		if err != nil {
			return err
		}
		if len(st.Snippets) == 0 {
			return fmt.Errorf("no snippets yet; add one with `snip add`")
		}

		if exportFormat != "json" && exportFormat != "sh" {
			return fmt.Errorf("unknown format %q (want json or sh)", exportFormat)
		}

		names, provided := parseExportArgs(args)
		switch {
		case exportAll && len(names) > 0:
			return fmt.Errorf("cannot combine --all with snippet names")
		case exportAll:
			names = sortedNames(st)
		case len(names) > 0:
			for _, n := range names {
				if _, ok := st.Snippets[n]; !ok {
					return fmt.Errorf("snippet %q not found", n)
				}
			}
		default:
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no snippets given; pass names or --all")
			}
			var ok bool
			names, ok, err = ui.RunMultiSelector(sortedNames(st), "Select snippets to export:")
			if err != nil {
				return err
			}
			if !ok || len(names) == 0 {
				fmt.Fprintln(os.Stderr, "No snippets selected.")
				return nil
			}
		}

		resolve := exportResolve || len(provided) > 0
		out := &snippet.Store{Snippets: make(map[string]*snippet.Snippet, len(names))}
		for _, n := range names {
			s := st.Snippets[n]
			e := &snippet.Snippet{Name: n, Description: s.Description, Template: s.Template}
			if resolve {
				rendered, ok, err := resolveForExport(s, provided)
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(os.Stderr, "Canceled.")
					return nil
				}
				e.Template = rendered
			}
			out.Snippets[n] = e
		}

		var data []byte
		if exportFormat == "sh" {
			data = renderSh(out)
		} else {
			data, err = json.MarshalIndent(out, "", "  ")
			if err != nil {
				return fmt.Errorf("could not encode snippets: %w", err)
			}
			data = append(data, '\n')
		}

		if exportOutput == "" || exportOutput == "-" {
			_, err := os.Stdout.Write(data)
			return err
		}
		if err := os.WriteFile(exportOutput, data, 0o644); err != nil {
			return fmt.Errorf("could not write %s: %w", exportOutput, err)
		}
		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("Exported %s to %s", plural(len(names), "snippet"), exportOutput)))
		return nil
	},
}

func init() {
	exportCmd.Flags().BoolVarP(&exportAll, "all", "a", false, "export every snippet")
	exportCmd.Flags().BoolVarP(&exportResolve, "resolve", "r", false, "fill in placeholders (last-used values, key=value args, or prompt)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "write to this file instead of stdout")
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "output format: json (importable) or sh (plain shell script)")
	_ = exportCmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions([]string{"json\timportable JSON", "sh\tplain shell script"}, cobra.ShellCompDirectiveNoFileComp))
}

// renderSh lays the snippets out as a shell script: each command under a
// comment naming it (and its description), separated by blank lines.
func renderSh(out *snippet.Store) []byte {
	var b strings.Builder
	for i, n := range sortedNames(out) {
		s := out.Snippets[n]
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("# " + n)
		if s.Description != "" {
			b.WriteString(": " + s.Description)
		}
		b.WriteString("\n" + strings.TrimRight(s.Template, "\n") + "\n")
	}
	return []byte(b.String())
}

// parseExportArgs splits args into snippet names (deduplicated, in order) and
// key=value pairs.
func parseExportArgs(args []string) (names []string, provided map[string]string) {
	provided = map[string]string{}
	seen := map[string]bool{}
	for _, a := range args {
		if k, v, ok := strings.Cut(a, "="); ok {
			provided[k] = v
			continue
		}
		if !seen[a] {
			seen[a] = true
			names = append(names, a)
		}
	}
	return names, provided
}

// exportValues chooses a value for each variable of s: a provided key=value
// wins, then the last-used value. Variables with neither are returned as missing.
func exportValues(s *snippet.Snippet, provided map[string]string) (values map[string]string, missing []string) {
	values = map[string]string{}
	for _, v := range s.Vars() {
		if val, ok := provided[v]; ok {
			values[v] = val
			continue
		}
		if val, ok := s.LastValues[v]; ok && val != "" {
			values[v] = val
			continue
		}
		missing = append(missing, v)
	}
	return values, missing
}

// resolveForExport renders s with exportValues, prompting for any variable that
// has neither a provided nor a last-used value. The bool is false if the user
// canceled the prompt. Without a terminal, a missing value is an error.
func resolveForExport(s *snippet.Snippet, provided map[string]string) (string, bool, error) {
	values, missing := exportValues(s, provided)
	if len(missing) > 0 {
		if !isatty.IsTerminal(os.Stdin.Fd()) {
			return "", false, fmt.Errorf("snippet %q: missing value(s) for %s and no terminal to prompt", s.Name, strings.Join(missing, ", "))
		}
		fmt.Fprintln(os.Stderr, ui.HeaderStyle.Render(s.Name))
		filled, ok, err := ui.RunForm(missing, nil)
		if err != nil || !ok {
			return "", false, err
		}
		for k, v := range filled {
			values[k] = v
		}
	}
	rendered, err := snippet.Render(s.Template, values)
	if err != nil {
		return "", false, fmt.Errorf("snippet %q: %w", s.Name, err)
	}
	return rendered, true, nil
}

// completeExport completes `snip export`: snippet names not yet listed, plus
// `var=` for the variables of the snippets named so far.
func completeExport(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, provided := parseExportArgs(args)
	st, err := store.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	given := map[string]bool{}
	for _, n := range names {
		given[n] = true
	}
	var nameComps []string
	for n, s := range st.Snippets {
		if given[n] || !strings.HasPrefix(n, toComplete) {
			continue
		}
		if s.Description != "" {
			nameComps = append(nameComps, n+"\t"+s.Description)
		} else {
			nameComps = append(nameComps, n)
		}
	}
	sort.Strings(nameComps)

	seen := map[string]bool{}
	var varComps []string
	for _, n := range names {
		s, ok := st.Snippets[n]
		if !ok {
			continue
		}
		for _, v := range s.Vars() {
			if _, done := provided[v]; done || seen[v] || !strings.HasPrefix(v, toComplete) {
				continue
			}
			seen[v] = true
			varComps = append(varComps, v+"=")
		}
	}
	sort.Strings(varComps)

	comps := append(nameComps, varComps...)
	directive := cobra.ShellCompDirectiveNoFileComp
	// Only when nothing but `var=` candidates remain can we safely suppress the
	// trailing space so the value can be typed right after the '='.
	if len(nameComps) == 0 && len(varComps) > 0 {
		directive |= cobra.ShellCompDirectiveNoSpace
	}
	return comps, directive
}
