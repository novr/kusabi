package manifest

import (
	"fmt"
	"os"
	"path/filepath"
)

const Filename = "kusabi.yaml"

type Manifest struct {
	Version         string                `yaml:"version"`
	Name            string                `yaml:"name"`
	Description     string                `yaml:"description,omitempty"`
	Context         ContextConfig         `yaml:"context,omitempty"`
	Repositories    map[string]Repository `yaml:"repositories,omitempty"`
	RepositoryOrder []string              `yaml:"-"`
}

type ContextConfig struct {
	Agents   string   `yaml:"agents,omitempty"`
	Includes []string `yaml:"includes,omitempty"`
	Paths    []string `yaml:"paths,omitempty"`
}

type Repository struct {
	Path        string   `yaml:"path"`
	URL         string   `yaml:"url"`
	Role        string   `yaml:"role,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Includes    []string `yaml:"includes,omitempty"`
	Branch      string   `yaml:"branch,omitempty"`
	SyncEnabled *bool    `yaml:"sync,omitempty"` // nil = enabled; *false = disabled
}

// IsSyncDisabled reports whether sync has been explicitly disabled for this repository.
func (r Repository) IsSyncDisabled() bool {
	return r.SyncEnabled != nil && !*r.SyncEnabled
}

// Find searches for kusabi.yaml by walking up from startDir.
func Find(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, Filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("kusabi.yaml not found (searched from %s)", startDir)
}

// FilterByNames returns only the named repositories. Returns an error if any name is unknown.
func (m *Manifest) FilterByNames(names []string) (map[string]Repository, error) {
	result := make(map[string]Repository, len(names))
	for _, name := range names {
		repo, ok := m.Repositories[name]
		if !ok {
			return nil, fmt.Errorf("no such repository: %q", name)
		}
		result[name] = repo
	}
	return result, nil
}

// FilterForExec returns repositories matching optional name and tag filters (intersection).
func (m *Manifest) FilterForExec(names []string, tag string) (map[string]Repository, error) {
	repos := m.Repositories
	if len(names) > 0 {
		var err error
		repos, err = m.FilterByNames(names)
		if err != nil {
			return nil, err
		}
	}
	if tag == "" {
		return repos, nil
	}

	tagged := make(map[string]Repository)
	for name, repo := range repos {
		for _, t := range repo.Tags {
			if t == tag {
				tagged[name] = repo
				break
			}
		}
	}
	if len(tagged) == 0 {
		if len(names) > 0 {
			return nil, fmt.Errorf("no repositories matched (--repo=%v --tag=%q)", names, tag)
		}
		return nil, fmt.Errorf("no repositories matched tag %q", tag)
	}
	return tagged, nil
}

// NamesInSet returns names from repos in declaration order.
func (m *Manifest) NamesInSet(repos map[string]Repository) []string {
	out := make([]string, 0, len(repos))
	for _, name := range m.RepositoryNames() {
		if _, ok := repos[name]; ok {
			out = append(out, name)
		}
	}
	return out
}

// FilterByTag returns repositories matching the given tag. Empty tag returns all.
func (m *Manifest) FilterByTag(tag string) map[string]Repository {
	if tag == "" {
		return m.Repositories
	}
	result := make(map[string]Repository)
	for name, repo := range m.Repositories {
		for _, t := range repo.Tags {
			if t == tag {
				result[name] = repo
				break
			}
		}
	}
	return result
}
