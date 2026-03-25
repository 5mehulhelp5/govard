# Laravel

Govard provides a pre-configured stack for Laravel applications with automated environment detection and optimized defaults.

## Requirements

- PHP 8.2, 8.3, or 8.4 (default 8.4)
- MariaDB 11.4
- Node.js for frontend assets (optional)

## Detection

Govard detects Laravel when `laravel/framework` is present in `composer.json`.

## Default Stack

- **Web Server**: nginx (optimized for Laravel)
- **Web Root**: `/public`
- **PHP**: `8.4`
- **Database**: MariaDB `11.4`

## Commands

Use `php artisan` for Laravel CLI tasks:

```bash
govard tool artisan list
govard tool artisan migrate
```

## Remote Database Operations

Govard auto-detects remote database credentials for Laravel by SSHing to the remote and parsing the `.env` file for `DB_*` variables.

## Examples

```bash
# Start the environment
govard env up

# Run migrations
govard tool artisan migrate

# Clear cache
govard tool artisan cache:clear
```

## Testing Integration

Govard provides a unified interface for running Laravel tests:

```bash
# Run unit and feature tests
govard test phpunit

# Run static analysis
govard test phpstan
```

Tests run as the web user inside the PHP container.
