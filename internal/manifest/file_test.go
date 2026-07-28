package manifest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/novr/kusabi/internal/manifest"
)

func TestFilePreservesCommentsAndOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)

	initial := `version: "1"
name: test-eco
# ecosystem note
repositories:
  zebra:
    path: packages/zebra
    url: git@github.com:org/zebra.git
  alpha:
    path: packages/alpha
    url: git@github.com:org/alpha.git
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	f, err := manifest.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Manifest.RepositoryOrder) != 2 {
		t.Fatalf("got order %v", f.Manifest.RepositoryOrder)
	}
	if f.Manifest.RepositoryOrder[0] != "zebra" || f.Manifest.RepositoryOrder[1] != "alpha" {
		t.Fatalf("unexpected order: %v", f.Manifest.RepositoryOrder)
	}

	f.Manifest.Repositories["alpha"] = manifest.Repository{
		Path: "packages/alpha",
		URL:  "git@github.com:org/alpha.git",
		Role: "Updated role",
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(saved)
	if !strings.Contains(text, "# ecosystem note") {
		t.Error("comment was not preserved")
	}
	zebraIdx := strings.Index(text, "zebra:")
	alphaIdx := strings.Index(text, "alpha:")
	if zebraIdx == -1 || alphaIdx == -1 || zebraIdx > alphaIdx {
		t.Errorf("repository key order not preserved:\n%s", text)
	}
	if !strings.Contains(text, "Updated role") {
		t.Error("updated role not saved")
	}
}

func TestFileAddPreservesOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)

	f := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"first": {Path: "packages/first", URL: "git@example.com/first.git"},
		},
		RepositoryOrder: []string{"first"},
	}
	if err := manifest.Save(f, path); err != nil {
		t.Fatal(err)
	}

	opened, err := manifest.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	opened.Manifest.Repositories["second"] = manifest.Repository{
		Path: "packages/second",
		URL:  "git@example.com/second.git",
	}
	opened.Manifest.RepositoryOrder = append(opened.Manifest.RepositoryOrder, "second")
	if err := opened.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := manifest.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Manifest.RepositoryOrder) != 2 || reloaded.Manifest.RepositoryOrder[1] != "second" {
		t.Fatalf("got order %v", reloaded.Manifest.RepositoryOrder)
	}
}

func TestSave_PreservesRepositoryOrderFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)

	initial := `version: "1"
name: test
repositories:
  first:
    path: packages/first
    url: git@example.com/first.git
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded.Repositories["first"] = manifest.Repository{
		Path: "packages/first",
		URL:  "git@example.com/first.git",
		Role: "updated",
	}
	if err := manifest.Save(loaded, path); err != nil {
		t.Fatal(err)
	}

	reopened, err := manifest.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened.Manifest.RepositoryOrder) != 1 || reopened.Manifest.RepositoryOrder[0] != "first" {
		t.Fatalf("order not preserved: %v", reopened.Manifest.RepositoryOrder)
	}
}
