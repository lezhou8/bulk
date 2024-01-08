package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Bulk rename files and folders",
	Long: `Bulk is a CLI tool that opens a text editor on temporary file that lists out your selected files and allows you to rename them.`,
	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("dry-run", "d", false, "Dry run")
}
