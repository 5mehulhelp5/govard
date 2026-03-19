# govard sync

Synchronize files, media, and databases between environments.

## Usage

```bash
govard sync --source staging --destination local --file
govard sync -s staging -d local --full --plan
govard sync -s dev --db --no-pii
```

## Options

- `-s, --source` Source environment (default: `staging`). Accepts remote name or alias (e.g. `dev`, `stg`, `prod`).
- `-d, --destination` Destination environment (default: `local`). Accepts remote name or alias.
- `--file` Sync source code/files
- `--media` Sync media files
- `--db` Sync database
- `--full` Sync files, media, and database
- `--delete` Remove destination files missing on source
- `--resume` Enable resumable rsync transfers (`--partial --append-verify`) (default: enabled)
- `--no-resume` Disable resumable rsync transfers
- `--no-compress` Disable rsync compression
- `--path` Sync a specific path
- `--include` Rsync include pattern (repeatable)
- `--exclude` Rsync exclude pattern (repeatable)
- `--plan` Print a dry-run summary (endpoints, scopes, risk, steps) and exit
- `-N, --no-noise` Exclude ephemeral/noise tables from DB sync (cron, cache, sessions, logs, etc.)
- `-S, --no-pii` Exclude PII/sensitive tables from DB sync (users, orders, passwords, etc.). Implies `--no-noise`.

## Remote Name Resolution

The `--source` (`-s`) and `--destination` (`-d`) flags support fuzzy remote name resolution:

- Exact key match in `config.Remotes` (e.g. `staging`).
- Alias matching via `NormalizeRemoteEnvironment` (e.g. `stg` → `staging`, `dev` → `development`).

This means `govard sync -s stg --db` and `govard sync --source staging --db` are equivalent if your config has a remote named `staging`.

## Notes

- Destination writes are blocked for protected remotes and remotes classified as `prod`.
- Remote capabilities are enforced per scope:
  - `files` for `--file`
  - `media` for `--media`
  - `db` for `--db`
- Current execution support is local-to-remote and remote-to-local.
- `--plan` prints a sync summary and all generated transfer steps.
- Resume mode is enabled by default for rsync scopes (`--file`, `--media`) to better tolerate interrupted transfers.
- Use `--no-resume` when you need a strict non-resumable transfer.
- `--include` and `--exclude` apply to rsync scopes (`--file`, `--media`) and are ignored for DB-only sync.
- `--no-noise` and `--no-pii` apply to DB sync only. They generate `--ignore-table` args in the `mysqldump` command.
- Policy warnings are included for risky options such as `--delete` and `--db`.
- Remote endpoint config used by sync supports secret references (`op://...`) for fields like host/user/path/auth, resolved via the 1Password CLI.
- Executions and plans are logged to `~/.govard/remote.log` for remote audit history.
- Sync runs also emit operation events to `~/.govard/operations.log` (used by desktop notifications and operation timeline views).

## Examples

```bash
# Sync media from staging to local
govard sync -s staging -d local --media

# Full sync plan (dry-run)
govard sync -s staging -d local --full --plan

# Sync code without resumable transfers
govard sync -s staging -d local --file --no-resume

# Sync code with include/exclude filters
govard sync -s staging -d local --file --include 'app/*' --exclude 'vendor/'

# Sync DB excluding logs and caches
govard sync -s dev --db --no-noise

# Sync DB excluding PII (implies --no-noise)
govard sync -s dev --db --no-pii

# Push code to staging (local → remote)
govard sync -d staging --file --path app/code
```
