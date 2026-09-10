package clipboardimage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const maxImageBytes = 5 * 1024 * 1024

var lookPath = exec.LookPath
var execute = executeClipboard

func Read(ctx context.Context) ([]byte, error) {
	readCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	data, err := execute(readCtx, runtime.GOOS)
	if errors.Is(readCtx.Err(), context.DeadlineExceeded) {
		return nil, errors.New("reading the clipboard took longer than 15 seconds")
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("the clipboard does not contain a PNG image")
	}
	if len(data) > maxImageBytes {
		return nil, errors.New("clipboard image is larger than 5 MB")
	}
	return data, nil
}

func executeClipboard(ctx context.Context, platform string) ([]byte, error) {
	switch platform {
	case "windows":
		command, err := lookPath("powershell.exe")
		if err != nil {
			return nil, errors.New("PowerShell is required to read a clipboard image")
		}
		temp, err := os.CreateTemp("", "torve-clipboard-*.png")
		if err != nil {
			return nil, err
		}
		path := temp.Name()
		_ = temp.Close()
		defer os.Remove(path)
		script := `$ErrorActionPreference='Stop'; Add-Type -AssemblyName System.Windows.Forms; Add-Type -AssemblyName System.Drawing; $image=[Windows.Forms.Clipboard]::GetImage(); if ($null -eq $image) { exit 3 }; $image.Save($env:TORVE_CLIPBOARD_IMAGE, [Drawing.Imaging.ImageFormat]::Png); $image.Dispose()`
		cmd := exec.CommandContext(ctx, command, "-NoProfile", "-Sta", "-Command", script) //nolint:gosec
		cmd.Env = append(os.Environ(), "TORVE_CLIPBOARD_IMAGE="+path)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 3 {
				return nil, errors.New("the clipboard does not contain an image")
			}
			if message := strings.TrimSpace(stderr.String()); message != "" {
				return nil, fmt.Errorf("cannot read clipboard image: %s", message)
			}
			return nil, fmt.Errorf("cannot read clipboard image: %w", err)
		}
		return os.ReadFile(path)
	case "darwin":
		return commandOutput(ctx, "pngpaste")
	case "linux":
		if _, err := lookPath("wl-paste"); err == nil {
			return commandOutput(ctx, "wl-paste", "--no-newline", "--type", "image/png")
		}
		return commandOutput(ctx, "xclip", "-selection", "clipboard", "-t", "image/png", "-o")
	default:
		return nil, fmt.Errorf("clipboard images are not supported on %s", platform)
	}
}

func commandOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	command, err := lookPath(name)
	if err != nil {
		return nil, fmt.Errorf("clipboard image support requires %s", name)
	}
	cmd := exec.CommandContext(ctx, command, args...) //nolint:gosec
	data, err := cmd.Output()
	if err != nil {
		return nil, errors.New("the clipboard does not contain an image")
	}
	return data, nil
}
