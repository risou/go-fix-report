package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/risou/go-fix-report/internal/report"
)

func TestWriteJSON(t *testing.T) {
	result := report.Result{
		Modules: []report.ModuleResult{
			{
				Repo:   "repo-a",
				Module: "./module-a",
				Counts: []report.Count{{Analyzer: "printf", Diagnostics: 2, Fixable: 1}},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, result); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "\n  \"modules\": [\n") {
		t.Fatalf("expected indented JSON, got: %q", got)
	}
	if !strings.Contains(got, "\"repo\": \"repo-a\"") {
		t.Fatalf("expected repo field in JSON, got: %q", got)
	}
}

func TestWriteTableIncludesSections(t *testing.T) {
	result := report.Result{
		Modules: []report.ModuleResult{
			{
				Repo:   "repo-a",
				Module: "./module-a",
				Counts: []report.Count{{Analyzer: "printf", Diagnostics: 2, Fixable: 1}},
			},
		},
		Repos: []report.RepoResult{
			{
				Repo:   "repo-a",
				Counts: []report.Count{{Analyzer: "printf", Diagnostics: 2, Fixable: 1, DuplicatesRemoved: 0}},
			},
		},
		Total: []report.Count{
			{Analyzer: "printf", Diagnostics: 2, Fixable: 1, DuplicatesRemoved: 0},
		},
	}

	var buf bytes.Buffer
	if err := WriteTable(&buf, result); err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	got := buf.String()
	for _, section := range []string{"MODULE RESULTS", "REPO TOTALS", "TOTAL"} {
		if !strings.Contains(got, section) {
			t.Fatalf("expected %q section in table output:\n%s", section, got)
		}
	}
	if strings.Contains(got, "ERRORS") {
		t.Fatalf("did not expect ERRORS section when there are no errors:\n%s", got)
	}
}

func TestWriteTableIncludesErrors(t *testing.T) {
	result := report.Result{
		Errors: []report.RunError{
			{
				Repo:     "repo-a",
				Module:   "./module-a",
				Command:  "go test ./...",
				ExitCode: 2,
				Stderr:   "failed to build",
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteTable(&buf, result); err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	got := buf.String()
	for _, want := range []string{"ERRORS", "repo-a", "./module-a", "go test ./...", "2", "failed to build"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in errors output:\n%s", want, got)
		}
	}
}
