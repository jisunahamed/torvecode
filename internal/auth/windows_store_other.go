//go:build !windows

package auth

import "errors"

func protectWindowsCredential(string, string) error {
	return errors.New("Windows credential protection is unavailable on this platform")
}

func unprotectWindowsCredential(string) ([]byte, error) {
	return nil, errors.New("Windows credential protection is unavailable on this platform")
}
