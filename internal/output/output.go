package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	"github.com/risou/go-fix-report/internal/report"
)

func WriteJSON(w io.Writer, result report.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func WriteTable(w io.Writer, result report.Result) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "MODULE RESULTS")
	fmt.Fprintln(tw, "repo\tmodule\tanalyzer\tdiagnostics\tfixable")
	for _, module := range result.Modules {
		for _, c := range module.Counts {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\n", module.Repo, module.Module, c.Analyzer, c.Diagnostics, c.Fixable)
		}
	}
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, "REPO TOTALS")
	fmt.Fprintln(tw, "repo\tanalyzer\tdiagnostics\tfixable\tduplicates_removed")
	for _, repo := range result.Repos {
		for _, c := range repo.Counts {
			fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\n", repo.Repo, c.Analyzer, c.Diagnostics, c.Fixable, c.DuplicatesRemoved)
		}
	}
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, "TOTAL")
	fmt.Fprintln(tw, "analyzer\tdiagnostics\tfixable\tduplicates_removed")
	for _, c := range result.Total {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", c.Analyzer, c.Diagnostics, c.Fixable, c.DuplicatesRemoved)
	}

	if len(result.Errors) > 0 {
		fmt.Fprintln(tw)
		fmt.Fprintln(tw, "ERRORS")
		fmt.Fprintln(tw, "repo\tmodule\tcommand\texit_code\tstderr")
		for _, e := range result.Errors {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", e.Repo, e.Module, e.Command, strconv.Itoa(e.ExitCode), e.Stderr)
		}
	}

	return tw.Flush()
}
