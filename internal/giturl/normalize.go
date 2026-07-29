package giturl

import (
	"net/url"
	"strings"
)

// Normalize converts any common git remote URL to a canonical https form
// for comparison purposes. Returns empty string on parse failure.
//
// Examples:
//   git@github.com:org/repo.git  →  https://github.com/org/repo
//   https://github.com/org/repo.git  →  https://github.com/org/repo
//   ssh://git@github.com/org/repo.git  →  https://github.com/org/repo
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
