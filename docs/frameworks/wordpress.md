# WordPress

Govard provides support for WordPress development with automated configuration and optimized nginx templates.

## Requirements

- PHP 8.1, 8.2, or 8.3 (default 8.3)
- MariaDB 11.4

## Detection

Govard detects WordPress when one of the following is present:

- `wordpress` in `composer.json` or `wp-config.php` file
- `wp-content` directory or `wordpress` folder exists

## Default Stack

- **Web Server**: nginx
- **Web Root**: `/wordpress` (falls back to project root if `/wordpress` is missing)
- **PHP**: `8.3`
- **Database**: MariaDB `11.4`

## Commands

Use `wp-cli` for WordPress tasks:

```bash
govard tool wp list
govard tool wp core version
```

## Remote Database Operations

Govard auto-detects remote database credentials for WordPress by SSHing to the remote and parsing `wp-config.php` for `DB_*` constants.

## Examples

```bash
# Start the environment
govard env up

# Update core
govard tool wp core update
```
