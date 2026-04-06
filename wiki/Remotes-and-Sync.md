# Remotes and Sync

This is the canonical guide for Govard remote environments, sync operations, and remote-backed database workflows.

---

## Remote Setup

### Add a Remote

```bash
govard remote add staging --host staging.example.com --user deploy --path /var/www/app
```

**`remote add` flags:**

| Flag | Description |
| :--- | :--- |
| `--host` | Remote hostname or IP |
| `--user` | SSH username |
| `--path` | Remote project path |
| `--port` | SSH port (default: 22) |
| `--capabilities` | Comma-separated capabilities (`files,media,db,deploy`) |
| `--auth-method` | Auth method (`keychain`, `ssh-agent`, `keyfile`) |
| `--key-path` | Path to SSH key (for `keyfile` method) |
| `--strict-host-key` | Enable strict host-key verification |
| `--known-hosts-file` | Custom known_hosts file |
| `--protected` | Write-protect this remote |

> [!TIP]
> To use the remote user's home directory, quote the path so the local shell does not expand it:
> ```bash
> govard remote add staging --host staging.example.com --user deploy --path '~/public_html'
> ```

### Validate Connectivity

```bash
govard remote copy-id staging    # Copy your SSH public key to remote authorized_keys
govard remote test staging       # Validate SSH + rsync, measure latency, classify failures
```

`remote test` identifies failure types: `network`, `auth`, `permission`, `host_key`, `dependency`.

### Exec and Audit

```bash
govard remote exec staging -- ls -la
govard remote audit tail --lines 20
govard remote audit tail --status failure --lines 50
govard remote audit stats --lines 200
```

**Audit log paths:**
- `~/.govard/remote.log`
- `~/.govard/operations.log`

---

## Remote Safety Model

| Protection | Behavior |
| :--- | :--- |
| **Production write protection** | `prod` remotes are write-protected by default |
| **Capability enforcement** | Each operation checks `files`, `media`, `db`, `deploy` scopes |
| **Strict host-key** | Opt-in per remote, not enforced by default |
| **1Password integration** | Remote fields support `op://...` secret references |

---

## Sync Overview

`govard sync` moves files, media, and database data between local and named remotes.

```bash
govard sync --source staging --destination local --full --plan
govard sync --from staging --to local --media
govard sync -s dev --db --no-noise --no-pii
govard sync -s prod --file --path app/etc/config.php
```

Auto-selects `staging` if no `--source` provided, falling back to `dev`.

### Endpoint Flags

| Flag | Description |
| :--- | :--- |
| `-s, --source` / `--from` | Source environment |
| `-d, --destination` / `--to` | Destination environment |
| `-e, --environment` | Alias for `--source` |

### Scope Flags

| Flag | Syncs |
| :--- | :--- |
| `-A, --full` | Everything (files + media + database) |
| `-f, --file` | Source code and generic files |
| `-m, --media` | Framework-specific media assets |
| `-b, --db` | Project database |

### Transfer Flags

| Flag | Description |
| :--- | :--- |
| `--plan` | Print plan and exit without executing |
| `-D, --delete` | Delete destination files missing from source |
| `-R, --resume` | Enable resumable transfers (default: `true`) |
| `--no-resume` | Disable resumable transfers |
| `-C, --no-compress` | Disable rsync compression |
| `-y, --yes` | Skip confirmation prompts |
| `-p, --path` | Specific file/directory relative to project root |
| `-I, --include` | Rsync include pattern (repeatable) |
| `-X, --exclude` | Rsync exclude pattern (repeatable) |

### Database Privacy Filters

| Flag | Category | Magento 2 Exclusions | Laravel | WordPress |
| :--- | :--- | :--- | :--- | :--- |
| `--no-noise` | Ephemeral data | `cron_schedule`, `session`, `cache_tag`, `report_event` | `cache`, `sessions`, `failed_jobs` | `redirection_404`, `wflogs` |
| `--no-pii` | Sensitive data | `customer_entity`, `sales_order`, `quote`, `admin_user` | `users`, `password_resets` | `users`, `usermeta`, `comments` |

> [!NOTE]
> Database filters are optimized for Magento 2. For other frameworks, safe default patterns are used when available.

---

## Sync Behavior

### Resumable Transfers

File and media sync use resumable rsync mode by default (`--partial` + `--append-verify`).

```bash
govard sync -s staging --file        # resumable by default
govard sync -s staging --file --no-resume  # disable resumable
```

### Include and Exclude Filters

`--include` and `--exclude` apply only to `-f, --file` and `-m, --media` scopes — they are ignored for DB-only sync.

### Protected Destinations

> [!WARNING]
> `--delete` combined with `--db` surfaces policy warnings. Production remotes are write-protected by default and will block destructive writes.

### Integration with `bootstrap`

`govard bootstrap` uses the same sync/filter flags for environment initialization:

```bash
govard bootstrap --clone -e staging --no-pii --no-noise --delete
```

---

## Remote Name Resolution

Remote flags support:
- Exact remote key lookup
- Normalized aliases (e.g. `stg` → `staging`)

### Auto-Select Priority

When no remote is specified for `bootstrap` or `sync`, Govard resolves:

1. **`staging`** (aliases: `stg`, `stage`, `qa`, `uat`, `test`)
2. **`dev`** (aliases: `development`, `local`)

These are equivalent when the `staging` remote exists:

```bash
govard sync -s stg --db
govard sync --source staging --db
govard sync --from staging --db
```

---

## Remote Snapshots

```bash
govard snapshot create -e staging
govard snapshot list -e staging
govard snapshot restore latest -e staging
govard snapshot delete latest -e staging
```

Remote snapshots run `mysqldump` and `tar` directly on the remote server without transferring data over the network. Stored in `~/.govard/snapshots/` within the remote project path.

### Bidirectional Transfer

```bash
# Pull from staging to local
govard snapshot pull before-upgrade -e staging

# Push local snapshot to production (blocked by default protection policy)
govard snapshot push fallback-state -e prod
```

---

## Remote Database Workflows

### Dump

```bash
govard db dump                        # Local DB to project var/
govard db dump -e staging             # Remote DB → saved on remote (~backup/)
govard db dump -e staging --local     # Remote DB → streamed to local var/
govard db dump --no-noise --no-pii    # With privacy filters
```

### Import

```bash
govard db import --file backup.sql --drop
govard db import --stream-db -e staging --drop
```

`--stream-db` pulls from the remote and imports into the local database. `--drop` performs a safe reset before import.

### Query, Info, and Live Monitoring

```bash
govard db query "SELECT COUNT(*) FROM sales_order"
govard db info -e staging
govard db top -e staging    # Live process monitoring
```

---

## Desktop Remote Actions

Desktop remote actions call the same backend paths as CLI commands:

- **Open Database (Remote)** → `govard open db -e <remote> --client`
- **Open SSH (Remote)** → native Linux terminal launchers, fallback to `ssh://`
- **Open SFTP (Remote)** → prefers FileZilla, fallback to `sftp://`

For `auth.method: ssh-agent`, Govard reuses `SSH_AUTH_SOCK` and probes `/run/user/<uid>/keyring/ssh` on Linux.

---

## Recommended Patterns

**Safe review before execution:**

```bash
govard sync --source staging --destination local --full --plan
```

**Target one file:**

```bash
govard sync --source prod --file --path app/etc/config.php
```

**Create a local-safe DB snapshot:**

```bash
govard db dump -e staging --local --no-noise --no-pii
```

**Full workflow: clone from staging with privacy:**

```bash
govard bootstrap --clone -e staging --no-pii --no-noise --yes
```

---

**[← Frameworks](Frameworks)** | **[SSL and Domains →](SSL-and-Domains)**
