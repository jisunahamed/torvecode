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
