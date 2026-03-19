# Magento 1 / OpenMage

Govard supports Magento 1 and OpenMage LTS projects with a lightweight stack and nginx template.

## Requirements

- PHP 7.4 or 8.1 (default 8.1)
- MariaDB 10.11
- A Magento 1/OpenMage codebase (e.g. OpenMage LTS or a legacy Magento 1 project)

## Detection

Govard detects Magento 1/OpenMage when one of the following is present:

- `openmage/magento-lts` in `composer.json`
- `magento-hackathon/magento-composer-installer` in `composer.json`
- `app/Mage.php`
- `app/etc/local.xml`

## Default Stack

- Web: nginx (`magento1.conf`)
- PHP: `8.1`
- DB: MariaDB `10.11`
- Cache: Redis (optional)

## Bootstrap (clone workflow)

When running `govard bootstrap --clone`, Govard performs the following post-clone steps automatically:

1. **`app/etc/local.xml`** — Created if missing. Uses the local container database credentials and a **randomly generated 32-character crypt key** (crypto/rand). This ensures each bootstrapped environment has a unique key.
2. **Directory setup** — Creates `var/cache`, `var/session`, and `media/` directories.

> For OpenMage LTS fresh installs, `govard bootstrap --fresh` also runs `composer create-project` and sets up `local.xml` with a random crypt key.

## Remote database operations

Govard auto-detects remote database credentials for Magento 1 / OpenMage by SSHing to the remote and parsing `app/etc/local.xml` with PHP. This means `govard db dump -e staging`, `govard db import --stream-db -e staging`, and `govard sync -e staging --db` all work without manual credential configuration.

## Set-config utilities

The bootstrap package exposes `RunMagento1SetConfigSQL` and `RunMagento1AdminUserSQL` for post-import configuration. These are used by hooks to:

- Update `core_config_data` base URLs (`web/secure/base_url`, `web/unsecure/base_url`, and related paths) to the local environment URL.
- Create the admin user with a salted MD5 password (Magento 1 compatible format).

## Commands

Use `n98-magerun` for Magento 1 CLI tasks:

```bash
govard tool magerun cache:flush
govard tool magerun admin:user:create --username admin --email admin@example.com
```

## Database dump filters

`--no-noise` and `--no-pii` flags work with Magento 1 projects exactly as with Magento 2:

```bash
# Exclude noise tables
govard db dump --no-noise --file backup.sql

# Exclude noise + PII tables
govard db dump --no-noise --no-pii --file backup.sql
```

See [`govard db`](../commands/db.md) for the full table lists.
