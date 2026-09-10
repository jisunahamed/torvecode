package chat

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jisunahamed/torvecode/internal/tui/components/dialog"
)

func TestBackspaceDeletesPromptText(t *testing.T) {
	model := NewEditorCmp(nil).(*editorCmp)
	model.textarea.SetValue("delete me")
	// Windows Terminal can encode Backspace as Ctrl+H.
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	if model.textarea.Value() != "delete m" {
		t.Fatalf("prompt = %q, want delete m", model.textarea.Value())
	}
}

func TestCtrlVPastesClipboardTextWithoutAddingImage(t *testing.T) {
	originalText := readClipboardText
	originalImage := readClipboardImage
	t.Cleanup(func() {
		readClipboardText = originalText
		readClipboardImage = originalImage
	})
	readClipboardText = func() (string, error) { return "pasted text", nil }
	readClipboardImage = func(context.Context) ([]byte, error) {
		t.Fatal("image reader called when clipboard contains text")
		return nil, nil
	}

	model := NewEditorCmp(nil).(*editorCmp)
	_, command := model.Update(tea.KeyMsg{Type: tea.KeyCtrlV})
	if command == nil {
		t.Fatal("Ctrl+V did not return a paste command")
	}
	model.Update(command())
	if model.textarea.Value() != "pasted text" {
		t.Fatalf("prompt = %q, want pasted text", model.textarea.Value())
	}
}

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
