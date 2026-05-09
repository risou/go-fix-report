package parser

import (
	"bytes"
	"encoding/json"

	"github.com/risou/go-fix-report/internal/report"
)

type Result struct {
	Diagnostics []report.Diagnostic
	Errors      []report.AnalyzerError
}

func Parse(stdout []byte) (Result, error) {
	if len(bytes.TrimSpace(stdout)) == 0 {
		return Result{}, nil
	}

	var packages map[string]map[string]json.RawMessage
	if err := json.Unmarshal(stdout, &packages); err != nil {
		return Result{}, err
	}

	var result Result
	for packageID, analyzers := range packages {
		for analyzer, raw := range analyzers {
			var diagnostics []report.Diagnostic
			diagErr := json.Unmarshal(raw, &diagnostics)
			if diagErr == nil {
				for i := range diagnostics {
					diagnostics[i].PackageID = packageID
					diagnostics[i].Analyzer = analyzer
				}
				result.Diagnostics = append(result.Diagnostics, diagnostics...)
				continue
			}

			var analyzerErr struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(raw, &analyzerErr); err == nil && analyzerErr.Error != "" {
				result.Errors = append(result.Errors, report.AnalyzerError{
					PackageID: packageID,
					Analyzer:  analyzer,
					Error:     analyzerErr.Error,
				})
				continue
			}

			return Result{}, diagErr
		}
	}

	return result, nil
}
