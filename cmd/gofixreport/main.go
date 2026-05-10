package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/risou/go-fix-report/internal/app"
	"github.com/risou/go-fix-report/internal/output"
)

func main() {
	opts, jsonOutput, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "usage: gofixreport [--json] [--jobs N] [path]")
		os.Exit(2)
	}

	result := app.Run(context.Background(), opts, app.Dependencies{})

	if jsonOutput {
		err = output.WriteJSON(os.Stdout, result.Report)
	} else {
		err = output.WriteTable(os.Stdout, result.Report)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "writing output: %v\n", err)
		os.Exit(1)
	}
	if result.HadErrors {
		os.Exit(1)
	}
}

func parseArgs(args []string) (app.Options, bool, error) {
	fs := flag.NewFlagSet("gofixreport", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON output")
	jobs := fs.Int("jobs", 1, "number of modules to process concurrently")
	if err := fs.Parse(args); err != nil {
		return app.Options{}, false, err
	}
	if fs.NArg() > 1 {
		return app.Options{}, false, fmt.Errorf("too many arguments")
	}
	if *jobs < 1 {
		return app.Options{}, false, fmt.Errorf("jobs must be greater than zero")
	}

	path := "."
	if fs.NArg() == 1 {
		path = fs.Arg(0)
	}

	return app.Options{
		Path: path,
		Jobs: *jobs,
	}, *jsonOutput, nil
}
