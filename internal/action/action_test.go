package action_test

import (
	"errors"
	"testing"

	"github.com/novr/kusabi/internal/action"
	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/manifest"
)

// fakeGit implements git.Runner for testing.
type fakeGit struct {
	repos      map[string]bool // path → isRepo
	worktrees  map[string]bool // path → isWorktree
	statusFunc func(path string) (git.StatusResult, error)
}

func (f *fakeGit) Clone(url, path string, depth int) error { return nil }
func (f *fakeGit) Pull(path string) error                  { return nil }
func (f *fakeGit) IsRepo(path string) bool                 { return f.repos[path] }
func (f *fakeGit) IsWorktree(path string) bool             { return f.worktrees[path] }
func (f *fakeGit) Status(path string) (git.StatusResult, error) {
	if f.statusFunc != nil {
		return f.statusFunc(path)
	}
	return git.StatusResult{Branch: "main"}, nil
}

func TestStatus_IsWorktreePopulated(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"repo-a": {Path: "repos/a", URL: "https://example.com/a"},
			"repo-b": {Path: "repos/b", URL: "https://example.com/b"},
		},
	}

	g := &fakeGit{
		repos:     map[string]bool{"repos/a": true, "repos/b": true},
		worktrees: map[string]bool{"repos/a": true, "repos/b": false},
	}

	results := action.Status("", m, g)

	byName := make(map[string]action.StatusResult, len(results))
	for _, r := range results {
		byName[r.Name] = r
	}

	if !byName["repo-a"].IsWorktree {
		t.Error("expected repo-a to be flagged as worktree")
	}
	if byName["repo-b"].IsWorktree {
		t.Error("expected repo-b to NOT be flagged as worktree")
	}
}

func TestStatus_NotCloned(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"repo-x": {Path: "repos/x", URL: "https://example.com/x"},
		},
	}

	g := &fakeGit{repos: map[string]bool{}}

	results := action.Status("", m, g)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Cloned {
		t.Error("expected Cloned=false for uncloned repo")
	}
	if r.Output != "NOT CLONED" {
		t.Errorf("expected output 'NOT CLONED', got %q", r.Output)
	}
}

func TestHasStatusErrors(t *testing.T) {
	withErr := []action.StatusResult{
		{RepoResult: action.RepoResult{Name: "x", Err: errors.New("fail")}},
	}
	if !action.HasStatusErrors(withErr) {
		t.Error("expected HasStatusErrors=true")
	}

	noErr := []action.StatusResult{
		{RepoResult: action.RepoResult{Name: "x"}},
	}
	if action.HasStatusErrors(noErr) {
		t.Error("expected HasStatusErrors=false")
	}
}
