# govard db

Database utilities for the local database container.

## Usage

```bash
govard db connect
govard db dump > backup.sql
cat backup.sql | govard db import
govard db import --stream-db --environment staging
govard db import --stream-db --environment staging --file staging.sql
```

## Options

- `-e, --environment` Target environment (default: local)
- `-f, --file` Database dump file (import or dump output)
- `--stream-db` For `db import`: stream dump from remote environment into local database
- `--full` For `db dump`: include routines, events, and triggers
- `--exclude-sensitive-data` Apply SQL sanitization pipeline (DEFINER/GTID cleanup)

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
cat backup.sql | govard db import
govard db dump -e staging --file staging.sql --exclude-sensitive-data
govard db import --stream-db -e staging
```
