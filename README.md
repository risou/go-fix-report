# go-fix-report

`gofixreport` runs `go fix -json ./...` for Go modules under target git repositories and reports analyzer counts.

`gofixreport` itself is a normal Go CLI, but the analyzed environment must provide a Go 1.26 or newer `go` command. Older Go versions do not support `go fix -json`.

## Usage

```bash
gofixreport [--json] [--jobs N] [path]
```

By default, output is a human-readable table. Use `--json` for machine-readable output.

By default, modules are processed sequentially. Use `--jobs N` to process up to `N` modules concurrently.
