package context_test

import (
	"encoding/json"
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
		Name:    "test",
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
		Name:    "test",
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

func TestBuild_ParentContextPaths(t *testing.T) {
	dir := t.TempDir()
	knowledgeDir := filepath.Join(dir, "team-knowledge")
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(knowledgeDir, "ADR.md"), []byte("adr-content"), 0644); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Paths: []string{"team-knowledge/ADR.md", "missing.md"},
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
	if !strings.Contains(out, "## Parent Context Files") {
		t.Errorf("expected Parent Context Files section, got:\n%s", out)
	}
	if !strings.Contains(out, "### team-knowledge/ADR.md") {
		t.Errorf("expected parent file heading, got:\n%s", out)
	}
	if !strings.Contains(out, "adr-content") {
		t.Errorf("expected ADR.md content in output, got:\n%s", out)
	}
	if !strings.Contains(out, "_(missing: missing.md)_") {
		t.Errorf("expected missing file indicator, got:\n%s", out)
	}
}

func TestBuild_ParentContextPathsEmptyOmitsSection(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Repositories: map[string]manifest.Repository{
			"app": {Path: "packages/app", URL: "git@example.com/app.git"},
		},
	}

	b := &kctx.Builder{Manifest: m, RootDir: dir}
	out, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Parent Context Files") {
		t.Errorf("empty paths must omit Parent Context Files section, got:\n%s", out)
	}
}

func TestBuild_ParentContextPathsRejectsEscape(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(filepath.Dir(dir), "kusabi-secret-outside.md")
	if err := os.WriteFile(secret, []byte("SECRET_OUTSIDE_CONTENT"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(secret) })

	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Paths: []string{"../" + filepath.Base(secret), "/etc/passwd"},
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
	if strings.Contains(out, "SECRET_OUTSIDE_CONTENT") {
		t.Errorf("path escape must not read outside workspace, got:\n%s", out)
	}
	if strings.Contains(out, "root:") {
		t.Errorf("absolute path must not read host files, got:\n%s", out)
	}
	if !strings.Contains(out, "_(missing: ../"+filepath.Base(secret)+")_") {
		t.Errorf("expected missing marker for escaped relative path, got:\n%s", out)
	}
	if !strings.Contains(out, "_(missing: /etc/passwd)_") {
		t.Errorf("expected missing marker for absolute path, got:\n%s", out)
	}
}

func TestBuildJSON_ParentContextPaths(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ADR.md"), []byte("adr-content"), 0644); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(dir, "packages", "app")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{
		Version: "1",
		Name:    "test",
		Context: manifest.ContextConfig{
			Paths: []string{"ADR.md", "missing.md"},
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

	var parsed struct {
		Meta struct {
			ParentContextFiles []struct {
				Path    string `json:"path"`
				Content string `json:"content"`
				Missing bool   `json:"missing"`
			} `json:"parent_context_files"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	files := parsed.Meta.ParentContextFiles
	if len(files) != 2 {
		t.Fatalf("expected 2 parent_context_files, got %d", len(files))
	}
	if files[0].Path != "ADR.md" || files[0].Content != "adr-content" || files[0].Missing {
		t.Errorf("unexpected present entry: %+v", files[0])
	}
	if files[1].Path != "missing.md" || files[1].Content != "" || !files[1].Missing {
		t.Errorf("unexpected missing entry: %+v", files[1])
	}
	if strings.Contains(string(data), `"content": ""`) {
		t.Errorf("missing entries must omit empty content, got:\n%s", data)
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
		Name:    "test",
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
