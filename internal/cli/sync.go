package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/manifest"
	"github.com/novr/kusabi/internal/runner"
)

func newSyncCmd() *cobra.Command {
	var depth int

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Clone missing repositories and pull existing ones",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath, err := manifest.Find(mustGetwd())
			if err != nil {
				return err
			}
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}

			rootDir := filepath.Dir(manifestPath)
			g := &git.SystemGit{}

			// Merge manifest paths into existing .gitignore entries (preserve user-added entries like packages/)
			existing := gitignoreEntries(rootDir)
			existingSet := make(map[string]bool, len(existing))
			for _, e := range existing {
				existingSet[e] = true
			}
			for _, repo := range m.Repositories {
				entry := repo.Path + "/"
				if !existingSet[entry] && !existingSet[repo.Path] {
					existing = append(existing, entry)
				}
			}
			if err := updateGitignore(rootDir, existing); err != nil {
				fmt.Fprintf(os.Stderr, "warning: .gitignore update failed: %v\n", err)
			}

			results := runner.Run(m.Repositories, 0, func(name string, repo manifest.Repository) runner.Result {
				absPath := filepath.Join(rootDir, repo.Path)
				var out string
				var opErr error

				if g.IsRepo(absPath) {
					opErr = g.Pull(absPath)
					if opErr == nil {
						out = "updated"
					}
				} else {
					if opErr = os.MkdirAll(filepath.Dir(absPath), 0755); opErr == nil {
						opErr = g.Clone(repo.URL, absPath, depth)
						if opErr == nil {
							out = "cloned"
						}
					}
				}
				return runner.Result{RepoName: name, Output: out, Err: opErr}
			})

			ok := color.New(color.FgGreen)
			fail := color.New(color.FgRed)

			for _, r := range results {
				if r.Err != nil {
					fail.Printf("  ✗ %-20s %v\n", r.RepoName, r.Err)
				} else {
					ok.Printf("  ✓ %-20s %s\n", r.RepoName, r.Output)
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&depth, "depth", 0, "Git clone depth (0 = full clone)")
	return cmd
}
