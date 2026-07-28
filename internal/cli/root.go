package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "kusabi",
	Aliases:      []string{"ksb"},
	Short:        "Bind multiple Git repositories and aggregate context for agents",
	Long:         "Kusabi declares how repositories are bound, operates on them, and outputs their documents as observed context.",
	SilenceUsage: true,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		newInitCmd(),
		newSyncCmd(),
		newStatusCmd(),
		newExecCmd(),
		newContextCmd(),
		newAddCmd(),
		newRemoveCmd(),
		newDoctorCmd(),
	)
}
