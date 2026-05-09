package runner

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
)

type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Err      error
}

type Executor interface {
	Run(ctx context.Context, dir string, name string, args ...string) Result
}

type ExecExecutor struct{}

func (e ExecExecutor) Run(ctx context.Context, dir string, name string, args ...string) Result {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}

	return Result{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: exitCode,
		Err:      err,
	}
}

type Runner struct {
	Executor Executor
}

func (r Runner) RunFixJSON(ctx context.Context, moduleRoot string) Result {
	executor := r.Executor
	if executor == nil {
		executor = ExecExecutor{}
	}
	name, args := fixJSONCommand()
	return executor.Run(ctx, moduleRoot, name, args...)
}

func CommandString() string {
	name, args := fixJSONCommand()
	return name + " " + strings.Join(args, " ")
}

func fixJSONCommand() (string, []string) {
	return "go", []string{"fix", "-json", "./..."}
}
