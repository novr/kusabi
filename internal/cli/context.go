package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	kctx "github.com/novr/kusabi/internal/context"
	"github.com/novr/kusabi/internal/declaration"
)

func newContextCmd() *cobra.Command {
	var (
		tree      bool
		asJSON    bool
		repoNames []string
		maxBytes  int
	)

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Print AI context aggregated from all repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := declaration.OpenWorkspace(mustGetwd())
			if err != nil {
				return err
			}

			if len(repoNames) > 0 {
				if _, err := f.Manifest.FilterByNames(repoNames); err != nil {
					return err
				}
			}

			b := &kctx.Builder{
				Manifest:    f.Manifest,
				RootDir:     f.RootDir(),
				IncludeTree: tree,
				RepoNames:   repoNames,
				MaxBytes:    maxBytes,
			}

			if asJSON {
				data, err := b.BuildJSON()
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			output, err := b.Build()
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(os.Stdout, output)
			return err
		},
	}
	cmd.Flags().BoolVar(&tree, "tree", false, "Include directory structure for each repository")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as structured JSON")
	cmd.Flags().StringSliceVar(&repoNames, "repo", nil, "Limit to named repositories (repeatable)")
	cmd.Flags().IntVar(&maxBytes, "max-bytes", 0, "Cap total context size in bytes (0 = unlimited)")
	return cmd
}
