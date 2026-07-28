package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// AppContext holds shared state injected via cobra command context.
type AppContext struct {
	ManifestPath string
	RootDir      string
}

var rootCmd = &cobra.Command{
	Use:   "kusabi",
	Short: "AI-native metarepo connector for Git",
	Long:  "Kusabi binds multiple Git repositories into a seamless, AI-native metarepo.",
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
