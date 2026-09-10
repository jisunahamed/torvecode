package chat

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jisunahamed/torvecode/internal/tui/components/dialog"
)

func TestSlashOnEmptyPromptOpensCommandPalette(t *testing.T) {
	model := NewEditorCmp(nil).(*editorCmp)
	_, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if command == nil {
		t.Fatal("slash did not return a command")
	}
	message := command()
	if _, ok := message.(dialog.OpenCommandDialogMsg); !ok {
		t.Fatalf("slash returned %T, want OpenCommandDialogMsg", message)
	}
	if model.textarea.Value() != "" {
		t.Fatalf("slash leaked into prompt: %q", model.textarea.Value())
	}
}

func TestSlashInsidePromptRemainsText(t *testing.T) {
	model := NewEditorCmp(nil).(*editorCmp)
	model.textarea.SetValue("path")
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if model.textarea.Value() != "path/" {
		t.Fatalf("prompt = %q, want path/", model.textarea.Value())
	}
}
