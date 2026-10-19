package giturl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/novr/kusabi/internal/giturl"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"owner/repo", "yumemi/coop9-android", "https://github.com/yumemi/coop9-android"},
		{"with .git", "org/repo.git", "https://github.com/org/repo"},
		{"trims space", "  org/repo  ", "https://github.com/org/repo"},
		{"https unchanged", "https://github.com/org/repo", "https://github.com/org/repo"},
		{"ssh unchanged", "git@github.com:org/repo.git", "git@github.com:org/repo.git"},
		{"nested not shorthand", "org/group/repo", "org/group/repo"},
		{"dot-relative path", "./local/repo", "./local/repo"},
		{"parent-relative path", "../other/repo", "../other/repo"},
		{"absolute path", "/tmp/repo", "/tmp/repo"},
		{"empty", "", ""},
		{"single segment", "repo", "repo"},
		{"leading hyphen owner", "-org/repo", "-org/repo"},
		// Bare relative paths are GitHub shorthand (gh semantics); use ./ for local.
		{"bare relative expands", "vendor/mirror", "https://github.com/vendor/mirror"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := giturl.Expand(tt.input)
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	root := t.TempDir()
	local := filepath.Join(root, "vendor", "mirror")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"missing bare path expands", "yumemi/coop9-android", "https://github.com/yumemi/coop9-android"},
		{"existing bare local kept", "vendor/mirror", "vendor/mirror"},
		{"dot-relative kept", "./vendor/mirror", "./vendor/mirror"},
		{"https unchanged", "https://github.com/org/repo", "https://github.com/org/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := giturl.Resolve(tt.input, root)
			if got != tt.want {
				t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}

	t.Run("empty rootDir acts like Expand", func(t *testing.T) {
		got := giturl.Resolve("vendor/mirror", "")
		want := "https://github.com/vendor/mirror"
		if got != want {
			t.Errorf("Resolve without root = %q, want %q", got, want)
		}
	})
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"scp with .git", "git@github.com:org/repo.git", "https://github.com/org/repo"},
		{"scp without .git", "git@github.com:org/repo", "https://github.com/org/repo"},
		{"https with .git", "https://github.com/org/repo.git", "https://github.com/org/repo"},
		{"https without .git", "https://github.com/org/repo", "https://github.com/org/repo"},
		{"https with user", "https://user@github.com/org/repo.git", "https://github.com/org/repo"},
		{"ssh scheme", "ssh://git@github.com/org/repo.git", "https://github.com/org/repo"},
		{"ssh scheme no .git", "ssh://git@github.com/org/repo", "https://github.com/org/repo"},
		{"trailing slash", "https://github.com/org/repo/", "https://github.com/org/repo"},
		{"self-hosted scp", "git@gitlab.example.com:team/project.git", "https://gitlab.example.com/team/project"},
		{"self-hosted https", "https://gitlab.example.com/team/project.git", "https://gitlab.example.com/team/project"},
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
		{"nested path", "git@github.com:org/group/repo.git", "https://github.com/org/group/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := giturl.Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"ssh vs https same repo", "git@github.com:org/repo.git", "https://github.com/org/repo", true},
		{"both ssh same", "git@github.com:org/repo.git", "git@github.com:org/repo.git", true},
		{"different repos", "git@github.com:org/foo.git", "git@github.com:org/bar.git", false},
		{"empty vs non-empty", "", "https://github.com/org/repo", false},
		{"both empty", "", "", false},
		{"with and without .git", "https://github.com/org/repo.git", "https://github.com/org/repo", true},
		// Shorthand must not Equal github via Normalize alone (would break local remotes).
		{"shorthand not equal via Normalize", "org/repo", "https://github.com/org/repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := giturl.Equal(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Equal(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
