# AGENTS.md

This document is the project-specific operating manual for AI coding agents working in `govard`.

## 1. Mission and Product Context

`Govard` is a Go-based local development orchestrator for PHP and web projects (Magento, Laravel, Symfony, WordPress, etc.), with:

- CLI orchestration (`govard ...`)
- Container runtime automation (Docker)
- Blueprint/template rendering
- Remote environment tooling
- SSL/proxy utilities
- Optional desktop app scaffolding

Primary goals for contributions:

1. Preserve CLI stability and predictable behavior.
2. Keep workflows fast for local developers.
3. Avoid regressions in bootstrap/proxy/db/sync/remote command families.
4. Maintain release quality (GoReleaser, checksums, install paths).

## 2. Runtime and Toolchain Requirements

- Go: `1.24+` (module uses `go 1.24.0`)
- Node.js: `20` (for frontend tests in CI)
- Docker: required for integration tests and runtime orchestration
- GitHub CLI (`gh`): useful for release inspection (optional)

Local sanity checks:

```bash
go version
node --version
docker --version
```

## 3. Repository Map

- `cmd/govard/main.go`: CLI entrypoint
- `internal/cmd/`: Cobra command implementations
- `internal/engine/`: orchestration, config, blueprint logic
- `internal/engine/bootstrap/`: framework bootstrap workflows
- `internal/engine/remote/`: remote sync/deploy/ssh helpers
- `internal/proxy/`: caddy/proxy route and TLS helpers
- `internal/updater/`: update-check notification logic
- `internal/ui/`: terminal rendering helpers
- `tests/`: unit/contract tests (default location for tests)
- `tests/integration/`: integration tests (tagged and heavier)
- `tests/frontend/`: Node test runner suite for frontend pieces
- `scripts/install.sh`: build-from-source installer
- `scripts/install-release.sh`: release-binary one-line installer
- `.goreleaser.yml`: release artifact config
- `.github/workflows/`: CI/release/security automation

## 4. Core Build and Test Commands

Preferred commands:

```bash
make test-fast           # frontend + unit tests
make test-unit           # unit tests only
make test-integration    # integration tests (requires build + docker)
make test                # full test suite
make vet                 # go vet
make fmt                 # go fmt ./...
make build               # build release binaries (linux amd64, darwin arm64)
```

Useful direct commands:

```bash
go test ./...
go test ./tests/... -v
go test -tags integration ./tests/integration/... -v -timeout 30m
```

## 5. CI and Quality Gates

Key workflows:

- `ci-pipeline.yml`: vet + gofmt check + fast tests + integration + binary build
- `release.yml`: triggered by tag `v*.*.*`, runs GoReleaser
- `codeql.yml`: code scanning
- `govulncheck.yml`: weekly vulnerability scan

Do not assume local success if you skipped:

1. formatting (`gofmt -s -w` / `make fmt`)
2. tests relevant to changed surface
3. at least one command-level smoke check for CLI behavior changes

## 6. Testing Conventions (Important)

Project convention is to keep most tests in `tests/` package `tests`.

When you need to test internal logic from `internal/cmd` (or other internal packages):

1. keep production helpers unexported where possible
2. add narrow exported wrappers suffixed with `ForTest` (example: `BuildLocalDBResetScriptForTest`)
3. consume those wrappers from test files in `tests/`

Example pattern:

- production function: `buildThing(...)`
- test wrapper: `BuildThingForTest(...)`
- test location: `tests/thing_test.go`

Avoid broad export of internals just for tests.

## 7. CLI Architecture and Command Work

`internal/cmd/root.go` owns root registration and `Version` binding.

When adding/modifying a command:

1. define command in `internal/cmd/<area>.go`
2. register with `rootCmd.AddCommand(...)` (or relevant subcommand group)
3. ensure flags are explicit and help text is actionable
4. return errors with context (`fmt.Errorf("operation: %w", err)`)
5. add/adjust tests in `tests/`
6. update docs for user-visible command/flag changes

For update/release-sensitive commands:

- confirm asset naming matches `.goreleaser.yml`
- verify checksums where possible
- avoid assuming `sudo` availability

## 8. Installer and Update Expectations

### Release artifacts

Current GoReleaser naming (non-Windows):

`govard_<version>_<OS>_<arch>.tar.gz`

Windows:

`govard_<version>_<OS>_<arch>.zip`

Checksum file:

`checksums.txt`

### Installer scripts

- `scripts/install.sh`: source build install
- `scripts/install-release.sh`: download release asset + checksum verification

### Self-update behavior

`govard self-update` should:

1. resolve target version (explicit or latest)
2. resolve platform-specific artifact name
3. download archive and checksum
4. verify SHA-256
5. extract binary
6. atomically replace executable (or fail with clear permissions guidance)

## 9. Coding Standards

- Always run `gofmt` after Go edits.
- Keep code ASCII unless file already requires Unicode.
- Prefer small pure helpers for parsing/formatting logic.
- Keep platform branching explicit (`runtime.GOOS`, `runtime.GOARCH`).
- Do not swallow errors silently for critical flows (network, file, process).
- Preserve existing UX tone from `pterm` output and command help strings.

## 10. Dependency and Security Guidance

- Avoid adding dependencies unless required by measurable benefit.
- Prefer Go stdlib for HTTP, archive, hash, file operations.
- Never log secrets, tokens, private keys, or DB passwords.
- For remote/ssh/db commands, keep safe defaults and explicit opt-in for write ops.

## 11. Documentation Update Rules

Update `README.md` when changes affect:

1. installation
2. upgrade/update flow
3. command names/flags
4. release consumption

If behavior changed and docs are stale, treat as incomplete work.

## 12. Git and Change Hygiene

- Keep commits focused by concern (installer, command logic, tests, docs).
- Do not revert unrelated working-tree changes.
- Avoid destructive git operations unless explicitly requested.
- Include test evidence in PR/hand-off notes.

## 13. Recommended Agent Workflow

1. Read impacted files and nearby tests.
2. Identify minimal change set.
3. Implement with small helpers and clear errors.
4. Add/adjust tests in `tests/`.
5. Run formatting and targeted tests.
6. Run broader suite as needed (`make test-fast` or `go test ./...`).
7. Update README/docs if user-facing behavior changed.
8. Summarize file-level changes and verification evidence.

## 14. Pre-Completion Checklist

Before declaring done:

1. `go test` on affected scope passes.
2. No formatting drift (`gofmt -s -l .` should be empty for changed Go files).
3. Command help/flags still coherent.
4. Docs updated for user-visible changes.
5. `git status` reviewed for unintended file changes.

## 15. Known Project-Specific Notes

- CI tracks `main`, `master`, and `develop`.
- Default branch currently appears as `master` in this repo.
- Release tags follow semantic style `vX.Y.Z`.
- Integration tests rely on built binary (`bin/govard-test`) and Docker.

When uncertain, prefer compatibility and least-surprise behavior over broad refactors.
