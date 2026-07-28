package declaration

import (
	"fmt"
	"path/filepath"

	"github.com/novr/kusabi/internal/manifest"
)

// AddRepoOptions configures a repository entry.
type AddRepoOptions struct {
	Path string
	Role string
	Tags []string
}

// AddRepo appends a repository to the manifest and updates gitignore.
func AddRepo(f *manifest.File, name, url string, opts AddRepoOptions) error {
	if err := manifest.ValidateRepoName(name); err != nil {
		return err
	}
	if _, exists := f.Manifest.Repositories[name]; exists {
		return fmt.Errorf("repository %q already exists", name)
	}

	path := opts.Path
	if path == "" {
		path = filepath.Join("packages", name)
	}
	if err := manifest.ValidateRepoPath(path); err != nil {
		return fmt.Errorf("repository %q: %w", name, err)
	}

	repo := manifest.Repository{
		Path: path,
		URL:  url,
		Role: opts.Role,
		Tags: opts.Tags,
	}

	f.Manifest.Repositories[name] = repo
	f.Manifest.RepositoryOrder = append(f.Manifest.RepositoryOrder, name)

	if err := EnsureGitignoreEntry(f.RootDir(), path); err != nil {
		delete(f.Manifest.Repositories, name)
		f.Manifest.RepositoryOrder = removeEntry(f.Manifest.RepositoryOrder, name)
		return err
	}
	if err := f.Save(); err != nil {
		_ = removeGitignoreEntry(f.RootDir(), path)
		delete(f.Manifest.Repositories, name)
		f.Manifest.RepositoryOrder = removeEntry(f.Manifest.RepositoryOrder, name)
		return err
	}
	return nil
}

// RemoveRepo deletes a repository from the manifest and gitignore.
func RemoveRepo(f *manifest.File, name string) (manifest.Repository, error) {
	repo, exists := f.Manifest.Repositories[name]
	if !exists {
		return manifest.Repository{}, fmt.Errorf("repository %q not found", name)
	}

	// Snapshot original order so rollback can restore exactly.
	originalOrder := make([]string, len(f.Manifest.RepositoryOrder))
	copy(originalOrder, f.Manifest.RepositoryOrder)

	delete(f.Manifest.Repositories, name)
	f.Manifest.RepositoryOrder = removeEntry(originalOrder, name)

	if err := removeGitignoreEntry(f.RootDir(), repo.Path); err != nil {
		f.Manifest.Repositories[name] = repo
		f.Manifest.RepositoryOrder = originalOrder
		return manifest.Repository{}, err
	}

	if err := f.Save(); err != nil {
		f.Manifest.Repositories[name] = repo
		f.Manifest.RepositoryOrder = originalOrder
		_ = EnsureGitignoreEntry(f.RootDir(), repo.Path)
		return manifest.Repository{}, err
	}
	return repo, nil
}
