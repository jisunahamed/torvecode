package auth

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWindowsCredentialProtectionRoundTrip(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows DPAPI test")
	}

	path := filepath.Join(t.TempDir(), "credential.dpapi")
	payload := `{"access_token":"secret","method":"website"}`
	if err := protectWindowsCredential(payload, path); err != nil {
		t.Fatal(err)
	}
	encrypted, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encrypted, []byte("secret")) {
		t.Fatal("credential was persisted as plaintext")
	}
	decrypted, err := unprotectWindowsCredential(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != payload {
		t.Fatalf("decrypted credential = %q, want %q", decrypted, payload)
	}
}
