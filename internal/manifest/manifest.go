package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const Filename = "kusabi.yaml"

type Manifest struct {
	Version      string                `yaml:"version"`
	Name         string                `yaml:"name"`
	Description  string                `yaml:"description,omitempty"`
	Context      ContextConfig         `yaml:"context,omitempty"`
	Repositories map[string]Repository `yaml:"repositories,omitempty"`
}

type ContextConfig struct {
	Agents   string   `yaml:"agents,omitempty"`
	Includes []string `yaml:"includes,omitempty"`
}

type Repository struct {
	Path string   `yaml:"path"`
	URL  string   `yaml:"url"`
	Role string   `yaml:"role,omitempty"`
	Tags []string `yaml:"tags,omitempty"`
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

// Load reads kusabi.yaml from path.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.Repositories == nil {
		m.Repositories = make(map[string]Repository)
	}
	return &m, nil
}

// Save writes the manifest to path atomically via a temp-file rename.
func Save(m *Manifest, path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".kusabi-*.yaml.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename to %s: %w", path, err)
	}
	return nil
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
