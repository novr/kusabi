package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/manifest"
)

func newAddCmd() *cobra.Command {
	var (
		path string
		role string
		tags []string
	)

	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a repository to kusabi.yaml",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, url := args[0], args[1]

			if err := validateRepoName(name); err != nil {
				return err
			}

			manifestPath, err := manifest.Find(mustGetwd())
			if err != nil {
				return err
			}
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}

			if _, exists := m.Repositories[name]; exists {
				return fmt.Errorf("repository %q already exists", name)
			}

			if path == "" {
				path = filepath.Join("packages", name)
			}

			m.Repositories[name] = manifest.Repository{
				Path: path,
				URL:  url,
				Role: role,
				Tags: tags,
			}

			if err := manifest.Save(m, manifestPath); err != nil {
				return err
			}

			// Ensure the path is excluded from .gitignore (accumulate, don't replace)
			rootDir := filepath.Dir(manifestPath)
			entries := gitignoreEntries(rootDir)
			entry := path + "/"
			if !contains(entries, entry) && !contains(entries, path) {
				entries = append(entries, entry)
				if err := updateGitignore(rootDir, entries); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to update .gitignore: %v\n", err)
				}
			}

			fmt.Printf("Added repository %q (%s)\n", name, url)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Local path relative to kusabi.yaml (default: packages/<name>)")
	cmd.Flags().StringVar(&role, "role", "", "Role description for AI context")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Comma-separated tags (e.g. frontend,ios)")
	return cmd
}

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a repository from kusabi.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			manifestPath, err := manifest.Find(mustGetwd())
			if err != nil {
				return err
			}
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}

			repo, exists := m.Repositories[name]
			if !exists {
				return fmt.Errorf("repository %q not found", name)
			}

			delete(m.Repositories, name)

			if err := manifest.Save(m, manifestPath); err != nil {
				return err
			}

			// Remove the repo's path from .gitignore
			rootDir := filepath.Dir(manifestPath)
			entries := gitignoreEntries(rootDir)
			entries = removeEntry(entries, repo.Path+"/")
			entries = removeEntry(entries, repo.Path)
			if err := updateGitignore(rootDir, entries); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to update .gitignore: %v\n", err)
			}

			fmt.Printf("Removed repository %q from kusabi.yaml\n", name)
			fmt.Println("Note: local files were not deleted.")
			return nil
		},
	}
}

// validateRepoName rejects names that would cause path traversal.
func validateRepoName(name string) error {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") || name == "" {
		return fmt.Errorf("repository name %q must not contain '/', '\\', or '..'", name)
	}
	return nil
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeEntry(slice []string, s string) []string {
	out := slice[:0]
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}
