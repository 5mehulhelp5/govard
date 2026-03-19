# Drupal

Govard provides support for Drupal projects with optimized nginx configurations and sensible defaults.

## Requirements

- PHP 8.3 or 8.4 (default 8.4)
- MariaDB 11.4

## Detection

Govard detects Drupal when one of the following is present:

- `drupal/core` in `composer.json`
- `web/core` or `core` directory exists

## Default Stack

- **Web Server**: nginx
- **Web Root**: `/web` (falls back to project root if `/web` is missing)
- **PHP**: `8.4`
- **Database**: MariaDB `11.4`

## Commands

Use `drush` for Drupal CLI tasks:

```bash
govard tool drush status
govard tool drush cache:rebuild
```

## Remote Database Operations

Govard auto-detects remote database credentials for Drupal by SSHing to the remote and parsing the `.env` file for database connection strings.

## Examples

```bash
# Start the environment
govard env up

# Rebuild cache
govard tool drush cr
```
