package cli

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/manifest"
	"github.com/novr/kusabi/internal/runner"
)

func newExecCmd() *cobra.Command {
	var tag string

	cmd := &cobra.Command{
		Use:   "exec <command>",
		Short: "Execute a shell command in all (or tagged) repositories",
		Long:  "Execute a shell command in all (or tagged) repositories.\nPass the entire command as a single quoted string: kusabi exec 'git log --oneline -5'",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := args[0]

			manifestPath, err := manifest.Find(mustGetwd())
			if err != nil {
				return err
			}
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}

			rootDir := filepath.Dir(manifestPath)
			repos := m.FilterByTag(tag)

			if len(repos) == 0 {
				return fmt.Errorf("no repositories matched (tag=%q)", tag)
			}

			results := runner.Run(repos, 0, func(name string, repo manifest.Repository) runner.Result {
				absPath := filepath.Join(rootDir, repo.Path)
				c := exec.Command("sh", "-c", command)
				c.Dir = absPath
				out, err := c.CombinedOutput()
				return runner.Result{RepoName: name, Output: string(out), Err: err}
			})

			sep := color.New(color.FgCyan, color.Bold)
			fail := color.New(color.FgRed)

			for _, r := range results {
				sep.Printf("\n=== %s ===\n", r.RepoName)
				if r.Err != nil {
					fail.Printf("error: %v\n", r.Err)
				}
				if r.Output != "" {
					fmt.Print(r.Output)
					if !strings.HasSuffix(r.Output, "\n") {
						fmt.Println()
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tag, "tag", "", "Filter repositories by tag")
	return cmd
}
