# Remotes and Sync

This is the canonical guide for Govard remote environments, sync operations, and remote-backed database workflows.

## Remote Setup

Add and validate a remote:

```bash
govard remote add staging --host staging.example.com --user deploy --path /var/www/app
govard remote copy-id staging
govard remote test staging
```

Useful `remote add` flags:

- `--host`
- `--user`
- `--path`
- `--port`
- `--capabilities`
- `--auth-method`
- `--key-path`
- `--strict-host-key`
- `--known-hosts-file`
- `--protected`

Auth methods:

- `keychain`
- `ssh-agent`
- `keyfile`

If you want to store a home-relative remote path, quote it in shell usage so the local shell does not expand `~/...` before Govard receives the flag:

```bash
govard remote add staging --host staging.example.com --user deploy --path '~/public_html'
```

### Audit and exec

```bash
govard remote exec staging -- ls -la
govard remote audit tail --lines 20
govard remote audit tail --status failure --lines 50
govard remote audit stats --lines 200
```

Audit logs are written to:

- `~/.govard/remote.log`
- `~/.govard/operations.log`

## Remote Safety Model

- production remotes are write-protected by default
- capabilities are enforced per operation: `files`, `media`, `db`, `deploy`
- strict host-key verification is opt-in per remote
- remote fields support `op://...` secret references via 1Password CLI

`remote test` validates SSH connectivity, checks `rsync`, measures probe latency, and classifies failures:

- `network`
- `auth`
- `permission`
- `host_key`
- `dependency`

## Sync Overview

`govard sync` moves files, media, and database data between local and named remotes.

```bash
govard sync --source staging --destination local --full --plan
govard sync --from staging --to local --media
govard sync -s dev --db --no-noise --no-pii
govard sync -s prod --file --path app/etc/config.php
```

### Endpoint flags

- `-s, --source`
- `--from`
- `-d, --destination`
- `--to`
- `-e, --environment`

`--source` and `--from` are interchangeable. `-e, --environment` remains supported as a source selector.

### Scope flags

- `--file`
- `--media`
- `--db`
- `--full`

### Transfer flags

- `--plan`
### Source & Destination

- `-s, --source`: The source environment (defaults to `staging`).
- `-d, --destination`: The target environment (defaults to `local`).
- `-e, -f, --environment, --from`: Aliases for `--source`.
- `--to`: Alias for `--destination`.

### Resource Scopes

- `-A, --full`: Sync all supported resources (files, media, database).
- `-f, --file`: Sync source code and generic files.
- `-m, --media`: Sync framework-specific media assets.
- `-b, --db`: Sync the project database.

### Filters & Paths

- `-p, --path`: Sync a specific file or directory relative to the project root.
- `-I, --include`: Rsync include pattern (repeatable).
- `-X, --exclude`: Rsync exclude pattern (repeatable).

### Database Privacy & Protection

- `-N, --no-noise`: Exclude ephemeral/noise tables (logs, cache, etc.).
### Transfer & Execution Control

- `-D, --delete`: Delete files on destination that are missing on source.
- `-R, --resume`: Enable resumable transfers (default: `true`).
- `--no-resume`: Disable resumable Transfers.
- `-C, --no-compress`: Disable rsync compression.
- `-y, --yes`: Skip confirmation prompts for all sync operations.
- `--plan`: Print the plan and exit without executing.

### Database Filter Policy

The `--no-noise` and `--no-pii` flags use context-aware rules to exclude specific tables and directories:

| Flag | Category | Key Exclusions (Magento 2) | Key Exclusions (Laravel) | Key Exclusions (WordPress) |
| :--- | :--- | :--- | :--- | :--- |
| `--no-noise` | **Ephemeral Data** | `cron_schedule`, `session`, `cache_tag`, `report_event`, `adminnotification_inbox` | `cache`, `sessions`, `failed_jobs`, `telescope_entries` | `redirection_404`, `wflogs` |
| `--no-pii` | **Sensitive Data** | `customer_entity`, `sales_order`, `quote`, `newsletter_subscriber`, `admin_user`, `wishlist` | `users`, `password_resets`, `personal_access_tokens` | `users`, `usermeta`, `comments` |

> [!NOTE]
> Database filters are currently optimized for Magento 2. For non-supported frameworks, `--no-noise` and `--no-pii` will fall back to safe default patterns when possible.

## Sync Behavior

Govard supports:

- remote to local
- local to remote

Current guardrails:

- protected destinations block destructive writes
- risky combinations such as `--delete` and `--db` surface policy warnings
- `--plan` prints endpoints, scopes, transfer steps, and risk before execution

### Resumable transfers

File and media sync use resumable rsync mode by default:

- `--partial`
- `--append-verify`

Disable that behavior with `--no-resume`.

### Include and exclude filters

`--include` and `--exclude` apply to rsync scopes only:

- `--file`
- `--media`

They are ignored for DB-only sync.

### Integration with `bootstrap`

The `govard bootstrap` command leverages these same sync and filter flags to perform safe, high-performance environment initialization:

```bash
govard bootstrap --clone -e staging --no-pii --no-noise --delete
```

This applies the privacy filters to the database import and enables the `--delete` policy for the initial file clone.

## Remote Name Resolution

Remote flags (`--source`, `--from`, `-e`, `--environment`) support:

- exact remote key lookup
- normalized aliases such as `stg` -> `staging`

### Auto-select Remote Priority

When no remote is explicitly provided for `bootstrap` or `sync`, Govard automatically attempts to resolve a source remote based on the following priority:

1. **`staging`** (and its aliases: `stg`, `stage`, `qa`, `uat`, `test`)
2. **`dev`** (and its aliases: `development`, `local`)

If your project only has a `staging` remote, Govard will use it automatically. If it has both, `staging` is prioritized. If neither exists, you must specify the remote explicitly or add one.

These are equivalent when the remote exists:

```bash
govard sync -s stg --db
govard sync --source staging --db
govard sync --from staging --db
```

## Remote Database Workflows

### Dump

```bash
govard db dump
govard db dump -e staging
govard db dump -e staging --local
govard db dump --no-noise
govard db dump --no-pii
```

Storage behavior:

- local environment: save into the project `var/` directory
- remote environment by default: save on the remote host, usually under `~/backup/`
- remote environment with `--local`: stream the dump back into local `var/`

### Import

```bash
govard db import --file backup.sql --drop
govard db import --stream-db -e staging --drop
```

`--stream-db` pulls from the remote and imports into the local database. `--drop` performs a safe reset before import.

### Query, info, and top

```bash
govard db query "SELECT COUNT(*) FROM sales_order"
govard db info -e staging
govard db top -e staging
```

## Desktop Remote Actions

Desktop remotes reuse the same backend flows:

- Open Database (Remote) -> `govard open db -e <remote> --client`
- Open SSH (Remote) prefers native Linux terminal launchers, then falls back to `ssh://`
- Open SFTP (Remote) prefers FileZilla, then falls back to `sftp://`

For `auth.method: ssh-agent`, Govard reuses `SSH_AUTH_SOCK` and also probes `/run/user/<uid>/keyring/ssh` on Linux.

## Recommended Patterns

Safe review before execution:

```bash
govard sync --source staging --destination local --full --plan
```

Target one file:

```bash
govard sync --source prod --file --path app/etc/config.php
```

Create a local-safe DB snapshot:

```bash
govard db dump -e staging --local --no-noise --no-pii
```

## Related Docs

- [Commands](commands.md)
- [Configuration](configuration.md)
- [Desktop](desktop.md)
