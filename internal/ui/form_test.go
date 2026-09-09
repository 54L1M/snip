/*
Copyright © 2026 54L1M
*/
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderCommandBoxWrapsToTerminal(t *testing.T) {
	long := "kubectl set image deployment ats ats-app=europe-west4-docker.pkg.dev/i2d-cloud/docker-containers/ats-production:69dca48 --namespace=production"
	for _, w := range []int{40, 80, 120} {
		out := renderCommandBox(long, w)
		if got := lipgloss.Width(out); got >= w {
			t.Errorf("termWidth %d: box is %d columns wide, expected < %d", w, got, w)
		}
		// Every line, borders included, must fit.
		for _, line := range strings.Split(out, "\n") {
			if lipgloss.Width(line) >= w {
				t.Errorf("termWidth %d: line too wide (%d): %q", w, lipgloss.Width(line), line)
			}
		}
	}
}

func TestRenderCommandBoxShortUnchanged(t *testing.T) {
	short := "ls -la"
	if got, want := renderCommandBox(short, 80), CommandBoxStyle.Render(short); got != want {
		t.Errorf("short command should not be stretched to the terminal width")
	}
	// Unknown width: never wrap.
	long := strings.Repeat("x", 300)
	if got, want := renderCommandBox(long, 0), CommandBoxStyle.Render(long); got != want {
		t.Errorf("width 0 should disable wrapping")
	}
}
