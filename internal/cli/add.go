package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/declaration"
)

func newAddCmd() *cobra.Command {
	var (
		path string
		role string
		tags []string
	)

	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a repository to kusabi.yaml",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, url := args[0], args[1]

			f, err := declaration.OpenWorkspace(mustGetwd())
			if err != nil {
				return err
			}

			if err := declaration.AddRepo(f, name, url, declaration.AddRepoOptions{
				Path: path,
				Role: role,
				Tags: tags,
			}); err != nil {
				return err
			}

			fmt.Printf("Added repository %q (%s)\n", name, url)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Local path relative to kusabi.yaml (default: packages/<name>)")
	cmd.Flags().StringVar(&role, "role", "", "Role description for AI context")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Comma-separated tags (e.g. frontend,ios)")
	return cmd
}

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a repository from kusabi.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			f, err := declaration.OpenWorkspace(mustGetwd())
			if err != nil {
				return err
			}

			repo, err := declaration.RemoveRepo(f, name)
			if err != nil {
				return err
			}

			fmt.Printf("Removed repository %q from kusabi.yaml\n", name)
			fmt.Printf("Note: %s/ is no longer gitignored — delete the directory or add it back to kusabi.yaml.\n", repo.Path)
			return nil
		},
	}
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}
