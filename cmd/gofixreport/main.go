package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/risou/go-fix-report/internal/app"
	"github.com/risou/go-fix-report/internal/output"
)

func main() {
	jsonOutput := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	path := "."
	if flag.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "usage: gofixreport [--json] [path]")
		os.Exit(2)
	}
	if flag.NArg() == 1 {
		path = flag.Arg(0)
	}

	result := app.Run(context.Background(), app.Options{Path: path}, app.Dependencies{})

	var err error
	if *jsonOutput {
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
