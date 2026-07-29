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
	Path     string   `yaml:"path"`
	URL      string   `yaml:"url"`
	Role     string   `yaml:"role,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
	Includes []string `yaml:"includes,omitempty"`
	Branch   string   `yaml:"branch,omitempty"`
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
