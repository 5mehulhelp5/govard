# Govard Real Environment Tests

This directory contains integration tests that run against real Docker environments with 3 Magento 2 instances (local, dev, staging). These tests validate govard commands in a realistic multi-environment setup.

## Overview

Unlike the existing integration tests that use runtime shims (mocked docker, ssh, rsync), these tests interact with actual Docker containers, real SSH connections, and genuine database operations.

### Test Environment Architecture

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   LOCAL ENV     │  │    DEV ENV      │  │  STAGING ENV    │
│                 │  │                 │  │                 │
│ • PHP-FPM       │  │ • PHP-FPM       │  │ • PHP-FPM       │
│ • MySQL :3306   │  │ • MySQL :3307   │  │ • MySQL :3308   │
│ • SSH :9022     │  │ • SSH :9023     │  │ • SSH :9024     │
│                 │  │                 │  │                 │
│ Recipe:         │  │ Recipe:         │  │ Recipe:         │
│ magento2        │  │ magento2        │  │ magento2        │
└─────────────────┘  └─────────────────┘  └─────────────────┘
         │                    │                    │
         └────────────────────┼────────────────────┘
                              │
                    Shared Docker Network
                    Shared SSH Keys Volume
```

## Prerequisites

- Docker 20.10+ and Docker Compose plugin
- Go 1.24+
- OpenSSH client (`ssh` command available)
- 8GB+ RAM (16GB recommended)
- Available ports:
  - 3306-3308 (MySQL)
  - 9022-9024 (SSH)

## Quick Start

### 1. Setup the Environment (One-time)

```bash
make test-realenv-setup
```

This will:

- Create SSH key pair for inter-container authentication
- Prepare Magento 2 fixtures for all 3 environments
- Start 9 Docker containers (3 PHP, 3 DB, 3 SSH)
- Generate .govard.yml configurations for each environment
- Verify SSH and database connectivity

### 2. Run All Real Environment Tests

```bash
make test-realenv
```

### 3. Run Specific Test Suites

```bash
# Bootstrap tests only
go test -tags realenv ./tests/integration/realenv/... -run TestBootstrap -v

# Sync tests only
go test -tags realenv ./tests/integration/realenv/... -run TestSync -v

# DB tests only
go test -tags realenv ./tests/integration/realenv/... -run TestDB -v

# Remote tests only
go test -tags realenv ./tests/integration/realenv/... -run TestRemote -v
```

### 4. Cleanup

```bash
make test-realenv-clean
```

Or full cycle (setup + test + cleanup):

```bash
make test-realenv-full
```

## Available Tests

### Bootstrap Tests (`bootstrap_real_test.go`)

| Test | Description |
|------|-------------|
| `TestBootstrapCloneFromDevToLocal` | Clone project from DEV to LOCAL via SSH |
| `TestBootstrapCloneCodeOnly` | Clone with --code-only flag (skip DB/media) |
| `TestBootstrapValidationFreshAndClone` | Verify --fresh and --clone are mutually exclusive |
| `TestBootstrapValidationCodeOnlyRequiresClone` | Verify --code-only requires --clone |
| `TestBootstrapCloneRequiresConfiguredRemote` | Verify error when remote not configured |

### Sync Tests (`sync_real_test.go`)

| Test | Description |
|------|-------------|
| `TestSyncFilesDevToLocal` | Sync files from DEV to LOCAL |
| `TestSyncPlanMode` | Generate sync plan without executing |
| `TestSyncSameSourceAndDestination` | Verify error on same source/dest |
| `TestSyncWithPatterns` | Sync with include/exclude patterns |
| `TestSyncLocalToStaging` | Bidirectional sync LOCAL to STAGING |
| `TestSyncWithDelete` | Sync with --delete flag |
| `TestSyncFull` | Full sync (files + media + db) |

### Database Tests (`db_real_test.go`)

| Test | Description |
|------|-------------|
| `TestDBDumpFromDev` | Dump database from DEV environment |
| `TestDBImportFromFile` | Import SQL dump file to LOCAL |
| `TestDBStreamFromRemote` | Stream database from remote to LOCAL |
| `TestDBDumpWithFullOption` | Full dump with routines/events |
| `TestDBImportInvalidFile` | Error handling for invalid file |
| `TestDBDumpToNonexistentDir` | Error handling for invalid path |

### Remote Tests (`remote_real_test.go`)

| Test | Description |
|------|-------------|
| `TestRemoteTestConnectionToDev` | Test SSH connection to DEV |
| `TestRemoteTestConnectionToStaging` | Test SSH connection to STAGING |
| `TestRemoteTestAllEnvironments` | Test all configured remotes |
| `TestRemoteTestInvalidRemote` | Error for non-configured remote |
| `TestRemoteExecOnDev` | Execute command on DEV via SSH |
| `TestRemoteExecOnStaging` | Execute command on STAGING |
| `TestRemoteAuditDev` | Audit DEV environment |

## Manual Testing

You can manually interact with the test environment:

```bash
# SSH into DEV environment (user: linuxserver.io)
ssh -i tests/integration/realenv/.ssh/id_rsa \
    -p 9023 -o StrictHostKeyChecking=no \
    linuxserver.io@localhost

# Access DEV database
docker exec -it govard-test-dev-db \
    mysql -umagento -pmagento magento

# Check container logs
docker logs govard-test-dev-ssh
docker logs govard-test-local-db

# List all test containers
docker ps | grep govard-test
```

## Test Data

Each environment is initialized with:

- **LOCAL**: Basic Magento 2 structure, minimal test data
- **DEV**: Magento 2 with sample data markers, multiple test products
- **STAGING**: Production-like setup with sanitized data

Database schemas include:

- `core_config_data` - Magento configuration
- `test_markers` - Environment identification
- `catalog_product_entity` - Sample products
- `customer_entity` - Sample customers (DEV only)

## Configuration Files

Reuses existing fixtures from `tests/integration/projects/magento2/`:

- **`options-local/`** - LOCAL workstation config (has remotes: dev, staging, production)
- **`options-dev/`** - DEV environment (no remotes - it's a target, not a workstation)
- **`options-staging/`** - STAGING environment (no remotes - it's a target, not a workstation)

**Important**: In real environment tests:

- The **LOCAL** fixture is used as the project (workstation)
- **DEV** and **STAGING** are remote targets accessed via SSH
- Tests run commands from LOCAL, syncing to/from DEV/STAGING

Each fixture includes:

- `.govard.yml` - Govard configuration
- `composer.json` - Magento 2 project definition
- `app/etc/env.php` - Magento environment configuration
- `init.sql` - Database initialization script

## Troubleshooting

### Port Already in Use

If you get port binding errors, check what's using the ports:

```bash
# Check ports
sudo lsof -i :3306
sudo lsof -i :3307
sudo lsof -i :3308
sudo lsof -i :9022
sudo lsof -i :9023
sudo lsof -i :9024

# Cleanup if needed
make test-realenv-clean
```

### SSH Connection Failing

```bash
# Regenerate SSH keys
cd tests/integration/environments
rm -rf .ssh
./setup-three-env.sh

# Test SSH manually (user: linuxserver.io)
ssh -i .ssh/id_rsa -p 9023 -o StrictHostKeyChecking=no linuxserver.io@localhost
```

### Database Not Starting

```bash
# Check container status
docker ps -a | grep govard-test

# Check logs
docker logs govard-test-dev-db

# Restart specific service
docker-compose -f tests/integration/realenv/docker-compose.three-env.yml restart dev-db
```

## CI/CD Integration

These tests can be run in CI but require:

- Docker-in-Docker or privileged mode
- Sufficient resources (8GB+ RAM)
- Extended timeout (30+ minutes)

Example GitHub Actions workflow is in `.github/workflows/real-env-tests.yml`

## Architecture

### Test Infrastructure (`real_env_test.go`)

The `RealEnvTest` struct provides:

- Automatic binary building
- Environment health checks
- Command execution with proper environment
- Result assertions

### Environment Setup (`setup-three-env.sh`)

The setup script handles:

1. SSH key generation
2. Fixture preparation
3. Docker Compose orchestration
4. Health check verification
5. Configuration generation

### Docker Compose (`docker-compose.three-env.yml`)

Defines:

- 3 PHP-FPM containers
- 3 MySQL containers  
- 3 SSH server containers
- Shared network and SSH volume

## Adding New Tests

To add a new real environment test:

1. Create test function in appropriate file:

```go
func TestMyNewFeature(t *testing.T) {
    env := NewRealEnvTest(t)
    env.Setup(t)
    
    localDir := env.CreateTempProject(t, "local")
    
    result := env.RunGovard(t, localDir, "my", "command", "--flag")
    result.AssertSuccess(t)
    result.AssertOutputContains(t, "expected output")
}
```

1. Run with realenv tag:

```bash
go test -tags realenv ./tests/integration/realenv/... -run TestMyNewFeature -v
```

## Contributing

When adding new tests:

- Use the `RealEnvTest` infrastructure
- Clean up any resources created during tests
- Add both success and failure cases
- Document expected behavior in test name
