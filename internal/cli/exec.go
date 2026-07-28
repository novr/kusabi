package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/action"
	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
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

			f, err := declaration.OpenWorkspace(mustGetwd())
			if err != nil {
				return err
			}

			if len(f.Manifest.Repositories) == 0 {
				fmt.Println("No repositories defined — run `kusabi add <name> <url>`")
				return nil
			}

			repos := f.Manifest.FilterByTag(tag)
			if len(repos) == 0 {
				return fmt.Errorf("no repositories matched tag %q", tag)
			}

			g := &git.SystemGit{}
			results := action.Exec(f.RootDir(), repos, command, g)

			sep := color.New(color.FgCyan, color.Bold)
			fail := color.New(color.FgRed)

			for _, r := range results {
				sep.Printf("\n=== %s ===\n", r.Name)
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
			if action.HasErrors(results) {
				return fmt.Errorf("exec failed for one or more repositories")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tag, "tag", "", "Filter repositories by tag")
	return cmd
}
