//go:build realenv
// +build realenv

// Package realenv provides integration tests for Govard that run against
// real Docker environments with 3 Magento 2 instances.
//
// These tests validate govard commands (bootstrap, sync, db, remote) in a
// realistic multi-environment setup with actual SSH connections and database
// operations.
//
// Usage:
//
//  1. Setup the environment:
//     cd tests/integration/realenv && ./setup-three-env.sh
//
//  2. Run all real environment tests:
//     go test -tags realenv ./tests/integration/realenv/... -v
//
//  3. Run specific test:
//     go test -tags realenv ./tests/integration/realenv/... -run TestBootstrapClone -v
//
//  4. Cleanup:
//     cd tests/integration/realenv && ./setup-three-env.sh cleanup
//
// Environment Structure:
//
//	LOCAL Environment (Workstation):
//	  - MySQL: localhost:3306
//	  - SSH: localhost:9022
//	  - Role: The project where govard commands are run
//	  - Has remotes: dev (9023), staging (9024)
//
//	DEV Environment (Remote Target):
//	  - MySQL: localhost:3307
//	  - SSH: localhost:9023
//	  - Role: Remote target for sync/clone operations
//	  - No remotes configured (it's a target, not a workstation)
//
//	STAGING Environment (Remote Target):
//	  - MySQL: localhost:3308
//	  - SSH: localhost:9024
//	  - Role: Remote target for sync/clone operations
//	  - No remotes configured (it's a target, not a workstation)
//
// Requirements:
//   - Docker 20.10+ and Docker Compose
//   - Go 1.24+
//   - OpenSSH client
//   - 8GB+ RAM (16GB recommended)
package realenv
