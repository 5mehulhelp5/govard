# Frameworks

Govard detects supported frameworks and applies runtime defaults plus version-aware overrides where available.

## Support Matrix

| Framework | Detection | Version-aware profile | Default web root |
| --- | --- | --- | --- |
| Magento 2 | yes | yes | `/pub` |
| Magento 1 / OpenMage | yes | framework defaults | project root |
| Laravel | yes | yes | `/public` |
| Next.js | yes | framework defaults | project root |
| Drupal | yes | yes | `/web` |
| Symfony | yes | yes | `/public` |
| Shopware | yes | framework defaults | `/public` |
| CakePHP | yes | framework defaults | `/webroot` |
| WordPress | yes | yes | `/wordpress` |
| Custom | manual | manual | project root |

## Runtime Defaults

| Framework | PHP | Node | DB | Cache | Search | Queue |
| --- | ---: | ---: | --- | --- | --- | --- |
| Magento 2 | 8.4 | 24 | mariadb 11.4 | valkey 8.0.0 | opensearch 2.19.0 | none |
| Magento 1 / OpenMage | 8.1 | - | mariadb 10.11 | none | none | none |
| Laravel | 8.4 | - | mariadb 11.4 | none | none | none |
| Next.js | - | 24 | none | none | none | none |
| Drupal | 8.4 | - | mariadb 11.4 | none | none | none |
| Symfony | 8.4 | - | mariadb 11.4 | none | none | none |
| Shopware | 8.4 | - | mariadb 11.4 | none | none | none |
| CakePHP | 8.4 | - | mariadb 11.4 | none | none | none |
| WordPress | 8.3 | - | mariadb 11.4 | none | none | none |
| Custom | 8.4 | - | mariadb 11.4 | none | none | none |

`-` means Govard does not force a default for that stack component.

## Version-aware Overrides

| Framework | Version | Override |
| --- | --- | --- |
| Laravel | 10 | PHP 8.2 |
| Laravel | 11 | PHP 8.3 |
| Laravel | 12 | PHP 8.4 |
| Symfony | 6 | PHP 8.2 |
| Symfony | 7 | PHP 8.3 |
| Drupal | 10 | PHP 8.3 |
| Drupal | 11 | PHP 8.4 |
| WordPress | 6 | PHP 8.3 |
| Magento 2 | 2.4.9+ | PHP 8.4, MariaDB 11.4, Redis 7.2, OpenSearch 3.0.0, RabbitMQ 4.1.0 |
| Magento 2 | 2.4.8 | PHP 8.4, MariaDB 11.4, Redis 7.2, OpenSearch 2.19.0 or 3.0.0, RabbitMQ 4.1.0 |
| Magento 2 | 2.4.7 | PHP 8.3, MariaDB 10.6 or 10.11, Redis 7.2, OpenSearch 2.12.0 or 2.19.0, RabbitMQ 3.13.7 or 4.1.0 |
| Magento 2 | 2.4.6 | PHP 8.2, MariaDB 10.6 or 10.11, Redis 7.0 or 7.2, OpenSearch 2.5.0 to 2.19.0, RabbitMQ 3.9.0 to 4.1.0 |

Use `govard config profile --json` to inspect the selected runtime profile.

## Magento 2

Magento 2 is the deepest supported workflow in Govard.

Highlights:

- `govard config auto` injects DB, cache, search, Varnish, and base URL settings into `app/etc/env.php`
- `govard tool magento [command]` runs Magento CLI commands
- optional Selenium support for MFTF
- optional frontend watcher support for Grunt or Vite workflows
- dedicated `php-debug` routing when Xdebug is enabled

Typical flow:

```bash
govard env up
govard config auto
govard tool magento cache:clean
govard db dump > dump.sql
govard test phpunit
```

Common services:

- web root: `/pub`
- search: OpenSearch by default
- cache/session: Redis or Valkey depending on selected runtime profile
- queue: optional RabbitMQ

## Magento 1 (OpenMage)

Use:

```bash
govard tool magerun [command]
```

Default runtime is conservative: PHP 8.1 with MariaDB 10.11 and no optional cache/search/queue service forced on.

## Laravel

Use:

```bash
govard tool artisan [command]
```

Defaults:

- web root: `/public`
- MariaDB 11.4
- PHP selected from the version-aware profile when available

## Drupal

Use:

```bash
govard tool drush [command]
```

Defaults:

- web root: `/web`
- MariaDB 11.4
- version-aware PHP selection

## Symfony

Use:

```bash
govard tool symfony [command]
```

Defaults:

- web root: `/public`
- MariaDB 11.4
- version-aware PHP selection

## Shopware

Use:

```bash
govard tool shopware [command]
```

Defaults:

- web root: `/public`
- MariaDB 11.4

## CakePHP

Use:

```bash
govard tool cake [command]
```

Defaults:

- web root: `/webroot`
- MariaDB 11.4

## WordPress

Use:

```bash
govard tool wp [command]
```

Defaults:

- web root: `/wordpress`
- MariaDB 11.4
- PHP 8.3 by default

## Next.js

Next.js uses Node-focused runtime defaults:

- project-root web serving
- Node 24
- no DB forced by default

## Custom

`govard init --framework custom` opens an interactive stack picker for:

- web server
- database
- cache
- search
- queue
- optional Varnish

## Verification Commands

```bash
govard config profile --json
govard config profile --framework laravel --framework-version 11 --json
govard config profile apply --framework laravel --framework-version 11
```

## Related Docs

- [Getting Started](getting-started.md)
- [Configuration](configuration.md)
- [Commands](commands.md)
