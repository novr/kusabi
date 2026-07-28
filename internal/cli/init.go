package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/declaration"
)

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new kusabi.yaml and AGENTS.md in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			result, err := declaration.Init(cwd, force)
			if err != nil {
				return err
			}
			if result.CreatedManifest {
				fmt.Println("Created kusabi.yaml")
			}
			if result.CreatedAgents {
				fmt.Println("Created AGENTS.md")
			}
			if result.UpdatedGitignore {
				fmt.Println("Updated .gitignore")
			}
			fmt.Println("\nDone! Run `kusabi add <name> <url>` to add repositories.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files")
	return cmd
}
