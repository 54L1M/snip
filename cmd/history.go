/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// zshPrefix matches the EXTENDED_HISTORY metadata that zsh prepends to each
// entry, e.g. ": 1700000000:0;git commit".
var zshPrefix = regexp.MustCompile(`^: \d+:\d+;`)

// bashTimestamp matches the "#<epoch>" lines bash writes when HISTTIMEFORMAT is set.
var bashTimestamp = regexp.MustCompile(`^#\d+$`)

// histFilePath locates the shell history file: $HISTFILE if exported, otherwise
// a per-shell default based on $SHELL, falling back to whichever of the common
// files exists.
func histFilePath() string {
	if hf := strings.TrimSpace(os.Getenv("HISTFILE")); hf != "" {
		return hf
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	zsh := filepath.Join(home, ".zsh_history")
	bash := filepath.Join(home, ".bash_history")

	switch filepath.Base(os.Getenv("SHELL")) {
	case "zsh":
		return zsh
	case "bash", "sh":
		return bash
	}
	// Unknown shell: prefer whichever file actually exists.
	if _, err := os.Stat(zsh); err == nil {
		return zsh
	}
	return bash
}

// lastShellCommand returns the most recent command from the shell history file,
// skipping blank lines and snip's own invocations.
func lastShellCommand() (string, error) {
	path := histFilePath()
	if path == "" {
		return "", fmt.Errorf("could not determine your shell history file (set $HISTFILE)")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read shell history (%s): %w", path, err)
	}

	cmds := parseHistory(data)
	for i := len(cmds) - 1; i >= 0; i-- {
		c := strings.TrimSpace(cmds[i])
		if c == "" || isSnipInvocation(c) {
			continue
		}
		return c, nil
	}
	return "", fmt.Errorf("no recent command found in %s", path)
}

// parseHistory extracts commands from a zsh or bash history file in order. It
// strips zsh EXTENDED_HISTORY prefixes and bash timestamp lines, and joins
// backslash-continued multi-line commands.
func parseHistory(data []byte) []string {
	var cmds []string
	var buf strings.Builder
	cont := false

	for _, raw := range strings.Split(string(data), "\n") {
		line := raw
		if !cont {
			if bashTimestamp.MatchString(line) {
				continue
			}
			line = zshPrefix.ReplaceAllString(line, "")
		}

		if strings.HasSuffix(line, "\\") {
			buf.WriteString(strings.TrimSuffix(line, "\\"))
			buf.WriteString("\n")
			cont = true
			continue
		}

		buf.WriteString(line)
		if s := buf.String(); strings.TrimSpace(s) != "" {
			cmds = append(cmds, s)
		}
		buf.Reset()
		cont = false
	}
	if s := buf.String(); strings.TrimSpace(s) != "" {
		cmds = append(cmds, s)
	}
	return cmds
}

// isSnipInvocation reports whether a history entry is a call to snip itself, so
// `snip add --last` doesn't capture its own command.
func isSnipInvocation(cmd string) bool {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return false
	}
	return filepath.Base(fields[0]) == "snip"
}
