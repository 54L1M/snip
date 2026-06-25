/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/54L1M/snip/internal/store"
)

// completeSnippetNames returns saved snippet names (with descriptions as
// completion hints) for shell completion.
func completeSnippetNames() ([]string, cobra.ShellCompDirective) {
	st, err := store.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, 0, len(st.Snippets))
	for n, s := range st.Snippets {
		if s.Description != "" {
			names = append(names, n+"\t"+s.Description)
		} else {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeSnippetName completes a single snippet-name argument: it offers names
// only for the first positional arg (commands like show/edit/rm take exactly one).
func completeSnippetName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSnippetNames()
}

// completeRun completes `snip run`: the snippet name for the first argument,
// then `var=` for each variable not yet supplied.
func completeRun(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	name, provided := parseRunArgs(args)
	if name == "" {
		return completeSnippetNames()
	}

	st, err := store.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	s, ok := st.Snippets[name]
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Don't re-suggest a value the user is mid-typing (toComplete past the '=').
	var comps []string
	for _, v := range s.Vars() {
		if _, done := provided[v]; done {
			continue
		}
		if strings.HasPrefix(v, toComplete) {
			comps = append(comps, v+"=")
		}
	}
	// NoSpace so the cursor stays put after `var=` for the value to be typed.
	return comps, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}
