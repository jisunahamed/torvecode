package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jisunahamed/torvecode/internal/permission"
)

func TestQuitDialogEscapeClosesWithoutQuitting(t *testing.T) {
	dialog := NewQuitCmp()
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Escape returned no command")
	}
	if _, ok := cmd().(CloseQuitMsg); !ok {
		t.Fatalf("Escape command returned %T, want CloseQuitMsg", cmd())
	}
}

func TestPermissionDialogEscapeDeniesAndCloses(t *testing.T) {
	request := permission.PermissionRequest{ID: "test-permission"}
	dialog := NewPermissionDialogCmp()
	_ = dialog.SetPermissions(request)
	_, cmd := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Escape returned no command")
	}
	msg, ok := cmd().(PermissionResponseMsg)
	if !ok {
		t.Fatalf("Escape command returned %T, want PermissionResponseMsg", cmd())
	}
	if msg.Action != PermissionDeny || msg.Permission.ID != request.ID {
		t.Fatalf("Escape response = %#v, want deny for %q", msg, request.ID)
	}
}
