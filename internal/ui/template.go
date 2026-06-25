/*
Copyright © 2026 54L1M
*/
package ui

import (
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// flagGap matches the run of two or more spaces cobra inserts between a flag's
// spec and its description (the column gap that keeps descriptions aligned).
var flagGap = regexp.MustCompile(` {2,}`)

const helpTemplate = `
{{styleCmdDesc .Long}}

{{styleHeader "USAGE"}}
  {{.UseLine}}{{if .HasAvailableSubCommands}}

{{styleHeader "COMMANDS"}}{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding}} {{.Short | styleSubtle}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{styleHeader "FLAGS"}}
{{.LocalFlags.FlagUsages | styleFlags}}{{end}}{{if .HasAvailableInheritedFlags}}

{{styleHeader "GLOBAL FLAGS"}}
{{.InheritedFlags.FlagUsages | styleFlags}}{{end}}{{if .HasExample}}

{{styleHeader "EXAMPLES"}}
{{styleExample .Example}}{{end}}

`

var templateFuncs = template.FuncMap{
	"styleHeader":  helpHeaderStyle.Render,
	"styleCmdDesc": helpCmdDescStyle.Render,
	"styleSubtle":  helpSubtleStyle.Render,
	"styleExample": helpExampleStyle.Render,
	"styleFlags": func(flagUsages string) string {
		lines := strings.Split(strings.TrimRight(flagUsages, "\n"), "\n")
		var styledLines []string

		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}

			// Preserve cobra's own leading indentation so flags with and without
			// a shorthand stay aligned in the same column.
			body := strings.TrimLeft(line, " ")
			indent := line[:len(line)-len(body)]

			// The flag spec and its description are separated by the first run of
			// 2+ spaces; splitting there keeps cobra's column alignment intact and
			// lets us color only the description.
			if loc := flagGap.FindStringIndex(body); loc != nil {
				flagPart := body[:loc[0]]
				gap := body[loc[0]:loc[1]]
				descPart := helpSubtleStyle.Render(body[loc[1]:])
				styledLines = append(styledLines, indent+flagPart+gap+descPart)
			} else {
				styledLines = append(styledLines, indent+body)
			}
		}
		return strings.Join(styledLines, "\n")
	},
}

// ApplyCustomHelpTemplate installs the styled help template on cmd.
func ApplyCustomHelpTemplate(cmd *cobra.Command) {
	cobra.AddTemplateFuncs(templateFuncs)
	cmd.SetHelpTemplate(helpTemplate)
}
