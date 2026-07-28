package declaration

import (
	"fmt"
	"os"
	"path/filepath"

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
	"(Declare roles in kusabi.yaml; `kusabi context` lists them in the overview section)\n"

// InitResult describes files created or updated by Init.
type InitResult struct {
	CreatedManifest bool
	CreatedAgents   bool
	UpdatedGitignore bool
}

// Init creates kusabi.yaml, AGENTS.md, and the gitignore block in dir.
func Init(dir string, force bool) (InitResult, error) {
	var result InitResult

	manifestPath := filepath.Join(dir, manifest.Filename)
	if _, err := os.Stat(manifestPath); err == nil && !force {
		return result, fmt.Errorf("%s already exists (use --force to overwrite)", manifest.Filename)
	}

	if err := os.WriteFile(manifestPath, []byte(kusabiYAMLTemplate), 0644); err != nil {
		return result, fmt.Errorf("write %s: %w", manifest.Filename, err)
	}
	result.CreatedManifest = true

	agentsPath := filepath.Join(dir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err != nil || force {
		if err := os.WriteFile(agentsPath, []byte(agentsMDTemplate), 0644); err != nil {
			return result, fmt.Errorf("write AGENTS.md: %w", err)
		}
		result.CreatedAgents = true
	}

	if err := UpdateGitignore(dir, []string{"packages/"}); err != nil {
		return result, fmt.Errorf("update .gitignore: %w", err)
	}
	result.UpdatedGitignore = true

	return result, nil
}
