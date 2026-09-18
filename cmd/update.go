package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an npm installation of Torvecode",
	Long: `Update Torvecode when it was installed from npm.

The npm launcher provides an interactive Update/Edit/Cancel flow. Direct binary,
Go and Homebrew installations should use their original installation channel.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return fmt.Errorf("automatic update is available in the npm installation; run `npm install -g @torveai/cli@beta`, or update through your original installation channel")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
