# govard db

Database utilities for the local database container.

## Usage

```bash
govard db connect
govard db dump > backup.sql
govard db query "SELECT * FROM core_config_data LIMIT 5"
govard db info
govard db top
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
- `-N, --no-noise` For `db dump`: exclude ephemeral/noise tables (cron, cache, session, logs, indexes…)
- `-S, --no-pii` For `db dump`: exclude PII/sensitive tables (customers, orders, quotes…) — implies `--no-noise`

> **Note:** SQL sanitization (DEFINER/GTID cleanup) is always applied automatically to local dump pipelines.

## Subcommands

### `connect`

Open an interactive MySQL/MariaDB shell to the database container (local or remote).

```bash
govard db connect
govard db connect -e staging
```

### `dump`

Create a database dump. Supports `--full` for including routines/events/triggers, `--file` for output to a file, and `--no-noise`/`--no-pii` to reduce dump size by excluding non-essential or sensitive tables.

```bash
govard db dump > backup.sql
govard db dump --file backup.sql --full

# Exclude noise tables (cron, cache, session, log tables, indexers…)
govard db dump --no-noise --file backup.sql

# Exclude noise + PII tables (customers, orders, quotes…)
govard db dump --no-noise --no-pii --file backup.sql

# Short flags
govard db dump -N -S --file backup.sql
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

### `top`

Real-time database process monitoring. Displays a live-updating table of active queries, their IDs, users, hosts, databases, commands, durations, and states.

```bash
govard db top
govard db top -e staging
```

## Notes

- **Progress Bar:** Database `import` (with `--file` or `--stream-db`) and `sync` operations now include a real-time progress bar to track data transfer.
- Use `-e <remote>` to run db operations over SSH for a configured remote.
- Local DB credentials are read from the local DB container (`MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`) with fallback `magento/magento/magento`.
- Local DB client fallback:
  - Govard uses `mysql` when available.
  - Falls back to `mariadb` client automatically when `mysql` binary is missing.
- Remote credential auto-detection:
  - Magento 2: from remote `app/etc/env.php`.
  - Magento 1/OpenMage: from remote `app/etc/local.xml`.
  - Symfony/Laravel/Drupal/WordPress/Shopware/CakePHP: from remote `.env*` (`DATABASE_URL` / `DB_*`).
  - Fallback remains `magento/magento/magento` when probing fails.
- Remote `db` capability is required for all remote database actions.
- Remote direct `import` is blocked when the target remote is protected or classified as `prod`.
- `db import --stream-db -e <remote>` treats the remote as a source and local as destination.
- `--file` can be used as file mode for both dump output and import input.
- Remote DB operations are logged to `~/.govard/remote.log` for audit and troubleshooting.

## Table filter flags

`--no-noise` and `--no-pii` add `--ignore-table` flags per excluded table to the `mysqldump` call. Tables excluded by each flag:

| Flag | Tables excluded |
| --- | --- |
| `--no-noise` (`-N`) | Ephemeral / high-churn tables: `cron_schedule`, `cache_tag`, `session`, index replicas, `report_*`, `queue_message`, `oauth_nonce`, `search_query`, log tables, third-party activity tables, etc. |
| `--no-pii` (`-S`) | All `--no-noise` tables **plus** PII tables: `customer_entity*`, `sales_order*`, `quote*`, `newsletter_subscriber`, `wishlist*`, `admin_user`, `admin_passwords`, `paypal_*`, `vault_payment_token*`, B2B company tables, etc. |

Use `--no-noise` for lightweight dumps shared with developers. Use `--no-pii` for public staging environments or any dump that must not contain customer data.

## Examples

```bash
govard db dump > backup.sql
govard db query "SELECT COUNT(*) FROM sales_order"
govard db info
govard db dump -e staging --file staging.sql -N
govard db dump -e staging --file staging-no-pii.sql -N -S
govard db import --stream-db -e staging
```
