package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/manifest"
)

const kusabiYAMLTemplate = `version: "1"
name: "my-ecosystem"
description: "Cross-platform ecosystem bound by Kusabi."

context:
  agents: "./AGENTS.md"
  includes:
    - "README.md"
    - "CLAUDE.md"

repositories: {}
`

const agentsMDTemplate = "# AGENTS.md \xe2\x80\x94 Global AI Policy\n\n" +
	"This project is managed by Kusabi (\xe6\xa5\x94). Multiple repositories are bound here.\n\n" +
	"## Architecture Overview\n\n" +
	"(Describe your system architecture here)\n\n" +
	"## Development Rules\n\n" +
	"- Commit and Git operations must be performed within each specific sub-repository directory.\n" +
	"- When implementing features that span repositories, coordinate changes explicitly.\n\n" +
	"## Repository Roles\n\n" +
	"(Kusabi will inject repository roles automatically via `kusabi context`)\n"

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new kusabi.yaml and AGENTS.md in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			manifestPath := filepath.Join(cwd, manifest.Filename)
			if _, err := os.Stat(manifestPath); err == nil && !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", manifest.Filename)
			}

			if err := os.WriteFile(manifestPath, []byte(kusabiYAMLTemplate), 0644); err != nil {
				return fmt.Errorf("write %s: %w", manifest.Filename, err)
			}
			fmt.Printf("Created %s\n", manifest.Filename)

			agentsPath := filepath.Join(cwd, "AGENTS.md")
			if _, err := os.Stat(agentsPath); err != nil || force {
				if err := os.WriteFile(agentsPath, []byte(agentsMDTemplate), 0644); err != nil {
					return fmt.Errorf("write AGENTS.md: %w", err)
				}
				fmt.Println("Created AGENTS.md")
			}

			if err := updateGitignore(cwd, []string{"packages/"}); err != nil {
				return fmt.Errorf("update .gitignore: %w", err)
			}
			fmt.Println("Updated .gitignore")

			fmt.Println("\nDone! Run `kusabi add <name> <url>` to add repositories.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files")
	return cmd
}

const gitignoreMarkerStart = "# kusabi managed — do not edit below"
const gitignoreMarkerEnd = "# kusabi managed — end"

// updateGitignore ensures the given entries exist within the kusabi-managed block.
func updateGitignore(dir string, entries []string) error {
	path := filepath.Join(dir, ".gitignore")

	var existing string
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	// Remove old managed block
	var head, tail string
	if start := strings.Index(existing, gitignoreMarkerStart); start != -1 {
		head = strings.TrimRight(existing[:start], "\n")
		if end := strings.Index(existing[start:], gitignoreMarkerEnd); end != -1 {
			tail = strings.TrimLeft(existing[start+end+len(gitignoreMarkerEnd):], "\n")
		}
	} else {
		head = strings.TrimRight(existing, "\n")
	}

	var block strings.Builder
	block.WriteString(gitignoreMarkerStart + "\n")
	for _, e := range entries {
		block.WriteString(e + "\n")
	}
	block.WriteString(gitignoreMarkerEnd + "\n")

	var out strings.Builder
	if head != "" {
		out.WriteString(head + "\n\n")
	}
	out.WriteString(block.String())
	if tail != "" {
		out.WriteString("\n" + tail)
	}

	return os.WriteFile(path, []byte(out.String()), 0644)
}

// gitignoreEntries extracts the list of entries in the kusabi-managed block.
func gitignoreEntries(dir string) []string {
	path := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)

	start := strings.Index(content, gitignoreMarkerStart)
	if start == -1 {
		return nil
	}
	end := strings.Index(content[start:], gitignoreMarkerEnd)
	if end == -1 {
		return nil
	}

	block := content[start+len(gitignoreMarkerStart) : start+end]
	var entries []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			entries = append(entries, line)
		}
	}
	return entries
}
