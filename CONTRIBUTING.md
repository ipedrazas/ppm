# Contributing to ppm

## Getting started

```sh
git clone https://github.com/ipedrazas/ppm.git
cd ppm
go build ./...
go test ./...
```

## Reporting issues

- Search existing issues before opening a new one.
- Include the `ppm` version (`ppm --version`), OS, and a minimal reproduction.
- For bugs, describe expected vs actual behaviour and include the exact command that failed.

## Making changes

1. Fork the repo and create a branch from `main`.
2. Make your change. Keep commits focused — one logical change per commit.
3. Add or update tests if you change behaviour.
4. Run `go vet ./...` and `go test ./...` before pushing.
5. Open a pull request against `main` with a clear description of what and why.

## Code conventions

- Output format contract: every command must emit the uniform `{ "ok", "message", "data" }` envelope. Do not break this.
- Entry types are a closed set — adding a new type requires updating the type registry, not just the parser.
- JSON is the default output mode; `-o text` is secondary. Keep both paths working.
- No external dependencies without discussion — the dependency footprint is intentionally small.

## Running the test suite

```sh
go test ./...
go vet ./...
```

There are no integration tests that require a live filesystem beyond what `go test` already exercises via temp directories.

## Commit messages

Use the conventional commits style:

```
feat: add question resolve --partial flag
fix: preserve frontmatter key order on update
docs: document memory root resolution order
```
