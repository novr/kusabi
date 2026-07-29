package giturl_test

import (
	"testing"

	"github.com/novr/kusabi/internal/giturl"
)

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
