package action

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/manifest"
	"github.com/novr/kusabi/internal/runner"
)

// RepoResult is the outcome of an operation on one repository.
type RepoResult struct {
	Name    string
	Output  string
	Err     error
	Skipped bool // true when the operation was skipped with a warning (not a failure)
}

// SyncProgress reports completion of one repository during sync.
type SyncProgress struct {
	Total  int
	Done   int
	Result RepoResult
}

// StatusResult is the outcome of a status check on one repository.
type StatusResult struct {
	Name         string
	Output       string
	Err          error
	IsWorktree   bool
	SyncDisabled bool
	Branch       string
	Ahead        int
	Behind       int
	Modified     int
	Untracked    int
	Cloned       bool
}

// Sync clones missing repositories and pulls existing ones.
func Sync(rootDir string, m *manifest.Manifest, depth, concurrency int, g git.Runner, onProgress func(SyncProgress)) []RepoResult {
	names := m.RepositoryNames()
	total := len(names)

	var hooks *runner.Hooks
	if onProgress != nil {
		var done atomic.Int32
		hooks = &runner.Hooks{
			OnDone: func(r runner.Result) {
				n := int(done.Add(1))
				onProgress(SyncProgress{
					Total:  total,
					Done:   n,
					Result: repoResultFromRunner(r),
				})
			},
		}
	}

	results := runner.Run(names, m.Repositories, concurrency, hooks, func(name string, repo manifest.Repository) runner.Result {
		absPath := filepath.Join(rootDir, repo.Path)
		var out string
		var opErr error

		if repo.IsSyncDisabled() {
			return runner.Result{RepoName: name, Output: "skipped: sync disabled", Skipped: true}
		}

		if g.IsRepo(absPath) {
			dirty, dirtyErr := g.IsDirty(absPath)
			if dirtyErr != nil {
				return runner.Result{RepoName: name, Err: dirtyErr}
			}
			if dirty {
				return runner.Result{RepoName: name, Output: "skipped: dirty working tree", Skipped: true}
			}
			switchMsg, alignErr := alignToDeclaredBranch(absPath, repo.Branch, g)
			if alignErr != nil {
				return runner.Result{RepoName: name, Err: alignErr}
			}
			if switchMsg == "detached" {
				return runner.Result{RepoName: name, Output: "skipped: detached HEAD", Skipped: true}
			}
			out = switchMsg

			var changed bool
			changed, opErr = g.Pull(absPath)
			if opErr == nil {
				if changed {
					out += "updated"
				} else {
					out += "updated: no change"
				}
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

// Status reports branch and working tree state for each repository.
func Status(rootDir string, m *manifest.Manifest, g git.Runner) []StatusResult {
	type enriched struct {
		name       string
		output     string
		err        error
		cloned     bool
		isWorktree   bool
		syncDisabled bool
		branch       string
		ahead      int
		behind     int
		modified   int
		untracked  int
	}

	raw := runner.RunTyped(m.RepositoryNames(), m.Repositories, 0, func(name string, repo manifest.Repository) enriched {
		absPath := filepath.Join(rootDir, repo.Path)
		r := enriched{
			name:         name,
			cloned:       g.IsRepo(absPath),
			isWorktree:   g.IsWorktree(absPath),
			syncDisabled: repo.IsSyncDisabled(),
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
		if r.isWorktree {
			r.output += "  [worktree]"
		}
		if r.syncDisabled {
			r.output += "  [sync off]"
		}
		return r
	})

	out := make([]StatusResult, len(raw))
	for i, r := range raw {
		out[i] = StatusResult{
			Name:         r.name,
			Output:       r.output,
			Err:          r.err,
			IsWorktree:   r.isWorktree,
			SyncDisabled: r.syncDisabled,
			Branch:       r.branch,
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
// If skipUncloned is true, uncloned repositories are skipped (Skipped=true) rather than failing.
func Exec(rootDir string, order []string, repos map[string]manifest.Repository, command string, skipUncloned bool, g git.Runner) []RepoResult {
	results := runner.Run(order, repos, 0, nil, func(name string, repo manifest.Repository) runner.Result {
		absPath := filepath.Join(rootDir, repo.Path)
		if !g.IsRepo(absPath) {
			if skipUncloned {
				return runner.Result{RepoName: name, Output: "skipped (not cloned)", Skipped: true}
			}
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

// alignToDeclaredBranch checks out the declared branch when needed.
// Returns a status message fragment (e.g. "switched main→develop, "), "detached" when
// HEAD is detached and no branch is declared, or "" when already aligned.
func alignToDeclaredBranch(absPath, declared string, g git.Runner) (string, error) {
	current, detached, err := g.CurrentBranch(absPath)
	if err != nil {
		return "", err
	}

	if declared == "" {
		if detached {
			return "detached", nil
		}
		return "", nil
	}

	if !detached && current == declared {
		return "", nil
	}

	if err := g.Fetch(absPath, declared); err != nil {
		return "", err
	}
	if err := g.CheckoutBranch(absPath, declared); err != nil {
		return "", err
	}

	if detached {
		return fmt.Sprintf("switched detached→%s, ", declared), nil
	}
	return fmt.Sprintf("switched %s→%s, ", current, declared), nil
}

func toRepoResults(in []runner.Result) []RepoResult {
	out := make([]RepoResult, len(in))
	for i, r := range in {
		out[i] = repoResultFromRunner(r)
	}
	return out
}

func repoResultFromRunner(r runner.Result) RepoResult {
	return RepoResult{Name: r.RepoName, Output: r.Output, Err: r.Err, Skipped: r.Skipped}
}
