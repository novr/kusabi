package declaration

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	GitignoreMarkerStart = "# kusabi managed — do not edit below"
	GitignoreMarkerEnd   = "# kusabi managed — end"
)

// UpdateGitignore ensures entries exist within the kusabi-managed block.
func UpdateGitignore(dir string, entries []string) error {
	path := filepath.Join(dir, ".gitignore")

	var existing string
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	var head, tail string
	if start := strings.Index(existing, GitignoreMarkerStart); start != -1 {
		head = strings.TrimRight(existing[:start], "\n")
		if end := strings.Index(existing[start:], GitignoreMarkerEnd); end != -1 {
			tail = strings.TrimLeft(existing[start+end+len(GitignoreMarkerEnd):], "\n")
		}
	} else {
		head = strings.TrimRight(existing, "\n")
	}

	var block strings.Builder
	block.WriteString(GitignoreMarkerStart + "\n")
	for _, e := range entries {
		block.WriteString(e + "\n")
	}
	block.WriteString(GitignoreMarkerEnd + "\n")

	var out strings.Builder
	if head != "" {
		out.WriteString(head + "\n\n")
	}
	out.WriteString(block.String())
	if tail != "" {
		out.WriteString("\n" + tail)
	}

	return os.WriteFile(path, []byte(out.String()), 0644)
}

// GitignoreEntries returns paths listed in the kusabi-managed block.
func GitignoreEntries(dir string) []string {
	path := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)

	start := strings.Index(content, GitignoreMarkerStart)
	if start == -1 {
		return nil
	}
	end := strings.Index(content[start:], GitignoreMarkerEnd)
	if end == -1 {
		return nil
	}

	block := content[start+len(GitignoreMarkerStart) : start+end]
	var entries []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			entries = append(entries, line)
		}
	}
	return entries
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeEntry(slice []string, s string) []string {
	out := slice[:0]
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

func ensureGitignoreEntry(dir, repoPath string) error {
	entries := GitignoreEntries(dir)
	entry := repoPath + "/"
	if contains(entries, entry) || contains(entries, repoPath) {
		return nil
	}
	return UpdateGitignore(dir, append(entries, entry))
}

func removeGitignoreEntry(dir, repoPath string) error {
	entries := GitignoreEntries(dir)
	entries = removeEntry(entries, repoPath+"/")
	entries = removeEntry(entries, repoPath)
	return UpdateGitignore(dir, entries)
}
