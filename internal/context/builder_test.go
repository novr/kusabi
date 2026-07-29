package context_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	kctx "github.com/novr/kusabi/internal/context"
	"github.com/novr/kusabi/internal/manifest"
)

func TestBuild_NoInstructions(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# App\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:  "test",
		Context: manifest.ContextConfig{
			Includes: []string{"README.md", "CLAUDE.md"},
		},
		Repositories: map[string]manifest.Repository{
			"app": {Path: "packages/app", URL: "git@example.com/app.git"},
		},
	}

	b := &kctx.Builder{Manifest: m, RootDir: dir}
	out, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Instructions for AI Agents") {
		t.Error("output must not contain fabricated Instructions section")
	}
}

func TestBuild_AllIncludes(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("readme-body"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "CLAUDE.md"), []byte("claude-body"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:  "test",
		Context: manifest.ContextConfig{
			Includes: []string{"README.md", "CLAUDE.md"},
		},
		Repositories: map[string]manifest.Repository{
			"app": {Path: "packages/app", URL: "git@example.com/app.git"},
		},
	}

	b := &kctx.Builder{Manifest: m, RootDir: dir}
	out, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "readme-body") || !strings.Contains(out, "claude-body") {
		t.Errorf("expected both include files in output, got:\n%s", out)
	}
	if !strings.Contains(out, "#### README.md") || !strings.Contains(out, "#### CLAUDE.md") {
		t.Errorf("expected per-file headings, got:\n%s", out)
	}
}

func TestBuild_PerRepoIncludes(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write both README.md and CHANGELOG.md
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("readme-body"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "CHANGELOG.md"), []byte("changelog-body"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Includes: []string{"README.md"}, // global includes only README
		},
		Repositories: map[string]manifest.Repository{
			"app": {
				Path:     "packages/app",
				URL:      "git@example.com/app.git",
				Includes: []string{"CHANGELOG.md"}, // per-repo overrides global
			},
		},
	}

	b := &kctx.Builder{Manifest: m, RootDir: dir}
	out, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	// Per-repo includes should be used (CHANGELOG.md), not global (README.md)
	if !strings.Contains(out, "changelog-body") {
		t.Errorf("expected per-repo include CHANGELOG.md in output, got:\n%s", out)
	}
	if strings.Contains(out, "readme-body") {
		t.Errorf("global include README.md should be overridden by per-repo includes, got:\n%s", out)
	}
}

func TestBuild_PerRepoIncludesFallbackToGlobal(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("readme-body"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Includes: []string{"README.md"},
		},
		Repositories: map[string]manifest.Repository{
			"app": {
				Path: "packages/app",
				URL:  "git@example.com/app.git",
				// No per-repo Includes → falls back to global
			},
		},
	}

	b := &kctx.Builder{Manifest: m, RootDir: dir}
	out, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "readme-body") {
		t.Errorf("expected global include README.md in output, got:\n%s", out)
	}
}

func TestBuildJSON_AllIncludes(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("readme-body"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "CLAUDE.md"), []byte("claude-body"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:  "test",
		Context: manifest.ContextConfig{
			Includes: []string{"README.md", "CLAUDE.md"},
		},
		Repositories: map[string]manifest.Repository{
			"app": {Path: "packages/app", URL: "git@example.com/app.git"},
		},
	}

	b := &kctx.Builder{Manifest: m, RootDir: dir}
	data, err := b.BuildJSON()
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "readme-body") || !strings.Contains(out, "claude-body") {
		t.Errorf("expected both include files in JSON, got:\n%s", out)
	}
	if !strings.Contains(out, `"context_files"`) {
		t.Errorf("expected context_files array in JSON, got:\n%s", out)
	}
}
