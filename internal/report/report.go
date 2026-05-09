package report

import "sort"

type TextEdit struct {
	Filename string `json:"filename"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	New      string `json:"new"`
}

type SuggestedFix struct {
	Message string     `json:"message"`
	Edits   []TextEdit `json:"edits"`
}

type Diagnostic struct {
	RepoAbsPath    string         `json:"-"`
	PackageID      string         `json:"package_id"`
	Analyzer       string         `json:"analyzer"`
	Posn           string         `json:"posn"`
	End            string         `json:"end"`
	Message        string         `json:"message"`
	SuggestedFixes []SuggestedFix `json:"suggested_fixes,omitempty"`
}

type Count struct {
	Analyzer          string `json:"analyzer"`
	Diagnostics       int    `json:"diagnostics"`
	Fixable           int    `json:"fixable"`
	DuplicatesRemoved int    `json:"duplicates_removed,omitempty"`
}

type ModuleResult struct {
	Repo        string       `json:"repo"`
	Module      string       `json:"module"`
	Diagnostics []Diagnostic `json:"-"`
	Counts      []Count      `json:"counts"`
}

type RepoResult struct {
	Repo   string  `json:"repo"`
	Counts []Count `json:"counts"`
}

type RunError struct {
	Repo     string `json:"repo"`
	Module   string `json:"module"`
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Stderr   string `json:"stderr"`
}

type AnalyzerError struct {
	PackageID string `json:"package_id"`
	Analyzer  string `json:"analyzer"`
	Error     string `json:"error"`
}

type Result struct {
	Modules []ModuleResult `json:"modules"`
	Repos   []RepoResult   `json:"repos"`
	Total   []Count        `json:"total"`
	Errors  []RunError     `json:"errors"`
}

func sortCounts(counts []Count) {
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].Analyzer < counts[j].Analyzer
	})
}
