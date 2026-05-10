package main

import "testing"

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
