package health_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/health"
	"github.com/novr/kusabi/internal/manifest"
)

type stubGit struct {
	repos map[string]bool
}

func (s *stubGit) Clone(url, path, branch string, depth int) error { return nil }
func (s *stubGit) Pull(path string) (bool, error)                   { return false, nil }
func (s *stubGit) Fetch(path, branch string) error                  { return nil }
func (s *stubGit) Status(path string) (git.StatusResult, error)     { return git.StatusResult{}, nil }
func (s *stubGit) IsRepo(path string) bool                          { return s.repos[path] }
func (s *stubGit) IsWorktree(path string) bool                      { return false }
func (s *stubGit) IsDetachedHEAD(path string) bool                  { return false }
func (s *stubGit) IsDirty(path string) (bool, error)                { return false, nil }
func (s *stubGit) RemoteURL(path, remote string) (string, error)    { return "", nil }
func (s *stubGit) SetRemoteURL(path, remote, url string) error      { return nil }
func (s *stubGit) CurrentBranch(path string) (string, bool, error)  { return "main", false, nil }
func (s *stubGit) CheckoutBranch(path, branch string) error         { return nil }

func TestDoctor_OrphanClone(t *testing.T) {
	dir := t.TempDir()
	orphan := filepath.Join(dir, "packages", "orphan")
	if err := os.MkdirAll(orphan, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(orphan, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(dir, manifest.Filename)
	if err := os.WriteFile(manifestPath, []byte(`version: "1"
name: test
repositories: {}
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := declaration.UpdateGitignore(dir, []string{"packages/"}); err != nil {
		t.Fatal(err)
	}

	f, err := manifest.Open(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	g := &stubGit{repos: map[string]bool{orphan: true}}
	checks := health.Doctor(f, g)

	var orphanClone, orphanPath int
	for _, c := range checks {
		switch c.Label {
		case "orphan clone":
			orphanClone++
		case "orphan path":
			orphanPath++
		}
	}
	if orphanClone != 1 {
		t.Fatalf("expected 1 orphan clone warning, got %d (%#v)", orphanClone, checks)
	}
	if orphanPath != 0 {
		t.Fatalf("expected no duplicate orphan path for same clone, got %d", orphanPath)
	}
}

func TestDoctor_OrphanPathNotClone(t *testing.T) {
	dir := t.TempDir()
	leftover := filepath.Join(dir, "packages", "leftover")
	if err := os.MkdirAll(leftover, 0755); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(dir, manifest.Filename)
	if err := os.WriteFile(manifestPath, []byte(`version: "1"
name: test
repositories: {}
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := declaration.UpdateGitignore(dir, []string{"packages/leftover/"}); err != nil {
		t.Fatal(err)
	}

	f, err := manifest.Open(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	g := &stubGit{repos: map[string]bool{}}
	checks := health.Doctor(f, g)

	found := false
	for _, c := range checks {
		if c.Label == "orphan path" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected orphan path warning, got %#v", checks)
	}
}
