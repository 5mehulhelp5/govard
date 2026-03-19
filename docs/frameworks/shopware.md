# Shopware

Govard supports Shopware projects with optimized nginx configurations and sensible defaults for PHP and MariaDB.

## Requirements

- PHP 8.4 (default)
- MariaDB 11.4

## Detection

Govard detects Shopware when `shopware/core` is present in `composer.json` or related Shopware files exist.

## Default Stack

- **Web Server**: nginx
- **Web Root**: `/public`
- **PHP**: `8.4`
- **Database**: MariaDB `11.4`

## Commands

Use the Shopware CLI:

```bash
govard tool console list
govard tool console cache:clear
```

## Remote Database Operations

Govard auto-detects remote database credentials for Shopware by SSHing to the remote and parsing the `.env` file.

## Examples

```bash
# Start the environment
govard env up

# Clear cache
govard tool console cache:clear
```
