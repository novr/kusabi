package health

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/giturl"
	"github.com/novr/kusabi/internal/manifest"
)

// Check describes one doctor finding.
type Check struct {
	OK     bool
	Warn   bool   // true means warning (not a failure)
	Label  string
	Detail string
}

// Doctor inspects workspace health without mutating repositories.
func Doctor(f *manifest.File, g git.Runner) []Check {
	rootDir := f.RootDir()
	var checks []Check

	if _, err := exec.LookPath("git"); err != nil {
		checks = append(checks, Check{Label: "git command", Detail: "not found in PATH"})
	} else {
		checks = append(checks, Check{OK: true, Label: "git command"})
	}

	if f.Manifest.Context.Agents != "" {
		agentsPath := filepath.Join(rootDir, f.Manifest.Context.Agents)
		if _, err := os.Stat(agentsPath); err != nil {
			checks = append(checks, Check{
				Label:  "AGENTS.md",
				Detail: fmt.Sprintf("not found at %s", agentsPath),
			})
		} else {
			checks = append(checks, Check{OK: true, Label: "AGENTS.md"})
		}
	}

	entries := declaration.GitignoreEntries(rootDir)
	if len(entries) == 0 {
		checks = append(checks, Check{
			Label:  ".gitignore",
			Detail: "kusabi managed block is missing (run `kusabi init` or `kusabi add`)",
		})
	} else {
		checks = append(checks, Check{OK: true, Label: ".gitignore"})
	}

	for _, name := range f.Manifest.RepositoryNames() {
		repo := f.Manifest.Repositories[name]
		absPath := filepath.Join(rootDir, repo.Path)
		if !g.IsRepo(absPath) {
			if repo.IsSyncDisabled() {
				// Uncloned + sync disabled: informational, not a failure.
				checks = append(checks, Check{
					OK:     true,
					Label:  fmt.Sprintf("[%s] cloned", name),
					Detail: "not cloned (sync disabled — clone manually if needed)",
					Warn:   true,
				})
			} else {
				checks = append(checks, Check{
					Label:  fmt.Sprintf("[%s] cloned", name),
					Detail: fmt.Sprintf("not found at %s (run `kusabi sync`)", repo.Path),
				})
			}
		} else {
			checks = append(checks, Check{OK: true, Label: fmt.Sprintf("[%s] cloned", name)})
		}

		if repo.IsSyncDisabled() {
			checks = append(checks, Check{
				OK:     true,
				Label:  fmt.Sprintf("[%s] sync", name),
				Detail: "disabled",
			})
		}

		entry := repo.Path + "/"
		inManagedBlock := slices.Contains(entries, entry) || slices.Contains(entries, repo.Path)
		if inManagedBlock {
			checks = append(checks, Check{OK: true, Label: fmt.Sprintf("[%s] .gitignore", name)})
		} else if declaration.IsExcludedInFullGitignore(rootDir, repo.Path) {
			checks = append(checks, Check{
				Label:  fmt.Sprintf("[%s] .gitignore", name),
				Detail: fmt.Sprintf("%s is manually excluded but not in kusabi managed block (run `kusabi doctor --migrate-gitignore`)", repo.Path),
				Warn:   true,
			})
		} else {
			checks = append(checks, Check{
				Label:  fmt.Sprintf("[%s] .gitignore", name),
				Detail: fmt.Sprintf("%s not excluded in .gitignore", repo.Path),
			})
		}

		// Remote URL check (only for cloned repos with a declared URL).
		if g.IsRepo(absPath) && repo.URL != "" {
			actual, err := g.RemoteURL(absPath, "origin")
			if err != nil {
				checks = append(checks, Check{
					Label:  fmt.Sprintf("[%s] remote", name),
					Detail: "no origin remote configured",
				})
			} else if !giturl.Equal(repo.URL, actual) {
				checks = append(checks, Check{
					Label:  fmt.Sprintf("[%s] remote", name),
					Detail: fmt.Sprintf("declared %q but origin is %q (run `kusabi doctor --fix-remote`)", repo.URL, actual),
				})
			} else {
				checks = append(checks, Check{OK: true, Label: fmt.Sprintf("[%s] remote", name)})
			}
		}

		// Worktree detection: informational only, not a failure.
		if g.IsRepo(absPath) {
			if g.IsWorktree(absPath) {
				checks = append(checks, Check{
					OK:     true,
					Label:  fmt.Sprintf("[%s] worktree", name),
					Detail: "linked worktree detected",
				})
			}
		}
	}

	return checks
}

// HasIssues reports whether any check failed (warnings do not count as failures).
func HasIssues(checks []Check) bool {
	for _, c := range checks {
		if !c.OK && !c.Warn {
			return true
		}
	}
	return false
}

