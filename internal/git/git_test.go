package git_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/novr/kusabi/internal/git"
)

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
