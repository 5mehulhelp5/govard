# govard db

Database utilities for the local database container.

## Usage

```bash
govard db connect
govard db dump > backup.sql
govard db query "SELECT * FROM core_config_data LIMIT 5"
govard db info
cat backup.sql | govard db import
govard db import --stream-db --environment staging
govard db import --stream-db --environment staging --file staging.sql
```

## Options

- `-e, --environment` Target environment (default: local)
- `-f, --file` Database dump file (import or dump output)
- `--profile` Environment scope (profile) to use. Loads `.govard.<profile>.yml` config layer and targets profile-specific containers/volumes.
- `--stream-db` For `db import`: stream dump from remote environment into local database
- `--full` For `db dump`: include routines, events, and triggers
- `--exclude-sensitive-data` Apply SQL sanitization pipeline (DEFINER/GTID cleanup)

## Subcommands

### `connect`

Open an interactive MySQL/MariaDB shell to the database container (local or remote).

```bash
govard db connect
govard db connect -e staging
```

### `dump`

Create a database dump. Supports `--full` for including routines/events/triggers and `--file` for output to a file.

```bash
govard db dump > backup.sql
govard db dump --file backup.sql --full
```

### `import`

Import SQL from stdin or a file. Use `--stream-db` to stream from a remote environment directly into local database.

```bash
cat backup.sql | govard db import
govard db import --file backup.sql
govard db import --stream-db -e staging
```

### `query`

Execute a single SQL query and display results. Useful for quick data checks without opening an interactive shell.

```bash
govard db query "SELECT COUNT(*) FROM sales_order"
govard db query "SELECT * FROM core_config_data WHERE path LIKE 'web/%'" -e staging
```

### `info`

Display database connection information (host, port, username, database) for local or remote environments.

```bash
govard db info
govard db info -e staging
```

## Notes

- Use `-e <remote>` to run db operations over SSH for a configured remote.
- Local DB credentials are read from the local DB container (`MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`) with fallback `magento/magento/magento`.
- Local DB client fallback:
  - Govard uses `mysql` when available.
  - Falls back to `mariadb` client automatically when `mysql` binary is missing.
- Remote credential auto-detection:
  - Magento 2: from remote `app/etc/env.php`.
  - Symfony/Laravel/Drupal/WordPress/Shopware/CakePHP: from remote `.env*` (`DATABASE_URL` / `DB_*`).
  - Fallback remains `magento/magento/magento` when probing fails.
- Remote `db` capability is required for all remote database actions.
- Remote direct `import` is blocked when the target remote is protected or classified as `prod`.
- `db import --stream-db -e <remote>` treats the remote as a source and local as destination.
- `--file` can be used as file mode for both dump output and import input.
- Remote DB operations are logged to `~/.govard/remote.log` for audit and troubleshooting.

## Examples

```bash
govard db dump > backup.sql
govard db query "SELECT COUNT(*) FROM sales_order"
govard db info
govard db dump -e staging --file staging.sql --exclude-sensitive-data
govard db import --stream-db -e staging
```
