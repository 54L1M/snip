/*
Copyright © 2026 54L1M
*/
package snippet

import (
	"reflect"
	"testing"
)

func TestVars(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{"none", "echo hello world", nil},
		{"single", "echo {{name}}", []string{"name"}},
		{"whitespace", "echo {{ name }}", []string{"name"}},
		{"multiple ordered", "{{a}} {{b}} {{c}}", []string{"a", "b", "c"}},
		{"dedup keeps first order", "{{ns}}-{{hash}}-{{ns}}", []string{"ns", "hash"}},
		{"underscores and digits", "{{my_var2}}", []string{"my_var2"}},
		{"ignores leading digit", "{{2bad}}", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Vars(tt.template); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vars(%q) = %v, want %v", tt.template, got, tt.want)
			}
		})
	}
}

func TestRender(t *testing.T) {
	got, err := Render(
		"docker run -d --name {{app}} -p {{port}}:8080 registry.example.com/{{app}}:{{tag}}",
		map[string]string{"app": "web", "port": "8080", "tag": "1.4.2"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "docker run -d --name web -p 8080:8080 registry.example.com/web:1.4.2"
	if got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRenderRepeatedVar(t *testing.T) {
	got, err := Render("{{x}}-{{x}}", map[string]string{"x": "v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "v-v" {
		t.Errorf("Render() = %q, want %q", got, "v-v")
	}
}

func TestRenderMissing(t *testing.T) {
	_, err := Render("echo {{a}} {{b}}", map[string]string{"a": "1"})
	if err == nil {
		t.Fatal("expected error for missing variable, got nil")
	}
}
