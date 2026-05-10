package output

import (
	"bytes"
	"errors"
	"io"
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

func TestWriteTablePropagatesWriterErrors(t *testing.T) {
	expected := errors.New("write failed")
	w := &failWriter{err: expected}

	err := WriteTable(w, report.Result{})
	if !errors.Is(err, expected) {
		t.Fatalf("expected writer error, got: %v", err)
	}
}

func TestWriteTableSanitizesErrorStderr(t *testing.T) {
	result := report.Result{
		Errors: []report.RunError{
			{
				Repo:     "repo-a",
				Module:   "./module-a",
				Command:  "go test ./...",
				ExitCode: 1,
				Stderr:   "line1\tfield\r\nline2\nline3",
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteTable(&buf, result); err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "\tfield\r\n") || strings.Contains(got, "line2\nline3") {
		t.Fatalf("stderr was not sanitized:\n%s", got)
	}
	if !strings.Contains(got, "line1 field line2 line3") {
		t.Fatalf("sanitized stderr not found:\n%s", got)
	}
}

func TestWriteTableSanitizesStringCells(t *testing.T) {
	result := report.Result{
		Modules: []report.ModuleResult{
			{
				Repo:   "repo\tname",
				Module: "module\nname",
				Counts: []report.Count{
					{Analyzer: "analyzer\rname", Diagnostics: 1},
				},
			},
		},
		Repos: []report.RepoResult{
			{
				Repo: "repo\tname",
				Counts: []report.Count{
					{Analyzer: "analyzer\rname", Diagnostics: 1},
				},
			},
		},
		Total: []report.Count{
			{Analyzer: "analyzer\rname", Diagnostics: 1},
		},
		Errors: []report.RunError{
			{
				Repo:     "repo\tname",
				Module:   "module\nname",
				Command:  "go\tfix\n-json",
				ExitCode: 1,
				Stderr:   "failed",
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteTable(&buf, result); err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "repo\tname") || strings.Contains(got, "module\nname") || strings.Contains(got, "analyzer\rname") || strings.Contains(got, "go\tfix\n-json") {
		t.Fatalf("table output contains unsanitized cell:\n%s", got)
	}
	for _, want := range []string{"repo name", "module name", "analyzer name", "go fix -json"} {
		if !strings.Contains(got, want) {
			t.Fatalf("sanitized cell %q not found:\n%s", want, got)
		}
	}
}

type failWriter struct {
	err error
}

func (w *failWriter) Write([]byte) (int, error) {
	return 0, w.err
}

var _ io.Writer = (*failWriter)(nil)
