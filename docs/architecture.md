# Architecture

This document describes the current system shape of Govard at a high level.

## Product Shape

Govard is a Go-native local development orchestrator with two user surfaces:

- CLI powered by Cobra
- Desktop app powered by Wails

Both surfaces reuse the same runtime engine for discovery, rendering, Docker orchestration, proxy integration, and remote workflows.

## Core Runtime

### CLI layer

- entrypoint: `cmd/govard/main.go`
- command registration: `internal/cmd/`
- terminal UX helpers: `internal/ui/`

### Engine layer

- discovery and config: `internal/engine/`
- bootstrap workflows: `internal/engine/bootstrap/`
- remote logic: `internal/engine/remote/`
- proxy and TLS: `internal/proxy/`
- updater: `internal/updater/`

### Startup pipeline

`govard env up` follows the same core phases across supported frameworks:

1. Detect framework context
2. Validate config and host prerequisites
3. Render project compose output into `~/.govard/compose/`
4. Start runtime containers
5. Verify proxy and host wiring

## Networking

- shared proxy: Caddy
- local DNS routing for `.test` domains via `dnsmasq`
- shared Docker networks for project and proxy communication

The proxy terminates HTTPS and forwards requests into the current project stack.

## Configuration Model

Govard composes layered config files on top of framework blueprints.

Important design points:

- `.govard.yml` is the main writable config surface
- profile and local overrides are read-only from the CLI perspective
- runtime defaults are framework-aware and optionally version-aware
- remote definitions, hooks, and project extensions live inside Govard config or `.govard/*`

See [Configuration](configuration.md) for the complete contract.

## Framework Support

Discovery inspects project manifests and maps them to framework defaults for:

- web root
- PHP and Node versions
- database engine and version
- optional cache, search, queue, and Varnish services

Magento 2 receives the deepest integration, including auto-configuration, search/cache defaults, and dedicated debug routing.

## Desktop Architecture

The desktop app is intentionally smaller than the historical concept and focuses on operational workflows.

Current desktop architecture centers on:

- quick actions (start/stop/open, PHPMyAdmin, Xdebug toggle, health)
- workspace grouping environment list, quick actions, and onboarding
- logs with service filtering and live streaming
- shell launcher with project/service/user/shell selection
- modular frontend split across feature modules and bridge/state services

Implementation map:

- desktop entrypoint: `cmd/govard-desktop`
- Wails bindings: `internal/desktop`
- frontend shell: `desktop/frontend/index.html`
- bootstrap/events: `desktop/frontend/main.js`
- backend bridge: `desktop/frontend/services/bridge.js`
- state: `desktop/frontend/state/store.js`
- feature modules: `desktop/frontend/modules/`

## Project Layout

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
└── docs/
```

## Extension Points

Common extension points:

- add framework detection in `internal/engine/discovery.go`
- extend runtime selection in the profile/config engine
- add blueprint fragments or compose template logic
- add project extensions in `.govard/commands`, `.govard/hooks`, or `.govard/docker-compose.override.yml`

## Release Shape

Release outputs currently include:

- platform archives for CLI
- Linux `.deb`
- macOS `.pkg`
- `checksums.txt`

`govard self-update` resolves the target release, downloads the platform artifact, verifies SHA-256, and replaces installed binaries atomically.

## Related Docs

- [Commands](commands.md)
- [Configuration](configuration.md)
- [Desktop](desktop.md)
- [Contributing](contributing.md)
