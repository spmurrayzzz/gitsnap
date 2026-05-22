# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go module for the `gitsnap` CLI. The executable entry
point lives in `cmd/gitsnap/`. Shared implementation packages are under
`internal/`: `alias` manages named snapshot references, `backend/gogit`
contains the go-git backend, `ref` resolves aliases or hashes, `store` manages
repository-local storage paths, and `treehash` handles tree hashing helpers.
Tests are colocated with the package they cover and use Go's `_test.go`
convention, including focused extra coverage files such as
`internal/store/store_extra_test.go`.

## Build, Test, and Development Commands

- `make build`: builds the CLI binary at `bin/gitsnap`.
- `make test`: runs `go test ./...` across all packages.
- `make clean`: removes generated build output from `bin/`.
- `go run ./cmd/gitsnap --help`: runs the CLI directly during development.

Use `go mod tidy` after dependency changes so `go.mod` and `go.sum` stay in
sync.

## Coding Style & Naming Conventions

Follow standard Go style: format code with `gofmt`, keep package names short
and lowercase, and prefer small functions with explicit error returns. Use tabs
for Go indentation as produced by `gofmt`. Keep lines readable in 80-column
terminals where practical. Avoid unnecessary abstractions and keep new code
close to the existing package boundaries.

Name tests after the behavior under test, for example
`TestResolveAlias` or `TestBackendSaveIncludesWorktreeChanges`. Keep helper
functions unexported unless another package genuinely needs them.

## Testing Guidelines

The project uses Go's built-in `testing` package. Add or update colocated
`*_test.go` files when behavior changes. Prefer table-driven tests for parsing,
resolution, and validation logic. Backend tests should isolate filesystem and
repository state with temporary directories rather than relying on developer
machine state.

Run `make test` before committing. For targeted work, run a package-specific
command such as `go test ./internal/ref`.

## Commit & Pull Request Guidelines

Git history uses Conventional Commits such as `feat: implement gitsnap
snapshot CLI`, `test: cover CLI and metadata helpers`, and `chore: initialize
Go module and build tooling`. Keep using short, imperative subjects with a
type prefix like `feat`, `fix`, `test`, or `chore`; add a scope when useful,
for example `test(gogit): cover backend error paths`.

Pull requests should describe the behavior change, list the validation command
run, and call out any user-visible CLI changes. Include terminal output or
screenshots only when they clarify command behavior.
