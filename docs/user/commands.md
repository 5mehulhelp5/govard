# CLI Command Reference

Complete reference for all Govard commands.

## Environment Commands

### `govard init`

Scans for `composer.json` or `package.json`, detects framework, and generates `.govard.yml`.

```bash
govard init
govard init --framework magento2
govard init --framework laravel
govard init --framework custom
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

- `--framework` Override detected framework (useful for empty projects or when starting a new app)
- `custom` framework opens an interactive prompt to choose web server, DB, cache, search, and varnish

### `govard bootstrap`

Bootstrap project local setup with clone/fresh workflows (Magento and other supported frameworks).

```bash
govard bootstrap
govard bootstrap --fresh --framework-version 2.4.8
```

Highlights:

- Auto-runs `govard init` if `.govard.yml` is missing
- Clone flow: file sync, optional composer install, DB/media sync, Magento configure, admin create
- Fresh flow: create-project, setup install, optional sample data, optional Hyva install
- Fresh mode does not require `--clone=false`; use `govard bootstrap --fresh ...` directly

See [env.md](file:///home/kai/Work/htdocs/ddtcorex/govard/docs/commands/env.md).

### `govard env up`

Renders the compose file at `~/.govard/compose/<project>-<hash>.yml` and starts Docker containers in detached mode.

```bash
govard env up
govard env up --pull
govard env up --fallback-local-build
govard env up --remove-orphans
govard env up --quickstart
```

**Process:**

1. `Detect` framework context from project files
2. `Validate` layered config and startup prerequisites (Docker, Compose, ports, disk/network sanity)
3. `Render` compose file (`~/.govard/compose/...`) and pre-up hooks
4. `Start` containers via Docker Compose
5. `Verify` hosts/proxy wiring and post-up hooks

On failure, `govard env up` prints a suggested next command such as `govard doctor` or `govard doctor fix-deps`.

**Options:**

- `--pull` Pull latest images from the registry before starting containers.
- `--fallback-local-build` When pull/start fails due to missing Govard-managed images, build missing Govard-managed images locally and retry once.
- `--remove-orphans` Remove containers for services not defined in the compose file.
- `--quickstart` applies a minimal runtime profile for the current startup (disables optional cache/search/queue/varnish/xdebug services) to reduce first-run time.

### `govard env` (alias: `project`)

Project-scoped lifecycle and service wrapper command.

```bash
govard env up
govard env start
govard env stop
govard env pull
govard env down
govard env restart
govard env ps
govard env logs [service] [-f]
```

Govard intelligently proxies almost all Docker Compose maintenance commands (ps, logs, stop, start, restart, pull, build, etc.) to the current project context.

Service access under project scope (now with smart proxying):

```bash
govard redis [command]
govard valkey [command]
govard elasticsearch [path|command]
govard opensearch [path|command]
govard varnish [log|ban|stats|command]
```

### `govard env stop`

Stops all project containers without removing them.

```bash
govard env stop
```

### `govard env pull`

Pull latest project images from the registry.

```bash
govard env pull
```

### `govard env down`

Tear down project containers and networks.

```bash
govard env down
govard env down --volumes
govard env down --rmi local --timeout 20
```

See [env.md](file:///home/kai/Work/htdocs/ddtcorex/govard/docs/commands/env.md).

### `govard domain`

Manage additional domains for the project.

```bash
govard domain add <domain>
govard domain remove <domain>
govard domain list
```

See `docs/commands/domain.md`.

### `govard svc`

Manage global services and workspace-wide sleep state.

```bash
govard svc up
govard svc up --pull
govard svc up --fallback-local-build
govard svc up --remove-orphans
govard svc up --auto-trust=false
govard svc up --trust-browsers=false
govard svc down
govard svc down --remove-orphans
govard svc restart
govard svc restart --pull
govard svc restart --fallback-local-build
govard svc restart --remove-orphans
govard svc pull
govard svc ps
govard svc logs
govard svc logs --tail 50
govard svc sleep
govard svc wake
```

`svc up`, `svc down`, and `svc restart` manage the global services suite (proxy, mailpit, phpmyadmin, and portainer).
By default, `svc up` and `svc restart` also auto-trust Govard Root CA on the host machine.
Set `--auto-trust=false` to skip CA trust, or `--trust-browsers=false` to skip browser NSS import.
`--fallback-local-build` allows `svc up/restart` to build missing Govard-managed images locally and retry startup once.

`svc sleep` stops all running Govard projects detected from Docker and persists wake state to
`~/.govard/sleep-state.json` (override with `GOVARD_SLEEP_STATE_PATH`).
`svc wake` restores the recorded projects and clears the state file when all wake operations succeed.

### `govard desktop`

Launch the Govard Desktop app (Wails-based).

```bash
govard desktop
govard desktop --dev
govard desktop --background
```

`--dev` runs Wails dev mode and requires the Wails CLI. Running without `--dev`
expects a built `govard-desktop` binary.
For source/manual builds, use production tags (Linux: `desktop production webkit2_41`,
macOS: `desktop production`) to avoid Wails startup tag errors.
`--background` starts desktop hidden, keeps the process alive when the window is
closed, and reuses the running instance on relaunch.

Desktop highlights:

- Environment dashboard with start/stop/open
- Project workspace layout (environments, quick actions, onboarding)
- Quick actions (PHPMyAdmin, Xdebug toggle, health)
- Additional quick actions: Mailpit and DB Client shortcuts
- Project onboarding supports:
  - local add/init flow (select folder and onboard)
  - optional Git clone onboarding (SSH/HTTPS URL)
  - pre-clone Git connection validation
  - explicit folder-override confirmation before clearing selected folder contents
  - step progress updates during onboarding (validate, prepare folder, clone, init)
- Remotes tab (list/add remotes, run remote test, open SSH/DB/SFTP, trigger sync plan presets)
- Resource monitoring (CPU/RAM/NET and OOM hints)
- Logs with multi-service selection, severity/text filtering, and live streaming
- Shell launcher (service, user, shell)
- Native notifications for operation success/failure updates
- Settings drawer (theme, proxy target, preferred browser)
- Background update check after app startup (then periodic checks) with bottom-right update prompt and one-click install flow

Desktop remote open behavior:

- Open SSH (Remote) on Linux prefers native terminal launchers (`x-terminal-emulator`, `gnome-terminal`, `konsole`, `xfce4-terminal`) and falls back to `ssh://` URL handoff when needed.
- Open SFTP (Remote) prefers FileZilla when available and falls back to standard `sftp://` URL handoff.
- Open Database (Remote) uses `govard open db -e <remote> --client`.
- Desktop remote selection and bootstrap prompt use configured remote names directly.
- For remotes with `auth.method: ssh-agent`, desktop open actions reuse `SSH_AUTH_SOCK` (Linux also probes `/run/user/<uid>/keyring/ssh`).
- If FileZilla asks for a password, load your key into ssh-agent and set remote auth method to `ssh-agent`.

### `govard status`

Lists running Govard project environments across the workspace.
Use `govard env ps` for current-project container status.

```bash
govard status
```

## Development Commands

### `govard shell`

Opens a bash shell in the PHP container.

```bash
govard shell
govard shell --no-tty
```

**User:** Runs as `www-data` for all frameworks.

### `govard env logs`

Streams container logs.

```bash
govard env logs      # All logs
govard env logs php  # Stream logs from PHP container only
govard env logs -e   # Error-only filtering
govard env logs --tail 200
```

### `govard test`

Run project testing frameworks within the application container.

```bash
govard test phpunit
govard test phpstan
govard test mftf
govard test integration
```

See `docs/commands/test.md`.

Manage Xdebug settings and debugging sessions (status, on, off, shell).

```bash
govard debug on      # Enable Xdebug
govard debug off     # Disable Xdebug
govard debug status  # Check Xdebug status
govard debug shell   # Open shell in php-debug container
```

When enabled, browser requests are routed to `php-debug` only if the `XDEBUG_SESSION`
cookie matches `stack.xdebug_session`.

### `govard config profile`

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

Manage remote environments (add, copy-id, exec, test).
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

Database utilities with subcommands `connect`, `dump`, `import`, `query`, `info`, and `top`.
Supports remote-source streaming for local imports (`db import --stream-db -e <remote>`),
file mode with `--file`, SQL query execution (`db query "SELECT ..."`), connection info (`db info`),
and real-time process monitoring (`db top`).

See `docs/commands/db.md`.

### `govard deploy`

Run deploy lifecycle hooks for the current project.

Current behavior:

- Executes `pre_deploy` and `post_deploy` hooks.
- Prints `Deploying (strategy: native)`.
- Strategy flags are accepted for forward compatibility and do not currently change execution behavior.

See `docs/commands/deploy.md`.

### `govard snapshot`

Create, list, and restore local snapshots for database and media. Supports export and delete operations.

See `docs/commands/snapshot.md`.

### `govard open`

Open common service URLs in your browser.

See `docs/commands/open.md`.

### `govard tunnel`

Manage public tunnels and automatic base URL registration.

```bash
govard tunnel start
govard tunnel status
govard tunnel stop
```

See `docs/commands/tunnel.md`.

## Tool Commands

### Magento 2

```bash
govard tool magento [command]     # Run Magento CLI (bin/magento)
```

### Magento 1 (OpenMage)

```bash
govard tool magerun [command]     # Run n98-magerun
```

### Laravel

```bash
govard tool artisan [command]     # Run Artisan commands
```

### Drupal

```bash
govard tool drush [command]       # Run Drush commands
```

### Symfony

```bash
govard tool symfony [command]     # Run Symfony CLI commands
```

### Shopware

```bash
govard tool shopware [command]    # Run Shopware CLI commands
```

### CakePHP

```bash
govard tool cake [command]        # Run CakePHP CLI commands
```

### WordPress

```bash
govard tool wp [command]          # Run WordPress CLI commands
```

### General PHP

```bash
govard tool composer [command]    # Run Composer
```

### Node.js

```bash
govard tool npm [command]         # Run npm
govard tool yarn [command]        # Run yarn
govard tool npx [command]         # Run npx
govard tool pnpm [command]        # Run pnpm
govard tool grunt [command]       # Run grunt
```

## Service Commands

### Database

```bash
govard db                    # Database utilities (connect, dump, import)
```

### Redis & Valkey

Manage Redis or Valkey cache. Supports custom utilities and smart maintenance proxy.

```bash
govard redis cli      # Open CLI
govard redis flush    # Flush all keys
govard redis info     # Show info
govard redis ps       # Check status (smart proxy)
govard redis logs -f  # Tail logs (smart proxy)
```

See [env.md](file:///home/kai/Work/htdocs/ddtcorex/govard/docs/commands/env.md).

### Search

```bash
govard env elasticsearch     # Project-scoped Elasticsearch (curl)
govard env opensearch        # Project-scoped OpenSearch (curl)
```

### Caching

```bash
govard env varnish           # Project-scoped Varnish commands
```

### Mail

```bash
govard open mail             # Mailpit target
```

### Database Admin

```bash
govard open db               # Generic DB target (opens PMA or Client according to Settings)
govard open db --pma         # Force open local PHPMyAdmin
govard open db --client      # Force open local Desktop Client (Beekeeper Studio)
govard open db -e staging    # Remote DB via SSH tunnel
```

Notes:

- Use `open db -e <remote>` for remote DB access.

## Configuration Commands

### `govard config auto`

Auto-injects stack settings into application config (Magento 2 only).

```bash
govard config auto
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

Common subcommands:

```bash
govard config get php_version
govard config set php_version 8.4
govard config set stack.php_version 8.3  # Nested keys supported
govard config auto
govard config profile
govard config profile apply
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
govard doctor trust
govard doctor fix-deps
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

### `govard doctor fix-deps`

Checks required host dependencies used by Govard (`docker`, `docker compose`, `ssh`, `rsync`).

```bash
govard doctor fix-deps
```

### `govard doctor trust`

Installs the Caddy CA into the system trust store for HTTPS.

```bash
govard svc up
govard doctor trust
```

Notes:

- Exports cert to `~/.govard/ssl/root.crt`.
- On Linux, browser NSS import is attempted automatically when `certutil` is available.
- This command is also run automatically by default during `govard svc up` and `govard svc restart`.

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
- `govard env up` performs a non-blocking lockfile warning check when `govard.lock` exists.
- Set `lock.strict: true` in `.govard.yml` to make `govard env up` fail on lock mismatches
  (and fail when the lock file is missing).

See `docs/commands/lock.md`.

### `govard self-update`

Checks for and installs the latest Govard binaries.

```bash
govard self-update
```

**Process:**

1. Queries GitHub API for latest release
2. Downloads release archives for `govard` and detected `govard-desktop` (on Linux, falls back to release `.deb` for desktop when standalone desktop archive is unavailable)
3. Verifies each archive against `checksums.txt`
4. Replaces installed binaries atomically

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

## Global Flags

All commands support:

- `--help` / `-h` - Show help
