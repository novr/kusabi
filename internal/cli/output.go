package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

func splitErrorMessage(err error) (summary string, details []string) {
	if err == nil {
		return "", nil
	}
	lines := strings.Split(strings.TrimSpace(err.Error()), "\n")
	summary = lines[0]
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Cloning into ") {
			continue
		}
		details = append(details, line)
	}
	return summary, details
}

func printRepoError(w io.Writer, fail *color.Color, prefix, name string, err error) {
	summary, details := splitErrorMessage(err)
	fail.Fprintf(w, "  ✗ %s %-20s %s\n", prefix, name, summary)
	for _, line := range details {
		fmt.Fprintf(w, "      %s\n", line)
	}
}
