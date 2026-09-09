package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOnboardingRendersConnectionChoices(t *testing.T) {
	model := newOnboardingModel(context.Background())
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	view := updated.(onboardingModel).View()
	for _, text := range []string{"TORVE", "CODE", "Login with website", "Use Torve AI API key", "ctrl+c quit"} {
		if !strings.Contains(view, text) {
			t.Fatalf("onboarding view does not contain %q", text)
		}
	}
}

func TestOnboardingSelectsAPIKeyEntry(t *testing.T) {
	model := newOnboardingModel(context.Background())
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(onboardingModel)
	if model.selected != 1 {
		t.Fatalf("selected option = %d, want 1", model.selected)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(onboardingModel)
	if model.stage != stageAPIKey || !model.apiKey.Focused() {
		t.Fatalf("API key input was not opened and focused")
	}
}
