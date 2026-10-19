package giturl

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// githubShorthand matches owner/repo, GitHub-style names only.
var githubShorthand = regexp.MustCompile(`^([A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?)/([A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?)$`)

// Expand resolves GitHub owner/repo shorthand to https://github.com/owner/repo.
// Other inputs are returned trimmed and unchanged.
//
// Bare "owner/repo" follows gh(1) semantics and is treated as GitHub.
// Local remotes must use "./path", "../path", an absolute path, or a file:// URL.
func Expand(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if strings.Contains(raw, "://") || strings.Contains(raw, "@") || strings.Contains(raw, ":") {
		return raw
	}
	if strings.HasPrefix(raw, ".") || strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "~") {
		return raw
	}
	m := githubShorthand.FindStringSubmatch(strings.TrimSuffix(raw, ".git"))
	if m == nil {
		return raw
	}
	return "https://github.com/" + m[1] + "/" + m[2]
}

// Resolve expands GitHub shorthand like Expand, but keeps the original string when
// it names an existing directory (relative to rootDir or absolute). That preserves
// bare local remotes such as "vendor/mirror" when the path already exists.
//
// When rootDir is empty, Resolve behaves like Expand.
func Resolve(raw, rootDir string) string {
	trimmed := strings.TrimSpace(raw)
	expanded := Expand(trimmed)
	if expanded == trimmed || rootDir == "" {
		return expanded
	}
	if localPathExists(rootDir, trimmed) {
		return trimmed
	}
	return expanded
}

func localPathExists(rootDir, raw string) bool {
	candidates := []string{raw}
	if !filepath.IsAbs(raw) {
		candidates = append(candidates, filepath.Join(rootDir, filepath.FromSlash(raw)))
	}
	for _, p := range candidates {
		st, err := os.Stat(p)
		if err == nil && st.IsDir() {
			return true
		}
	}
	return false
}

// Normalize converts any common git remote URL to a canonical https form
// for comparison purposes. Returns empty string on parse failure.
//
// Examples:
//   git@github.com:org/repo.git  →  https://github.com/org/repo
//   https://github.com/org/repo.git  →  https://github.com/org/repo
//   ssh://git@github.com/org/repo.git  →  https://github.com/org/repo
//
// GitHub owner/repo shorthand is not handled here — use Expand or Resolve
// before persisting or cloning. Folding shorthand into Normalize would make
// bare local paths compare equal to github.com URLs.
func Normalize(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	// SCP-like syntax: git@github.com:org/repo.git
	if !strings.Contains(rawURL, "://") && strings.Contains(rawURL, "@") && strings.Contains(rawURL, ":") {
		// Strip leading user@ if present
		at := strings.Index(rawURL, "@")
		rest := rawURL[at+1:]
		// Replace first ":" with "/" to convert host:path → host/path
		rest = strings.Replace(rest, ":", "/", 1)
		rawURL = "https://" + rest
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Normalize scheme to https
	u.Scheme = "https"
	// Strip user info (ssh://git@host → https://host)
	u.User = nil
	// Remove .git suffix and trailing slash
	u.Path = strings.TrimSuffix(u.Path, ".git")
	u.Path = strings.TrimRight(u.Path, "/")

	return u.String()
}

// Equal reports whether two git remote URLs refer to the same repository,
// using Normalize for comparison.
func Equal(a, b string) bool {
	na, nb := Normalize(a), Normalize(b)
	return na != "" && na == nb
}
