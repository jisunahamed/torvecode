//go:build windows

package auth

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func protectWindowsCredential(payload, path string) error {
	plain := []byte(payload)
	in := dataBlob(plain)
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return fmt.Errorf("Windows credential encryption failed: %w", err)
	}
	encrypted := copyAndFreeDataBlob(out)
	return os.WriteFile(path, encrypted, 0600)
}

func unprotectWindowsCredential(path string) ([]byte, error) {
	encrypted, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	in := dataBlob(encrypted)
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, fmt.Errorf("Windows credential decryption failed: %w", err)
	}
	return copyAndFreeDataBlob(out), nil
}

func dataBlob(data []byte) windows.DataBlob {
	if len(data) == 0 {
		return windows.DataBlob{}
	}
	return windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
}

func copyAndFreeDataBlob(blob windows.DataBlob) []byte {
	if blob.Data == nil || blob.Size == 0 {
		return nil
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(blob.Data)))
	return append([]byte(nil), unsafe.Slice(blob.Data, int(blob.Size))...)
}
