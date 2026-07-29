package declaration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/novr/kusabi/internal/declaration"
)

func TestIsExcludedInFullGitignore(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")

	write := func(content string) {
		t.Helper()
		if err := os.WriteFile(gitignorePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("path in managed block is detected", func(t *testing.T) {
		write("# kusabi managed — do not edit below\npackages/app/\n# kusabi managed — end\n")
		if !declaration.IsExcludedInFullGitignore(dir, "packages/app") {
			t.Error("expected true for path in managed block")
		}
	})

	t.Run("manually excluded path is detected", func(t *testing.T) {
		write("# hand-written\npackages/app/\n")
		if !declaration.IsExcludedInFullGitignore(dir, "packages/app") {
			t.Error("expected true for manually excluded path")
		}
	})

	t.Run("path without trailing slash is detected", func(t *testing.T) {
		write("packages/app\n")
		if !declaration.IsExcludedInFullGitignore(dir, "packages/app") {
			t.Error("expected true for path without trailing slash")
		}
	})

	t.Run("unrelated path returns false", func(t *testing.T) {
		write("packages/other/\n")
		if declaration.IsExcludedInFullGitignore(dir, "packages/app") {
			t.Error("expected false for unrelated path")
		}
	})

	t.Run("comment lines are not matched", func(t *testing.T) {
		write("# packages/app/\n")
		if declaration.IsExcludedInFullGitignore(dir, "packages/app") {
			t.Error("expected false for comment line")
		}
	})

	t.Run("no .gitignore returns false", func(t *testing.T) {
		dir2 := t.TempDir()
		if declaration.IsExcludedInFullGitignore(dir2, "packages/app") {
			t.Error("expected false when no .gitignore")
		}
	})
}
