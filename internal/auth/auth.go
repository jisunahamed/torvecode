package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	defaultAPIURL = "https://api.torveai.com"
	defaultWebURL = "https://torveai.com"
	serviceName   = "Torvecode"
	accountName   = "default"
)

type Credential struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Method       string    `json:"method"`
}

type DeviceAuthorization struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func APIURL() string {
	if value := strings.TrimRight(os.Getenv("TORVE_API_URL"), "/"); value != "" {
		return value
	}
	return defaultAPIURL
}

func WebURL() string {
	if value := strings.TrimRight(os.Getenv("TORVE_WEB_URL"), "/"); value != "" {
		return value
	}
	return defaultWebURL
}

// Resolve never silently falls back when TORVE_API_KEY is explicitly supplied.
func Resolve() (Credential, error) {
	if token, ok := os.LookupEnv("TORVE_API_KEY"); ok {
		if strings.TrimSpace(token) == "" {
			return Credential{}, errors.New("TORVE_API_KEY is set but empty")
		}
		return Credential{AccessToken: strings.TrimSpace(token), Method: "environment"}, nil
	}
	credential, err := loadSecure()
	if err != nil {
		return Credential{}, errors.New("not signed in; run `torve auth login` or set TORVE_API_KEY")
	}
	if credential.RefreshToken != "" && !credential.ExpiresAt.IsZero() && time.Now().Add(time.Minute).After(credential.ExpiresAt) {
		credential, err = Refresh(context.Background(), credential.RefreshToken)
		if err != nil {
			return Credential{}, fmt.Errorf("saved session expired: %w", err)
		}
		if err := Save(credential); err != nil {
			return Credential{}, err
		}
	}
	return credential, nil
}

func Save(credential Credential) error {
	payload, err := json.Marshal(credential)
	if err != nil {
		return err
	}
	return storeSecure(string(payload))
}

func Delete() error { return deleteSecure() }

func StartDevice(ctx context.Context, deviceName string) (DeviceAuthorization, error) {
	var result DeviceAuthorization
	err := requestJSON(ctx, http.MethodPost, APIURL()+"/cli/auth/device", "", map[string]string{"device_name": deviceName}, &result)
	return result, err
}

func PollDevice(ctx context.Context, grant DeviceAuthorization) (Credential, error) {
	interval := time.Duration(grant.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	deadline := time.NewTimer(time.Duration(grant.ExpiresIn) * time.Second)
	defer deadline.Stop()
	for {
		select {
		case <-ctx.Done():
			return Credential{}, ctx.Err()
		case <-deadline.C:
			return Credential{}, errors.New("authorization expired")
		case <-time.After(interval):
			var result Credential
			err := requestJSON(ctx, http.MethodPost, APIURL()+"/cli/auth/token", "", map[string]string{"grant_type": "urn:ietf:params:oauth:grant-type:device_code", "device_code": grant.DeviceCode}, &result)
			if err == nil {
				result.Method = "website"
				return result, nil
			}
			var remote *RemoteError
			if errors.As(err, &remote) {
				switch remote.Code {
				case "authorization_pending":
					continue
				case "slow_down":
					interval += 5 * time.Second
					continue
				}
			}
			return Credential{}, err
		}
	}
}

func Refresh(ctx context.Context, refreshToken string) (Credential, error) {
	var result Credential
	err := requestJSON(ctx, http.MethodPost, APIURL()+"/cli/auth/token", "", map[string]string{"grant_type": "refresh_token", "refresh_token": refreshToken}, &result)
	if err == nil {
		result.Method = "website"
	}
	return result, err
}

func Revoke(ctx context.Context, token string) error {
	return requestJSON(ctx, http.MethodPost, APIURL()+"/cli/auth/revoke", token, nil, nil)
}

type RemoteError struct {
	Status        int
	Code, Message string
}

func (e *RemoteError) Error() string {
	return fmt.Sprintf("Torve AI returned %s (%d): %s", e.Code, e.Status, e.Message)
}

func requestJSON(ctx context.Context, method, endpoint, token string, body any, target any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")
	if body != nil {
		req.Header.Set("content-type", "application/json")
	}
	if token != "" {
		req.Header.Set("authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var value struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(payload, &value)
		if value.Error == "" {
			value.Error = http.StatusText(resp.StatusCode)
		}
		return &RemoteError{Status: resp.StatusCode, Code: value.Error, Message: value.Message}
	}
	if target != nil && len(payload) > 0 {
		return json.Unmarshal(payload, target)
	}
	return nil
}

func OpenBrowser(rawURL string) error {
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command, args = "rundll32", []string{"url.dll,FileProtocolHandler", rawURL}
	case "darwin":
		command, args = "open", []string{rawURL}
	default:
		command, args = "xdg-open", []string{rawURL}
	}
	return exec.Command(command, args...).Start()
}

func DeviceName() string {
	host, _ := os.Hostname()
	if host == "" {
		host = runtime.GOOS
	}
	return host
}

func storeSecure(payload string) error {
	switch runtime.GOOS {
	case "windows":
		script := `$b=[Text.Encoding]::UTF8.GetBytes($env:TORVE_SECRET);$e=[Security.Cryptography.ProtectedData]::Protect($b,$null,[Security.Cryptography.DataProtectionScope]::CurrentUser);[IO.File]::WriteAllBytes($env:TORVE_SECRET_FILE,$e)`
		path, err := secretPath()
		if err != nil {
			return err
		}
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
		cmd.Env = append(os.Environ(), "TORVE_SECRET="+payload, "TORVE_SECRET_FILE="+path)
		return cmd.Run()
	case "darwin":
		return exec.Command("security", "add-generic-password", "-U", "-s", serviceName, "-a", accountName, "-w", payload).Run()
	default:
		cmd := exec.Command("secret-tool", "store", "--label=Torvecode", "service", serviceName, "account", accountName)
		cmd.Stdin = strings.NewReader(payload)
		return cmd.Run()
	}
}

func loadSecure() (Credential, error) {
	var output []byte
	var err error
	switch runtime.GOOS {
	case "windows":
		path, pathErr := secretPath()
		if pathErr != nil {
			return Credential{}, pathErr
		}
		script := `$e=[IO.File]::ReadAllBytes($env:TORVE_SECRET_FILE);$b=[Security.Cryptography.ProtectedData]::Unprotect($e,$null,[Security.Cryptography.DataProtectionScope]::CurrentUser);[Text.Encoding]::UTF8.GetString($b)`
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
		cmd.Env = append(os.Environ(), "TORVE_SECRET_FILE="+path)
		output, err = cmd.Output()
	case "darwin":
		output, err = exec.Command("security", "find-generic-password", "-s", serviceName, "-a", accountName, "-w").Output()
	default:
		output, err = exec.Command("secret-tool", "lookup", "service", serviceName, "account", accountName).Output()
	}
	if err != nil {
		return Credential{}, err
	}
	var result Credential
	if err := json.Unmarshal(bytes.TrimSpace(output), &result); err != nil {
		return Credential{}, err
	}
	return result, nil
}

func deleteSecure() error {
	switch runtime.GOOS {
	case "windows":
		path, err := secretPath()
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	case "darwin":
		return exec.Command("security", "delete-generic-password", "-s", serviceName, "-a", accountName).Run()
	default:
		return exec.Command("secret-tool", "clear", "service", serviceName, "account", accountName).Run()
	}
}

func secretPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = dir + string(os.PathSeparator) + "torvecode"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir + string(os.PathSeparator) + "credential.dpapi", nil
}

func AddCodeToURL(rawURL, code string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	query.Set("code", code)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
