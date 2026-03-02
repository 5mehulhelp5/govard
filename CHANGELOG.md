# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.11.0] - 2026-03-02

### Added

- **Embedded Frontend Assets**: The desktop application now embeds all frontend assets (CSS, JS, Fonts) directly into the binary, enabling standalone distribution and simplified packaging.
- **Terminal Integration**: Integrated Xterm.js with backend PTY support for a full interactive shell experience directly within the desktop application.
- **Debian Packaging Support**: Official support for generating `.deb` packages for Linux distributions, complete with application menu integration and icons.

### Improved

- **Asset Resolution**: Enhanced the path resolution logic to automatically fall back to embedded assets when running in production/installed mode.
- **Development Environment Setup**: Significant documentation updates for local toolchain installation and prerequisites.
- **Frontend Bundle Consolidation**: Merged third-party CSS (Inter, Material Symbols, Xterm) into a single optimized bundle for faster loading.

### Fixed

- **Production Asset Loading**: Fixed a critical issue where the desktop application failed to locate assets after being installed via package managers.


## [1.10.0] - 2026-03-02

### Added

- **Tailwind CSS Integration**: Fully migrated the frontend to Tailwind CSS for improved maintainability and modern aesthetics.
- **Shared Reference Architecture**: Implemented a shared `refs` object across all frontend controllers, ensuring resilient DOM binding and immediate UI updates for dynamically injected elements.
- **Tab Selection Persistence**: The desktop app now preserves the active tab (e.g., Logs & Shell) when switching between environments.

### Fixed

- **Terminal Mounting**: Resolved the "Terminal requires a parent element" error by standardizing controller initialization with live DOM references.
- **Log Service Filtering**: Corrected state management in `main.js` to ensure log output is correctly filtered for the selected service.
- **Process Management**: Improved development server stability by adding robust cleanup logic for redundant `wails` and `govard` instances.

### Improved

- **Testing Infrastructure**: Expanded integration test suite for environment selection and log retrieval.
- **Build Automation**: Enhanced release workflows for multi-platform distribution.

## [1.9.0] - 2026-03-01

### Added

- **Portainer Integration**: Portainer is now available as a new global service, integrated with the `svc` and `open` commands.
- **Streaming Toasts**: Introduced streaming toasts and enhanced desktop UI messages for improved clarity and user feedback.

### Changed

- **Remote Environments**: Removed direct remote connection management from the UI and backend, adding a plan-only bootstrap option instead.
- **System Improvements**: Various system improvements and technical debt reduction.

## [1.8.0] - 2026-02-28

### Added

- **Interactive SSH Sessions**: `govard open shell` now supports fully interactive terminal handoff for remote environments using `syscall.Exec`.
- **Image Pulling Support**: New `govard env pull` command and added `--pull` / `--remove-orphans` flags to `govard env up`.
- **Search Engine Health**: Automatic detection and resolution for Elasticsearch/OpenSearch "read-only" index blocks caused by low disk space.
- **Environment Scopes (Profiles)**: Run isolated environment variants with `--profile <name>`. Config layers merge as Base → Profile (`.govard.<profile>.yml`) → Local (`.govard.local.yml`). Each profile gets its own Docker Compose file and database volumes for full isolation.
- **Network Isolation Mode**: Set `isolated: true` in `features` to prevent containers from reaching the internet.
- **MFTF & Selenium Support**: Set `mftf: true` in `features` to auto-start a Selenium Standalone Chrome container.
- **Frontend LiveReload / Watcher**: Set `livereload: true` in `features` to start a dedicated Node.js watcher container.
- **`govard open mftf`**: New target to open the Selenium VNC viewer in-browser.

### Improved

- **Remote Environments**: Added support for deriving environment configurations from name and making protected status optional.
- **Embedded Terminal**: Consistently handle sync operations and interactive commands within the UI.

## [1.7.0] - 2026-02-27

### Added

- **dnsmasq Service**: Built-in `dnsmasq` service for automatic `.test` domain resolution.
- **Interactive Recipe Selection**: Prompt for recipe (framework) selection during `init` if detection fails.
- **Restructured CLI**: Organized environment commands under a unified `env` subcommand (`govard env up`, `govard env stop`, etc.).
- **Multi-domain Support**: Enhanced support for multiple domains per project.

### Changed

- **Standardized Terminology**: Consistent use of "recipe" and "framework" across CLI and documentation.
- **Bootstrap Logic**: Refined to auto-clone if no source is present and improved remote connectivity tests.
- **Flag Renames**:
  - Renamed `--version` to `--framework-version` in `bootstrap`.
  - Renamed `--framework` to `--recipe` in `profile`.

### Fixed

- **Bootstrap Remote Sync**: Fixed issues with remote synchronization when source already exists.
- **CI Reliability**: Fixed potential CI recursion issues in bootstrap flows.

### Improved

- **Makefile**: Added `fmt-check` for better code quality control.
- **Proxy Naming**: Standardized proxy container naming for better visibility.

## [1.6.0] - 2026-02-26

### Added

- **Composer Cache**: Share the host's Composer cache directory with PHP containers to vastly speed up dependency installation.
- **Log Tailing**: Added `--tail` flag to `govard env logs` and `govard svc logs` for better control over log output.
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
