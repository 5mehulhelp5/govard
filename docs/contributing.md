# Contributing

This guide covers the expected workflow for contributors working on Govard.

## Toolchain

Required:

- Go `1.25+`
- Node.js `20+`
- Docker Engine + Docker Compose plugin
- Wails `v2.11+` for desktop development

Common local checks:

```bash
go version
node --version
wails version
docker --version
```

## Repository Map

```text
.
├── cmd/govard/
├── cmd/govard-desktop/
├── desktop/
├── internal/cmd/
├── internal/engine/
├── internal/proxy/
├── internal/ui/
├── internal/updater/
├── docker/
├── tests/
├── scripts/
└── docs/
```

## Build Commands

```bash
make build
make install
make install-release
go build -o govard cmd/govard/main.go
```

## Test Commands

Preferred commands:

```bash
make test
make test-unit
make test-integration
make vet
make fmt
```

Useful direct commands:

```bash
go test ./...
go test ./tests/... -v
go test -tags integration ./tests/integration/... -v -timeout 30m
```

## Desktop Development

Run Wails dev mode through the CLI wrapper:

```bash
DISPLAY=:1 govard desktop --dev
```

For browser-based UI testing, use the Wails dev server:

```text
http://localhost:34115
```

## Test Layout

Canonical test fixture locations:

- `tests/fixtures/`: shared fixture files used by unit tests
- `tests/integration/projects/`: integration project fixtures per framework or workflow

Keep tests hermetic:

- do not rely on user-local projects
- do not rely on real containers unless the test is explicitly integration-tagged
- prefer mocks and temp directories
- isolate `GOVARD_HOME_DIR` in tests that need runtime state

When a test needs access to internal logic from `internal/cmd`, add a narrow exported wrapper with a `ForTest` suffix rather than broadening the production API.

## Contribution Rules

Before declaring work done:

1. run tests relevant to the changed area
2. run `gofmt` on changed Go files
3. update canonical docs in `docs/*.md` when behavior changes
4. check `git status` for unintended file changes

## Command and Docs Hygiene

When changing CLI behavior:

1. update the command implementation in `internal/cmd/`
2. update tests in `tests/`
3. update the canonical top-level docs file for that topic

Current canonical docs:

- [Getting Started](getting-started.md)
- [Commands](commands.md)
- [Configuration](configuration.md)
- [Remotes and Sync](remotes-and-sync.md)
- [Frameworks](frameworks.md)
- [Desktop](desktop.md)
- [Architecture](architecture.md)

## Dependency and Security Expectations

- prefer stdlib unless a new dependency has clear payoff
- do not log secrets, tokens, or private keys
- keep remote and database commands conservative by default
- preserve release artifact naming and checksum expectations

## Related Docs

- [Architecture](architecture.md)
- [Commands](commands.md)
