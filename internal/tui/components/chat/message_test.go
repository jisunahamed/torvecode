package chat

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jisunahamed/torvecode/internal/message"
)

func TestLongMessageNeverExceedsPanelWidth(t *testing.T) {
	const width = 40
	rendered := renderMessage(strings.Repeat("https://example.test/no-break/", 8), false, true, width)
	for index, line := range strings.Split(rendered, "\n") {
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("line %d width=%d, want <=%d", index, got, width)
		}
	}
}

func TestReasoningRendersThinkingActivity(t *testing.T) {
	msg := message.Message{
		ID:   "assistant-1",
		Role: message.Assistant,
		Parts: []message.ContentPart{
			message.ReasoningContent{Thinking: "Inspecting the project"},
		},
	}
	rendered := renderAssistantMessage(msg, 0, []message.Message{msg}, nil, msg.ID, false, 60, 0)
	if len(rendered) != 1 || !strings.Contains(rendered[0].content, "Thinking") {
		t.Fatalf("reasoning activity was not rendered: %#v", rendered)
	}
}

func TestViewToolUsesReadLabel(t *testing.T) {
	if got := toolName("view"); got != "Read" {
		t.Fatalf("toolName(view)=%q, want Read", got)
	}
}
