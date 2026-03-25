# CLI Commands

This is the canonical CLI reference for Govard.

## Aliases and Shortcuts

Root lifecycle shortcuts:

- `govard up` -> `govard env up`
- `govard down` -> `govard env down`
- `govard restart` -> `govard env restart`
- `govard ps` -> `govard env ps`
- `govard logs` -> `govard env logs`

Common command aliases:

- `govard boot` -> `govard bootstrap`
- `govard cfg` -> `govard config`
- `govard dbg` -> `govard debug`
- `govard gui` -> `govard desktop`
- `govard diag` -> `govard doctor`
- `govard ext` -> `govard extensions`
- `govard prj` -> `govard projects`
- `govard rmt` -> `govard remote`
- `govard sh` -> `govard shell`
- `govard snap` -> `govard snapshot`

`govard sync` aliases:

- `--from` is an alias for `--source`
- `--to` is an alias for `--destination`
- `-e, --environment` remains a supported source-environment option

## Environment Commands

### `govard init`

Detect the project framework and generate `.govard.yml`.

```bash
govard init
govard init --framework magento2
govard init --framework custom
```

### `govard bootstrap`

Run bootstrap flows for clone or fresh-install setups.

```bash
govard bootstrap
govard bootstrap --clone --environment dev --yes
govard bootstrap --framework magento2 --fresh --framework-version 2.4.9
```

Bootstrap shows a review step before environment startup unless you skip prompts with `--yes`.

### `govard env`

Project lifecycle and service wrapper.

```bash
govard env up
govard env start
govard env stop
govard env restart
govard env down
govard env ps
govard env logs php -f
govard env pull
govard env build
```

`govard env up` supports:

- `--pull`
- `--fallback-local-build`
- `--remove-orphans`
- `--quickstart`

`govard env down` supports:

- `-v, --volumes`
- `--rmi local`

### `govard svc`

Manage global services and workspace sleep state.

```bash
govard svc up
govard svc restart --no-trust
govard svc logs --tail 50
govard svc sleep
govard svc wake
```

Global services include proxy, Mailpit, PHPMyAdmin, and Portainer.

### `govard domain`

Manage extra local domains for the current project.

```bash
govard domain add brand-b.test
govard domain remove brand-b.test
govard domain list
```

### `govard status`

List running Govard environments across the workspace.

```bash
govard status
```

### `govard desktop`

Launch the Wails desktop app.

```bash
govard desktop
govard desktop --dev
govard desktop --background
```

Desktop highlights:

- Environment dashboard with start/stop/open
- Project workspace layout (environments, quick actions, onboarding)
- Quick actions (PHPMyAdmin, Xdebug toggle, health)
- Additional quick actions: Mailpit and DB Client shortcuts
- Remotes tab for add/test/open/sync presets
- Logs with service filtering and live streaming
- Shell launcher (service, user, shell)
- Native notifications and settings drawer

For the full desktop surface and dev-mode notes, see [Desktop](desktop.md).

## Development Commands

### `govard shell`

Open a shell in the application container.

```bash
govard shell
govard shell --no-tty
```

### `govard env logs`

Stream project logs.

```bash
govard env logs
govard env logs php
govard env logs -e
govard env logs --tail 200
```

### `govard debug`

Manage Xdebug status and sessions.

```bash
govard debug status
govard debug on
govard debug off
govard debug shell
```

Requests route to `php-debug` only when the `XDEBUG_SESSION` cookie matches `stack.xdebug_session`.

### `govard test`

Run project test tools inside the application container.

```bash
govard test phpunit
govard test phpstan
govard test mftf
govard test integration
```

### `govard config profile`

Inspect or apply version-aware runtime profiles.

```bash
govard config profile
govard config profile --json
govard config profile apply --framework laravel --framework-version 11
```

See [Configuration](configuration.md) and [Frameworks](frameworks.md).

### `govard custom`

Run custom commands from `.govard/commands` or `~/.govard/commands`.

```bash
govard custom list
govard custom hello
govard custom deploy -- --dry-run
```

### `govard projects`

Query the local project registry.

```bash
govard projects open billing
cd "$(govard projects open demo)"
```

## Remote, Sync, and Data Commands

### `govard remote`

Manage named remotes for sync, deploy, shell, and database workflows.

```bash
govard remote add staging --host staging.example.com --user deploy --path /var/www/app
govard remote copy-id staging
govard remote test staging
govard remote exec staging -- ls -la
govard remote audit tail --status failure --lines 50
```

Key features:

- remote capabilities: `files`, `media`, `db`, `deploy`
- auth methods: `keychain`, `ssh-agent`, `keyfile`
- optional strict host-key verification
- production write protection by default
- audit logs in `~/.govard/remote.log`

If you want to use the remote user's home directory, quote the value so your local shell does not expand it first:

```bash
govard remote add staging --host staging.example.com --user deploy --path '~/public_html'
```

### `govard sync`

Synchronize files, media, or databases between local and named remotes.

```bash
govard sync --source staging --destination local --full --plan
govard sync --from staging --to local --media
govard sync -s prod --file --path app/etc/config.php
govard sync -s dev --db --no-noise --no-pii
```

Key flags:

- `--source`, `--from`
- `--destination`, `--to`
- `-e, --environment`
- `--file`, `--media`, `--db`, `--full`
- `--plan`
- `--resume`, `--no-resume`
- `--include`, `--exclude`
- `--no-noise`
- `-P, --no-pii`

### `govard db`

Database utilities for local and remote-backed workflows.

```bash
govard db connect
govard db dump
govard db dump -e staging --local
govard db query "SELECT COUNT(*) FROM sales_order"
govard db info
govard db top
govard db import --file backup.sql --drop
govard db import --stream-db -e staging --drop
```

Notable flags:

- `-e, --environment`
- `--profile`
- `--stream-db`
- `--drop`
- `--local`
- `-N, --no-noise`
- `-P, --no-pii`
- `-S, --sanitize`

### `govard deploy`

Run deploy lifecycle hooks for the current project.

```bash
govard deploy
```

### `govard snapshot`

Create, list, restore, export, and delete local snapshots for DB and media.

```bash
govard snapshot create
govard snapshot list
govard snapshot restore latest
```

### `govard open`

Open common browser and utility targets.

```bash
govard open app
govard open mail
govard open db
govard open db --pma
govard open db --client
govard open db -e staging
```

### `govard tunnel`

Manage public tunnels and automatic base URL registration.

```bash
govard tunnel start
govard tunnel status
govard tunnel stop
```

For deeper remote policy and sync behavior, see [Remotes and Sync](remotes-and-sync.md).

## Tool Commands

### Magento 2

```bash
govard tool magento [command]
```

### Magento 1 (OpenMage)

```bash
govard tool magerun [command]
```

### Laravel

```bash
govard tool artisan [command]
```

### Drupal

```bash
govard tool drush [command]
```

### Symfony

```bash
govard tool symfony [command]
```

### Shopware

```bash
govard tool shopware [command]
```

### CakePHP

```bash
govard tool cake [command]
```

### WordPress

```bash
govard tool wp [command]
```

### General PHP and Node tools

```bash
govard tool composer [command]
govard tool npm [command]
govard tool yarn [command]
govard tool npx [command]
govard tool pnpm [command]
govard tool grunt [command]
```

## Configuration Commands

### `govard config auto`

Auto-inject runtime settings into Magento 2 `app/etc/env.php`.

```bash
govard config auto
```

### `govard config`

Inspect or write Govard config keys.

```bash
govard config get php_version
govard config set php_version 8.4
govard config set stack.php_version 8.3
```

### `govard extensions`

Initialize `.govard/*` extension scaffolding.

```bash
govard extensions init
govard extensions init --force
```

## Diagnostics and Security

### `govard doctor`

Run startup diagnostics with actionable remediation hints.

```bash
govard doctor
govard doctor --fix
govard doctor --json
govard doctor --pack
govard doctor fix-deps
govard doctor trust
```

Checks include Docker, Compose, ports, disk sanity, Govard home readiness, and outbound connectivity.

### `govard doctor fix-deps`

Check host dependencies used by Govard:

- `docker`
- `docker compose`
- `ssh`
- `rsync`

### `govard doctor trust`

Install the Govard Root CA into the system trust store and attempt browser NSS import when possible.

## Utility Commands

### `govard lock`

Generate or validate `govard.lock` snapshots.

```bash
govard lock generate
govard lock check
govard lock generate --file .govard/govard.lock
```

### `govard self-update`

Download release artifacts, verify checksums, and replace installed binaries atomically.

```bash
govard self-update
```

### `govard upgrade`

Framework version upgrade entrypoint.

```bash
govard upgrade
```

### `govard version`

Print Govard version information.

```bash
govard version
```

## Global Flags

All commands support:

- `-h, --help`
