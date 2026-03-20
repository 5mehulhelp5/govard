# govard snapshot

Create and restore local compressed snapshots of database and media.

## Usage

```bash
govard snapshot create
govard snapshot create before-upgrade
govard snapshot list
govard snapshot restore before-upgrade
govard snapshot restore before-upgrade --db-only
govard snapshot restore before-upgrade --media-only
govard snapshot export before-upgrade backup.tar.gz
govard snapshot delete temporary-snap
```

## Subcommands

- `create [name]` Create a compressed snapshot in `./.govard/snapshots/<name>`
- `list` List available snapshots with their estimated disk size
- `restore <name>` Restore an existing snapshot (automatically detects compression)
- `delete <name>` Permanently delete a snapshot
- `export <name> [file]` Export a snapshot to a compressed `.tar.gz` file (defaults to `name.tar.gz`)

## Restore Options

- `--db-only` Restore database only
- `--media-only` Restore media only

## Notes

- Snapshot creation uses **Gzip compression** for the database dump to save disk space (typically 70-90% reduction).
- Restore automatically detects whether a snapshot is compressed or legacy plain SQL.
- Media is copied from the framework-specific local media path.
