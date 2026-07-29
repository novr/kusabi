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
	var repoNames []string
	var skipUncloned bool

	cmd := &cobra.Command{
		Use:   "exec <command>",
		Short: "Execute a shell command in all (or filtered) repositories",
		Long:  "Execute a shell command in all (or filtered) repositories.\nPass the entire command as a single quoted string: kusabi exec 'git log --oneline -5'",
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

			repos, err := f.Manifest.FilterForExec(repoNames, tag)
			if err != nil {
				return err
			}

			g := &git.SystemGit{}
			results := action.Exec(f.RootDir(), f.Manifest.NamesInSet(repos), repos, command, skipUncloned, g)

			sep := color.New(color.FgCyan, color.Bold)
			warn := color.New(color.FgYellow)
			fail := color.New(color.FgRed)

			for _, r := range results {
				sep.Printf("\n=== %s ===\n", r.Name)
				switch {
				case r.Err != nil:
					fail.Printf("error: %v\n", r.Err)
				case r.Skipped:
					warn.Printf("%s\n", r.Output)
				default:
					if r.Output != "" {
						fmt.Print(r.Output)
						if !strings.HasSuffix(r.Output, "\n") {
							fmt.Println()
						}
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
	cmd.Flags().StringArrayVar(&repoNames, "repo", nil, "Filter repositories by name (repeatable)")
	cmd.Flags().BoolVar(&skipUncloned, "skip-uncloned", false, "Skip uncloned repositories instead of failing")
	return cmd
}
