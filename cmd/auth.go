package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jisunahamed/torvecode/internal/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{Use: "auth", Short: "Connect Torvecode to your Torve AI account"}
var apiKeyLogin bool

var authLoginCmd = &cobra.Command{Use: "login", Short: "Sign in through torveai.com", RunE: func(cmd *cobra.Command, _ []string) error {
	if apiKeyLogin {
		fmt.Fprint(cmd.OutOrStdout(), "Torve AI API key: ")
		secret, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(cmd.OutOrStdout())
		if err != nil {
			return err
		}
		key := strings.TrimSpace(string(secret))
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}
		credential := auth.Credential{AccessToken: key, Method: "api-key"}
		ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
		defer cancel()
		if _, err := auth.FetchModels(ctx, credential); err != nil {
			return fmt.Errorf("API key rejected: %w", err)
		}
		if err := auth.Save(credential); err != nil {
			return fmt.Errorf("secure credential storage unavailable; use TORVE_API_KEY for this session: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Connected with a Torve AI API key.")
		return nil
	}
	grant, err := auth.StartDevice(cmd.Context(), auth.DeviceName())
	if err != nil {
		return err
	}
	loginURL := grant.VerificationURIComplete
	if loginURL == "" {
		loginURL = auth.AddCodeToURL(grant.VerificationURI, grant.UserCode)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Open %s\nCode: %s\nWaiting for approval...\n", loginURL, grant.UserCode)
	_ = auth.OpenBrowser(loginURL)
	credential, err := auth.PollDevice(cmd.Context(), grant)
	if err != nil {
		return err
	}
	if err := auth.Save(credential); err != nil {
		return fmt.Errorf("signed in, but secure storage failed: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Torvecode is connected.")
	return nil
}}

var authStatusCmd = &cobra.Command{Use: "status", Short: "Show the active credential source", RunE: func(cmd *cobra.Command, _ []string) error {
	credential, err := auth.Resolve()
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Not connected.")
		return err
	}
	models, err := auth.FetchModels(cmd.Context(), credential)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Connected via %s. %d models available.\n", credential.Method, len(models))
	return nil
}}

var authLogoutCmd = &cobra.Command{Use: "logout", Short: "Remove the saved credential", RunE: func(cmd *cobra.Command, _ []string) error {
	credential, err := auth.Resolve()
	if err == nil && credential.Method == "website" {
		_ = auth.Revoke(cmd.Context(), credential.AccessToken)
	}
	if os.Getenv("TORVE_API_KEY") != "" {
		return fmt.Errorf("TORVE_API_KEY is set in the environment; unset it to log out")
	}
	if err := auth.Delete(); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Torvecode signed out.")
	return nil
}}

func init() {
	authLoginCmd.Flags().BoolVar(&apiKeyLogin, "api-key", false, "Connect with a Torve AI API key")
	authCmd.AddCommand(authLoginCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
