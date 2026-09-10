package clipboardimage

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestReadReturnsClipboardPNG(t *testing.T) {
	original := execute
	t.Cleanup(func() { execute = original })
	execute = func(context.Context, string) ([]byte, error) { return []byte("png"), nil }
	data, err := Read(context.Background())
	if err != nil || string(data) != "png" {
		t.Fatalf("data=%q err=%v", data, err)
	}
}

func TestReadRejectsOversizedClipboardImage(t *testing.T) {
	original := execute
	t.Cleanup(func() { execute = original })
	execute = func(context.Context, string) ([]byte, error) { return make([]byte, maxImageBytes+1), nil }
	_, err := Read(context.Background())
	if err == nil || !strings.Contains(err.Error(), "5 MB") {
		t.Fatalf("error=%v", err)
	}
}

func TestReadPreservesActionableError(t *testing.T) {
	original := execute
	t.Cleanup(func() { execute = original })
	execute = func(context.Context, string) ([]byte, error) { return nil, errors.New("install pngpaste") }
	_, err := Read(context.Background())
	if err == nil || err.Error() != "install pngpaste" {
		t.Fatalf("error=%v", err)
	}
}
