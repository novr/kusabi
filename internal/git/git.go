package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Runner abstracts git operations for testability.
type Runner interface {
	Clone(url, path, branch string, depth int) error
	Pull(path string) error
	Fetch(path, branch string) error
	Status(path string) (StatusResult, error)
	IsRepo(path string) bool
	IsWorktree(path string) bool
	IsDetachedHEAD(path string) bool
	IsDirty(path string) (bool, error)
	RemoteURL(path, remote string) (string, error)
	SetRemoteURL(path, remote, url string) error
	CurrentBranch(path string) (branch string, detached bool, err error)
	CheckoutBranch(path, branch string) error
}

type StatusResult struct {
	Branch    string
	Ahead     int
	Behind    int
	Modified  int
	Untracked int
}

// SystemGit executes real git commands.
type SystemGit struct{}

var _ Runner = (*SystemGit)(nil)

func (g *SystemGit) Clone(url, path, branch string, depth int) error {
	args := []string{"clone"}
	if depth > 0 {
		args = append(args, fmt.Sprintf("--depth=%d", depth))
	}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, url, path)
	return run("", args...)
}

func (g *SystemGit) Pull(path string) error {
	return run(path, "pull", "--ff-only")
}

func (g *SystemGit) Status(path string) (StatusResult, error) {
	var result StatusResult

	out, err := output(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		result.Branch = "(no commits)"
	} else {
		result.Branch = strings.TrimSpace(out)
	}

	abOut, err := output(path, "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(abOut))
		if len(parts) == 2 {
			result.Behind, _ = strconv.Atoi(parts[0])
			result.Ahead, _ = strconv.Atoi(parts[1])
		}
	}

	statusOut, err := output(path, "status", "--porcelain")
	if err != nil {
		return result, err
	}
	for _, line := range strings.Split(strings.TrimSpace(statusOut), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			result.Untracked++
		} else {
			result.Modified++
		}
	}

	return result, nil
}

// IsWorktree reports whether path is a linked git worktree (vs a main clone).
// A linked worktree has .git as a regular file; a main clone has .git as a directory.
func (g *SystemGit) IsWorktree(path string) bool {
	dotGit := filepath.Join(path, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func (g *SystemGit) IsDetachedHEAD(path string) bool {
	out, err := output(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "HEAD"
}

// IsDirty reports whether the working tree has staged or unstaged changes to tracked files.
// Untracked files are not considered dirty (pull --ff-only succeeds in that state).
func (g *SystemGit) IsDirty(path string) (bool, error) {
	out, err := output(path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line != "" && !strings.HasPrefix(line, "??") {
			return true, nil
		}
	}
	return false, nil
}

// RemoteURL returns the URL configured for the named remote.
func (g *SystemGit) RemoteURL(path, remote string) (string, error) {
	out, err := output(path, "remote", "get-url", remote)
	if err != nil {
		return "", fmt.Errorf("git remote get-url %s: %w", remote, err)
	}
	return strings.TrimSpace(out), nil
}

// SetRemoteURL updates the URL for the named remote.
func (g *SystemGit) SetRemoteURL(path, remote, url string) error {
	return run(path, "remote", "set-url", remote, url)
}

// CurrentBranch returns the current branch name and whether HEAD is detached.
func (g *SystemGit) CurrentBranch(path string) (string, bool, error) {
	out, err := output(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", false, err
	}
	name := strings.TrimSpace(out)
	return name, name == "HEAD", nil
}

// Fetch fetches a specific branch from origin to ensure local remote-tracking refs are current.
// If branch is empty, fetches all refs.
func (g *SystemGit) Fetch(path, branch string) error {
	if branch != "" {
		return run(path, "fetch", "origin", branch)
	}
	return run(path, "fetch", "origin")
}

// CheckoutBranch runs git checkout <branch> in the given path.
func (g *SystemGit) CheckoutBranch(path, branch string) error {
	return run(path, "checkout", branch)
}

func (g *SystemGit) IsRepo(path string) bool {
	if _, err := output(path, "rev-parse", "--git-dir"); err == nil {
		return true
	}

	dotGit := filepath.Join(path, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return true
	}
	if !info.Mode().IsRegular() {
		return false
	}

	data, err := os.ReadFile(dotGit)
	if err != nil {
		return false
	}
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "gitdir: ") {
		return false
	}
	gitdir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir: "))
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(path, gitdir)
	}
	_, err = os.Stat(gitdir)
	return err == nil
}

func run(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

func output(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
