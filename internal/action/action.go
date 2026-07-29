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
	Name    string
	Output  string
	Err     error
	Skipped bool // true when the operation was skipped with a warning (not a failure)
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
			if dirty, err := g.IsDirty(absPath); err == nil && dirty {
				return runner.Result{RepoName: name, Output: "skipped: dirty working tree", Skipped: true}
			}
			if repo.Branch != "" {
				current, _, err := g.CurrentBranch(absPath)
				if err != nil {
					return runner.Result{RepoName: name, Err: err}
				}
				if current != repo.Branch {
					// Fetch before checkout so DWIM can resolve remote-tracking refs.
					if opErr = g.Fetch(absPath, repo.Branch); opErr != nil {
						return runner.Result{RepoName: name, Err: opErr}
					}
					if opErr = g.CheckoutBranch(absPath, repo.Branch); opErr != nil {
						return runner.Result{RepoName: name, Err: opErr}
					}
					out = fmt.Sprintf("switched %s→%s, ", current, repo.Branch)
				}
			}
			opErr = g.Pull(absPath)
			if opErr == nil {
				out += "updated"
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
func Status(rootDir string, m *manifest.Manifest, g git.Runner) []RepoResult {
	results := runner.Run(m.Repositories, 0, func(name string, repo manifest.Repository) runner.Result {
		absPath := filepath.Join(rootDir, repo.Path)
		if !g.IsRepo(absPath) {
			return runner.Result{RepoName: name, Output: "NOT CLONED"}
		}
		status, err := g.Status(absPath)
		if err != nil {
			return runner.Result{RepoName: name, Err: err}
		}
		out := fmt.Sprintf("%-20s  +%d ~%d", status.Branch, status.Untracked, status.Modified)
		if status.Ahead > 0 || status.Behind > 0 {
			out += fmt.Sprintf("  (↑%d ↓%d)", status.Ahead, status.Behind)
		}
		return runner.Result{RepoName: name, Output: out}
	})
	return toRepoResults(results)
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

func toRepoResults(in []runner.Result) []RepoResult {
	out := make([]RepoResult, len(in))
	for i, r := range in {
		out[i] = RepoResult{Name: r.RepoName, Output: r.Output, Err: r.Err, Skipped: r.Skipped}
	}
	return out
}
