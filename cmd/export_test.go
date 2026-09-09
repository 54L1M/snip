/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"reflect"
	"testing"

	"github.com/54L1M/snip/internal/snippet"
)

func TestParseExportArgs(t *testing.T) {
	names, provided := parseExportArgs([]string{"deploy", "ns=prod", "worker", "deploy", "hash=abc=1"})
	if want := []string{"deploy", "worker"}; !reflect.DeepEqual(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
	if want := map[string]string{"ns": "prod", "hash": "abc=1"}; !reflect.DeepEqual(provided, want) {
		t.Errorf("provided = %v, want %v", provided, want)
	}
}

func TestExportValues(t *testing.T) {
	s := &snippet.Snippet{
		Name:       "deploy",
		Template:   "deploy {{app}}:{{tag}} to {{ns}} {{empty}}",
		LastValues: map[string]string{"app": "web", "ns": "prod", "empty": ""},
	}
	values, missing := exportValues(s, map[string]string{"ns": "staging", "unused": "x"})

	// provided wins over last value; last value fills the rest; an empty last
	// value and a var with no value at all are both missing.
	wantValues := map[string]string{"app": "web", "ns": "staging"}
	if !reflect.DeepEqual(values, wantValues) {
		t.Errorf("values = %v, want %v", values, wantValues)
	}
	if want := []string{"tag", "empty"}; !reflect.DeepEqual(missing, want) {
		t.Errorf("missing = %v, want %v", missing, want)
	}
}

func TestExportValuesNoVars(t *testing.T) {
	s := &snippet.Snippet{Name: "db", Template: "pgcli -h localhost"}
	values, missing := exportValues(s, nil)
	if len(values) != 0 || len(missing) != 0 {
		t.Errorf("expected nothing to resolve, got values=%v missing=%v", values, missing)
	}
}

func TestRenderSh(t *testing.T) {
	out := &snippet.Store{Snippets: map[string]*snippet.Snippet{
		"deploy": {Name: "deploy", Description: "Roll out", Template: "kubectl set image x\n"},
		"db":     {Name: "db", Template: "pgcli -h localhost\n  -p 5432"},
	}}
	got := string(renderSh(out))
	want := "# db\npgcli -h localhost\n  -p 5432\n\n# deploy: Roll out\nkubectl set image x\n"
	if got != want {
		t.Errorf("renderSh() =\n%q\nwant\n%q", got, want)
	}
}
