package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRunRealGoFixJSONAnyAnalyzer(t *testing.T) {
	goVersion := strings.TrimSpace(string(runCommand(t, "", "go", "env", "GOVERSION")))
	if !supportsGoFixJSON(goVersion) {
		t.Skipf("requires go1.26+ for go fix -json, got %q", goVersion)
	}

	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "sample.go"), "package sample\n\nfunc F(x interface{}) interface{} { return x }\n")

	runCommand(t, repoRoot, "git", "init")
	runCommand(t, repoRoot, "go", "mod", "init", "example.com/sample")

	result := Run(context.Background(), Options{Path: repoRoot}, Dependencies{})
	if result.HadErrors {
		t.Fatalf("result.HadErrors = true, errors=%v", result.Report.Errors)
	}

	if got := len(result.Report.Total); got != 1 {
		t.Fatalf("len(result.Report.Total) = %d, want 1", got)
	}

	count := result.Report.Total[0]
	if got := count.Analyzer; got != "any" {
		t.Fatalf("count.Analyzer = %q, want %q", got, "any")
	}
	if got := count.Diagnostics; got != 2 {
		t.Fatalf("count.Diagnostics = %d, want 2", got)
	}
	if got := count.Fixable; got != 2 {
		t.Fatalf("count.Fixable = %d, want 2", got)
	}
}

func runCommand(t *testing.T, dir string, name string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, output)
	}
	return output
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", path, err)
	}
}

func supportsGoFixJSON(version string) bool {
	version = strings.TrimPrefix(version, "go")
	if !strings.HasPrefix(version, "1.") {
		return false
	}

	rest := strings.TrimPrefix(version, "1.")
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return false
	}

	minor, err := strconv.Atoi(rest[:end])
	if err != nil {
		return false
	}
	return minor >= 26
}
