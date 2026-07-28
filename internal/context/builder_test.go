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
