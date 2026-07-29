package cli

import (
	"fmt"
	"os"
	"sync"

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

			ok := color.New(color.FgGreen)
			warn := color.New(color.FgYellow)
			fail := color.New(color.FgRed)

			total := len(f.Manifest.Repositories)
			fmt.Fprintf(os.Stderr, "Syncing %d repositories…\n", total)

			var mu sync.Mutex
			printResult := func(r action.RepoResult, done, total int) {
				prefix := fmt.Sprintf("[%d/%d]", done, total)
				switch {
				case r.Err != nil:
					printRepoError(os.Stderr, fail, prefix, r.Name, r.Err)
				case r.Skipped:
					warn.Printf("  ⚠ %s %-20s %s\n", prefix, r.Name, r.Output)
				default:
					ok.Printf("  ✓ %s %-20s %s\n", prefix, r.Name, r.Output)
				}
			}

			results := action.Sync(f.RootDir(), f.Manifest, depth, g, func(p action.SyncProgress) {
				if p.Started {
					return
				}
				mu.Lock()
				defer mu.Unlock()
				printResult(p.Result, p.Done, p.Total)
			})
			if action.HasErrors(results) {
				return fmt.Errorf("sync failed for one or more repositories")
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 0, "Git clone depth (0 = full clone)")
	return cmd
}
