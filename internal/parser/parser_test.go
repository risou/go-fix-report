package parser

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseDiagnostics(t *testing.T) {
	stdout := []byte(`{
  "example.com/x": {
    "any": [
      {
        "posn": "/tmp/x.go:3:10",
        "end": "/tmp/x.go:3:21",
        "message": "interface{} can be replaced by any",
        "suggested_fixes": [
          {
            "message": "Replace interface{} by any",
            "edits": [
              {"filename":"/tmp/x.go","start":33,"end":44,"new":"any"}
            ]
          }
        ]
      }
    ]
  }
}`)

	result, err := Parse(stdout)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got := len(result.Errors); got != 0 {
		t.Fatalf("len(result.Errors) = %d, want 0", got)
	}
	if got := len(result.Diagnostics); got != 1 {
		t.Fatalf("len(result.Diagnostics) = %d, want 1", got)
	}

	d := result.Diagnostics[0]
	if d.PackageID != "example.com/x" {
		t.Fatalf("d.PackageID = %q, want %q", d.PackageID, "example.com/x")
	}
	if d.Analyzer != "any" {
		t.Fatalf("d.Analyzer = %q, want %q", d.Analyzer, "any")
	}
	if got := len(d.SuggestedFixes); got != 1 {
		t.Fatalf("len(d.SuggestedFixes) = %d, want 1", got)
	}
}

func TestParseDiagnosticWithoutSuggestedFix(t *testing.T) {
	stdout := []byte(`{
  "example.com/x": {
    "any": [
      {
        "posn": "/tmp/x.go:3:10",
        "end": "/tmp/x.go:3:21",
        "message": "interface{} can be replaced by any"
      }
    ]
  }
}`)

	result, err := Parse(stdout)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got := len(result.Diagnostics); got != 1 {
		t.Fatalf("len(result.Diagnostics) = %d, want 1", got)
	}
	if got := len(result.Diagnostics[0].SuggestedFixes); got != 0 {
		t.Fatalf("len(result.Diagnostics[0].SuggestedFixes) = %d, want 0", got)
	}
}

func TestParseAnalyzerError(t *testing.T) {
	stdout := []byte(`{
  "example.com/x": {
    "any": {
      "error": "analysis failed"
    }
  }
}`)

	result, err := Parse(stdout)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got := len(result.Diagnostics); got != 0 {
		t.Fatalf("len(result.Diagnostics) = %d, want 0", got)
	}
	if got := len(result.Errors); got != 1 {
		t.Fatalf("len(result.Errors) = %d, want 1", got)
	}

	e := result.Errors[0]
	if e.PackageID != "example.com/x" {
		t.Fatalf("e.PackageID = %q, want %q", e.PackageID, "example.com/x")
	}
	if e.Analyzer != "any" {
		t.Fatalf("e.Analyzer = %q, want %q", e.Analyzer, "any")
	}
	if e.Error != "analysis failed" {
		t.Fatalf("e.Error = %q, want %q", e.Error, "analysis failed")
	}
}

func TestParseEmptyOutput(t *testing.T) {
	result, err := Parse([]byte(" \n\t "))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got := len(result.Diagnostics); got != 0 {
		t.Fatalf("len(result.Diagnostics) = %d, want 0", got)
	}
	if got := len(result.Errors); got != 0 {
		t.Fatalf("len(result.Errors) = %d, want 0", got)
	}
}

func TestParseUnexpectedAnalyzerObjectShape(t *testing.T) {
	stdout := []byte(`{
  "example.com/x": {
    "any": {
      "foo": "bar"
    }
  }
}`)

	_, err := Parse(stdout)
	if err == nil {
		t.Fatal("Parse returned nil error, want unmarshal error")
	}
	var unmarshalTypeErr *json.UnmarshalTypeError
	if !errors.As(err, &unmarshalTypeErr) {
		t.Fatalf("Parse error type = %T, want *json.UnmarshalTypeError", err)
	}
}

func TestParseUnexpectedAnalyzerScalarValue(t *testing.T) {
	stdout := []byte(`{
  "example.com/x": {
    "any": 123
  }
}`)

	_, err := Parse(stdout)
	if err == nil {
		t.Fatal("Parse returned nil error, want unmarshal error")
	}
	var unmarshalTypeErr *json.UnmarshalTypeError
	if !errors.As(err, &unmarshalTypeErr) {
		t.Fatalf("Parse error type = %T, want *json.UnmarshalTypeError", err)
	}
}
