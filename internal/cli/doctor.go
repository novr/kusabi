package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/git"
	"github.com/novr/kusabi/internal/health"
	"github.com/novr/kusabi/internal/manifest"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check health of the kusabi workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath, err := manifest.Find(mustGetwd())
			if err != nil {
				printCheck(false, "kusabi.yaml", err.Error())
				return fmt.Errorf("doctor found issues")
			}
			printCheck(true, "kusabi.yaml", manifestPath)

			f, err := manifest.Open(manifestPath)
			if err != nil {
				printCheck(false, "kusabi.yaml parse", err.Error())
				return fmt.Errorf("doctor found issues")
			}

			g := &git.SystemGit{}
			checks := health.Doctor(f, g)

			fmt.Println()
			for _, c := range checks {
				printCheck(c.OK, c.Label, c.Detail)
			}

			if health.HasIssues(checks) {
				return fmt.Errorf("doctor found issues")
			}
			return nil
		},
	}
}

func printCheck(ok bool, label, detail string) {
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
