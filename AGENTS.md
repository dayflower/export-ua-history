# AGENTS.md

## Overview

- Project: `export-ua-history`
- Type: Go CLI for exporting Google Chrome browsing history as JSON
- Entry point: `./cmd/export-ua-history`

## Repository Layout

- `cmd/export-ua-history`: CLI bootstrap
- `internal/cli`: command parsing and user-facing CLI behavior
- `internal/browser`: browser-specific discovery and history access
- `internal/output`: JSON output shaping
- `internal/buildinfo`: version metadata
- `docs/SPEC.md`: product and behavior specification

## Common Commands

- Build: `make build`
- Test: `make test`
- Format: `make fmt`
- Tidy modules: `make tidy`
- Run locally: `make run ARGS='help'`
- Print version: `make version`

## Development Notes

- Prefer `make` targets over ad hoc commands when an equivalent target exists.
- Keep changes scoped and consistent with the current package boundaries under `internal/`.
- Update `README.md` when user-visible CLI behavior changes.
- Update `docs/SPEC.md` when behavior or requirements change materially.

## Testing Expectations

- Run `make test` after behavior changes.
- Run `make fmt` for Go source changes before finishing.
- Add or update tests alongside code changes when logic changes are introduced.

## CLI and Output Constraints

- Preserve the existing CLI shape unless the task explicitly requires a breaking change.
- Keep JSON output stable and UTF-8 encoded with a trailing newline.
- Treat local time and profile/path handling carefully, especially in platform-specific files.

## Platform Notes

- Platform-specific behavior lives under files such as `*_darwin.go`, `*_linux.go`, and `*_windows.go`.
- Keep cross-platform interfaces aligned when changing browser path or platform resolution logic.

## Change Hygiene

- Make the smallest change that satisfies the request.
- Avoid unrelated refactors in the same patch.
- Follow existing naming and package organization patterns.
