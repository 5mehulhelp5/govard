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
- Varnish: optional and supported in front of `nginx`, `apache`, and `hybrid` web server modes

### Magento 2 multiple websites / stores

Govard handles local routing for every hostname in `domain`, `extra_domains`, and `store_domains`. For Magento 2, you can additionally ask `govard config auto` to set scoped base URLs with `store_domains`.

Example:

```yaml
framework: "magento2"
domain: "primary.test"
store_domains:
  store-a.test:
    code: base
    type: website
  store-b.test:
    code: store_b
    type: store
```

Recommended flow:

```bash
govard domain add store-a.test
govard domain add store-b.test
govard config auto
govard tool magento cache:flush
```

You only need `govard domain add ...` when you want to persist those hostnames into `extra_domains` separately. Rendering and routing already include the `store_domains` hostnames.

What Govard does:

- routes `primary.test`, `store-a.test`, and `store-b.test` through the shared proxy
- keeps local HTTPS working for every listed domain
- sets the global base URL from `domain`
- adds optional scoped `bin/magento config:set` commands for each `store_domains` entry during `govard config auto`
- emits `MAGE_RUN_CODE` / `MAGE_RUN_TYPE` host mappings when a `store_domains` entry uses the object form with explicit `type`

What you still need to do in Magento 2:

- create the websites, stores, and store views in Magento admin or with `bin/magento`
- ensure each `store_domains.<host>.code` matches the intended Magento website or store code
- clear config/cache after changing domains or store mappings

If a referenced scope code does not exist yet, `govard config auto` keeps going and leaves that scoped command as optional rather than failing the whole setup.

## Magento 1 (OpenMage)

Use:

```bash
govard tool magerun [command]
```

Default runtime is conservative: PHP 8.1 with MariaDB 10.11 and no optional cache/search/queue service forced on.

### Magento 1 multiple websites / stores

Govard can route multiple local hostnames to the same Magento 1 project, and it can now emit host-based `MAGE_RUN_CODE` / `MAGE_RUN_TYPE` mappings when you declare `store_domains` with explicit `type`.

Example `.govard.yml`:

```yaml
framework: "magento1"
domain: "primary.test"
store_domains:
  store-a.test:
    code: base
    type: website
  store-b.test:
    code: store_b
    type: store
  store-c.test: store_c
```

Recommended flow:

```bash
govard domain add store-a.test
govard domain add store-b.test
govard domain add store-c.test
govard env up
govard bootstrap --clone --yes
```

Like Magento 2, `store_domains` hostnames are routed automatically during render. Persist them to `extra_domains` only if you want them listed there explicitly.

What Govard does:

- routes every configured hostname through the shared proxy and local CA
- configures Magento 1 to trust `HTTP_X_FORWARDED_PROTO` for HTTPS detection during bootstrap
- runs `govard config auto` as part of remote bootstrap unless you use `--skip-up`
- sets the global base URL from `domain` during `govard config auto`
- tries each scalar `store_domains` entry as both a website code and a store code when updating scoped base URLs in `core_config_data`
- respects `store_domains.<host>.type=website|store` for both scoped base URL updates and generated runtime host mappings
- injects host-based `MAGE_RUN_CODE` / `MAGE_RUN_TYPE` mapping into nginx or Apache automatically when you use typed `store_domains` entries

What you still need to do in Magento 1:

- ensure each `store_domains.<host>.code` matches an existing Magento website code or store code
- use the object form with explicit `type` when you want deterministic runtime routing per hostname

You do not need manual `SetEnvIf` rules in `.htaccess` for the standard one-hostname-to-one-scope case when `store_domains` uses the object form.

Keep custom host switching in `.htaccess`, `index.php`, or `get.php` only when you want logic beyond Govard's generated mapping, for example:

```php
switch ($_SERVER['HTTP_HOST']) {
    case 'store-b.test':
        Mage::run('store_b', 'store');
        break;
}
```

For Magento 1, scalar `store_domains` entries still keep the legacy behavior of trying both scopes. Use the object form when you want Govard to emit deterministic `MAGE_RUN_CODE` / `MAGE_RUN_TYPE` routing.

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
