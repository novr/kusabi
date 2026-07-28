package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/manifest"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check health of the kusabi workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath, err := manifest.Find(mustGetwd())
			if err != nil {
				check(false, "kusabi.yaml", err.Error())
				return nil
			}
			check(true, "kusabi.yaml", manifestPath)

			m, err := manifest.Load(manifestPath)
			if err != nil {
				check(false, "kusabi.yaml parse", err.Error())
				return nil
			}

			rootDir := filepath.Dir(manifestPath)

			// git in PATH
			_, gitErr := exec.LookPath("git")
			check(gitErr == nil, "git command", "not found in PATH")

			// AGENTS.md
			if m.Context.Agents != "" {
				agentsPath := filepath.Join(rootDir, m.Context.Agents)
				_, err := os.Stat(agentsPath)
				check(err == nil, "AGENTS.md", fmt.Sprintf("not found at %s", agentsPath))
			}

			// .gitignore managed block
			entries := gitignoreEntries(rootDir)
			hasGitignore := len(entries) > 0
			check(hasGitignore, ".gitignore", "kusabi managed block is missing (run `kusabi init` or `kusabi sync`)")

			// Per-repository checks
			fmt.Println()
			for name, repo := range m.Repositories {
				absPath := filepath.Join(rootDir, repo.Path)
				_, err := os.Stat(filepath.Join(absPath, ".git"))
				cloned := err == nil
				check(cloned, fmt.Sprintf("[%s] cloned", name), fmt.Sprintf("not found at %s (run `kusabi sync`)", repo.Path))

				entry := repo.Path + "/"
				ignored := contains(entries, entry) || contains(entries, repo.Path)
				check(ignored, fmt.Sprintf("[%s] .gitignore", name), fmt.Sprintf("%s not excluded in .gitignore", repo.Path))
			}
			return nil
		},
	}
}

func check(ok bool, label, detail string) {
	okColor := color.New(color.FgGreen, color.Bold)
	warnColor := color.New(color.FgRed, color.Bold)

	status := okColor.Sprint("  [OK]   ")
	if !ok {
		status = warnColor.Sprint("  [ERROR]")
	}

	msg := fmt.Sprintf("%s %s", status, label)
	if !ok && detail != "" && !strings.HasPrefix(detail, "not found") {
		msg += fmt.Sprintf(": %s", detail)
	} else if !ok && detail != "" {
		msg += fmt.Sprintf(" (%s)", detail)
	}
	fmt.Println(msg)
}
