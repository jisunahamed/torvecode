package documents

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const maxMarkdownBytes = 512 * 1024

var supportedExtensions = map[string]bool{
	".pdf": true, ".docx": true, ".pptx": true, ".xlsx": true, ".xls": true,
	".html": true, ".htm": true, ".csv": true, ".json": true, ".xml": true,
	".epub": true, ".zip": true, ".msg": true,
}

var findCommand = exec.LookPath
var runCommand = runMarkItDown

func Supported(path string) bool {
	return supportedExtensions[strings.ToLower(filepath.Ext(path))]
}

func Available() bool {
	_, err := findCommand("markitdown")
	return err == nil
}

func InstallHint() string {
	return "Install Python 3.10+, then run: pip install 'markitdown[pdf,docx,pptx,xlsx,xls,outlook]'"
}

func Convert(ctx context.Context, path string) (string, error) {
	if !Supported(path) {
		return "", fmt.Errorf("MarkItDown does not support %s", filepath.Ext(path))
	}
	command, err := findCommand("markitdown")
	if err != nil {
		return "", fmt.Errorf("Microsoft MarkItDown is not installed. %s", InstallHint())
	}
	conversionCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	markdown, err := runCommand(conversionCtx, command, path)
	if errors.Is(conversionCtx.Err(), context.DeadlineExceeded) {
		return "", errors.New("document conversion took longer than 60 seconds")
	}
	if err != nil {
		return "", fmt.Errorf("MarkItDown could not convert %s: %w", filepath.Base(path), err)
	}
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return "", fmt.Errorf("MarkItDown found no readable content in %s", filepath.Base(path))
	}
	return markdown, nil
}

func runMarkItDown(ctx context.Context, command, path string) (string, error) {
	cmd := exec.CommandContext(ctx, command, path) //nolint:gosec -- command comes from PATH; path is one argument.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	limited, readErr := io.ReadAll(io.LimitReader(stdout, maxMarkdownBytes+1))
	waitErr := cmd.Wait()
	if readErr != nil {
		return "", readErr
	}
	if len(limited) > maxMarkdownBytes {
		return "", errors.New("converted document exceeds the 512 KiB context limit; select a smaller document")
	}
	if waitErr != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return "", errors.New(message)
		}
		return "", waitErr
	}
	return string(limited), nil
}
