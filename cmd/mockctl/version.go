package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version and build metadata",
	Long:  `Prints the version, git commit hash, build date, and Go runtime version of the Mock:ctl CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Mock:ctl CLI v%s (Internal Testing Harness)\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Build Date: %s\n", Date)
		fmt.Printf("Go Version: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
