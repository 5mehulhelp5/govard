# Symfony

Govard supports Symfony projects with a standard nginx/PHP-FPM stack and automated configuration detection.

## Requirements

- PHP 8.2, 8.3, or 8.4 (default 8.4)
- MariaDB 11.4

## Detection

Govard detects Symfony when one of the following is present:

- `symfony/framework-bundle` in `composer.json`
- `bin/console` file in the project root

## Default Stack

- **Web Server**: nginx
- **Web Root**: `/public`
- **PHP**: `8.4`
- **Database**: MariaDB `11.4`

## Commands

Use `bin/console` for Symfony CLI tasks:

```bash
govard tool console list
govard tool console cache:clear
```

## Remote Database Operations

Govard auto-detects remote database credentials for Symfony by SSHing to the remote and parsing the `.env` file for `DATABASE_URL` or `DB_*` variables.

## Examples

```bash
# Start the environment
govard env up

# Clear cache
govard tool console cache:clear
```
