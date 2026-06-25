/*
Copyright © 2026 54L1M
*/
package cmd

import (
	"reflect"
	"testing"
)

func TestParseHistoryZshExtended(t *testing.T) {
	data := []byte(": 1700000000:0;git status\n: 1700000001:0;docker ps -a\n")
	want := []string{"git status", "docker ps -a"}
	if got := parseHistory(data); !reflect.DeepEqual(got, want) {
		t.Errorf("parseHistory() = %v, want %v", got, want)
	}
}

func TestParseHistoryPlain(t *testing.T) {
	data := []byte("ls -la\ncd /tmp\n")
	want := []string{"ls -la", "cd /tmp"}
	if got := parseHistory(data); !reflect.DeepEqual(got, want) {
		t.Errorf("parseHistory() = %v, want %v", got, want)
	}
}

func TestParseHistoryBashTimestamps(t *testing.T) {
	data := []byte("#1700000000\nls\n#1700000001\necho hi\n")
	want := []string{"ls", "echo hi"}
	if got := parseHistory(data); !reflect.DeepEqual(got, want) {
		t.Errorf("parseHistory() = %v, want %v", got, want)
	}
}

func TestParseHistoryContinuation(t *testing.T) {
	data := []byte(": 1700000000:0;rsync -avz ./src/ \\\n  user@host:/dest/\n")
	want := []string{"rsync -avz ./src/ \n  user@host:/dest/"}
	if got := parseHistory(data); !reflect.DeepEqual(got, want) {
		t.Errorf("parseHistory() = %q, want %q", got, want)
	}
}

func TestIsSnipInvocation(t *testing.T) {
	tests := map[string]bool{
		"snip add deploy --last": true,
		"/usr/local/bin/snip ls": true,
		"git commit -m snip":     false,
		"docker ps -a":          false,
		"":                       false,
	}
	for cmd, want := range tests {
		if got := isSnipInvocation(cmd); got != want {
			t.Errorf("isSnipInvocation(%q) = %v, want %v", cmd, got, want)
		}
	}
}
