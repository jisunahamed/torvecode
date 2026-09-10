package documents

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestConvertUsesMarkItDownForDocuments(t *testing.T) {
	originalFind, originalRun := findCommand, runCommand
	t.Cleanup(func() { findCommand, runCommand = originalFind, originalRun })
	findCommand = func(string) (string, error) { return "markitdown", nil }
	runCommand = func(_ context.Context, command, path string) (string, error) {
		if command != "markitdown" || path != "report.pdf" {
			t.Fatalf("command=%q path=%q", command, path)
		}
		return "# Report\n\nA table", nil
	}

	markdown, err := Convert(context.Background(), "report.pdf")
	if err != nil || markdown != "# Report\n\nA table" {
		t.Fatalf("markdown=%q err=%v", markdown, err)
	}
}

func TestConvertExplainsMissingDependency(t *testing.T) {
	original := findCommand
	t.Cleanup(func() { findCommand = original })
	findCommand = func(string) (string, error) { return "", errors.New("missing") }

	_, err := Convert(context.Background(), "report.docx")
	if err == nil || !strings.Contains(err.Error(), "pip install") {
		t.Fatalf("error=%v", err)
	}
}
