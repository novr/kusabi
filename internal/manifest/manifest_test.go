package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/novr/kusabi/internal/manifest"
)

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)

	m := &manifest.Manifest{
		Version:     "1",
		Name:        "test-eco",
		Description: "test",
		Repositories: map[string]manifest.Repository{
			"app": {
				Path: "packages/app",
				URL:  "git@github.com:org/app.git",
				Role: "App",
				Tags: []string{"frontend"},
			},
		},
	}

	if err := manifest.Save(m, path); err != nil {
		t.Fatal(err)
	}

	loaded, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != "test-eco" {
		t.Errorf("got Name=%q, want %q", loaded.Name, "test-eco")
	}
	repo, ok := loaded.Repositories["app"]
	if !ok {
		t.Fatal("repository 'app' not found")
	}
	if repo.URL != "git@github.com:org/app.git" {
		t.Errorf("unexpected URL: %s", repo.URL)
	}
}

func TestFind(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, manifest.Filename)
	if err := os.WriteFile(manifestPath, []byte("version: \"1\"\nname: test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Find from a nested subdirectory
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	found, err := manifest.Find(nested)
	if err != nil {
		t.Fatal(err)
	}
	if found != manifestPath {
		t.Errorf("got %q, want %q", found, manifestPath)
	}
}

func TestFind_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := manifest.Find(dir)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestBranchField_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)

	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"with-branch":    {Path: "packages/wb", URL: "git@example.com/wb.git", Branch: "develop"},
			"without-branch": {Path: "packages/wo", URL: "git@example.com/wo.git"},
		},
	}

	if err := manifest.Save(m, path); err != nil {
		t.Fatal(err)
	}
	loaded, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Repositories["with-branch"].Branch != "develop" {
		t.Errorf("Branch not persisted: got %q", loaded.Repositories["with-branch"].Branch)
	}
	if loaded.Repositories["without-branch"].Branch != "" {
		t.Errorf("Branch should be empty for repo without branch declaration")
	}
}

func TestSyncDisabled_IsSyncDisabled(t *testing.T) {
	f := false
	tr := true

	rDisabled := manifest.Repository{SyncEnabled: &f}
	rEnabled := manifest.Repository{SyncEnabled: &tr}
	rDefault := manifest.Repository{}

	if !rDisabled.IsSyncDisabled() {
		t.Error("expected IsSyncDisabled=true for SyncEnabled=false")
	}
	if rEnabled.IsSyncDisabled() {
		t.Error("expected IsSyncDisabled=false for SyncEnabled=true")
	}
	if rDefault.IsSyncDisabled() {
		t.Error("expected IsSyncDisabled=false for nil SyncEnabled")
	}
}

func TestSyncDisabled_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)

	f := false
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"disabled": {Path: "pkg/d", URL: "https://example.com/d.git", SyncEnabled: &f},
			"enabled":  {Path: "pkg/e", URL: "https://example.com/e.git"},
		},
	}

	if err := manifest.Save(m, path); err != nil {
		t.Fatal(err)
	}
	loaded, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Repositories["disabled"].IsSyncDisabled() {
		t.Error("expected disabled repo to have sync: false after round-trip")
	}
	if loaded.Repositories["enabled"].IsSyncDisabled() {
		t.Error("expected enabled repo to NOT have sync: false after round-trip")
	}
}

func TestFilterByNames(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"alpha": {Path: "pkg/alpha", URL: "https://example.com/alpha"},
			"beta":  {Path: "pkg/beta", URL: "https://example.com/beta"},
			"gamma": {Path: "pkg/gamma", URL: "https://example.com/gamma"},
		},
	}

	result, err := m.FilterByNames([]string{"alpha", "gamma"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("got %d repos, want 2", len(result))
	}
	if _, ok := result["beta"]; ok {
		t.Error("beta should not be included")
	}

	_, err = m.FilterByNames([]string{"nonexistent"})
	if err == nil {
		t.Error("expected error for unknown repository name")
	}
}

func TestFilterByTag(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"ios":     {Tags: []string{"frontend", "ios"}},
			"backend": {Tags: []string{"backend"}},
			"web":     {Tags: []string{"frontend"}},
		},
	}

	result := m.FilterByTag("frontend")
	if len(result) != 2 {
		t.Errorf("got %d repos, want 2", len(result))
	}
	if _, ok := result["backend"]; ok {
		t.Error("backend should not be included")
	}
}

func TestFilterForExec(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"ios":     {Path: "pkg/ios", Tags: []string{"frontend", "ios"}},
			"backend": {Path: "pkg/backend", Tags: []string{"backend"}},
			"web":     {Path: "pkg/web", Tags: []string{"frontend"}},
		},
	}

	repos, err := m.FilterForExec([]string{"ios", "web"}, "frontend")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("got %d repos, want 2", len(repos))
	}

	_, err = m.FilterForExec(nil, "missing")
	if err == nil {
		t.Fatal("expected error for unknown tag")
	}

	_, err = m.FilterForExec([]string{"ios"}, "backend")
	if err == nil {
		t.Fatal("expected error for empty intersection")
	}
}

func TestNamesInSet(t *testing.T) {
	m := &manifest.Manifest{
		RepositoryOrder: []string{"gamma", "alpha", "beta"},
		Repositories: map[string]manifest.Repository{
			"alpha": {Path: "pkg/alpha"},
			"beta":  {Path: "pkg/beta"},
			"gamma": {Path: "pkg/gamma"},
		},
	}

	names := m.NamesInSet(map[string]manifest.Repository{
		"beta":  m.Repositories["beta"],
		"gamma": m.Repositories["gamma"],
	})
	if len(names) != 2 || names[0] != "gamma" || names[1] != "beta" {
		t.Fatalf("unexpected order: %v", names)
	}
}
