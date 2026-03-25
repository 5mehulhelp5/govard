# Govard Documentation

Govard documentation is now intentionally flat: one topic, one canonical file, no audience tree to chase through.

## Start Here

- **[Getting Started](getting-started.md)**: install Govard, initialize a project, bring the stack up, and learn the daily workflow.
- **[Commands](commands.md)**: CLI reference, shortcuts, framework tool commands, desktop launch, diagnostics, and utilities.
- **[Configuration](configuration.md)**: config layering, ownership rules, profiles, remotes, and blueprint registry settings.
- **[Remotes and Sync](remotes-and-sync.md)**: remote setup, sync policies, audit logs, resumable transfers, and remote DB workflows.
- **[Frameworks](frameworks.md)**: support matrix, runtime defaults, version-aware overrides, and framework-specific notes.
- **[SSL and Domains](ssl-and-domains.md)**: `.test` DNS, local CA trust, HTTPS, extra domains, and troubleshooting.
- **[Desktop](desktop.md)**: Wails desktop surface, launch modes, live backend dev flow, and feature boundaries.
- **[Architecture](architecture.md)**: system layout, runtime modules, networking, desktop integration, and extension points.
- **[Contributing](contributing.md)**: toolchain, build/test workflow, fixture conventions, and contribution hygiene.

## Reading Order

1. Start with [Getting Started](getting-started.md).
2. Use [Commands](commands.md) for day-to-day CLI work.
3. Read [Configuration](configuration.md) before changing stack behavior.
4. Jump to [Frameworks](frameworks.md), [Remotes and Sync](remotes-and-sync.md), or [Desktop](desktop.md) only when needed.

## Scope

This folder is for maintained project documentation. Old per-command and per-framework pages were consolidated to reduce duplication and stale cross-links. If behavior changes, update the canonical top-level file for that topic instead of creating a new subtree.
