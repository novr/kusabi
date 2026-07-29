package action

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/manifest"
	"github.com/novr/kusabi/internal/runner"
)

// RepoResult is the outcome of an operation on one repository.
type RepoResult struct {
	Name       string
	Output     string
	Err        error
	Skipped    bool // true when the operation was skipped with a warning (not a failure)
	IsWorktree bool // populated by Status; indicates a linked git worktree
}

// Sync clones missing repositories and pulls existing ones.
func Sync(rootDir string, m *manifest.Manifest, depth int, g git.Runner) []RepoResult {
	results := runner.Run(m.Repositories, 0, func(name string, repo manifest.Repository) runner.Result {
		absPath := filepath.Join(rootDir, repo.Path)
		var out string
		var opErr error

		if g.IsRepo(absPath) {
			// Skip gracefully on detached HEAD or dirty working tree.
			if g.IsDetachedHEAD(absPath) {
				return runner.Result{RepoName: name, Output: "skipped: detached HEAD", Skipped: true}
			}
			dirty, dirtyErr := g.IsDirty(absPath)
			if dirtyErr != nil {
				return runner.Result{RepoName: name, Err: dirtyErr}
			}
			if dirty {
				return runner.Result{RepoName: name, Output: "skipped: dirty working tree", Skipped: true}
			}
			opErr = g.Pull(absPath)
			if opErr == nil {
				out = "updated"
			}
		} else {
			if opErr = os.MkdirAll(filepath.Dir(absPath), 0755); opErr == nil {
				opErr = g.Clone(repo.URL, absPath, repo.Branch, depth)
				if opErr == nil {
					out = "cloned"
				}
			}
		}
		return runner.Result{RepoName: name, Output: out, Err: opErr}
	})
	return toRepoResults(results)
}

// StatusResult extends RepoResult with per-repo status details for JSON output.
type StatusResult struct {
	RepoResult
	Branch    string
	Ahead     int
	Behind    int
	Modified  int
	Untracked int
	Cloned    bool
}

// Status reports branch and working tree state for each repository.
func Status(rootDir string, m *manifest.Manifest, g git.Runner) []StatusResult {
	type enriched struct {
		name       string
		output     string
		err        error
		cloned     bool
		isWorktree bool
		branch     string
		ahead      int
		behind     int
		modified   int
		untracked  int
	}

	raw := runner.RunTyped(m.Repositories, 0, func(name string, repo manifest.Repository) enriched {
		absPath := filepath.Join(rootDir, repo.Path)
		r := enriched{
			name:       name,
			cloned:     g.IsRepo(absPath),
			isWorktree: g.IsWorktree(absPath),
		}
		if !r.cloned {
			r.output = "NOT CLONED"
			return r
		}
		s, err := g.Status(absPath)
		if err != nil {
			r.err = err
			return r
		}
		r.branch = s.Branch
		r.ahead = s.Ahead
		r.behind = s.Behind
		r.modified = s.Modified
		r.untracked = s.Untracked
		r.output = fmt.Sprintf("%-20s  +%d ~%d", s.Branch, s.Untracked, s.Modified)
		if s.Ahead > 0 || s.Behind > 0 {
			r.output += fmt.Sprintf("  (↑%d ↓%d)", s.Ahead, s.Behind)
		}
		return r
	})

	out := make([]StatusResult, len(raw))
	for i, r := range raw {
		out[i] = StatusResult{
			RepoResult: RepoResult{Name: r.name, Output: r.output, Err: r.err, IsWorktree: r.isWorktree},
			Branch:     r.branch,
			Ahead:      r.ahead,
			Behind:     r.behind,
			Modified:   r.modified,
			Untracked:  r.untracked,
			Cloned:     r.cloned,
		}
	}
	return out
}

// Exec runs a shell command in each selected repository.
func Exec(rootDir string, repos map[string]manifest.Repository, command string, g git.Runner) []RepoResult {
	results := runner.Run(repos, 0, func(name string, repo manifest.Repository) runner.Result {
		absPath := filepath.Join(rootDir, repo.Path)
		if !g.IsRepo(absPath) {
			return runner.Result{RepoName: name, Err: fmt.Errorf("not cloned — run `kusabi sync` first")}
		}
		c := exec.Command("sh", "-c", command)
		c.Dir = absPath
		out, err := c.CombinedOutput()
		return runner.Result{RepoName: name, Output: string(out), Err: err}
	})
	return toRepoResults(results)
}

// HasErrors reports whether any result failed.
func HasErrors(results []RepoResult) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}

// HasStatusErrors reports whether any status result failed.
func HasStatusErrors(results []StatusResult) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}

func toRepoResults(in []runner.Result) []RepoResult {
	out := make([]RepoResult, len(in))
	for i, r := range in {
		out[i] = RepoResult{Name: r.RepoName, Output: r.Output, Err: r.Err, Skipped: r.Skipped}
	}
	return out
}
