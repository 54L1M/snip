/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/54L1M/snip/internal/snippet"
)

func TestParseImport(t *testing.T) {
	data := []byte(`{"snippets": {
		"deploy": {"description": "Roll out", "template": "kubectl set image {{img}}", "last_values": {"img": "x"}},
		"db": {"name": "ignored", "template": "pgcli"}
	}}`)
	in, err := parseImport(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(in.Snippets) != 2 {
		t.Fatalf("got %d snippets, want 2", len(in.Snippets))
	}
	// The map key is authoritative for the name.
	if in.Snippets["db"].Name != "db" {
		t.Errorf("name = %q, want %q", in.Snippets["db"].Name, "db")
	}
	if in.Snippets["deploy"].LastValues["img"] != "x" {
		t.Errorf("last_values not preserved")
	}
}

func TestParseImportErrors(t *testing.T) {
	tests := map[string]string{
		"not json":        `{`,
		"no snippets key": `{"foo": 1}`,
		"empty snippets":  `{"snippets": {}}`,
		"null snippet":    `{"snippets": {"a": null}}`,
		"empty template":  `{"snippets": {"a": {"template": "  "}}}`,
		"empty name":      `{"snippets": {" ": {"template": "ls"}}}`,
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := parseImport([]byte(data)); err == nil {
				t.Errorf("expected error for %s", data)
			}
		})
	}
}

func TestMergeImport(t *testing.T) {
	dst := &snippet.Store{Snippets: map[string]*snippet.Snippet{
		"deploy": {Name: "deploy", Template: "old {{ns}}", LastValues: map[string]string{"ns": "prod"}},
		"keep":   {Name: "keep", Template: "untouched"},
	}}
	src := &snippet.Store{Snippets: map[string]*snippet.Snippet{
		"deploy": {Name: "deploy", Description: "new", Template: "new {{ns}}"},
		"fresh":  {Name: "fresh", Template: "echo hi", LastValues: map[string]string{"a": "1"}},
		"other":  {Name: "other", Template: "not selected"},
	}}

	imported, skipped := mergeImport(dst, src, []string{"deploy", "fresh"}, false)
	if want := []string{"fresh"}; !reflect.DeepEqual(imported, want) {
		t.Errorf("imported = %v, want %v", imported, want)
	}
	if want := []string{"deploy"}; !reflect.DeepEqual(skipped, want) {
		t.Errorf("skipped = %v, want %v", skipped, want)
	}
	if dst.Snippets["deploy"].Template != "old {{ns}}" {
		t.Errorf("existing snippet was overwritten without --force")
	}
	if _, ok := dst.Snippets["other"]; ok {
		t.Errorf("unselected snippet was imported")
	}
	if dst.Snippets["fresh"].LastValues["a"] != "1" {
		t.Errorf("imported last_values not kept")
	}

	imported, skipped = mergeImport(dst, src, []string{"deploy"}, true)
	if len(imported) != 1 || len(skipped) != 0 {
		t.Fatalf("force: imported=%v skipped=%v", imported, skipped)
	}
	d := dst.Snippets["deploy"]
	if d.Template != "new {{ns}}" || d.Description != "new" {
		t.Errorf("force did not overwrite: %+v", d)
	}
	// A forced overwrite with no last_values of its own keeps the local defaults.
	if d.LastValues["ns"] != "prod" {
		t.Errorf("local last_values were dropped on overwrite: %v", d.LastValues)
	}
	if !strings.Contains(dst.Snippets["keep"].Template, "untouched") {
		t.Errorf("unrelated snippet modified")
	}
}

func TestPlural(t *testing.T) {
	if got := plural(1, "snippet"); got != "1 snippet" {
		t.Errorf("got %q", got)
	}
	if got := plural(3, "snippet"); got != "3 snippets" {
		t.Errorf("got %q", got)
	}
}
