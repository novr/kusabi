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
	Clone(url, path string, depth int) error
	Pull(path string) error
	Status(path string) (StatusResult, error)
	IsRepo(path string) bool
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

func (g *SystemGit) Clone(url, path string, depth int) error {
	args := []string{"clone"}
	if depth > 0 {
		args = append(args, fmt.Sprintf("--depth=%d", depth))
	}
	args = append(args, url, path)
	return run("", args...)
}

func (g *SystemGit) Pull(path string) error {
	return run(path, "pull", "--ff-only")
}

func (g *SystemGit) Status(path string) (StatusResult, error) {
	var result StatusResult

	// Branch name — falls back to "(no commits)" on a freshly-init'd repo
	out, err := output(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		result.Branch = "(no commits)"
	} else {
		result.Branch = strings.TrimSpace(out)
	}

	// Ahead/behind upstream
	abOut, err := output(path, "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(abOut))
		if len(parts) == 2 {
			result.Behind, _ = strconv.Atoi(parts[0])
			result.Ahead, _ = strconv.Atoi(parts[1])
		}
	}

	// Modified/untracked files
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

func (g *SystemGit) IsRepo(path string) bool {
	dotGit := filepath.Join(path, ".git")
	_, err := os.Stat(dotGit)
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
