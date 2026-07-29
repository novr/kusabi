package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/novr/kusabi/internal/git"
)

func mustRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func tryRunGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func TestIsRepo_DirectoryGit(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	g := &git.SystemGit{}
	if !g.IsRepo(dir) {
		t.Fatal("expected directory .git to be detected as repo")
	}
}

func TestIsRepo_GitFileWorktree(t *testing.T) {
	dir := t.TempDir()
	gitdir := filepath.Join(dir, "actual-git")
	if err := os.Mkdir(gitdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: "+gitdir+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	g := &git.SystemGit{}
	if !g.IsRepo(dir) {
		t.Fatal("expected gitfile worktree to be detected as repo")
	}
}

func TestIsRepo_NotARepo(t *testing.T) {
	dir := t.TempDir()
	g := &git.SystemGit{}
	if g.IsRepo(dir) {
		t.Fatal("expected non-repo directory to return false")
	}
}

func TestIsDetachedHEAD_NotARepo(t *testing.T) {
	dir := t.TempDir()
	g := &git.SystemGit{}
	// Non-repo should return false, not panic
	if g.IsDetachedHEAD(dir) {
		t.Fatal("expected false for non-repo directory")
	}
}

func TestIsDirty_NotARepo(t *testing.T) {
	dir := t.TempDir()
	g := &git.SystemGit{}
	// Non-repo: git status fails → err != nil, but should not panic
	_, err := g.IsDirty(dir)
	if err == nil {
		t.Fatal("expected error for non-repo directory")
	}
}

func TestCurrentBranch_NotARepo(t *testing.T) {
	dir := t.TempDir()
	g := &git.SystemGit{}
	_, _, err := g.CurrentBranch(dir)
	if err == nil {
		t.Fatal("expected error for non-repo directory")
	}
}

func initTestRepo(t *testing.T, dir string) {
	t.Helper()
	if err := tryRunGit(dir, "init", "-b", "main"); err != nil {
		t.Skip("git init -b not supported (requires git >= 2.28)")
	}
	mustRunGit(t, dir, "config", "user.email", "test@example.com")
	mustRunGit(t, dir, "config", "user.name", "Test")
	mustRunGit(t, dir, "config", "commit.gpgsign", "false")
	mustRunGit(t, dir, "commit", "--allow-empty", "-m", "init")
}

func TestCurrentBranch_RealRepo(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	g := &git.SystemGit{}
	branch, detached, err := g.CurrentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("expected branch 'main', got %q", branch)
	}
	if detached {
		t.Error("expected not detached")
	}
}

func TestCheckoutBranch(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)
	mustRunGit(t, dir, "checkout", "-b", "feature")

	g := &git.SystemGit{}
	if err := g.CheckoutBranch(dir, "main"); err != nil {
		t.Fatal(err)
	}
	branch, _, err := g.CurrentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("expected 'main' after checkout, got %q", branch)
	}
}
