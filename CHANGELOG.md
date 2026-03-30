# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.37.3] - 2026-03-31

### 🐛 Bug Fixes

- **Mailpit Communication**: Added the mandatory `-t` flag to `sendmail_path` in the standard PHP configuration. This resolves an `InvalidArgumentException` in Symfony Mailer (used by Magento, Shopware, Laravel, etc.) and ensures reliable email delivery to Mailpit.

## [1.37.2] - 2026-03-30

### 🛠 Improvements

- **Release Automation**: Significantly improved the automated changelog generation in `.goreleaser.yml` to eliminate redundant bullet points and improve commit categorization.
- **Improved Grouping**: Commits using the `refactor:` prefix are now correctly grouped under the **Improvements** section.
- **Documentation Visibility**: Re-enabled the **Documentation** section in release notes by refining the global exclusion filters.
- **Changelog Hygiene**: Cleaned up the `chore:` exclusion list to only target release-specific commits.

## [1.37.1] - 2026-03-30

### 🐛 Bug Fixes

- **Nested Vendor Sync**: Resolved a bug where a nested `vendor/vendor` directory was created when fallback synchronization was triggered during project bootstrap.
- **Improved Directory Detection**: Significantly enhanced the `sync` command's directory detection logic to correctly handle `--path` arguments even if trailing slashes are omitted.

### 🔄 Refactors

- **Testing Infrastructure**: Exported internal bootstrap and synchronization types to facilitate robust unit testing across project packages.

## [1.37.0] - 2026-03-30

### ✨ New Features

- **Automated Upgrade Pipelines**: Introduced dedicated upgrade workflows for Magento, WordPress, Laravel, and Symfony, enabling seamless version transitions within the environment.
- **Project Sync Status Tracking**: Implemented persistent tracking for synchronization status, providing better visibility into out-of-date remote data.
- **Improved Project Resolution**: Refactored internal project path resolution logic to handle complex symlinked and non-standard directory structures more reliably.
- **Composer Cache Configuration**: Added the `COMPOSER_CACHE_DIR` environment variable to all base configurations, ensuring optimized dependency persistence across container rebuilds.

### 🛠 Improvements

- **Blueprint Lifecycle**: Incremented internal `BlueprintVersion` to 1.27, triggering automatic environment re-renders to ensure all projects receive the latest configuration optimizations.

## [1.36.0] - 2026-03-29

### ✨ New Features

- **Desktop Auto-Update Notifications**: Implemented a comprehensive notification system for the desktop application to alert users when a new version is available.
- **Project Deletion**: Added the ability to delete projects directly from the Desktop UI with safety confirmation and backend cleanup.
- **Visual Sync Progress**: Introduced a dedicated UI module for remote synchronization with real-time progress bars and a live terminal console.
- **OS Terminal Support**: Integrated support for launching native OS terminals for service shells and remote operations.
- **Git Branch Visibility**: The dashboard now displays the active Git branch for each project in the environment details.
- **Desktop Doctor**: Added a `doctor` command to the Desktop app for automated troubleshooting and health checks.

### 🛠 Improvements

- **Project Selection Flow**: Refined the onboarding and project selection experience for better clarity and speed.
- **Sync Log Throttling**: Implemented event throttling for synchronization logs to improve UI responsiveness during high-volume data transfers.
- **Installer Compatibility**: Enhanced the system installer to support a wider range of Linux distributions and environment configurations.
- **Production Lock ID**: Switched the Desktop application's single-instance lock to a stable production identifier.

### 🔄 Refactors

- **Remote Configuration Management**: Implemented a new `RemoteConfigMap` type for deterministic YAML marshaling and priority-based sorting of remote environments.
- **UI Style Consolidation**: Removed redundant Tailwind CSS base styles and reset configurations, shifting to a more flexible and lightweight vanilla CSS architecture.
- **Project Name Normalization**: Standardized internal project naming conventions to ensure consistency across CLI and Desktop modules.

## [1.35.0] - 2026-03-28

### ✨ New Features

- **Real-time Desktop Logs**: Re-engineered the desktop log streaming engine to support carriage returns (`\r`). This enables real-time progress bar animations and spinner updates in the Desktop UI, providing immediate visual feedback for long-running synchronization tasks.
- **Enhanced Database Probing**: Extended the remote environment probe to support `MYSQL_` prefixed environment variables in `.env` files. This improves compatibility with Symfony, ensuring seamless database credentials resolution.

### 🛠 Improvements

- **Optimized Environment Actions**: The Desktop dashboard now automatically includes `--force-recreate` and `--remove-orphans` flags for all "Start" and "Restart" actions, ensuring a clean and predictable environment state.
- **Streamlined Sync Workflow**: Defaulted all automated "Pull" and "Sync" actions to use the `--yes` flag to bypass interactive prompts in background tasks. Removed the redundant "Assume Yes" option from the UI modals for a cleaner experience.

## [1.34.1] - 2026-03-28

### 🐛 Bug Fixes

- **Desktop Auto-Restart**: Fixed an issue where the desktop application failed to restart seamlessly after an update on Linux due to the `SingleInstanceLock`. The app now correctly yields the lock and relaunches via `gtk-launch` to preserve dock integration.
- **Update UI Alignment**: Fixed the flex alignment of the "Installing..." button in the desktop settings pane.

## [1.34.0] - 2026-03-28

### ✨ New Features

- **Store Domain Management**: Enhanced the `domain list` command with sorted tables and clear distinction between primary and extra domains for multi-store Magento environments.
- **Improved Onboarding UX**: Added explicit framework version selection (e.g., Magento 2.4.7, Laravel 11) to the project onboarding flow, ensuring the stack is correctly tuned from the start.

### 🛠 Improvements

- **Desktop Stability**: Implemented automatic panic recovery for desktop bridge proxies, preventing the application from crashing due to unexpected backend errors.
- **Image Fallback Engineering**: Refactored the local image fallback logic into the core engine, improving the reliability of environment startups in offline or air-gapped scenarios.

### 🔄 Refactors

- **Unified Synchronization Options**: Consolidated the legacy `sanitize` and `excludeLogs` options into a single, high-performance `--no-noise` flag for both CLI and Desktop sync operations. This simplifies the interface while providing more robust data protection and smaller transfer sizes.


## [1.33.0] - 2026-03-27

### ✨ New Features

- **Smart Data Synchronization**: Introduced the `--no-noise` flag for `bootstrap` and `sync` commands. It automatically excludes ephemeral data (caches, logs, sessions, and media thumbnails) for Magento, Laravel, and WordPress, drastically reducing transfer volume.
- **Centralized Credential Management**: Transitioned to using the global `~/.composer/auth.json` on the host, eliminating project-level `auth.json` files and stopping automatic `.gitignore` modifications for better security and hygiene.
- **Advanced Debugging**: Integrated Xdebug routing for Apache environments and added Varnish bypass logic for active debug sessions.

### 🛠 Improvements

- **Database Stream Compression**: Implemented automatic `gzip` streaming for all remote database transfers. This reduces network bandwidth usage by 80-90% during synchronization projects.
- **High-Performance SQL Sanitization**: Refactored the SQL sanitization engine using fast-path string matching, significantly reducing CPU overhead and memory usage during database imports.
- **SSH Transfer Tuning**: Standardized on the `aes128-ctr` cipher for remote operations and disabled redundant SSH-level compression to maximize throughput on high-speed networks.
- **Intelligent Progress Tracking**: Enhanced database import progress bars to distinguish between compressed file sizes and logical database volume for more accurate ETAs.

### 🐛 Bug Fixes

- **Stream Integrity**: Resolved edge cases involving data corruption during piped database dumps by ensuring atomic stream handling and proper gzip detection.

## [1.32.0] - 2026-03-27

### ✨ New Features

- **Automated Remote Selection**: `bootstrap` and `sync` commands now automatically prioritize `staging` then `dev` environments if not specified.
- **Blueprint Inspection**: New `blueprint` command for enhanced environment fingerprinting and template review.
- **Lock UX Enhancements**: Added `lock diff` command and `--update-lock` flag to `env up` for easier dependency tracking.
- **Compose Hygiene**: Introduced background cleanup for stale Docker Compose files and a new `env cleanup` command to manage directory saturation.

### 🛠 Improvements

- **Intelligent Update Notifier**: Refined update checks to suppress redundant warnings for development and pre-release builds.
- **Doctor Diagnostics**: Enhanced `govard doctor` with new checks for Docker Compose storage health.

### 🐛 Bug Fixes

- **CLI Stability**: Resolved various minor edge cases in command execution and internal configuration layering.

## [1.31.0] - 2026-03-27

### ✨ New Features

- **Comprehensive Web Server Support**: Introduced new Nginx and Apache templates with improved asset management and framework bootstrapping.
- **OpenMage Support**: Added dedicated support for the OpenMage framework and adjusted Magento 1 media paths.
- **Magento Cron Support**: Added `cronie` and `crond` to the PHP container for automated cron task execution.

### 🛠 Improvements

- **Nginx Backend Resolution**: Standardized PHP backend resolution and enhanced Varnish support across all web server modes.
- **Dashboard UI Refinement**: Modernized the desktop dashboard environment list using CSS Grid and optimized color contrast for status indicators.
- **HTTPS & TLS Enhancements**: Improved HTTPS detection for Magento 1 and added configurable TLS policies for local development domains.

### 🐛 Bug Fixes

- **Header Cleanup**: Removed redundant `X-Forwarded-Proto` directives for cleaner protocol detection.
- **Test Stability**: Enhanced integration test coverage and adjusted assertions for blueprint content verification.

## [1.30.0] - 2026-03-26

### ✨ New Features

- **Command Aliases**: Introduced command aliases and shortcuts for improved CLI usability, allowing users to run `up`, `down`, `ps`, and `logs` as top-level shortcuts.
- **Enhanced Service Management**: Added new flags and Portainer integration to the `svc` command for better observability of global services.

### 🛠 Improvements

- **DB Command Refactor**: Refactored `db` command flags for consistency. The `--no-pii` shorthand is now `-P`, and `--sanitize` is introduced as a `-S` alias.
- **Improved Help Output**: Enhanced help output by dynamically filtering compose flags and adding Govard-specific options.
- **Documentation Restructuring**: Restructured and consolidated documentation into broader topics for better navigation and clarity.
- **Framework Detection & Config**: Simplified framework detection logic and added profile-based config loading for more flexible environment setups.

## [1.29.0] - 2026-03-25

### ✨ New Features

- **Interactivity Control**: Introduced the `-y, --yes` flag for `bootstrap` and `sync` commands. In non-interactive environments (CI/CD), these commands now require the `--yes` flag to proceed, preventing unexpected hangs.
- **Improved Headers**: Redesigned all CLI headers with a bold, blue-boxed style and standardized vertical margins for better focus and readability.
- **Elasticsearch Alias**: Added the `opensearch` hostname alias to the `elasticsearch` service in blueprints to ensure backward compatibility for projects expecting the OpenSearch hostname.

### 🛠 Improvements

- **Bootstrap Flow**: Reordered the bootstrap execution flow to display the full synchronization plan review *before* starting environment containers, giving users a clear overview of the intended operations.
- **Sync Plan Visibility**: The synchronization plan review now explicitly lists endpoints, scopes (files, media, db), risk assessment, and detailed rsync/shell steps.
- **Sync Progress UI**: Integrated a new live-scrolling 10-line window for `rsync` progress during file and media synchronization, providing better real-time feedback without overwhelming the terminal.
- **Single File Sync**: Improved `--path` handling in `sync` to correctly distinguish between single files and directories, ensuring precise `rsync` behavior.
- **Non-Interactive Self-Update**: The `self-update` command now intelligently skips heavy system dependency checks when running in CI/non-interactive mode.

### 🐛 Bug Fixes

- **Integration Test Stability**: Resolved multiple test hangs in CI by enforcing the `--yes` flag and disabling terminal color (`NO_COLOR=1`) for consistent assertion matching.
- **Varnish CI Path Fix**: Corrected Varnish VCL path references in integration tests to align with the decentralized engine storage architecture.
- **Rsync Path Sanitization**: Correctly handles trailing slashes in sync operations to prevent duplicated subdirectories when syncing specific paths.

## [1.28.1] - 2026-03-24

### Fixed

- **SSH Key Mounting**: Fixed an issue where SSH keys from the host were missing in the container when a safe SSH configuration copy was being used.
- **Architectural Update**: Incremented internal blueprint version to ensure all project environments automatically receive the SSH mounting fix during the next `env up`.

## [1.28.0] - 2026-03-24

### Added

- **CLI Flag `--force-recreate`**: Added to `govard env up` and `govard svc up` commands, allowing users to force container recreation.
- **Apache Hybrid Mode Improvements**: Configured essential Apache modules (`mod_alias`, `mod_remoteip`, `mod_expires`, etc.) and added `X-Backend-Server: apache` header for easier debugging of hybrid environments.

### Improved

- **Global Nginx Proxy Support**: Updated all Nginx framework templates to include `fastcgi_param HTTPS 'on';`, ensuring correct HTTPS detection for all project types when running behind Govard's reverse proxy.
- **Service Management**: `govard svc up` now correctly parses additional arguments and always runs in detached mode (`-d`).
- **Magento 1 Bootstrap**: `RunMagento1SetConfigSQL` now sets `web/secure/offloader_header` to `X-Forwarded-Proto`, ensuring Magento 1 trusts the forwarded protocol header from Govard's proxy — a prerequisite for correct HTTPS detection.

## [1.27.0] - 2026-03-24

### Added

- **phpMyAdmin Database Access**: Database containers are now connected to the `govard-proxy` network, enabling phpMyAdmin (`pma.govard.test`) to directly reach all running project databases without additional configuration.
- **Magento 1 HTTPS Fix**: The `magento1.conf` Nginx template now includes `fastcgi_param HTTPS 'on';` so Magento 1 correctly identifies HTTPS requests behind Govard's Caddy reverse proxy, eliminating infinite redirect loops.
- **Composer Version Config**: Added `composer_version` field to `.govard.yml` stack config, allowing projects to pin a specific Composer version (e.g. `2.2`, `2`, `latest`).

### Improved

- **Auto Composer Downgrade**: Govard automatically selects Composer 2.2 LTS for projects running PHP < 7.2.5 when `composer_version` is not explicitly set, preventing plugin-blocking errors on legacy stacks.
- **DB Import Validation**: Improved `db import` command to correctly validate flag combinations (`--drop`, `--local`) and restrict incompatible options (`--no-noise`, `--no-pii`).

### Fixed

- **Self-Update CI Safety**: Avoided system dependency checks in non-interactive/CI environments to prevent test hangs and improve `TestSelfUpdateAutoConfirmViaEnv` reliability.

## [1.26.0] - 2026-03-24

### Added

- **Enhanced Magento 1 Support**: Dedicated bootstrap logic for Magento 1 / OpenMage with automated `local.xml` generation, admin user creation, and base URL configuration. Added support for `--no-noise` and `--no-pii` flags in Magento 1 database dumps.
- **Improved Mailpit Persistence**: Added a dedicated `mail_data` volume to the global Mailpit service and configured the `mail` network alias for reliable internal mail routing.

### Changed

- **Blueprint Architecture Standardizing**: Refactored blueprints to centralize shared service definitions (Redis, Varnish, RabbitMQ, etc.) into unified includes for better consistency.
- **Blueprint Version Update**: Updated internal blueprints to version 2 (V2 architecture) with optimized networking.

### Improved

- **Caddy Stability**: Enabled `--resume` flag for the global Caddy proxy, ensuring routes persist across container or host restarts.
- **Self-update Robustness**: Improved dependency installation checks and error handling in the `self-update` command.

### Fixed

- **Network Isolation**: Fully isolated PHP and Database networks from the global `govard-proxy` to resolve hostname conflicts and improve security.
- **Mail Connectivity**: Fixed mail routing issues by using `host-gateway` for the `mail` alias in all project environments.

## [1.25.0] - 2026-03-23

### Added

- **Optimized Self-Update**: The `govard self-update` command now includes automated system dependency checks (Linux-specific), post-update global service refreshes, and SSL trust verification, ensuring a consistent state after upgrades.
- **Debian Package Priority**: On Ubuntu/Debian, `self-update` now prioritizes using `.deb` packages for better dependency management and parity with the installation script.

### Fixed

- **Mail Server Connectivity**: Resolved an issue where PHP containers were isolated from the global Mailpit service, causing mail sending failures in projects like Magento 2.
- **Database Networking**: Added `govard-proxy` network to database services in blueprints for consistent phpMyAdmin access.
- **Magento 1 Credentials**: Restored the default database password in Magento 1 blueprints.

## [1.24.0] - 2026-03-23

### Added

- **Global HTTP Redirect**: Implemented a global 308 Permanent Redirect from port 80 to 443 in the Caddy proxy. All `.test` and `.govard.test` domains (including Mailpit and phpMyAdmin) now force HTTPS by default.
- **Zero-Config Installer**:
    - Automatic detection and installation of `libnss3-tools` (certutil) and `libwebkit2gtk-4.1-0` on Linux.
    - Post-installation hooks: Automatically starts global services (`svc up`) and configures SSL trust (`doctor trust`) for a "Green Lock" experience immediately after install.
    - Pipe compatibility: Optimized `install.sh` for `curl | bash` execution with `/dev/tty` redirection for interactive prompts.
- **WordPress Remote Support**: Added dedicated SSH-based database credential probing and site URL auto-correction for WordPress projects.
- **Framework Detection**: Added WordPress to the default list of auto-detected frameworks.

### Improved

- **Bootstrap Hygiene**: The `bootstrap` command now defaults to `--remove-orphans`, ensuring a clean environment state without requiring manual flags.
- **Package Integrity**: Elevated `libwebkit2gtk-4.1-0` and `libnss3-tools` to mandatory dependencies in the `.deb` package for seamless offline installation.
- **phpMyAdmin Reliability**: Switched to permanent directory-based mounting for the project registry, resolving "stale data" issues in phpMyAdmin.
- **Remote Shell Robustness**: Improved remote command execution with `bash -l` login shell invocation and `sh` fallback.
- **Database Tooling**: Added Gzip compression for `db dump` output and removed environment restrictions for the `--local` flag.

### Fixed

- **Installer Path Resolution**: Fixed a `BASH_SOURCE` edge case in `install.sh` when executing via pipes.
- **PHPMyAdmin Visibility**: Resolved inode-related search failures in phpMyAdmin by using more stable Docker volume configurations.

## [1.23.0] - 2026-03-23

### Added

- **Independent Monitoring Flags**: Separated `--no-noise` (`-N`) and `--no-pii` (`-S`) flags. They are now independent and can be used individually or together to fine-tune database synchronization and dumps.
- **Canonical Remote Name Display**: The CLI now consistently resolves and displays the canonical remote name (from `remotes` config) in all output messages, even when using aliases or environment names.

### Improved

- **Bootstrap & Sync Visibility**: Added clear, immediate `INFO` messages at the start of `bootstrap` and `sync` operations to provide instant feedback on the source, destination, and scope of the action.
- **Remote DB Dump Feedback**: Database dump operations now explicitly include the target remote file path in the success message for better traceability.

### Fixed

- **Remote Path Expansion**: Fixed an issue where paths starting with `~/` were not expanded on remote servers due to shell quoting. Introduced a `quoteRemotePath` helper to safely handle home-relative paths.
- **PHPMyAdmin Visibility**: Resolved a race condition where `pma.govard.test` would not show project databases after a fresh `bootstrap` or `env up`. Implemented an explicit active project refresh in the verification stage.
- **Remote Identity Resolution**: Refactored `ensureRemoteKnown` to consistently resolve and return the confirmed remote name across all CLI commands.

## [1.22.1] - 2026-03-21

### Added

- **Local Image Build Fallback**: Govard now automatically attempts to build missing Docker images from embedded blueprints if pulling fails.
- **Dependency-Aware Image Building**: Implemented a resolution algorithm for local image builds that correctly handles parent-child image dependencies.

### Changed

- **Command Refactoring**: Centralized Docker Compose execution logic in the `engine` package for better maintainability and consistency.
- **Standardized CLI Signature**: Unified `RunE` signatures and context handling across various command implementations.
- **Bootstrap Stability**: Improved Magento bootstrap reliability by clearing generated code and simplifying autoloader generation.
- **Log Management**: Refined log tailing and follow logic in `env logs` commands for more predictable behavior.

### Fixed

- **Composer Workflow**: Removed the `-o` (optimize) flag from `composer dump-autoload` in development flows to align with development best practices and resolve test failures.
- **Documentation Paths**: Corrected documentation links and existence checks in test suites.

## [1.22.0] - 2026-03-20

### Added

- **Snapshot Compression**: Database snapshots are now automatically compressed using Gzip, reducing disk usage by 70-90%.
- **Enhanced Tunnel Management**:
  - Added `tunnel stop` and `tunnel status` commands.
  - Automatic base URL update/revert for all frameworks when a tunnel starts or stops.
- **Database Operations**:
  - New `db top` command for real-time database process monitoring.
  - Real-time progress bar for database imports (`db import`, `bootstrap`) and `sync` operations.
- **Project Testing Integration**:
  - New `test` command with subcommands for `phpunit`, `phpstan`, `mftf`, and `integration` tests.
  - Standardized container execution and user resolution across all test types.
- **Log Service Filtering**: `govard env logs` now supports an optional `<service>` argument to stream logs from a specific service only.
- **Capability Scopes**: Added `cache` capability for remote environments to explicitly allow/deny Redis operations.

### Changed

- **Redis Command Refactoring**: Migrated `redis` command to a structured subcommand pattern: `redis cli`, `redis flush`, and `redis info`.
  - Added support for both Redis and Valkey providers.
  - Added remote execution support for Redis commands via SSH.
- **Debug Command Refactoring**: Migrated `debug` command to a structured subcommand pattern: `debug on`, `debug off`, `debug status`, and `debug shell` (inspired by Warden).

### Fixed

- **Proxy Networking**: Remedied an issue where framework-specific compose overrides (such as `magento2/services.yml`) accidentally detached the `web` service from the `govard-proxy` network, resulting in 502 Bad Gateway errors.
- **Undefined References**: Fixed multiple lint errors and unreachable code in `bootstrap`, `debug`, and `open` commands by exporting and unifying container execution helpers.

## [1.21.1] - 2026-03-20

### Fixed

- **Visual Assets**: Updated the desktop application icon to use the new branding logo in the Linux package.

## [1.21.0] - 2026-03-20

### Added

- **Sync Data Obfuscation Flags**: Implemented `--no-noise` (`-N`) and `--no-pii` (`-S`) flags for `govard sync` and `govard bootstrap`.
  - `--no-noise`: Excludes ephemeral/noise tables (cron schedules, caches, sessions, logs) from `mysqldump`.
  - `--no-pii`: Excludes sensitive/PII tables (customers, orders, passwords, etc.) and implies `--no-noise`.
  - Supports framework-specific table lists: Magento 2, Laravel, and WordPress.
- **Sync Shortcut Flags**: Added `-s` / `-d` as short aliases for `--source` / `--destination` in `govard sync`.
- **Smart Remote Name Resolution**: `--source`, `--destination` (sync) and `--environment` (bootstrap) now resolve remote name aliases automatically (e.g. `-s dev` → matches a remote named `development`, or any remote whose name normalizes to `dev`).
- **SSH Agent Diagnostics**: `govard doctor` now checks SSH agent connectivity and reports the number of loaded keys with remediation guidance.
- **Secure SSH Config Mounting**: SSH config files with overly broad permissions are now copied to a safe temporary file with restricted permissions (`0600`) before mounting into containers, preventing SSH warnings in remote operations.
- **Magento Search Auto-fix**: Bootstrap now automatically configures the search engine host and port (`es`, `opensearch`) from container labels, and unblocks read-only Elasticsearch/OpenSearch indices to prevent post-install search failures.

### Changed

- **Sync Endpoint Display**: Removed redundant "environment" field from remote/sync endpoint summaries. Environment context is now derived from the remote name itself.
- **Bootstrap Examples Updated**: Updated `-e` flag description in help text to clarify it accepts remote name aliases.

### Fixed

- **GoReleaser Changelog Grouping**: Refined changelog group configuration in `.goreleaser.yml` for cleaner release notes.

## [1.20.0] - 2026-03-19


### Added

- **Enhanced Database Dump Flags**: Replaced `--exclude-sensitive-data` with `--no-noise` (`-N`) and `--no-pii` (`-S`) flags for `govard db dump`.
  - `--no-noise`: Excludes ephemeral tables (cron, cache, session, logs, etc.).
  - `--no-pii`: Excludes PII/sensitive tables (customers, orders, etc.) and implies `--no-noise`.
- **Locale Auto-detection**: Implemented automatic locale detection for `govard deploy` in Magento 2 projects when the `--locales` flag is omitted.
- **Magento 1 / OpenMage Improvements**:
  - Secure bootstrap with randomly generated 32-hex crypt keys in `local.xml`.
  - Automated post-clone setup (base URL configuration and admin user creation).
  - Remote database credential probing from `local.xml` via SSH.
- **Desktop Application Enhancements**:
  - New theme support with **Dark Mode**.
  - Added "Run in Background" preference and standard Quit functionality.
  - Refactored service action buttons to be icon-only with detailed tooltips.
- **Comprehensive Documentation**: Added dedicated documentation files for all supported frameworks (Laravel, Symfony, Drupal, WordPress, Shopware, CakePHP, Next.js).
- **SVG Logo**: Updated project logo to a modern SVG format.

### Changed

- **Go Version Upgrade**: Updated project Go requirement to **1.25.0**.
- **Linter Evolution**: Upgraded `golangci-lint` to **v2.11.3** and modernized configuration.
- **Refined Documentation**: Major updates to `README.md` and CLI command references.

### Improved

- **Database Reliability**: Enhanced local DB operations with larger max packet size and foreign key check safeguards.
- **Remote Stability**: Improved SSH connection handling and remote probe accuracy.

### Fixed

- **Elasticsearch Safety**: Fixed "read-only" index block issues during post-install operations.
- **Code Quality**: Fixed various linter warnings and improved code structure across the bootstrap engine.

## [1.19.0] - 2026-03-04

### Added

- **Environment Pull**: Introduced `govard env pull` command to pull Docker images for the current environment.

### Changed

- **Database Proxy**: Refactored `open db` and generic database access to support a shared containerized proxy, improving connectivity and client-url resolution.

### Testing

- **Desktop & Integration**: Added comprehensive integration tests for environment compose flows, database client URL resolution, and PMA proxy configurations.
- **Frontend**: Expanded unit tests for dashboard actions and global services modules.

## [1.18.0] - 2026-03-04

### Fixed

- **Updater**: Unified desktop and self-update binary handling and hardened permissions.
- **Update Targets**: Refined desktop update target resolution to prefer sibling binaries and support explicit environment overrides.

## [1.17.0] - 2026-03-04

### Added

- **Service Start/Restart Flow**: Reworked global service start/restart flow to include proxy readiness, route registration, and enhanced UI feedback with message summarization.

### Improved

- **UI Feedback**: Improved user feedback during service lifecycle operations with summarized messages and real-time status updates.
- **Service Stability**: Enhanced reliability of service startup by ensuring proxy readiness before route registration.

### Testing

- **Backend Tests**: Added comprehensive integration tests for desktop global services and service startup logic.
- **Frontend Tests**: Added core tests for global services frontend module.
- **Mocking**: Introduced additional test helpers for mocking backend services in tests.

## [1.16.1] - 2026-03-04

### Added

- **Update Formatting**: Unified update message formatting across the desktop app.

### Improved

- **Notifier UX**: Synchronized update notifier visibility with the settings drawer to avoid UI overlapping.

### Fixed

- **Redundant Config**: Removed unnecessary Darwin build configuration from GoReleaser.

## [1.16.0] - 2026-03-04

### Added

- **Self-Update Logic**: Implemented desktop application self-update functionality.
- **Log Management**: Introduced desktop log export and management features.
- **Caddy Stability**: Implemented Caddy command retry logic for better proxy reliability.

### Improved

- **UI Responsiveness**: Improved desktop UI layout responsiveness and log display for better readability.
- **Diagnostics**: Added additional test helpers and integration tests for desktop services and update logic.

## [1.15.0] - 2026-03-04

### Added

- **Git Project Onboarding**: Introduced the ability to clone Git repositories directly during the project onboarding process.
- **Onboarding UI Enhancements**: Added support for Git URL, branch selection, and progress tracking for repository cloning.
- **Terminal Integration Improvements**: Enhanced terminal output and progress monitoring for long-running onboarding operations.

### Improved

- **Testing Infrastructure**: Expanded integration and frontend tests to cover Git cloning and complex onboarding flows.
- **Desktop UI**: Refined onboarding styles and logic for a smoother repository setup experience.

### Fixed

- **Remotes Cleaning**: Fixed issues where some remote environment references were not correctly cleaned up in the UI.

## [1.14.1] - 2026-03-04

### Improved

- **Linux App Icons**: Optimized SVG logo size and distributed multiple hicolor icon sizes (16x16 up to 256x256) in the Debian package to ensure correct visual weight and visibility in the Ubuntu launcher.

## [1.14.0] - 2026-03-03

### Added

- **Desktop UI Revamp**: Major overhaul of global services controls and header UX for a more premium experience.
- **Modular Controllers**: Introduced `global-services.js` for better state management and dynamic service status handling.
- **Enhanced Bridge Proxies**: Improved backend-to-frontend communication for workspace-wide services.

## [1.13.0] - 2026-03-03

### Added

- **Local Image Fallback**: Introduced the `--fallback-local-build` flag for `up`, `svc up`, and `svc restart` commands. This allows Govard to automatically build missing ddtcorex/govard images locally from embedded blueprints if they cannot be pulled from Docker Hub, ensuring environments can start even without internet access or registry availability.

### Improved

- **`open db` Command UX**: Updated the default behavior of `govard open db` to launch phpMyAdmin (PMA) in the browser for a more immediate visual experience. A new `--client` flag was added to explicitly launch external database client protocols (e.g., `mysql://`).

### Quality & Testing

- **Additional Test Gates**: Added comprehensive tests for image reference parsing, local build spec resolution, and command-level flag existence to maintain high stability.

## [1.12.0] - 2026-03-03

### Added

- **Sync Presets Remote Capabilities**: Added support for remote synchronization presets, enhancing cross-environment workflow flexibility.
- **SSL Auto-Trust**: Automated root CA trust for `svc` lifecycle commands, simplifying SSL management on Linux systems.

### Improved

- **Log Stream Sanitization**: Significantly improved terminal output reliability by stripping ANSI escape codes, control characters, and invalid UTF-8 from log streams.
- **Desktop Stability**: Refined "open" actions and default configurations for a more consistent desktop experience.
- **Sync Options**: Refined synchronization options for better control over remote environment updates.

### Fixed

- **ANSI Fragment Cleaning**: Resolved issues with trailing or orphan ANSI fragments disrupting stream output.

### Quality & Documentation

- **Test Coverage**: Expanded command runtime coverage and refreshed the integration test suite.
- **Project Documentation**: Updated README to highlight remote management features and core differentiators.

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
