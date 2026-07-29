package action_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/novr/kusabi/internal/action"
	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/manifest"
)

// fakeGit implements git.Runner for testing.
type fakeGit struct {
	repos              map[string]bool
	worktrees          map[string]bool
	statusFunc         func(path string) (git.StatusResult, error)
	currentBranchFunc  func(path string) (string, bool, error)
	fetchCalls         []string
	checkoutCalls      []string
	pullFunc         func(path string) (bool, error)
}

func (f *fakeGit) Clone(url, path, branch string, depth int) error { return nil }
func (f *fakeGit) Pull(path string) (bool, error) {
	if f.pullFunc != nil {
		return f.pullFunc(path)
	}
	return true, nil
}
func (f *fakeGit) Fetch(path, branch string) error {
	f.fetchCalls = append(f.fetchCalls, path+":"+branch)
	return nil
}
func (f *fakeGit) IsRepo(path string) bool                                   { return f.repos[path] }
func (f *fakeGit) IsWorktree(path string) bool                               { return f.worktrees[path] }
func (f *fakeGit) IsDetachedHEAD(path string) bool                          { return false }
func (f *fakeGit) IsDirty(path string) (bool, error)                        { return false, nil }
func (f *fakeGit) RemoteURL(path, remote string) (string, error)             { return "", nil }
func (f *fakeGit) SetRemoteURL(path, remote, url string) error               { return nil }
func (f *fakeGit) CurrentBranch(path string) (string, bool, error) {
	if f.currentBranchFunc != nil {
		return f.currentBranchFunc(path)
	}
	return "main", false, nil
}
func (f *fakeGit) CheckoutBranch(path, branch string) error {
	f.checkoutCalls = append(f.checkoutCalls, path+":"+branch)
	return nil
}
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
	if !strings.Contains(byName["repo-a"].Output, "[worktree]") {
		t.Errorf("expected worktree marker in output, got %q", byName["repo-a"].Output)
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
		{Name: "x", Err: errors.New("fail")},
	}
	if !action.HasStatusErrors(withErr) {
		t.Error("expected HasStatusErrors=true")
	}

	noErr := []action.StatusResult{
		{Name: "x"},
	}
	if action.HasStatusErrors(noErr) {
		t.Error("expected HasStatusErrors=false")
	}
}

func TestSync_AlignsDeclaredBranch(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"app": {Path: "pkg/app", URL: "https://example.com/app.git", Branch: "develop"},
		},
	}

	g := &fakeGit{
		repos: map[string]bool{"pkg/app": true},
		currentBranchFunc: func(path string) (string, bool, error) {
			return "main", false, nil
		},
	}

	results := action.Sync("", m, 0, g, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if len(g.fetchCalls) != 1 || g.fetchCalls[0] != "pkg/app:develop" {
		t.Errorf("expected fetch pkg/app:develop, got %v", g.fetchCalls)
	}
	if len(g.checkoutCalls) != 1 || g.checkoutCalls[0] != "pkg/app:develop" {
		t.Errorf("expected checkout pkg/app:develop, got %v", g.checkoutCalls)
	}
	if !strings.Contains(results[0].Output, "switched main→develop") {
		t.Errorf("expected switch message, got %q", results[0].Output)
	}
}

func TestSync_DetachedHEADWithDeclaredBranch_ChecksOut(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"app": {Path: "pkg/app", URL: "https://example.com/app.git", Branch: "develop"},
		},
	}

	g := &fakeGit{
		repos: map[string]bool{"pkg/app": true},
		currentBranchFunc: func(path string) (string, bool, error) {
			return "HEAD", true, nil
		},
	}

	results := action.Sync("", m, 0, g, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Skipped {
		t.Fatalf("expected checkout from detached HEAD, got skip: %q", results[0].Output)
	}
	if len(g.checkoutCalls) != 1 {
		t.Errorf("expected checkout, got %v", g.checkoutCalls)
	}
}

func TestSync_DetachedHEADWithoutDeclaredBranch_Skips(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"app": {Path: "pkg/app", URL: "https://example.com/app.git"},
		},
	}

	g := &fakeGit{
		repos: map[string]bool{"pkg/app": true},
		currentBranchFunc: func(path string) (string, bool, error) {
			return "HEAD", true, nil
		},
	}

	results := action.Sync("", m, 0, g, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Skipped {
		t.Fatalf("expected skip, got %q err=%v", results[0].Output, results[0].Err)
	}
}

func TestSync_PullNoChange(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"app": {Path: "pkg/app", URL: "https://example.com/app.git"},
		},
	}

	g := &fakeGit{
		repos: map[string]bool{"pkg/app": true},
		pullFunc: func(path string) (bool, error) {
			return false, nil
		},
	}

	results := action.Sync("", m, 0, g, nil)
	if results[0].Output != "updated: no change" {
		t.Errorf("expected updated: no change, got %q", results[0].Output)
	}
}
