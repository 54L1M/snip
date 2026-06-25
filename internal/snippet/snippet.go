/*
Copyright © 2026 54L1M
*/

// Package snippet defines the snippet model and the template/variable handling
// used to store and render parameterized commands.
package snippet

import (
	"fmt"
	"regexp"
)

// varPattern matches {{name}} placeholders, allowing surrounding whitespace,
// e.g. {{ hash }}. Names are Go-identifier-like: a letter/underscore followed
// by letters, digits or underscores.
var varPattern = regexp.MustCompile(`{{\s*([A-Za-z_]\w*)\s*}}`)

// Snippet is a single saved command template.
type Snippet struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Template    string            `json:"template"`
	// LastValues remembers the most recent value used for each variable so it
	// can be offered as the default next time.
	LastValues map[string]string `json:"last_values,omitempty"`
}

// Store is the on-disk collection of snippets, keyed by name.
type Store struct {
	Snippets map[string]*Snippet `json:"snippets"`
}

// Vars returns the variable names found in template, in first-seen order with
// duplicates removed.
func Vars(template string) []string {
	matches := varPattern.FindAllStringSubmatch(template, -1)
	seen := make(map[string]bool, len(matches))
	var vars []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			vars = append(vars, name)
		}
	}
	return vars
}

// Vars returns the variables declared in the snippet's template.
func (s *Snippet) Vars() []string { return Vars(s.Template) }

// Render substitutes every {{name}} placeholder in template with its value from
// values. It returns an error if any placeholder has no corresponding value.
func Render(template string, values map[string]string) (string, error) {
	var missing []string
	out := varPattern.ReplaceAllStringFunc(template, func(match string) string {
		name := varPattern.FindStringSubmatch(match)[1]
		val, ok := values[name]
		if !ok {
			missing = append(missing, name)
			return match
		}
		return val
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("missing value(s) for: %v", missing)
	}
	return out, nil
}
