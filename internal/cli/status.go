package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/action"
	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/manifest"
)

type statusJSON struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Branch     string `json:"branch"`
	Ahead      int    `json:"ahead"`
	Behind     int    `json:"behind"`
	Modified   int    `json:"modified"`
	Untracked  int    `json:"untracked"`
	IsWorktree bool   `json:"is_worktree"`
	Cloned     bool   `json:"cloned"`
	Error      string `json:"error,omitempty"`
}

func newStatusCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
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

			if jsonOut {
				return printStatusJSON(results, f.Manifest)
			}

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
			if action.HasStatusErrors(results) {
				return fmt.Errorf("status failed for one or more repositories")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output status as JSON")
	return cmd
}

func printStatusJSON(results []action.StatusResult, m *manifest.Manifest) error {
	entries := make([]statusJSON, len(results))
	hasErr := false
	for i, r := range results {
		e := statusJSON{
			Name:       r.Name,
			Path:       m.Repositories[r.Name].Path,
			Branch:     r.Branch,
			Ahead:      r.Ahead,
			Behind:     r.Behind,
			Modified:   r.Modified,
			Untracked:  r.Untracked,
			IsWorktree: r.IsWorktree,
			Cloned:     r.Cloned,
		}
		if r.Err != nil {
			e.Error = r.Err.Error()
			hasErr = true
		}
		entries[i] = e
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return err
	}
	if hasErr {
		return fmt.Errorf("status failed for one or more repositories")
	}
	return nil
}
