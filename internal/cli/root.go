package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "kusabi",
	Short:        "Bind multiple Git repositories and aggregate context for agents",
	SilenceUsage: true,
}

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
		newVersionCmd(),
	)
}
