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
- `-y, --yes`
- `--delete`
- `--resume`
- `--no-resume`
- `--no-compress`
- `--path`
- `--include`
- `--exclude`

### Database filter flags

- `-N, --no-noise`
- `-P, --no-pii`

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

## Remote Name Resolution

Remote flags support:

- exact remote key lookup
- normalized aliases such as `stg` -> `staging`

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
