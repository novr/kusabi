package health

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/giturl"
	"github.com/novr/kusabi/internal/manifest"
)

// Check describes one doctor finding.
type Check struct {
	OK     bool
	Warn   bool // true means warning (not a failure)
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
		agentsLabel := f.Manifest.Context.Agents
		if _, err := os.Stat(agentsPath); err != nil {
			checks = append(checks, Check{
				Label:  agentsLabel,
				Detail: fmt.Sprintf("not found at %s", agentsPath),
			})
		} else {
			checks = append(checks, Check{OK: true, Label: agentsLabel})
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

		if repo.Branch != "" && g.IsRepo(absPath) {
			current, detached, err := g.CurrentBranch(absPath)
			if err != nil {
				checks = append(checks, Check{
					Label:  fmt.Sprintf("[%s] branch", name),
					Detail: fmt.Sprintf("could not determine current branch: %v", err),
				})
			} else if detached {
				checks = append(checks, Check{
					Label:  fmt.Sprintf("[%s] branch", name),
					Detail: fmt.Sprintf("detached HEAD (declared: %s)", repo.Branch),
				})
			} else if current != repo.Branch {
				checks = append(checks, Check{
					Label:  fmt.Sprintf("[%s] branch", name),
					Detail: fmt.Sprintf("%q (declared: %q — run `kusabi sync`)", current, repo.Branch),
				})
			} else {
				checks = append(checks, Check{OK: true, Label: fmt.Sprintf("[%s] branch", name)})
			}
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

		if g.IsRepo(absPath) && g.IsWorktree(absPath) {
			checks = append(checks, Check{
				OK:     true,
				Label:  fmt.Sprintf("[%s] worktree", name),
				Detail: "linked worktree detected",
			})
		}
	}

	checks = append(checks, orphanChecks(rootDir, f, g, entries)...)

	return checks
}

func orphanChecks(rootDir string, f *manifest.File, g git.Runner, gitignoreEntries []string) []Check {
	declared := make(map[string]bool, len(f.Manifest.Repositories))
	for _, name := range f.Manifest.RepositoryNames() {
		declared[filepath.ToSlash(f.Manifest.Repositories[name].Path)] = true
	}

	var checks []Check
	seen := make(map[string]bool)

	warnOrphan := func(path, label, detail string) {
		path = filepath.ToSlash(path)
		if path == "" || path == "packages" || declared[path] || seen[path] {
			return
		}
		seen[path] = true
		checks = append(checks, Check{
			OK:     true,
			Warn:   true,
			Label:  label,
			Detail: detail,
		})
	}

	// Prefer clone-specific warning for undeclared git repos under packages/.
	packagesDir := filepath.Join(rootDir, "packages")
	if children, err := os.ReadDir(packagesDir); err == nil {
		for _, child := range children {
			if !child.IsDir() {
				continue
			}
			path := filepath.ToSlash(filepath.Join("packages", child.Name()))
			abs := filepath.Join(rootDir, filepath.FromSlash(path))
			if !g.IsRepo(abs) {
				continue
			}
			warnOrphan(path, "orphan clone", fmt.Sprintf("%q exists on disk but is not declared in kusabi.yaml", path))
		}
	}

	for _, entry := range gitignoreEntries {
		path := filepath.ToSlash(strings.TrimSuffix(entry, "/"))
		abs := filepath.Join(rootDir, filepath.FromSlash(path))
		if _, err := os.Stat(abs); err != nil {
			continue
		}
		warnOrphan(path, "orphan path", fmt.Sprintf("%q is gitignored but not declared in kusabi.yaml", path))
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
