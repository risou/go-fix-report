package app

import (
	"context"
	"errors"
	"testing"

	"github.com/risou/go-fix-report/internal/module"
	"github.com/risou/go-fix-report/internal/repo"
	"github.com/risou/go-fix-report/internal/runner"
)

func TestRunSkipsReposWithoutModules(t *testing.T) {
	repos := []repo.Repository{
		{Root: "/repos/a", Name: "a"},
		{Root: "/repos/b", Name: "b"},
	}

	result := Run(context.Background(), Options{}, Dependencies{
		DiscoverRepos: func(string) ([]repo.Repository, error) { return repos, nil },
		DiscoverModules: func(string) ([]module.Module, error) {
			return nil, nil
		},
		Runner: runner.Runner{Executor: fakeExecutor{}},
	})

	if result.HadErrors {
		t.Fatal("result.HadErrors = true, want false")
	}
	if got := len(result.Report.Modules); got != 0 {
		t.Fatalf("len(result.Report.Modules) = %d, want 0", got)
	}
	if got := len(result.Report.Repos); got != 0 {
		t.Fatalf("len(result.Report.Repos) = %d, want 0", got)
	}
	if got := len(result.Report.Total); got != 0 {
		t.Fatalf("len(result.Report.Total) = %d, want 0", got)
	}
	if got := len(result.Report.Errors); got != 0 {
		t.Fatalf("len(result.Report.Errors) = %d, want 0", got)
	}
}

func TestRunContinuesWhenOneModuleFails(t *testing.T) {
	result := Run(context.Background(), Options{}, Dependencies{
		DiscoverRepos: func(string) ([]repo.Repository, error) {
			return []repo.Repository{{Root: "/repos/a", Name: "a"}}, nil
		},
		DiscoverModules: func(string) ([]module.Module, error) {
			return []module.Module{
				{Root: "/repos/a/m1", Display: "./m1"},
				{Root: "/repos/a/m2", Display: "./m2"},
			}, nil
		},
		Runner: runner.Runner{
			Executor: fakeExecutor{
				resultsByDir: map[string]runner.Result{
					"/repos/a/m1": {Stdout: []byte(`{"pkg":{"printf":[{"posn":"f.go:1:1","end":"f.go:1:2","message":"m"}]}}`)},
					"/repos/a/m2": {Err: errors.New("boom"), ExitCode: 2, Stderr: []byte("failed")},
				},
			},
		},
	})

	if !result.HadErrors {
		t.Fatal("result.HadErrors = false, want true")
	}
	if got := len(result.Report.Modules); got != 1 {
		t.Fatalf("len(result.Report.Modules) = %d, want 1", got)
	}
	if got := result.Report.Modules[0].Module; got != "./m1" {
		t.Fatalf("result.Report.Modules[0].Module = %q, want %q", got, "./m1")
	}
	if got := len(result.Report.Errors); got != 1 {
		t.Fatalf("len(result.Report.Errors) = %d, want 1", got)
	}
	if got := result.Report.Errors[0].Module; got != "./m2" {
		t.Fatalf("result.Report.Errors[0].Module = %q, want %q", got, "./m2")
	}
}

func TestRunAggregatesSuccessfulModules(t *testing.T) {
	result := Run(context.Background(), Options{}, Dependencies{
		DiscoverRepos: func(string) ([]repo.Repository, error) {
			return []repo.Repository{
				{Root: "/repos/a", Name: "a"},
				{Root: "/repos/b", Name: "b"},
			}, nil
		},
		DiscoverModules: func(repoRoot string) ([]module.Module, error) {
			switch repoRoot {
			case "/repos/a":
				return []module.Module{
					{Root: "/repos/a/m1", Display: "./m1"},
				}, nil
			case "/repos/b":
				return []module.Module{
					{Root: "/repos/b/m2", Display: "./m2"},
				}, nil
			default:
				return nil, nil
			}
		},
		Runner: runner.Runner{
			Executor: fakeExecutor{
				resultsByDir: map[string]runner.Result{
					"/repos/a/m1": {
						Stdout: []byte(`{"pkg":{"printf":[{"posn":"f.go:1:1","end":"f.go:1:2","message":"x","suggested_fixes":[{"message":"fix","edits":[{"filename":"f.go","start":1,"end":2,"new":"y"}]}]}]}}`),
					},
					"/repos/b/m2": {
						Stdout: []byte(`{"pkg":{"printf":[{"posn":"g.go:2:1","end":"g.go:2:2","message":"z"}]}}`),
					},
				},
			},
		},
	})

	if result.HadErrors {
		t.Fatal("result.HadErrors = true, want false")
	}
	if got := len(result.Report.Modules); got != 2 {
		t.Fatalf("len(result.Report.Modules) = %d, want 2", got)
	}
	if got := len(result.Report.Repos); got != 2 {
		t.Fatalf("len(result.Report.Repos) = %d, want 2", got)
	}
	if got := len(result.Report.Total); got != 1 {
		t.Fatalf("len(result.Report.Total) = %d, want 1", got)
	}
	if got := result.Report.Total[0].Diagnostics; got != 2 {
		t.Fatalf("result.Report.Total[0].Diagnostics = %d, want 2", got)
	}
}

func TestRunReturnsHadErrorsForExecutionFailure(t *testing.T) {
	result := Run(context.Background(), Options{}, Dependencies{
		DiscoverRepos: func(string) ([]repo.Repository, error) {
			return []repo.Repository{{Root: "/repos/a", Name: "a"}}, nil
		},
		DiscoverModules: func(string) ([]module.Module, error) {
			return []module.Module{{Root: "/repos/a/m1", Display: "./m1"}}, nil
		},
		Runner: runner.Runner{
			Executor: fakeExecutor{
				resultsByDir: map[string]runner.Result{
					"/repos/a/m1": {Err: errors.New("exec failed"), ExitCode: -1},
				},
			},
		},
	})

	if !result.HadErrors {
		t.Fatal("result.HadErrors = false, want true")
	}
	if got := len(result.Report.Errors); got != 1 {
		t.Fatalf("len(result.Report.Errors) = %d, want 1", got)
	}
	if got := result.Report.Errors[0].ExitCode; got != -1 {
		t.Fatalf("result.Report.Errors[0].ExitCode = %d, want -1", got)
	}
}

type fakeExecutor struct {
	resultsByDir map[string]runner.Result
}

func (f fakeExecutor) Run(_ context.Context, dir string, _ string, _ ...string) runner.Result {
	if f.resultsByDir == nil {
		return runner.Result{}
	}
	if result, ok := f.resultsByDir[dir]; ok {
		return result
	}
	return runner.Result{}
}
