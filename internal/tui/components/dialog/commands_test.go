package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandPaletteFiltersAfterSlash(t *testing.T) {
	palette := NewCommandDialogCmp().(*commandDialogCmp)
	palette.SetCommands([]Command{
		{Title: "/models", Description: "Switch model"},
		{Title: "/skills", Description: "Manage skills"},
	})
	palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("mod")})
	items := palette.listView.GetItems()
	if len(items) != 1 || items[0].Title != "/models" {
		t.Fatalf("filtered commands = %#v", items)
	}
}

func TestEmptyBackspaceClosesCommandPalette(t *testing.T) {
	palette := NewCommandDialogCmp().(*commandDialogCmp)
	_, command := palette.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if command == nil {
		t.Fatal("Backspace did not close an empty command palette")
	}
	if _, ok := command().(CloseCommandDialogMsg); !ok {
		t.Fatalf("Backspace returned %T, want CloseCommandDialogMsg", command())
	}
}

func TestSlashClosesCommandPalette(t *testing.T) {
	palette := NewCommandDialogCmp().(*commandDialogCmp)
	_, command := palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if command == nil {
		t.Fatal("Slash did not close the command palette")
	}
	if _, ok := command().(CloseCommandDialogMsg); !ok {
		t.Fatalf("Slash returned %T, want CloseCommandDialogMsg", command())
	}
}
