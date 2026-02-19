# govard snapshot

Create and restore local snapshots of database and media.

## Usage

```bash
govard snapshot create
govard snapshot create before-upgrade
govard snapshot list
govard snapshot restore before-upgrade
govard snapshot restore before-upgrade --db-only
govard snapshot restore before-upgrade --media-only
```

## Subcommands

- `create [name]` Create a snapshot in `./.govard/snapshots/<name>`
- `list` List available snapshots
- `restore <name>` Restore an existing snapshot

## Restore Options

- `--db-only` Restore database only
- `--media-only` Restore media only

## Notes

- Snapshot creation attempts to dump the local DB container `<project>-db-1`.
- Media is copied from the framework-specific local media path.
