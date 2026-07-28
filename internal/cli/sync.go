package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/action"
	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
)

func newSyncCmd() *cobra.Command {
	var depth int

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Clone missing repositories and pull existing ones",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := declaration.OpenWorkspace(mustGetwd())
			if err != nil {
				return err
			}

			if len(f.Manifest.Repositories) == 0 {
				fmt.Println("No repositories defined — run `kusabi add <name> <url>`")
				return nil
			}

			// Ensure all manifest paths are present in .gitignore before cloning.
			for _, repo := range f.Manifest.Repositories {
				if err := declaration.EnsureGitignoreEntry(f.RootDir(), repo.Path); err != nil {
					fmt.Fprintf(os.Stderr, "warning: .gitignore update failed for %s: %v\n", repo.Path, err)
				}
			}

			g := &git.SystemGit{}
			results := action.Sync(f.RootDir(), f.Manifest, depth, g)

			ok := color.New(color.FgGreen)
			fail := color.New(color.FgRed)

			for _, r := range results {
				if r.Err != nil {
					fail.Printf("  ✗ %-20s %v\n", r.Name, r.Err)
				} else {
					ok.Printf("  ✓ %-20s %s\n", r.Name, r.Output)
				}
			}
			if action.HasErrors(results) {
				return fmt.Errorf("sync failed for one or more repositories")
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 0, "Git clone depth (0 = full clone)")
	return cmd
}
