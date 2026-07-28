package declaration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/manifest"
)

func TestAddRepo_RollsBackOnInvalidPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)
	if err := os.WriteFile(path, []byte(`version: "1"
name: test
repositories: {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	f, err := manifest.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	err = declaration.AddRepo(f, "app", "git@example.com/app.git", declaration.AddRepoOptions{
		Path: "../escape",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if len(f.Manifest.Repositories) != 0 {
		t.Fatalf("manifest mutated: %#v", f.Manifest.Repositories)
	}
}
