/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

// editorBinary resolves the editor to use: viper EDITOR, then $EDITOR, then
// $VISUAL, then "vi".
func editorBinary() string {
	for _, v := range []string{viper.GetString("EDITOR"), os.Getenv("EDITOR"), os.Getenv("VISUAL")} {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "vi"
}

// openEditor opens the resolved editor on a temp file seeded with initial and
// returns the edited contents, trimmed of a single trailing newline.
func openEditor(initial string) (string, error) {
	tmp, err := os.CreateTemp("", "snip-*.sh")
	if err != nil {
		return "", fmt.Errorf("could not create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(initial); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()

	editor := editorBinary()
	// Support editors invoked with flags, e.g. EDITOR="code --wait".
	parts := strings.Fields(editor)
	args := append(parts[1:], tmp.Name())
	c := exec.Command(parts[0], args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\n"), nil
}

// runCommand executes the resolved command via the shell, inheriting stdio so
// pipes, quoting and redirects behave exactly as typed. It returns the child's
// exit code (0 on success) and any error.
func runCommand(resolved string) (int, error) {
	c := exec.Command("sh", "-c", resolved)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if ok := asExitError(err, &exitErr); ok {
		return exitErr.ExitCode(), err
	}
	return 1, err
}

// asExitError is a tiny errors.As helper kept local to avoid an extra import in
// callers.
func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}
