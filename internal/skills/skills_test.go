package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWritePreservesExistingAndRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	path, err := Write(dir, "review", false, []byte("Review tests"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Write(dir, "review", false, []byte("overwrite")); err == nil {
		t.Fatal("overwrote existing skill")
	}
	body, _ := os.ReadFile(path)
	if string(body) != "Review tests" {
		t.Fatal("changed existing skill")
	}
	if _, err = Write(dir, "../escape", false, []byte("bad")); err == nil {
		t.Fatal("accepted traversal")
	}
	if filepath.Base(path) != "SKILL.md" {
		t.Fatal(path)
	}
}
