package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
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

	if err := writeLine(tw, "MODULE RESULTS"); err != nil {
		return err
	}
	if err := writeLine(tw, "repo\tmodule\tanalyzer\tdiagnostics\tfixable"); err != nil {
		return err
	}
	for _, module := range result.Modules {
		for _, c := range module.Counts {
			if err := writef(tw, "%s\t%s\t%s\t%d\t%d\n", module.Repo, module.Module, c.Analyzer, c.Diagnostics, c.Fixable); err != nil {
				return err
			}
		}
	}
	if err := writeLine(tw); err != nil {
		return err
	}

	if err := writeLine(tw, "REPO TOTALS"); err != nil {
		return err
	}
	if err := writeLine(tw, "repo\tanalyzer\tdiagnostics\tfixable\tduplicates_removed"); err != nil {
		return err
	}
	for _, repo := range result.Repos {
		for _, c := range repo.Counts {
			if err := writef(tw, "%s\t%s\t%d\t%d\t%d\n", repo.Repo, c.Analyzer, c.Diagnostics, c.Fixable, c.DuplicatesRemoved); err != nil {
				return err
			}
		}
	}
	if err := writeLine(tw); err != nil {
		return err
	}

	if err := writeLine(tw, "TOTAL"); err != nil {
		return err
	}
	if err := writeLine(tw, "analyzer\tdiagnostics\tfixable\tduplicates_removed"); err != nil {
		return err
	}
	for _, c := range result.Total {
		if err := writef(tw, "%s\t%d\t%d\t%d\n", c.Analyzer, c.Diagnostics, c.Fixable, c.DuplicatesRemoved); err != nil {
			return err
		}
	}

	if len(result.Errors) > 0 {
		if err := writeLine(tw); err != nil {
			return err
		}
		if err := writeLine(tw, "ERRORS"); err != nil {
			return err
		}
		if err := writeLine(tw, "repo\tmodule\tcommand\texit_code\tstderr"); err != nil {
			return err
		}
		for _, e := range result.Errors {
			if err := writef(tw, "%s\t%s\t%s\t%s\t%s\n", e.Repo, e.Module, e.Command, strconv.Itoa(e.ExitCode), sanitizeCell(e.Stderr)); err != nil {
				return err
			}
		}
	}

	return tw.Flush()
}

func writeLine(w io.Writer, a ...any) error {
	_, err := fmt.Fprintln(w, a...)
	return err
}

func writef(w io.Writer, format string, a ...any) error {
	_, err := fmt.Fprintf(w, format, a...)
	return err
}

func sanitizeCell(s string) string {
	replacer := strings.NewReplacer("\r", " ", "\n", " ", "\t", " ")
	return strings.Join(strings.Fields(replacer.Replace(s)), " ")
}
