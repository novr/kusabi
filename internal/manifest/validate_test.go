package manifest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/novr/kusabi/internal/manifest"
)

func TestValidate_OK(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"app": {
				Path: "packages/app",
				URL:  "git@github.com:org/app.git",
			},
		},
	}
	if err := manifest.Validate(m); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_MissingVersion(t *testing.T) {
	m := &manifest.Manifest{Name: "test"}
	if err := manifest.Validate(m); err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestValidate_MissingName(t *testing.T) {
	m := &manifest.Manifest{Version: "1"}
	if err := manifest.Validate(m); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidate_MissingRepoPath(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"app": {URL: "git@github.com:org/app.git"},
		},
	}
	if err := manifest.Validate(m); err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestValidate_MissingRepoURL(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"app": {Path: "packages/app"},
		},
	}
	if err := manifest.Validate(m); err == nil {
		t.Fatal("expected error for missing url")
	}
}

func TestValidate_PathTraversal(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"app": {
				Path: "../escape",
				URL:  "git@github.com:org/app.git",
			},
		},
	}
	if err := manifest.Validate(m); err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestValidate_InvalidRepoName(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"../bad": {
				Path: "packages/app",
				URL:  "git@github.com:org/app.git",
			},
		},
	}
	if err := manifest.Validate(m); err == nil {
		t.Fatal("expected error for invalid repo name")
	}
}

func TestValidateRepoName(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"app", false},
		{"", true},
		{"app/ios", true},
		{`app\ios`, true},
		{"..", true},
	}
	for _, tc := range cases {
		err := manifest.ValidateRepoName(tc.name)
		if tc.wantErr && err == nil {
			t.Errorf("ValidateRepoName(%q) expected error", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ValidateRepoName(%q) unexpected error: %v", tc.name, err)
		}
	}
}

func TestValidate_DuplicatePath(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"a": {Path: "packages/shared", URL: "git@github.com:org/a.git"},
			"b": {Path: "packages/shared", URL: "git@github.com:org/b.git"},
		},
	}
	if err := manifest.Validate(m); err == nil {
		t.Fatal("expected error for duplicate path")
	}
}

func TestValidateRepoPath_Absolute(t *testing.T) {
	if err := manifest.ValidateRepoPath("/tmp/app"); err == nil {
		t.Fatal("expected error for absolute path")
	}
}

func TestRepositoryNames_IncludesUnorderedRepos(t *testing.T) {
	m := &manifest.Manifest{
		Repositories: map[string]manifest.Repository{
			"beta":  {Path: "packages/beta", URL: "git@example.com/beta.git"},
			"alpha": {Path: "packages/alpha", URL: "git@example.com/alpha.git"},
		},
		RepositoryOrder: []string{"beta"},
	}
	names := m.RepositoryNames()
	if len(names) != 2 || names[0] != "beta" || names[1] != "alpha" {
		t.Fatalf("got %v", names)
	}
}

func TestValidate_ContextPathsTraversal(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Paths: []string{"../escape.md"},
		},
	}
	err := manifest.Validate(m)
	if err == nil || !strings.Contains(err.Error(), "context.paths[0]") {
		t.Fatalf("expected context.paths validation error, got: %v", err)
	}
}

func TestValidate_ContextPathsAbsolute(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Paths: []string{"/etc/passwd"},
		},
	}
	err := manifest.Validate(m)
	if err == nil || !strings.Contains(err.Error(), "context.paths[0]") {
		t.Fatalf("expected context.paths validation error, got: %v", err)
	}
}

func TestValidate_ContextPathsOK(t *testing.T) {
	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Paths: []string{"team-knowledge/ADR.md", ".agents/skills/deploy.md"},
		},
	}
	if err := manifest.Validate(m); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_Validates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.Filename)
	content := "version: \"1\"\nname: test\nrepositories:\n  app:\n    path: \"../bad\"\n    url: git@example.com/app.git\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := manifest.Load(path)
	if err == nil || !strings.Contains(err.Error(), "validate") {
		t.Fatalf("expected validation error on load, got: %v", err)
	}
}
