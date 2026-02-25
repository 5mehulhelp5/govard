# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.0] - 2026-02-26

### Added

- **Composer Cache**: Share the host's Composer cache directory with PHP containers to vastly speed up dependency installation.
- **Log Tailing**: Added `--tail` flag to `govard logs` and `govard svc logs` for better control over log output.
- **Snapshot Management**: New `snapshot export` and `snapshot delete` subcommands.
- **CI Support**: Added `--no-tty` flag to `govard shell` for non-interactive environments.
- **Docker Images**: Use a non-alpine base image for `varnish:6.0` to resolve libc compatibility issues.
- **Route Revival**: Automatically re-registers domains for all running project containers when global services (`svc`) start or restart.

### Changed

- **Bootstrap Command**: `--clone` is now disabled by default. Running `govard bootstrap` will import DB/media and start containers without downloading the remote source. Pass `--clone` to override.
- **Docker Prefix**: Updated default Docker image prefix to `ddtcorex/govard-`.
- **PHP Redis**: Pinned Redis extension version specifically for PHP 7.1, 7.2, & 7.3 to resolve compilation failures.

### Fixed

- **Docker Error Messages**: Made Docker daemon connection errors significantly clearer and easier to diagnose.
- **Proxy Stability**: Improved `govard svc up` to handle stopped proxy containers and provide better port conflict diagnostics.
- **Configuration**: Support for dot notation (nested keys) in `govard config set` (e.g., `stack.php_version`).
- **SSL Trust**: Fixed diagnostics and instructions for Linux system trust store.

## [1.5.0] - 2026-02-25

### Added

- **Database Commands**: Added `db query` and `db info` commands for easier direct database interaction.
- **Enhanced Integration Tests**: Comprehensive realenv integration tests for bootstrap, sync, db, and open commands.
- **Improved Warden Migration**: Support for modern remotes and stack versions.

### Changed

- **Docker Organization**: Introduced `DOCKER_ORG` variable for better flexibility in image naming.
- **Help Documentation**: Major refactor of help commands with detailed examples and case studies.

## [1.4.0] - 2026-02-25

### Added

- **Migration Suite**: New automatic migration from DDEV and Warden environments during `govard init`.
- **Embedded Blueprints**: Blueprints are now embedded directly into the binary, enabling standalone installation without external dependencies.

### Changed

- **Rename Configuration File**: Standardized on `.govard.yml` (previously `govard.yml`) for consistent hidden file convention.

### Fixed

- **Docker Build Stability**: Enhanced PHP Dockerfiles with version-specific capping for PECL extensions (`redis`, `imagick`, `amqp`) and standard shell compatibility for older PHP versions (7.1, 7.2, 7.3).
- **Integration Test Reliability**: Synchronized test suite with embedded blueprints and updated file naming expectations.

## [1.3.0] - 2026-02-24

### Added

- **Local Snapshots**: Introduced the `snapshot` command to create, list, and restore database and media snapshots for rapid environment switching.
- **Public Tunnels**: New `tunnel` command with Cloudflare Tunnel integration to securely expose local projects to the internet.
- **Project Extensions**: Support for project-specific custom commands and extensions in `.govard/commands`.
- **Enhanced SSL Trust**: Automated root CA management and browser trust via `govard doctor trust`.
- **Deployment & Sync**: Improved `deploy` and `sync` commands for better remote environment orchestration.
- **Observability**: New events tracking for better CLI audit and telemetry.

### Changed

- **Proxy Architecture**: Significant refactor of Caddy route management for better performance and flexibility.
- **Framework Discovery**: Refined detection logic and runtime profiles for Magento 2, Laravel, Symfony, and WordPress.
- **CLI Robustness**: Added comprehensive input validation and clearer error messaging across all commands.

### Fixed

- Improved handling of `sudo` requirements for certificate installation.
- Fixed numerous edge cases in Docker Compose template rendering and service dependency management.
- Corrected various stability issues in remote environment synchronization.

### Quality & Testing

- **Massive Test Suite Expansion**: Added over 50,000 lines of unit, integration, and frontend tests.
- **Automated Quality Gates**: Integrated comprehensive coverage analysis and enhanced CI pipelines.

## [1.2.0] - 2026-02-24

### Added

- **Centralized Docker Image Management**: Introduced custom Elasticsearch and OpenSearch Docker images optimized for Magento 2.
- **Extended Magento 2 Support**: Enhanced Magento directory setup and refactored database import functions.
- **`tool` Subcommand**: New command for runtime wrappers to execute framework-specific CLI tools within project containers.
- **Desktop App Transformation**:
  - Full UI redesign with updated styling and modular dashboard.
  - Added Log viewer, Remote management, and Onboarding modules.
  - Integrated Remote Shell and Sync-plan wiring.
  - Redesigned Toast notification system with message deduplication logic.

### Changed

- **Command Architecture**: Restructured command groups and refactored `config` and `bootstrap` for consistency.
- **Desktop Layout**: Migrated from embedded Mailpit to a dedicated project workspace layout.
- **Docker Efficiency**: Optimized container listing performance for port availability checks.

### Fixed

- Fixed YAML syntax error in Elasticsearch blueprint.
- Corrected test expectations and CLI completion defaults.

## [1.1.2] - 2026-02-20

- Patch release with minor stability fixes.

## [1.1.0] - 2026-02-20

- Added initial Desktop app framework support.
- Enhanced framework discovery for Laravel and Next.js.

## [1.0.0] - 2026-02-20

- Initial professional-grade release of Govard.
