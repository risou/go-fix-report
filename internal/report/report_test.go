package report

import "testing"

func TestModelConstruction(t *testing.T) {
	result := Result{
		Modules: []ModuleResult{{Repo: "repo", Module: "."}},
		Repos:   []RepoResult{{Repo: "repo"}},
		Total:   []Count{{Analyzer: "any", Diagnostics: 1, Fixable: 1}},
	}
	if result.Total[0].Analyzer != "any" {
		t.Fatalf("unexpected analyzer: %q", result.Total[0].Analyzer)
	}
}

func TestCountModuleDiagnosticsAndFixable(t *testing.T) {
	diags := []Diagnostic{
		{Analyzer: "b", SuggestedFixes: []SuggestedFix{{Message: "fix"}}},
		{Analyzer: "a"},
		{Analyzer: "b"},
		{Analyzer: "a", SuggestedFixes: []SuggestedFix{{Message: "fix"}}},
	}

	got := BuildModuleCounts(diags)
	if len(got) != 2 {
		t.Fatalf("unexpected count length: got=%d want=2", len(got))
	}

	if got[0].Analyzer != "a" || got[0].Diagnostics != 2 || got[0].Fixable != 1 || got[0].DuplicatesRemoved != 0 {
		t.Fatalf("unexpected count for analyzer a: %+v", got[0])
	}
	if got[1].Analyzer != "b" || got[1].Diagnostics != 2 || got[1].Fixable != 1 || got[1].DuplicatesRemoved != 0 {
		t.Fatalf("unexpected count for analyzer b: %+v", got[1])
	}
}

func TestBuildRepoTotalsDeduplicatesDiagnostics(t *testing.T) {
	dup := Diagnostic{
		RepoAbsPath: "/repo",
		Analyzer:    "a",
		Posn:        "f.go:1:1",
		End:         "f.go:1:2",
		Message:     "m1",
		SuggestedFixes: []SuggestedFix{{
			Message: "fix",
			Edits: []TextEdit{{
				Filename: "f.go",
				Start:    0,
				End:      1,
				New:      "x",
			}},
		}},
	}
	uniqA := Diagnostic{
		RepoAbsPath: "/repo",
		Analyzer:    "a",
		Posn:        "f.go:2:1",
		End:         "f.go:2:2",
		Message:     "m2",
	}
	uniqB := Diagnostic{
		RepoAbsPath: "/repo",
		Analyzer:    "b",
		Posn:        "g.go:1:1",
		End:         "g.go:1:2",
		Message:     "m3",
	}

	modules := []ModuleResult{
		{Repo: "repo", Module: "m1", Diagnostics: []Diagnostic{dup, uniqB}},
		{Repo: "repo", Module: "m1", Diagnostics: []Diagnostic{dup, uniqA}},
	}
	got := BuildRepoCounts(modules)
	if len(got) != 2 {
		t.Fatalf("unexpected count length: got=%d want=2", len(got))
	}

	if got[0].Analyzer != "a" || got[0].Diagnostics != 2 || got[0].Fixable != 1 || got[0].DuplicatesRemoved != 1 {
		t.Fatalf("unexpected repo count for analyzer a: %+v", got[0])
	}
	if got[1].Analyzer != "b" || got[1].Diagnostics != 1 || got[1].Fixable != 0 || got[1].DuplicatesRemoved != 0 {
		t.Fatalf("unexpected repo count for analyzer b: %+v", got[1])
	}
}

func TestBuildRepoTotalsDoesNotDeduplicateDifferentPackages(t *testing.T) {
	diagA := Diagnostic{
		RepoAbsPath: "/repo",
		PackageID:   "example.com/repo/a",
		Analyzer:    "a",
		Posn:        "f.go:1:1",
		End:         "f.go:1:2",
		Message:     "m1",
	}
	diagB := Diagnostic{
		RepoAbsPath: "/repo",
		PackageID:   "example.com/repo/b",
		Analyzer:    "a",
		Posn:        "f.go:1:1",
		End:         "f.go:1:2",
		Message:     "m1",
	}

	got := BuildRepoCounts([]ModuleResult{
		{Repo: "repo", Module: "m1", Diagnostics: []Diagnostic{diagA}},
		{Repo: "repo", Module: "m2", Diagnostics: []Diagnostic{diagB}},
	})

	if len(got) != 1 {
		t.Fatalf("unexpected count length: got=%d want=1", len(got))
	}
	if got[0].Diagnostics != 2 || got[0].DuplicatesRemoved != 0 {
		t.Fatalf("unexpected repo count: %+v", got[0])
	}
}

func TestBuildRepoTotalsDoesNotDeduplicateDifferentModules(t *testing.T) {
	diag := Diagnostic{
		RepoAbsPath: "/repo",
		PackageID:   "example.com/repo/pkg",
		Analyzer:    "a",
		Posn:        "f.go:1:1",
		End:         "f.go:1:2",
		Message:     "m1",
	}

	got := BuildRepoCounts([]ModuleResult{
		{Repo: "repo", Module: "m1", Diagnostics: []Diagnostic{diag}},
		{Repo: "repo", Module: "m2", Diagnostics: []Diagnostic{diag}},
	})

	if len(got) != 1 {
		t.Fatalf("unexpected count length: got=%d want=1", len(got))
	}
	if got[0].Diagnostics != 2 || got[0].DuplicatesRemoved != 0 {
		t.Fatalf("unexpected repo count: %+v", got[0])
	}
}

func TestBuildTotalDeduplicatesAcrossRepos(t *testing.T) {
	dup := Diagnostic{
		RepoAbsPath: "/repo1",
		Analyzer:    "a",
		Posn:        "f.go:1:1",
		End:         "f.go:1:2",
		Message:     "m1",
	}
	uniqA := Diagnostic{
		RepoAbsPath: "/repo2",
		Analyzer:    "a",
		Posn:        "f.go:3:1",
		End:         "f.go:3:2",
		Message:     "m2",
		SuggestedFixes: []SuggestedFix{{
			Message: "fix",
		}},
	}
	uniqB := Diagnostic{
		RepoAbsPath: "/repo1",
		Analyzer:    "b",
		Posn:        "g.go:1:1",
		End:         "g.go:1:2",
		Message:     "m3",
	}
	modules := []ModuleResult{
		{Repo: "repo1", Module: "m1", Diagnostics: []Diagnostic{dup}},
		{Repo: "repo1", Module: "m1", Diagnostics: []Diagnostic{dup, uniqB}},
		{Repo: "repo2", Module: "m3", Diagnostics: []Diagnostic{uniqA}},
	}

	got := BuildTotalCounts(modules)
	if len(got) != 2 {
		t.Fatalf("unexpected count length: got=%d want=2", len(got))
	}

	if got[0].Analyzer != "a" || got[0].Diagnostics != 2 || got[0].Fixable != 1 || got[0].DuplicatesRemoved != 1 {
		t.Fatalf("unexpected total count for analyzer a: %+v", got[0])
	}
	if got[1].Analyzer != "b" || got[1].Diagnostics != 1 || got[1].Fixable != 0 || got[1].DuplicatesRemoved != 0 {
		t.Fatalf("unexpected total count for analyzer b: %+v", got[1])
	}
}
