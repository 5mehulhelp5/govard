# CLI Command Reference

Complete reference for all Govard commands.

## Environment Commands

### `govard init`

Scans for `composer.json` or `package.json`, detects framework, and generates `govard.yml`.

```bash
govard init
govard init --recipe magento2
govard init -r laravel
govard init -r custom
```

**Detection Logic:**
- Magento 1/OpenMage: `openmage/magento-lts` in composer.json or `app/Mage.php`
- Magento 2: `magento/product-community-edition` in composer.json
- Laravel: `laravel/framework` in composer.json
- Next.js: `next` in package.json dependencies
- Drupal: `drupal/core` in composer.json
- Symfony: `symfony/framework-bundle` in composer.json
- Shopware: `shopware/core` in composer.json
- CakePHP: `cakephp/cakephp` in composer.json
- WordPress: `johnpbloch/wordpress` in composer.json

**Options:**
- `-r, --recipe` Override detected framework (useful for empty projects or when starting a new app)
- `custom` recipe opens an interactive prompt to choose web server, DB, cache, search, and varnish

### `govard bootstrap`

Bootstrap project local setup with clone/fresh workflows (Magento and other supported recipes).

```bash
govard bootstrap
govard bootstrap --environment dev
govard bootstrap --fresh --version 2.4.8
```

Highlights:
- Auto-runs `govard init` if `govard.yml` is missing
- Clone flow: file sync, optional composer install, DB/media sync, Magento configure, admin create
- Fresh flow: create-project, setup install, optional sample data, optional Hyva install
- Fresh mode does not require `--clone=false`; use `govard bootstrap --fresh ...` directly
- Supports legacy bootstrap-compatible flags (`--download-source`, `--skip-db-import`, etc.)

See `docs/commands/bootstrap.md`.

### `govard up`

Renders the compose file at `~/.govard/compose/<project>-<hash>.yml` and starts Docker containers in detached mode.

```bash
govard up
govard up --quickstart
```

**Process:**
1. `Detect` framework context from project files
2. `Validate` layered config and startup prerequisites (Docker, Compose, ports, disk/network sanity)
3. `Render` compose file (`~/.govard/compose/...`) and pre-up hooks
4. `Start` containers via Docker Compose
5. `Verify` hosts/proxy wiring and post-up hooks

On failure, `govard up` prints a suggested next command such as `govard doctor` or `govard deps`.

`--quickstart` applies a minimal runtime profile for the current startup (disables optional cache/search/queue/varnish/xdebug services) to reduce first-run time.

### `govard stop`

Stops all project containers without removing them.

```bash
govard stop
```

### `govard down`

Tear down project containers and networks.

```bash
govard down
govard down --volumes
govard down --rmi local --timeout 20
```

See `docs/commands/down.md`.

### `govard sleep`

Stop all running Govard projects detected from Docker and persist a wake snapshot.

```bash
govard sleep
```

Sleep state is saved to `~/.govard/sleep-state.json` (override via `GOVARD_SLEEP_STATE_PATH`).
Only projects found in the project registry are included in the saved state.

### `govard wake`

Start all projects recorded by the latest sleep snapshot.

```bash
govard wake
```

When wake succeeds for all projects, the sleep-state file is removed.
If some projects fail to start, only failed entries are kept for the next retry.

### `govard proxy`

Start, stop, restart, or check status of the global Caddy proxy.

```bash
govard proxy start
govard proxy stop
govard proxy restart
govard proxy status
govard proxy routes
```

### `govard desktop`

Launch the Govard Desktop app (Wails-based).

```bash
govard desktop
govard desktop --dev
govard desktop --background
```

`--dev` runs Wails dev mode and requires the Wails CLI. Running without `--dev`
expects a built `govard-desktop` binary.
`--background` starts desktop hidden, keeps the process alive when the window is
closed, and reuses the running instance on relaunch.

Desktop highlights:
- Environment dashboard with start/stop/open
- Project workspace layout (environments, quick actions, onboarding)
- Quick actions (PHPMyAdmin, Xdebug toggle, health)
- Manual project onboarding (select folder and add/init project)
- Remotes tab (list/add remotes, run remote test, trigger sync plan presets)
- Resource monitoring (CPU/RAM/NET and OOM hints)
- Logs with multi-service selection, severity/text filtering, and live streaming
- Shell launcher (service, user, shell)
- Native notifications for operation success/failure updates
- Settings drawer (theme, proxy target, preferred browser)

### `govard status`

Lists all running Govard environments.

```bash
govard status
```

## Development Commands

### `govard shell`

Opens a bash shell in the PHP container.

```bash
govard shell
```

**User:** Runs as `www-data` for all frameworks.

### `govard logs`

Streams container logs.

```bash
govard logs          # All logs
govard logs -e       # Error-only filtering
```

### `govard debug`

Toggle Xdebug 3.

```bash
govard debug on      # Enable Xdebug
govard debug off     # Disable Xdebug
govard debug status  # Check Xdebug status
govard debug shell   # Open shell in php-debug container
```

When enabled, browser requests are routed to `php-debug` only if the `XDEBUG_SESSION`
cookie matches `stack.xdebug_session`.

### `govard profile`

Inspect or apply framework-aware runtime profile selection.

See `docs/commands/profile.md`.
See `docs/frameworks/support-matrix.md` for version-specific profile mappings.

### `govard custom`

Run custom commands from:
- project scope: `.govard/commands`
- global scope: `~/.govard/commands` (override with `GOVARD_GLOBAL_COMMANDS_DIR`)

```bash
govard custom list
govard custom hello
govard custom deploy -- --dry-run
```

Conflict rule: project command names take precedence over global commands.
Custom commands are namespaced under `govard custom` to avoid conflicts with built-in commands.

### `govard projects`

Query known projects from the local project registry.

```bash
govard projects open billing
cd "$(govard projects open demo)"
```

`projects open <query>` performs fuzzy matching across `project_name`, domain, and path,
then prints the matched project path to stdout.

## Remote & Deployment Commands

### `govard remote`

Manage remote environments (add, exec, test).
Supports remote environment classification (`dev`, `staging`, `prod`) and per-remote capabilities (`files`, `media`, `db`, `deploy`).
Supports auth method selection (`keychain`, `ssh-agent`, `keyfile`) and SSH key-path overrides.
Supports optional strict SSH host key verification per remote.
Supports `op://...` secret references in remote host/user/path and SSH path fields via 1Password CLI integration.
Writes remote operation audit events to `~/.govard/remote.log`.
Includes `remote audit tail` and `remote audit stats` for querying and summarizing recent remote audit events.
Audit query commands support `--since` and `--until` time-window filters.

See `docs/commands/remote.md`.

### `govard sync`

Synchronize files, media, and databases between environments.
Supports rsync include/exclude filters via repeatable `--include` and `--exclude` flags.
Uses resumable rsync mode by default for file/media transfers (`--resume`, disable via `--no-resume`).

See `docs/commands/sync.md`.

### `govard db`

Database utilities with subcommands `connect`, `dump`, and `import`.
Supports remote-source streaming for local imports (`db import --stream-db -e <remote>`)
and file mode with `--file`.

See `docs/commands/db.md`.

### `govard deploy`

Run deploy lifecycle hooks for the current project.

Current behavior:
- Executes `pre_deploy` and `post_deploy` hooks.
- Prints `Deploying (strategy: native)`.
- Strategy flags are accepted for forward compatibility and do not currently change execution behavior.

See `docs/commands/deploy.md`.

### `govard snapshot`

Create, list, and restore local snapshots for database and media.

See `docs/commands/snapshot.md`.

### `govard open`

Open common service URLs in your browser.

See `docs/commands/open.md`.

### `govard tunnel`

Create a public tunnel for your local project URL.
Current provider support: Cloudflare (`cloudflared`) quick tunnels.

See `docs/commands/tunnel.md`.

## Framework Commands

### Magento 2

```bash
govard magento [command]     # Run Magento CLI (bin/magento)
```

### Magento 1 (OpenMage)

```bash
govard magerun [command]     # Run n98-magerun
```

### Laravel

```bash
govard artisan [command]     # Run Artisan commands
```

### Drupal

```bash
govard drush [command]       # Run Drush commands
```

### Symfony

```bash
govard symfony [command]     # Run Symfony CLI commands
```

### Shopware

```bash
govard shopware [command]    # Run Shopware CLI commands
```

### CakePHP

```bash
govard cake [command]        # Run CakePHP CLI commands
```

### WordPress

```bash
govard wp [command]          # Run WordPress CLI commands
```

### General PHP

```bash
govard composer [command]    # Run Composer
```

### Node.js

```bash
govard npm [command]         # Run npm
govard yarn [command]        # Run yarn
govard npx [command]         # Run npx
govard pnpm [command]        # Run pnpm
govard grunt [command]       # Run grunt
```

## Service Commands

### Database

```bash
govard db                    # Database utilities (connect, dump, import)
```

### Redis

```bash
govard redis                 # Interact with Redis (redis-cli)
govard valkey                # Interact with Valkey (valkey-cli)
```

### Search

```bash
govard elasticsearch         # Query Elasticsearch API (curl)
govard opensearch            # Query OpenSearch API (curl)
```

### Caching

```bash
govard varnish               # Varnish management commands
```

### Mail

```bash
govard mail                  # Open Mailpit web interface
```

### Database Admin

```bash
govard pma                   # Open PHPMyAdmin interface
```

## Configuration Commands

### `govard configure`

Auto-injects stack settings into application config (Magento 2 only).

```bash
govard configure
```

**Configures:**
- Database connection
- Redis cache and sessions
- Varnish page cache
- Elasticsearch/OpenSearch
- Base URLs

### `govard config`

Manage Govard configuration.

```bash
govard config [subcommand]
```

### `govard extensions`

Manage extension scaffolding for `.govard/*`.

```bash
govard extensions init
govard extensions init --force
```

## Diagnostics & Security

### `govard doctor`

Runs startup diagnostics with actionable remediation hints.

```bash
govard doctor
govard doctor --fix
govard doctor --json
govard doctor --pack
```

**Checks:**
- Docker Desktop/Daemon connectivity
- Docker Compose plugin availability
- Port conflicts on host machine (80/443)
- Disk scratch write sanity
- Govard home directory readiness (`~/.govard`)
- Outbound network probe sanity

**Output:**
- Each check is reported as `pass`, `warn`, or `fail`.
- `fail` checks return a non-zero exit code.
- `--json` prints a machine-readable diagnostics report.
- `--fix` applies explicit safe remediations when available and prints each action taken.
- `--pack` exports a diagnostics support pack zip (report, environment metadata, config/compose snapshots).
- `--pack-dir` overrides output directory (default: `~/.govard/diagnostics`).
- Pack also includes best-effort runtime command snapshots and remote audit log snapshot when available.
- JSON report includes `issue_cards` for warn/fail checks, intended for Desktop troubleshooting card rendering.

### `govard deps`

Checks required host dependencies used by Govard (`docker`, `docker compose`, `ssh`, `rsync`).

```bash
govard deps
```

`govard fix-deps` remains available as a compatibility alias.

### `govard trust`

Installs the Caddy CA into the system trust store for HTTPS.

```bash
govard proxy start
govard trust
```

Notes:
- On Linux, this exports certs to `~/.govard/ssl/root.crt`.
- On macOS, export the cert first: `docker cp proxy-caddy-1:/data/caddy/pki/authorities/local/root.crt /tmp/govard-ca.crt`.

## Utility Commands

### `govard lock`

Generate and validate `govard.lock` snapshots for team environment consistency.

```bash
govard lock generate
govard lock check
govard lock generate --file .govard/govard.lock
```

Behavior:
- `lock generate` captures current Govard/Docker toolchain values and runtime stack metadata.
- `lock check` compares current environment values against the lock file and reports mismatches.
- `govard up` performs a non-blocking lockfile warning check when `govard.lock` exists.
- Set `lock.strict: true` in `govard.yml` to make `govard up` fail on lock mismatches
  (and fail when the lock file is missing).

See `docs/commands/lock.md`.

### `govard self-update`

Checks for and installs the latest Govard binary.

```bash
govard self-update
```

**Process:**
1. Queries GitHub API for latest release
2. Downloads new binary
3. Replaces current binary

### `govard upgrade`

Upgrade framework version (placeholder).

```bash
govard upgrade
```

See `docs/commands/upgrade.md`.

### `govard version`

Display version information.

```bash
govard version
```

### `govard completion`

Generate shell completion scripts.

```bash
govard completion bash
govard completion zsh
govard completion fish
```

## Global Flags

All commands support:

- `--help` / `-h` - Show help
