package manifest

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Validate checks that the manifest satisfies required constraints.
func Validate(m *Manifest) error {
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}

	for i, p := range m.Context.Paths {
		if err := ValidateRepoPath(p); err != nil {
			return fmt.Errorf("context.paths[%d]: %w", i, err)
		}
	}

	seenPaths := make(map[string]string, len(m.Repositories))
	for name, repo := range m.Repositories {
		if err := ValidateRepoName(name); err != nil {
			return err
		}
		if err := ValidateRepoPath(repo.Path); err != nil {
			return fmt.Errorf("repository %q: %w", name, err)
		}
		if repo.URL == "" {
			return fmt.Errorf("repository %q: url is required", name)
		}
		if other, ok := seenPaths[repo.Path]; ok {
			return fmt.Errorf("repositories %q and %q share path %q", other, name, repo.Path)
		}
		seenPaths[repo.Path] = name
	}

	return nil
}

// ValidateRepoName rejects names that would cause path traversal.
func ValidateRepoName(name string) error {
	if name == "" {
		return fmt.Errorf("repository name must not be empty")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return fmt.Errorf("repository name %q must not contain '/', '\\', or '..'", name)
	}
	return nil
}

// ValidateRepoPath rejects repository paths outside the workspace root.
func ValidateRepoPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("path must be relative")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path must not contain '..'")
	}
	clean := filepath.Clean(path)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path must stay within the workspace")
	}
	return nil
}
