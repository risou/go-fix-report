# go-fix-report

`gofixreport` runs `go fix -json ./...` for Go modules under target git repositories and reports analyzer counts.

The first version requires a Go 1.26 or newer `go` command at runtime because it depends on `go fix -json`.

## Usage

```bash
gofixreport [--json] [path]
```

By default, output is a human-readable table. Use `--json` for machine-readable output.
