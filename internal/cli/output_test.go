package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestSplitErrorMessage(t *testing.T) {
	err := errors.New("git clone url path: exit status 128\nCloning into 'path'...\nERROR: Repository not found.")
	summary, details := splitErrorMessage(err)
	if summary != "git clone url path: exit status 128" {
		t.Fatalf("summary = %q", summary)
	}
	if len(details) != 1 || details[0] != "ERROR: Repository not found." {
		t.Fatalf("details = %v", details)
	}
}

func TestPrintRepoError(t *testing.T) {
	var buf strings.Builder
	fail := color.New()
	printRepoError(&buf, fail, "[1/9]", "infra", errors.New("git clone: exit status 128\nERROR: Repository not found."))
	out := buf.String()
	if !strings.Contains(out, "✗ [1/9] infra") {
		t.Fatalf("missing summary line: %q", out)
	}
	if !strings.Contains(out, "      ERROR: Repository not found.") {
		t.Fatalf("missing indented detail: %q", out)
	}
}
