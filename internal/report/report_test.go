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
