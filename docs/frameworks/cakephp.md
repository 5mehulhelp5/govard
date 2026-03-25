# CakePHP

Govard provides a pre-configured stack for CakePHP applications with automated environment detection and optimized defaults.

## Requirements

- PHP 8.4 (default)
- MariaDB 11.4

## Detection

Govard detects CakePHP when `cakephp/cakephp` is present in `composer.json`.

## Default Stack

- **Web Server**: nginx
- **Web Root**: `/webroot`
- **PHP**: `8.4`
- **Database**: MariaDB `11.4`

## Commands

Use `bin/cake` for CakePHP CLI tasks:

```bash
govard tool cake version
```

## Remote Database Operations

Govard auto-detects remote database credentials for CakePHP by SSHing to the remote and parsing the `.env` file.

## Examples

```bash
# Start the environment
govard env up

# Run migrations
govard tool cake migrations migrate
```
