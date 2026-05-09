package report

import (
	"sort"
	"strconv"
	"strings"
)

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

func BuildModuleCounts(diags []Diagnostic) []Count {
	countByAnalyzer := map[string]*Count{}
	for _, d := range diags {
		c, ok := countByAnalyzer[d.Analyzer]
		if !ok {
			c = &Count{Analyzer: d.Analyzer}
			countByAnalyzer[d.Analyzer] = c
		}
		c.Diagnostics++
		if len(d.SuggestedFixes) > 0 {
			c.Fixable++
		}
	}
	return collectCounts(countByAnalyzer)
}

func BuildRepoCounts(modules []ModuleResult) []Count {
	return buildDeduplicatedCounts(modules)
}

func BuildTotalCounts(repos []ModuleResult) []Count {
	return buildDeduplicatedCounts(repos)
}

func buildDeduplicatedCounts(modules []ModuleResult) []Count {
	countByAnalyzer := map[string]*Count{}
	seen := map[string]struct{}{}

	for _, m := range modules {
		for _, d := range m.Diagnostics {
			c, ok := countByAnalyzer[d.Analyzer]
			if !ok {
				c = &Count{Analyzer: d.Analyzer}
				countByAnalyzer[d.Analyzer] = c
			}

			fp := Fingerprint(d)
			if _, ok := seen[fp]; ok {
				c.DuplicatesRemoved++
				continue
			}
			seen[fp] = struct{}{}

			c.Diagnostics++
			if len(d.SuggestedFixes) > 0 {
				c.Fixable++
			}
		}
	}

	return collectCounts(countByAnalyzer)
}

func collectCounts(countByAnalyzer map[string]*Count) []Count {
	counts := make([]Count, 0, len(countByAnalyzer))
	for _, c := range countByAnalyzer {
		counts = append(counts, *c)
	}
	sortCounts(counts)
	return counts
}

func Fingerprint(diag Diagnostic) string {
	var b strings.Builder
	writeField(&b, diag.RepoAbsPath)
	writeField(&b, diag.Analyzer)
	writeField(&b, diag.Posn)
	writeField(&b, diag.End)
	writeField(&b, diag.Message)
	writeField(&b, strconv.Itoa(len(diag.SuggestedFixes)))
	for _, fix := range diag.SuggestedFixes {
		writeField(&b, fix.Message)
		writeField(&b, strconv.Itoa(len(fix.Edits)))
		for _, edit := range fix.Edits {
			writeField(&b, edit.Filename)
			writeField(&b, strconv.Itoa(edit.Start))
			writeField(&b, strconv.Itoa(edit.End))
			writeField(&b, edit.New)
		}
	}
	return b.String()
}

func writeField(b *strings.Builder, s string) {
	b.WriteString(strconv.Itoa(len(s)))
	b.WriteByte(':')
	b.WriteString(s)
	b.WriteByte('|')
}
