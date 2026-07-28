package cli

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/manifest"
	"github.com/novr/kusabi/internal/runner"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show branch and working tree status for all repositories",
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

			results := runner.Run(m.Repositories, 0, func(name string, repo manifest.Repository) runner.Result {
				absPath := filepath.Join(rootDir, repo.Path)
				if !g.IsRepo(absPath) {
					return runner.Result{RepoName: name, Output: "NOT CLONED"}
				}
				status, err := g.Status(absPath)
				if err != nil {
					return runner.Result{RepoName: name, Err: err}
				}
				out := fmt.Sprintf("%-20s  +%d ~%d", status.Branch, status.Untracked, status.Modified)
				if status.Ahead > 0 || status.Behind > 0 {
					out += fmt.Sprintf("  (↑%d ↓%d)", status.Ahead, status.Behind)
				}
				return runner.Result{RepoName: name, Output: out}
			})

			bold := color.New(color.Bold)
			warn := color.New(color.FgYellow)
			fail := color.New(color.FgRed)

			for _, r := range results {
				prefix := fmt.Sprintf("  [%-15s]  ", r.RepoName)
				switch {
				case r.Err != nil:
					fail.Printf("%s%v\n", prefix, r.Err)
				case r.Output == "NOT CLONED":
					warn.Printf("%s%s\n", prefix, r.Output)
				default:
					bold.Printf("%s", prefix)
					fmt.Println(r.Output)
				}
			}
			return nil
		},
	}
}
