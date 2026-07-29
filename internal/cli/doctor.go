package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/declaration"
	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/health"
	"github.com/novr/kusabi/internal/manifest"
)

func newDoctorCmd() *cobra.Command {
	var migrateGitignore bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check health of the kusabi workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath, err := manifest.Find(mustGetwd())
			if err != nil {
				printCheck(false, false, "kusabi.yaml", err.Error())
				return fmt.Errorf("doctor found issues")
			}
			printCheck(true, false, "kusabi.yaml", manifestPath)

			f, err := manifest.Open(manifestPath)
			if err != nil {
				printCheck(false, false, "kusabi.yaml parse", err.Error())
				return fmt.Errorf("doctor found issues")
			}

			g := &git.SystemGit{}
			checks := health.Doctor(f, g)

			fmt.Println()
			for _, c := range checks {
				printCheck(c.OK, c.Warn, c.Label, c.Detail)
			}

			if migrateGitignore {
				fmt.Println()
				if err := runMigrateGitignore(f); err != nil {
					return err
				}
				// Re-run checks after migration
				checks = health.Doctor(f, g)
				fmt.Println()
				for _, c := range checks {
					printCheck(c.OK, c.Warn, c.Label, c.Detail)
				}
			}

			if health.HasIssues(checks) {
				return fmt.Errorf("doctor found issues")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&migrateGitignore, "migrate-gitignore", false,
		"Adopt manually-excluded paths into the kusabi managed block")
	return cmd
}

// runMigrateGitignore moves manually-excluded-but-not-managed paths into the managed block.
func runMigrateGitignore(f *manifest.File) error {
	rootDir := f.RootDir()
	migrateColor := color.New(color.FgCyan)
	for _, name := range f.Manifest.RepositoryNames() {
		repo := f.Manifest.Repositories[name]
		if declaration.IsExcludedInFullGitignore(rootDir, repo.Path) {
			if err := declaration.EnsureGitignoreEntry(rootDir, repo.Path); err != nil {
				return fmt.Errorf("migrate gitignore for %s: %w", name, err)
			}
			migrateColor.Printf("  ↳ migrated [%s] %s to kusabi managed block\n", name, repo.Path)
		}
	}
	return nil
}

func printCheck(ok, warn bool, label, detail string) {
	okColor := color.New(color.FgGreen, color.Bold)
	warnColor := color.New(color.FgYellow, color.Bold)
	errColor := color.New(color.FgRed, color.Bold)

	var status string
	switch {
	case ok:
		status = okColor.Sprint("  [OK]   ")
	case warn:
		status = warnColor.Sprint("  [WARN] ")
	default:
		status = errColor.Sprint("  [ERROR]")
	}

	msg := fmt.Sprintf("%s %s", status, label)
	if !ok && detail != "" && !strings.HasPrefix(detail, "not found") {
		msg += fmt.Sprintf(": %s", detail)
	} else if !ok && detail != "" {
		msg += fmt.Sprintf(" (%s)", detail)
	}
	fmt.Println(msg)
}
