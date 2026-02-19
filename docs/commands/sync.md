# govard sync

Synchronize files, media, and databases between environments.

## Usage

```bash
govard sync --source staging --destination local --file
govard sync --source staging --destination local --full --plan
```

## Options

- `--source` Source environment (default: staging)
- `--destination` Destination environment (default: local)
- `--file` Sync source code/files
- `--media` Sync media files
- `--db` Sync database
- `--full` Sync files, media, and database
- `--delete` Remove destination files missing on source
- `--resume` Enable resumable rsync transfers (`--partial --append-verify`) (default: enabled)
- `--no-resume` Disable resumable rsync transfers
- `--path` Sync a specific path
- `--include` Rsync include pattern (repeatable)
- `--exclude` Rsync exclude pattern (repeatable)
- `--plan` Print a dry-run summary (endpoints, scopes, risk, steps) and exit

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
- Policy warnings are included for risky options such as `--delete` and `--db`.
- Executions and plans are logged to `~/.govard/remote.log` for remote audit history.

## Examples

```bash
govard sync --source staging --destination local --media
govard sync --source staging --destination local --full --plan
govard sync --source staging --destination local --file --no-resume
govard sync --source staging --destination local --file --include 'app/*' --exclude 'vendor/'
```
