package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/risou/go-fix-report/internal/report"
)

func TestParseArgsDefaultsJobsToOne(t *testing.T) {
	options, jsonOutput, err := parseArgs([]string{})
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}
	if jsonOutput {
		t.Fatal("jsonOutput = true, want false")
	}
	if options.Jobs != 1 {
		t.Fatalf("options.Jobs = %d, want 1", options.Jobs)
	}
}

func TestParseArgsAcceptsJobs(t *testing.T) {
	options, _, err := parseArgs([]string{"--jobs", "3", "/tmp/repos"})
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}
	if options.Jobs != 3 {
		t.Fatalf("options.Jobs = %d, want 3", options.Jobs)
	}
	if options.Path != "/tmp/repos" {
		t.Fatalf("options.Path = %q, want %q", options.Path, "/tmp/repos")
	}
}

func TestParseArgsRejectsNonPositiveJobs(t *testing.T) {
	_, _, err := parseArgs([]string{"--jobs", "0"})
	if err == nil {
		t.Fatal("parseArgs() error = nil, want non-nil")
	}
}

func TestMainE2ERejectsInvalidJobs(t *testing.T) {
	result := runCLI(t, "--jobs", "0")

	if result.exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2\nstdout:\n%s\nstderr:\n%s", result.exitCode, result.stdout, result.stderr)
	}
	if !strings.Contains(result.stderr, "jobs must be greater than zero") {
		t.Fatalf("stderr does not contain jobs error:\n%s", result.stderr)
	}
	if !strings.Contains(result.stderr, "usage: gofixreport [--json] [--jobs N] [path]") {
		t.Fatalf("stderr does not contain usage:\n%s", result.stderr)
	}
}

func TestMainE2ERejectsTooManyArgs(t *testing.T) {
	result := runCLI(t, "a", "b")

	if result.exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2\nstdout:\n%s\nstderr:\n%s", result.exitCode, result.stdout, result.stderr)
	}
	if !strings.Contains(result.stderr, "too many arguments") {
		t.Fatalf("stderr does not contain too many arguments error:\n%s", result.stderr)
	}
}

func TestMainE2EReportsMissingPath(t *testing.T) {
	missing := t.TempDir() + "/missing"

	result := runCLI(t, missing)

	if result.exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1\nstdout:\n%s\nstderr:\n%s", result.exitCode, result.stdout, result.stderr)
	}
	if !strings.Contains(result.stdout, "ERRORS") {
		t.Fatalf("stdout does not contain errors table:\n%s", result.stdout)
	}
	if !strings.Contains(result.stdout, "no such file or directory") {
		t.Fatalf("stdout does not contain missing path error:\n%s", result.stdout)
	}
}

func TestMainE2EWritesJSON(t *testing.T) {
	result := runCLI(t, "--json", t.TempDir())

	if result.exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0\nstdout:\n%s\nstderr:\n%s", result.exitCode, result.stdout, result.stderr)
	}
	if result.stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}

	var parsed report.Result
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("stdout is not valid report JSON: %v\n%s", err, result.stdout)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GOFIXREPORT_TEST_MAIN") != "1" {
		return
	}

	args := []string{}
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	os.Args = append([]string{"gofixreport"}, args...)
	main()
	os.Exit(0)
}

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func runCLI(t *testing.T, args ...string) cliResult {
	t.Helper()

	cmdArgs := append([]string{"-test.run=TestHelperProcess", "--"}, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = append(os.Environ(), "GOFIXREPORT_TEST_MAIN=1")

	output, err := cmd.Output()
	stderr := ""
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = string(exitErr.Stderr)
		return cliResult{
			stdout:   string(output),
			stderr:   stderr,
			exitCode: exitErr.ExitCode(),
		}
	}
	if err != nil {
		t.Fatalf("running CLI helper: %v", err)
	}

	return cliResult{
		stdout:   string(output),
		stderr:   stderr,
		exitCode: 0,
	}
}
