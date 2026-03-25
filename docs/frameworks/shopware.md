# Shopware

Govard supports Shopware projects with optimized nginx configurations and sensible defaults for PHP and MariaDB.

## Requirements

- PHP 8.4 (default)
- MariaDB 11.4

## Detection

Govard detects Shopware when `shopware/core` or `shopware/platform` is present in `composer.json`.

## Default Stack

- **Web Server**: nginx
- **Web Root**: `/public`
- **PHP**: `8.4`
- **Database**: MariaDB `11.4`

## Commands

Use the Shopware CLI:

```bash
govard tool shopware list
govard tool shopware cache:clear
```

## Remote Database Operations

Govard auto-detects remote database credentials for Shopware by SSHing to the remote and parsing the `.env` file.

## Examples

```bash
# Start the environment
govard env up

# Clear cache
govard tool shopware cache:clear
```
