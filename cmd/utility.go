package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jisunahamed/torvecode/internal/auth"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{Use: "models", Short: "List coding models available to this account", RunE: func(cmd *cobra.Command, _ []string) error {
	credential, err := auth.Resolve()
	if err != nil {
		return err
	}
	items, err := auth.FetchModels(cmd.Context(), credential)
	if err != nil {
		return err
	}
	for _, item := range items {
		fmt.Fprintf(cmd.OutOrStdout(), "%-32s  %-10s  %-18s  tools=%t\n", item.ID, item.Protocol, item.Operation, item.ToolUse)
	}
	return nil
}}

var doctorCmd = &cobra.Command{Use: "doctor", Short: "Check the Torvecode installation and connection", RunE: func(cmd *cobra.Command, _ []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Torvecode %s/%s\nAPI: %s\nWebsite: %s\n", runtime.GOOS, runtime.GOARCH, auth.APIURL(), auth.WebURL())
	credential, err := auth.Resolve()
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Authentication: failed (%v)\n", err)
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Authentication: %s\n", credential.Method)
	items, err := auth.FetchModels(cmd.Context(), credential)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Model catalog: %d available\n", len(items))
	return nil
}}

var runCmd = &cobra.Command{Use: "run PROMPT", Short: "Run one non-interactive task", Args: cobra.MinimumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	if err := rootCmd.Flags().Set("prompt", strings.Join(args, " ")); err != nil {
		return err
	}
	return rootCmd.RunE(rootCmd, args)
}}

func init() { _ = os.Getenv("NO_COLOR"); rootCmd.AddCommand(modelsCmd, doctorCmd, runCmd) }
