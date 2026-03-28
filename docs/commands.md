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
- `govard prj` -> `govard project`
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
govard bootstrap --clone --environment staging --yes
govard bootstrap --framework magento2 --fresh --framework-version 2.4.9
govard bootstrap -e staging --no-pii --no-noise
```

### Mode and Framework Selection

Two primary ways to use `bootstrap`:

- **Fresh Install**: Use `--fresh` with `--framework` and `--framework-version` to create a clean, vanilla project from a composer meta-package.
- **Remote Clone**: Use `--clone` with `--environment` to rsync the entire source code from a remote server. Use this when you don't have a local git repository.

### Source Selection

- `-e, --environment`: The source remote name (e.g., `staging`, `production`, `dev`). Govard auto-selects `staging` or `dev` if omitted.
- `--remote`: Alias for `--environment`.
- `--db-dump`: Import the database from a local SQL file path rather than syncing from a remote.

### Interactive Confirmation and Plan Preview

Bootstrap shows a **Review Plan** step before starting sync or environment operations. This summary includes:
- Source and Destination endpoints
- Scopes (Code, DB, Media)
- Risks (e.g., if local files will be overwritten or DB reset)

You must explicitly confirm the plan to proceed. Use the `-y, --yes` flag to skip this check in CI or non-interactive environments.

Use `--plan` to print the summary and exit immediately without starting any services or transfers.

### Privacy and Performance Filters

- `-N, --no-noise`: Excludes ephemeral data (logs, session tables, cache tags, cron history).
- `-S, --no-pii`: Excludes sensitive information (customers, orders, admin users, passwords).
- `--delete`: Deletes files on the destination that do not exist on the source (media/file sync).
- `--no-compress`: Disables rsync compression (useful if CPU is a bottleneck).
- `--exclude`: Pass custom rsync-style patterns to exclude specific directories or files.
- `--no-db`: Skip database import entirely.
- `--no-media`: Skip media asset sync entirely.
- `--no-composer`: Skip automated `composer install`.
- `--no-admin`: Skip admin user creation (Magento 2 only).
- `--no-stream-db`: Disable piped DB transfer (use local temp file).

### Framework Special (Magento 2 / Magento 1)

- `--include-sample`: Install sample data during a fresh install.
- `--hyva-install`: Automatically install the Hyva theme during bootstrap.
- `--include-product`: Specifically sync catalog product images during media sync.
- `--fix-deps`: Run a project-specific `fix-deps` command before bootstrap starts.

During remote bootstrap flows, Govard runs `govard config auto` automatically for Magento 2, Magento 1, and OpenMage unless you use `--skip-up`.

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
govard env cleanup
```

`govard env up` supports:

- `--pull`
- `--fallback-local-build`
- `--remove-orphans`
- `--quickstart`
- `--update-lock`: automatically update `govard.lock` on mismatches

Before starting containers, `govard env up` re-renders the per-project compose file and the generated web-server config assets under `~/.govard/`, including:

- `~/.govard/compose/<project-hash>.yml`
- `~/.govard/nginx/<project>/default.conf`
- `~/.govard/apache/<project>/httpd.conf`
- `~/.govard/nginx/<project>/mage-run-map.conf`
- `~/.govard/apache/<project>/mage-run-map.conf`

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

## Blueprint Commands

### `govard blueprint cache`

Manage the remote blueprint registry cache.

```bash
govard blueprint cache list
govard blueprint cache clear
```

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

### `govard project`

Browse and manage known projects from the registry.

Aliases: `project`, `prj`, `projects`, `registry`

```bash
# List all available projects
govard project list

# Get the path of a project by name
govard project open billing

# Delete a project and its environment resources
govard project delete demo
```

Use `govard project list` to see all known projects, their framework, and paths.

### `govard project delete`

Completely remove a project's orchestration resources and unregister it from Govard. This command is resilient and can clean up "ghost" projects even if their `.govard.yml` configuration is missing.

```bash
govard project delete demo
govard project delete --yes demo
```

The deletion process:
1. Runs `pre-delete` lifecycle hooks (skipped if config is missing).
2. Executes `docker compose down -v` to remove containers and **volumes** (databases, etc.).
3. Unregisters any proxy domains (skipped if config is missing).
4. Removes the project from the registry (`projects.json`).
5. Runs `post-delete` lifecycle hooks (skipped if config is missing).

> [!CAUTION]
> Deletion removes persistent database volumes by default. Project source code is **never** deleted.

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
govard sync --db --no-noise --no-pii
```

By default, if no `--source` is provided, Govard **auto-selects** your `staging` remote if it exists, falling back to `dev` (or its aliases).

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

Magento CLI runs inside the PHP container as the project user. This includes `govard tool magento cron:install`, which installs crontab entries inside that container.

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

If `.govard.yml` includes `store_domains`, `govard config auto` also attempts scoped base URL updates for Magento projects:

- Magento 2: `bin/magento config:set` using `--scope=stores` by default, or `--scope=websites` when `store_domains.<host>.type=website`
- Magento 1 / OpenMage: scoped `core_config_data` updates using website/store codes, with explicit `type` respected when provided

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

Checks include Docker, Compose, ports, disk sanity, Govard home readiness, compose directory saturation, and outbound connectivity.

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
govard lock diff
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
