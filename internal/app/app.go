package app

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/risou/go-fix-report/internal/module"
	"github.com/risou/go-fix-report/internal/parser"
	"github.com/risou/go-fix-report/internal/repo"
	"github.com/risou/go-fix-report/internal/report"
	"github.com/risou/go-fix-report/internal/runner"
)

type Dependencies struct {
	DiscoverRepos   func(string) ([]repo.Repository, error)
	DiscoverModules func(string) ([]module.Module, error)
	Runner          runner.Runner
}

type Options struct {
	Path string
	Jobs int
}

type RunResult struct {
	Report    report.Result
	HadErrors bool
}

type moduleJob struct {
	repo   repo.Repository
	module module.Module
}

type jobOutcome struct {
	moduleResult *report.ModuleResult
	runErrors    []report.RunError
	repoRoot     string
}

func Run(ctx context.Context, opts Options, deps Dependencies) RunResult {
	path := opts.Path
	if path == "" {
		path = "."
	}

	discoverRepos := deps.DiscoverRepos
	if discoverRepos == nil {
		discoverRepos = repo.Discover
	}
	discoverModules := deps.DiscoverModules
	if discoverModules == nil {
		discoverModules = module.Discover
	}

	repos, err := discoverRepos(path)
	if err != nil {
		return withSingleError(report.RunError{
			Command:  runner.CommandString(),
			ExitCode: -1,
			Stderr:   err.Error(),
		})
	}

	jobs := make([]moduleJob, 0)
	discoveryErrors := make([]report.RunError, 0)
	for _, r := range repos {
		modules, discoverErr := discoverModules(r.Root)
		if discoverErr != nil {
			discoveryErrors = append(discoveryErrors, report.RunError{
				Repo:     r.Name,
				Command:  runner.CommandString(),
				ExitCode: -1,
				Stderr:   discoverErr.Error(),
			})
			continue
		}
		for _, m := range modules {
			jobs = append(jobs, moduleJob{repo: r, module: m})
		}
	}

	workerCount := opts.Jobs
	if workerCount < 1 {
		workerCount = 1
	}

	jobCh := make(chan moduleJob)
	resultCh := make(chan jobOutcome, len(jobs))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				if ctx.Err() != nil {
					return
				}
				resultCh <- runModule(ctx, deps.Runner, job)
			}
		}()
	}

	go func() {
		for _, job := range jobs {
			select {
			case <-ctx.Done():
				close(jobCh)
				wg.Wait()
				close(resultCh)
				return
			case jobCh <- job:
			}
		}
		close(jobCh)
		wg.Wait()
		close(resultCh)
	}()

	result := report.Result{}
	result.Errors = append(result.Errors, discoveryErrors...)
	repoModules := map[string][]report.ModuleResult{}
	repoNames := map[string]string{}
	for outcome := range resultCh {
		if outcome.moduleResult != nil {
			result.Modules = append(result.Modules, *outcome.moduleResult)
			repoModules[outcome.repoRoot] = append(repoModules[outcome.repoRoot], *outcome.moduleResult)
			repoNames[outcome.repoRoot] = outcome.moduleResult.Repo
		}
		if len(outcome.runErrors) > 0 {
			result.Errors = append(result.Errors, outcome.runErrors...)
		}
	}

	sort.Slice(result.Modules, func(i, j int) bool {
		if result.Modules[i].Repo == result.Modules[j].Repo {
			return result.Modules[i].Module < result.Modules[j].Module
		}
		return result.Modules[i].Repo < result.Modules[j].Repo
	})
	sort.Slice(result.Errors, func(i, j int) bool {
		if result.Errors[i].Repo == result.Errors[j].Repo {
			if result.Errors[i].Module == result.Errors[j].Module {
				if result.Errors[i].Command == result.Errors[j].Command {
					if result.Errors[i].ExitCode == result.Errors[j].ExitCode {
						return result.Errors[i].Stderr < result.Errors[j].Stderr
					}
					return result.Errors[i].ExitCode < result.Errors[j].ExitCode
				}
				return result.Errors[i].Command < result.Errors[j].Command
			}
			return result.Errors[i].Module < result.Errors[j].Module
		}
		return result.Errors[i].Repo < result.Errors[j].Repo
	})

	repoRoots := make([]string, 0, len(repoModules))
	for repoRoot := range repoModules {
		repoRoots = append(repoRoots, repoRoot)
	}
	sort.Strings(repoRoots)

	for _, repoRoot := range repoRoots {
		result.Repos = append(result.Repos, report.RepoResult{
			Repo:   repoNames[repoRoot],
			Counts: report.BuildRepoCounts(repoModules[repoRoot]),
		})
	}
	result.Total = report.BuildTotalCounts(result.Modules)

	return RunResult{
		Report:    result,
		HadErrors: len(result.Errors) > 0,
	}
}

func runModule(ctx context.Context, r runner.Runner, job moduleJob) jobOutcome {
	runResult := r.RunFixJSON(ctx, job.module.Root)
	if runResult.Err != nil {
		return jobOutcome{
			repoRoot: job.repo.Root,
			runErrors: []report.RunError{
				{
					Repo:     job.repo.Name,
					Module:   job.module.Display,
					Command:  runner.CommandString(),
					ExitCode: runResult.ExitCode,
					Stderr:   string(runResult.Stderr),
				},
			},
		}
	}

	parsed, err := parser.Parse(runResult.Stdout)
	if err != nil {
		return jobOutcome{
			repoRoot: job.repo.Root,
			runErrors: []report.RunError{
				{
					Repo:     job.repo.Name,
					Module:   job.module.Display,
					Command:  runner.CommandString(),
					ExitCode: -1,
					Stderr:   err.Error(),
				},
			},
		}
	}

	for i := range parsed.Diagnostics {
		parsed.Diagnostics[i].RepoAbsPath = job.repo.Root
	}

	outcome := jobOutcome{
		repoRoot: job.repo.Root,
		moduleResult: &report.ModuleResult{
			Repo:        job.repo.Name,
			Module:      job.module.Display,
			Diagnostics: parsed.Diagnostics,
			Counts:      report.BuildModuleCounts(parsed.Diagnostics),
		},
	}
	for _, analyzerErr := range parsed.Errors {
		outcome.runErrors = append(outcome.runErrors, report.RunError{
			Repo:     job.repo.Name,
			Module:   job.module.Display,
			Command:  runner.CommandString(),
			ExitCode: -1,
			Stderr:   fmt.Sprintf("%s (%s): %s", analyzerErr.PackageID, analyzerErr.Analyzer, analyzerErr.Error),
		})
	}

	return outcome
}

func withSingleError(err report.RunError) RunResult {
	return RunResult{
		Report: report.Result{
			Errors: []report.RunError{err},
		},
		HadErrors: true,
	}
}
