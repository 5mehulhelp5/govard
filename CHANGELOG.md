# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
