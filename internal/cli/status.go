package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/action"
	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show branch and working tree status for all repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := declaration.OpenWorkspace(mustGetwd())
			if err != nil {
				return err
			}

			if len(f.Manifest.Repositories) == 0 {
				fmt.Println("No repositories defined — run `kusabi add <name> <url>`")
				return nil
			}

			g := &git.SystemGit{}
			results := action.Status(f.RootDir(), f.Manifest, g)

			bold := color.New(color.Bold)
			warn := color.New(color.FgYellow)
			fail := color.New(color.FgRed)

			for _, r := range results {
				prefix := fmt.Sprintf("  [%-15s]  ", r.Name)
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
			if action.HasErrors(results) {
				return fmt.Errorf("status failed for one or more repositories")
			}
			return nil
		},
	}
}
