package runner

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

type fakeExecutor struct {
	gotDir  string
	gotName string
	gotArgs []string
	result  Result
}

func (f *fakeExecutor) Run(_ context.Context, dir string, name string, args ...string) Result {
	f.gotDir = dir
	f.gotName = name
	f.gotArgs = append([]string(nil), args...)
	return f.result
}

func TestRunFixJSONUsesModuleRoot(t *testing.T) {
	fake := &fakeExecutor{}
	r := Runner{Executor: fake}
	moduleRoot := "/tmp/module-root"

	r.RunFixJSON(context.Background(), moduleRoot)

	if fake.gotDir != moduleRoot {
		t.Fatalf("dir = %q, want %q", fake.gotDir, moduleRoot)
	}
	if fake.gotName != "go" {
		t.Fatalf("name = %q, want %q", fake.gotName, "go")
	}
	wantArgs := []string{"fix", "-json", "./..."}
	if len(fake.gotArgs) != len(wantArgs) {
		t.Fatalf("len(args) = %d, want %d", len(fake.gotArgs), len(wantArgs))
	}
	for i := range wantArgs {
		if fake.gotArgs[i] != wantArgs[i] {
			t.Fatalf("args[%d] = %q, want %q", i, fake.gotArgs[i], wantArgs[i])
		}
	}
}

func TestRunFixJSONReturnsStdout(t *testing.T) {
	want := Result{
		Stdout:   []byte("ok"),
		Stderr:   []byte("warn"),
		ExitCode: 0,
	}
	r := Runner{
		Executor: &fakeExecutor{result: want},
	}

	got := r.RunFixJSON(context.Background(), "/tmp/module-root")

	if string(got.Stdout) != string(want.Stdout) {
		t.Fatalf("stdout = %q, want %q", string(got.Stdout), string(want.Stdout))
	}
	if string(got.Stderr) != string(want.Stderr) {
		t.Fatalf("stderr = %q, want %q", string(got.Stderr), string(want.Stderr))
	}
	if got.ExitCode != want.ExitCode {
		t.Fatalf("exitCode = %d, want %d", got.ExitCode, want.ExitCode)
	}
	if got.Err != nil {
		t.Fatalf("err = %v, want nil", got.Err)
	}
}

func TestRunFixJSONCapturesFailure(t *testing.T) {
	runErr := errors.New("boom")
	r := Runner{
		Executor: &fakeExecutor{
			result: Result{
				Stderr:   []byte("failed"),
				ExitCode: 3,
				Err:      runErr,
			},
		},
	}

	got := r.RunFixJSON(context.Background(), "/tmp/module-root")

	if got.ExitCode != 3 {
		t.Fatalf("exitCode = %d, want 3", got.ExitCode)
	}
	if !errors.Is(got.Err, runErr) {
		t.Fatalf("err = %v, want %v", got.Err, runErr)
	}
	if string(got.Stderr) != "failed" {
		t.Fatalf("stderr = %q, want %q", string(got.Stderr), "failed")
	}
}

func TestCommandStringMatchesInvocation(t *testing.T) {
	fake := &fakeExecutor{}
	r := Runner{Executor: fake}
	moduleRoot := "/tmp/module-root"

	r.RunFixJSON(context.Background(), moduleRoot)

	got := CommandString()
	want := fake.gotName + " " + strings.Join(fake.gotArgs, " ")
	if got != want {
		t.Fatalf("CommandString() = %q, want %q", got, want)
	}
}

func TestExecExecutorNonProcessFailureUsesMinusOne(t *testing.T) {
	exec := ExecExecutor{}

	got := exec.Run(context.Background(), t.TempDir(), "command-that-does-not-exist-12345")

	if got.Err == nil {
		t.Fatalf("err = nil, want non-nil")
	}
	if got.ExitCode != -1 {
		t.Fatalf("exitCode = %d, want -1", got.ExitCode)
	}
}

func TestExecExecutorCapturesProcessExitCode(t *testing.T) {
	if hasArg(os.Args[1:], "--helper-exit-7") {
		os.Exit(7)
	}

	exec := ExecExecutor{}
	args := []string{
		"-test.run=TestExecExecutorCapturesProcessExitCode",
		"--",
		"--helper-exit-7",
	}

	got := exec.Run(context.Background(), t.TempDir(), os.Args[0], args...)

	if got.Err == nil {
		t.Fatalf("err = nil, want non-nil")
	}
	if got.ExitCode != 7 {
		t.Fatalf("exitCode = %d, want 7", got.ExitCode)
	}
}

func hasArg(args []string, target string) bool {
	for _, arg := range args {
		if arg == target {
			return true
		}
	}
	return false
}
