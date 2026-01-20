# Agent Guide (ignr)

This repository is a Go CLI + TUI app (Cobra + Bubble Tea v2).
Keep changes small, match existing patterns, and let the linters drive style.

## Quick Commands

### Build

- Build local binary:
  - `go build ./cmd/ignr`
- Install to your Go bin:
  - `go install ./cmd/ignr`
- Run via `go run`:
  - `go run ./cmd/ignr --help`

### Test

- Run a single package (preferred while iterating):
  - `go test ./internal/templates`
- Run a single test (regex on test name):
  - `go test -run '^TestMergeTemplates$' ./internal/templates`
- Verbose + single test:
  - `go test -v -run '^TestMergeTemplates$' ./internal/templates`
- Run tests with coverage:
  - `go test ./internal/templates -coverprofile=coverage.out`

Notes:
- `go test ./...` currently fails due to pre-existing compile errors in some `_test.go` files (unused imports / bad table types).
- Prefer `go test ./internal/<pkg>` or `go test -run '^TestX$' ./internal/<pkg>` while iterating.

### Lint / Format

This repo is configured for `golangci-lint` and expects goimports-style formatting.

- Lint:
  - `golangci-lint run`
- Lint + autofix (only for fixable issues):
  - `golangci-lint run --fix`

Formatter expectations come from `golangci.yml`:
- `goimports` enabled (with local-prefix `go.seanlatimer.dev/ignr`)
- `golines` enabled (max line length 120)

Notes:
- `golangci-lint run` currently fails because typechecking fails in broken test files.

### Release (Local)

GoReleaser is configured in `.goreleaser.yaml`.

- Validate a snapshot build:
  - `goreleaser build --snapshot --clean`

## Repo Layout (How to Place Code)

- `cmd/ignr/`: entry point only (`main.go` calls into `pkg/cli`).
- `pkg/cli/`: Cobra commands and CLI wiring; keep command functions small and delegate real work.
- `internal/`: core implementation (cache, templates, presets, config, tui).
- `testutil/`: helpers for tests (spawning CLI via `go run`, creating temp repos/caches, etc.).
- `ref/`: reference Bubble Tea examples; may not compile under this module; ignore for builds/lint unless explicitly updating them.

Guideline: prefer putting business logic in `internal/` and keeping `pkg/cli/` as thin glue.

## Code Style Rules (Inferred + Enforced)

### Formatting

- Go files use tabs (see `.editorconfig`).
- Keep lines <= 120 chars (see `golines` config).
- Run goimports-equivalent formatting; imports should be grouped:
  1) standard library
  2) blank line
  3) third-party
  4) blank line
  5) module-local imports (`go.seanlatimer.dev/ignr/...`)

### Imports

- Prefer explicit imports; avoid dot imports.
- Aliases are fine when conventional or required (e.g. `tea "charm.land/bubbletea/v2"`).
- `golangci.yml` sets `go.seanlatimer.dev/ignr` as the local prefix for import grouping.

### Naming

- Packages: short, lowercase, single-word (`cache`, `templates`, `tui`).
- Exported identifiers: PascalCase.
- Unexported identifiers: camelCase.
- Constructors:
  - exported: `NewX(...)`
  - unexported: `newX(...)`
- Keep receiver names short (often 1 letter) and consistent.

### Errors

- Prefer early returns.
- Wrap errors with context using `%w`:
  - `return fmt.Errorf("read config: %w", err)`
- Avoid stringly-typed error comparisons; use sentinel errors when callers must branch.
- Don’t drop errors; `errcheck` is enabled.

### CLI (Cobra)

- Global flags live in a shared `Options` struct (`pkg/cli/root.go`).
- New commands should follow existing factory pattern:
  - `func newThingCommand(opts *Options) *cobra.Command`
- Avoid doing heavy work directly in cobra handlers; call into `internal/*`.

### TUI (Bubble Tea)

- Follow MVU conventions: model state, `Init`, `Update`, `View`.
- Keep styles centralized (see `internal/tui/styles.go`).
- Avoid mixing rendering/styling logic into command handlers.

### Types and Safety

- Keep types explicit; avoid `interface{}` unless you have to.
- Avoid global mutable state; `gochecknoglobals` is enabled.
- Avoid `init()`; `gochecknoinits` is enabled.
- Avoid `log` outside `main`; `depguard` blocks `log` in non-main files (use `log/slog` if you introduce logging).

### Concurrency / Context

- If you add long-running or cancellable operations, consider threading `context.Context` through `internal/*` APIs.
- If you add loops with closures, be mindful of Go loop-var capture; `copyloopvar` is enabled.

## Linting Constraints Worth Knowing

`golangci.yml` is strict (based on maratori config). Expect checks for:

- error handling (`errcheck`, `errorlint`, `nilerr`, `nilnil`)
- complexity (`cyclop`, `funlen`, `gocognit`)
- forbidden patterns (`forbidigo`, `gochecknoglobals`, `gochecknoinits`)
- modern Go idioms (`modernize`, `usestdlibvars`)
- security (`gosec`, `bidichk`)

When adding code, run `golangci-lint run` locally and fix issues in the code you touched.

## Existing Repo Gotchas

- The checked-in `.gitignore` contains a broken glob (`Icon[` / `]` split across lines). This can break tools that parse gitignore patterns (e.g. `rg` on Windows). Prefer excluding `.gitignore` or fixing it if you are working on repo hygiene.
- Some tests currently do not compile (type mismatches / unused imports). Don’t assume `go test ./...` or `golangci-lint run` is green unless you’ve verified and/or fixed them.

## Cursor / Copilot Rules

No Cursor rules found in `.cursor/rules/` or `.cursorrules`.
No Copilot instructions found in `.github/copilot-instructions.md`.

