package tui

import (
	"context"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
	"testing"
)

func TestBrandedOnboardingFitsTerminal(t *testing.T) {
	for _, width := range []int{40, 80, 120} {
		m := newOnboardingModel(context.Background())
		m.width, m.height = width, 40
		view := m.View()
		for _, line := range strings.Split(view, "\n") {
			if lipgloss.Width(line) > width {
				t.Fatalf("width %d overflow: %d", width, lipgloss.Width(line))
			}
		}
	}
	if path := os.Getenv("TORVE_DESIGN_PREVIEW"); path != "" {
		m := newOnboardingModel(context.Background())
		m.width, m.height = 100, 38
		if err := os.WriteFile(path, []byte(m.View()), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
