package app

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

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

func TestRunRepoTotalsDoNotMergeSameBasenameRepos(t *testing.T) {
	result := Run(context.Background(), Options{}, Dependencies{
		DiscoverRepos: func(string) ([]repo.Repository, error) {
			return []repo.Repository{
				{Root: "/workspace/a/shared", Name: "shared"},
				{Root: "/workspace/b/shared", Name: "shared"},
			}, nil
		},
		DiscoverModules: func(repoRoot string) ([]module.Module, error) {
			switch repoRoot {
			case "/workspace/a/shared":
				return []module.Module{{Root: repoRoot + "/m1", Display: "./m1"}}, nil
			case "/workspace/b/shared":
				return []module.Module{{Root: repoRoot + "/m1", Display: "./m1"}}, nil
			default:
				return nil, nil
			}
		},
		Runner: runner.Runner{
			Executor: fakeExecutor{
				resultsByDir: map[string]runner.Result{
					"/workspace/a/shared/m1": {Stdout: []byte(`{"pkg":{"printf":[{"posn":"a.go:1:1","end":"a.go:1:2","message":"a"}]}}`)},
					"/workspace/b/shared/m1": {Stdout: []byte(`{"pkg":{"printf":[{"posn":"b.go:1:1","end":"b.go:1:2","message":"b"}]}}`)},
				},
			},
		},
	})

	if got := len(result.Report.Repos); got != 2 {
		t.Fatalf("len(result.Report.Repos) = %d, want 2", got)
	}
	for i, repoResult := range result.Report.Repos {
		if got := repoResult.Repo; got != "shared" {
			t.Fatalf("result.Report.Repos[%d].Repo = %q, want %q", i, got, "shared")
		}
		if got := len(repoResult.Counts); got != 1 {
			t.Fatalf("len(result.Report.Repos[%d].Counts) = %d, want 1", i, got)
		}
		if got := repoResult.Counts[0].Diagnostics; got != 1 {
			t.Fatalf("result.Report.Repos[%d].Counts[0].Diagnostics = %d, want 1", i, got)
		}
	}
}

func TestRunSortsErrorsDeterministicallyForSameRepoModule(t *testing.T) {
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
					"/repos/a/m1": {
						Stdout: []byte(`{"p":{"z":{"error":"zee"},"a":{"error":"ayy"}}}`),
					},
				},
			},
		},
	})

	if got := len(result.Report.Errors); got != 2 {
		t.Fatalf("len(result.Report.Errors) = %d, want 2", got)
	}
	stderrs := []string{
		result.Report.Errors[0].Stderr,
		result.Report.Errors[1].Stderr,
	}
	want := []string{
		"p (a): ayy",
		"p (z): zee",
	}
	if !reflect.DeepEqual(stderrs, want) {
		t.Fatalf("error order mismatch:\n got=%v\nwant=%v", stderrs, want)
	}
}

func TestRunContinuesAfterModuleDiscoveryFailure(t *testing.T) {
	result := Run(context.Background(), Options{}, Dependencies{
		DiscoverRepos: func(string) ([]repo.Repository, error) {
			return []repo.Repository{
				{Root: "/repos/fail", Name: "fail"},
				{Root: "/repos/ok", Name: "ok"},
			}, nil
		},
		DiscoverModules: func(repoRoot string) ([]module.Module, error) {
			if repoRoot == "/repos/fail" {
				return nil, errors.New("discover failed")
			}
			if repoRoot == "/repos/ok" {
				return []module.Module{{Root: "/repos/ok/m1", Display: "./m1"}}, nil
			}
			return nil, nil
		},
		Runner: runner.Runner{
			Executor: fakeExecutor{
				resultsByDir: map[string]runner.Result{
					"/repos/ok/m1": {Stdout: []byte(`{"pkg":{"printf":[{"posn":"ok.go:1:1","end":"ok.go:1:2","message":"ok"}]}}`)},
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
	if got := result.Report.Modules[0].Repo; got != "ok" {
		t.Fatalf("result.Report.Modules[0].Repo = %q, want %q", got, "ok")
	}
	if got := len(result.Report.Errors); got != 1 {
		t.Fatalf("len(result.Report.Errors) = %d, want 1", got)
	}
	if got := result.Report.Errors[0].Repo; got != "fail" {
		t.Fatalf("result.Report.Errors[0].Repo = %q, want %q", got, "fail")
	}
}

func TestRunDefaultsToSerialExecution(t *testing.T) {
	executor := newBlockingExecutor()
	done := make(chan RunResult, 1)
	go func() {
		done <- Run(context.Background(), Options{}, Dependencies{
			DiscoverRepos: func(string) ([]repo.Repository, error) {
				return []repo.Repository{{Root: "/repos/a", Name: "a"}}, nil
			},
			DiscoverModules: func(string) ([]module.Module, error) {
				return []module.Module{
					{Root: "/repos/a/m1", Display: "m1"},
					{Root: "/repos/a/m2", Display: "m2"},
				}, nil
			},
			Runner: runner.Runner{Executor: executor},
		})
	}()

	waitForEntered(t, executor.entered)
	assertNoAdditionalEntry(t, executor.entered)
	close(executor.release)
	<-done
}

func TestRunHonorsJobsLimit(t *testing.T) {
	executor := newBlockingExecutor()
	done := make(chan RunResult, 1)
	go func() {
		done <- Run(context.Background(), Options{Jobs: 2}, Dependencies{
			DiscoverRepos: func(string) ([]repo.Repository, error) {
				return []repo.Repository{{Root: "/repos/a", Name: "a"}}, nil
			},
			DiscoverModules: func(string) ([]module.Module, error) {
				return []module.Module{
					{Root: "/repos/a/m1", Display: "m1"},
					{Root: "/repos/a/m2", Display: "m2"},
					{Root: "/repos/a/m3", Display: "m3"},
				}, nil
			},
			Runner: runner.Runner{Executor: executor},
		})
	}()

	waitForEntered(t, executor.entered)
	waitForEntered(t, executor.entered)
	assertNoAdditionalEntry(t, executor.entered)
	close(executor.release)
	<-done
}

func TestRunStopsSchedulingJobsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	executor := newCancelingExecutor(cancel)

	result := Run(ctx, Options{Jobs: 1}, Dependencies{
		DiscoverRepos: func(string) ([]repo.Repository, error) {
			return []repo.Repository{{Root: "/repos/a", Name: "a"}}, nil
		},
		DiscoverModules: func(string) ([]module.Module, error) {
			return []module.Module{
				{Root: "/repos/a/m1", Display: "m1"},
				{Root: "/repos/a/m2", Display: "m2"},
			}, nil
		},
		Runner: runner.Runner{Executor: executor},
	})

	if result.HadErrors {
		t.Fatalf("result.HadErrors = true, errors=%v", result.Report.Errors)
	}
	if got := executor.calls; got != 1 {
		t.Fatalf("executor.calls = %d, want 1", got)
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

type blockingExecutor struct {
	entered chan string
	release chan struct{}
}

func newBlockingExecutor() *blockingExecutor {
	return &blockingExecutor{
		entered: make(chan string, 10),
		release: make(chan struct{}),
	}
}

func (e *blockingExecutor) Run(_ context.Context, dir string, _ string, _ ...string) runner.Result {
	e.entered <- dir
	<-e.release
	return runner.Result{Stdout: []byte(`{}`)}
}

func waitForEntered(t *testing.T, entered <-chan string) {
	t.Helper()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for module execution to start")
	}
}

func assertNoAdditionalEntry(t *testing.T, entered <-chan string) {
	t.Helper()
	select {
	case dir := <-entered:
		t.Fatalf("unexpected additional module execution started: %s", dir)
	case <-time.After(50 * time.Millisecond):
	}
}

type cancelingExecutor struct {
	cancel context.CancelFunc
	calls  int
}

func newCancelingExecutor(cancel context.CancelFunc) *cancelingExecutor {
	return &cancelingExecutor{cancel: cancel}
}

func (e *cancelingExecutor) Run(_ context.Context, _ string, _ string, _ ...string) runner.Result {
	e.calls++
	e.cancel()
	return runner.Result{Stdout: []byte(`{}`)}
}
