package cmd

import (
	"fmt"
	"os"

	"github.com/jisunahamed/torvecode/internal/documents"
	"github.com/spf13/cobra"
)

var documentOutput string

var documentsCmd = &cobra.Command{
	Use:   "documents",
	Short: "Convert documents with Microsoft MarkItDown",
}

var documentsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check local MarkItDown support",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		if documents.Available() {
			fmt.Fprintln(cmd.OutOrStdout(), "Microsoft MarkItDown is available.")
			return
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Microsoft MarkItDown is not installed. %s\n", documents.InstallHint())
	},
}

var documentsConvertCmd = &cobra.Command{
	Use:   "convert FILE",
	Short: "Convert a local document to Markdown",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		markdown, err := documents.Convert(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		if documentOutput == "" {
			fmt.Fprintln(cmd.OutOrStdout(), markdown)
			return nil
		}
		if err := os.WriteFile(documentOutput, []byte(markdown+"\n"), 0600); err != nil {
			return fmt.Errorf("write Markdown output: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Converted document: %s\n", documentOutput)
		return nil
	},
}

func init() {
	documentsConvertCmd.Flags().StringVarP(&documentOutput, "output", "o", "", "Write Markdown to this file")
	documentsCmd.AddCommand(documentsStatusCmd, documentsConvertCmd)
	rootCmd.AddCommand(documentsCmd)
}
