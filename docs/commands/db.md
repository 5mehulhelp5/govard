# govard db

Database utilities for the local database container.

## Usage

```bash
govard db connect
govard db dump
govard db dump --local
govard db query "SELECT * FROM core_config_data LIMIT 5"
govard db info
govard db top
govard db import --file backup.sql --drop
govard db import --stream-db --environment staging --drop
```

## Options

- `-e, --environment` Target environment (default: local)
- `-f, --file` Database dump file (import or dump output)
- `--profile` Environment scope (profile) to use. Loads `.govard.<profile>.yml` config layer.
- `--stream-db` For `db import`: stream dump from remote environment into local database
- `--drop` For `db import`: drop and recreate the database before importing
- `--local` For `db dump/import`: force local file operations for remote environments
- `-N, --no-noise` For `db dump`: exclude ephemeral/noise tables (cron, cache, session, logs, indexes…)
- `-S, --no-pii` For `db dump`: exclude PII/sensitive tables (customers, orders, quotes…)
- `-y, --yes` For `db import`: skip interactive confirmation when dropping or streaming databases

> **Note:** SQL sanitization (DEFINER/GTID cleanup) and tiered dumping (metadata + data) are always applied automatically for Magento 2.

## Subcommands

### `connect`

Open an interactive MySQL/MariaDB shell to the database container (local or remote).

```bash
govard db connect
govard db connect -e staging
```

### `dump`

Create a database dump. Dumps are comprehensive (including routines and triggers) by default.

**Storage Behavior:**
- **Local Environment**: Saved to project's local `var/` directory.
- **Remote Environment (default)**: Stored on the remote server (usually `~/backup/`).
- **Remote Environment (+ --local)**: Streamed directly to the project's local `var/` directory.

```bash
# Save to remote ~/backup/
govard db dump -e staging

# Stream from remote to local var/
govard db dump -e staging --local

# Exclude noise tables (cron, cache, session, log tables, indexers…)
govard db dump --no-noise

# Exclude PII tables (customers, orders, quotes…)
govard db dump --no-pii

# Exclude noise + PII tables
govard db dump --no-noise --no-pii
```

### `import`

Import SQL from stdin or a file. Use `--stream-db` to stream from a remote environment directly into local database. Use `--drop` to ensure a clean slate before importing.

```bash
# Import a local file
govard db import --file backup.sql --drop

# Stream from remote with clean reset
govard db import --stream-db -e staging --drop
```

### `query`

Execute a single SQL query and display results.

```bash
govard db query "SELECT COUNT(*) FROM sales_order"
govard db query "SELECT * FROM core_config_data WHERE path LIKE 'web/%'" -e staging
```

### `info`

Display database connection information (host, port, username, database).

```bash
govard db info
govard db info -e staging
```

### `top`

Real-time database process monitoring.

```bash
govard db top
govard db top -e staging
```

## Notes

- **Progress Bar:** Database `import` (with `--file` or `--stream-db`) operations include a real-time progress bar.
- **Reset Logic:** The `--drop` flag uses a robust reset script that kills active connections before dropping the database to avoid lock issues.
- **Warden Parity**: Table exclusion lists and command-line flags are synchronized with Warden's latest specialized Magento 2 tools.
- Local DB client fallback: automatically switches between `mysql` and `mariadb` clients.
- Remote credential auto-detection for Magento 1/2, Symfony, Laravel, Drupal, WordPress, Shopware, and more.

## Table filter flags

`--no-noise` and `--no-pii` use an expanded list of tables ported directly from Warden.

| Flag | Tables excluded |
| --- | --- |
| `--no-noise` (`-N`) | Ephemeral / high-churn tables: `cron_schedule`, `cache_tag`, `session`, index replicas, `report_*`, `queue_message`, `oauth_nonce`, `search_query`, log tables, third-party activity tables (Amasty, ElasticSuite, Klaviyo, Mailchimp, etc.). |
| `--no-pii` (`-S`) | PII tables: `customer_entity*`, `sales_order*`, `quote*`, `newsletter_subscriber`, `wishlist*`, `admin_user`, `admin_passwords`, `paypal_*`, `vault_payment_token*`, B2B company tables, etc. |

## Examples

```bash
# Quick local backup
govard db dump

# Clean sync from production
govard db import --stream-db -e prod --drop

# Dump from staging to local for debugging
govard db dump -e staging --local -N
```
