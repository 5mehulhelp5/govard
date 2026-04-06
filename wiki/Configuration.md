# Configuration

Govard uses layered project configuration plus framework blueprints.

---

## Config Layer Order

Govard loads config in this order (later layers override earlier ones):

| Priority | File | Description |
| :---: | :--- | :--- |
| 1 | `.govard.yml` | Base team config — main writable file |
| 2 | `.govard.<profile>.yml` | Team-shared profile override |
| 3 | `.govard.local.yml` | Legacy developer-local override |
| 4 | `.govard/.govard.local.yml` | **Preferred** developer-local override |
| 5 | `.govard.<env>.yml` | Legacy environment override |
| 6 | `.govard/.govard.<env>.yml` | **Preferred** environment override |

### Ownership Model

- **`.govard.yml`** — team-owned base config; target for all `govard config set` writes
- **Profile/local/env overrides** — read-only from CLI perspective; never auto-written by Govard

---

## Profiles

Use profiles when a team needs multiple runtime shapes for the same project.

```bash
govard env up --profile upgrade
govard db dump --profile perf
```

Govard loads `.govard.<profile>.yml` and creates an isolated compose file + separate data volumes, so profile switching does not contaminate existing data.

---

## Environment Override

```bash
export GOVARD_ENV=staging
govard env up
```

With `GOVARD_ENV=staging`, Govard additionally loads:
- `.govard.staging.yml`
- `.govard/.govard.staging.yml`

---

## Global Environment Variables

| Variable | Effect |
| :--- | :--- |
| `GOVARD_HOME_DIR` | Override `~/.govard` |
| `GOVARD_BLUEPRINTS_DIR` | Override blueprint lookup location |
| `GOVARD_IMAGE_REPOSITORY` | Override managed image repository prefix |
| `GOVARD_DOCKER_DIR` | Override local Docker build contexts for fallback builds |

---

## Example `.govard.yml`

```yaml
project_name: "my_project"
framework: "magento2"
framework_version: "2.4.7"
domain: "myproject.test"
lock:
  strict: false
blueprint_registry:
  provider: "http"
  url: "https://example.com/govard-blueprints.tar.gz"
  checksum: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  trusted: false
stack:
  php_version: "8.4"
  node_version: "24"
  db_type: "mariadb"
  db_version: "11.4"
  web_root: "/public"
  cache_version: "7.4"
  search_version: "3.4.0"
  queue_version: "3.13.7"
  xdebug_session: "PHPSTORM"
  services:
    web_server: "nginx"
    search: "opensearch"
    cache: "redis"
    queue: "none"
  features:
    xdebug: true
    varnish: false
    isolated: false
    mftf: false
    livereload: false
```

---

## Key Fields

### Project Identity

| Field | Description |
| :--- | :--- |
| `project_name` | Unique project name (must be unique across all tracked projects) |
| `framework` | Detected or forced framework |
| `framework_version` | Framework version (used for version-aware profiles) |
| `domain` | Primary project domain (e.g. `myproject.test`) |
| `extra_domains` | Additional hostnames routed through the local proxy |
| `store_domains` | Magento multi-store hostname → scope code map |

> [!IMPORTANT]
> `project_name` and `domain` must be **unique** across all tracked Govard projects. Govard blocks `init` and `env up` when another project uses the same identity.

#### `store_domains` — Scalar Form (Legacy)

```yaml
store_domains:
  brand-b.test: brand_b
  brand-c.test: brand_c
```

#### `store_domains` — Object Form (Explicit Routing)

```yaml
store_domains:
  brand-b.test:
    code: base
    type: website
  brand-c.test:
    code: brand_c
    type: store
```

Object form instructs Govard to emit `MAGE_RUN_CODE` / `MAGE_RUN_TYPE` host mappings automatically.

---

### Runtime Stack

| Field | Options | Description |
| :--- | :--- | :--- |
| `stack.services.web_server` | `nginx`, `apache`, `hybrid` | Web server |
| `stack.services.search` | `opensearch`, `elasticsearch`, `none` | Search engine |
| `stack.services.cache` | `redis`, `valkey`, `none` | Cache service |
| `stack.services.queue` | `rabbitmq`, `none` | Queue service |
| `stack.php_version` | e.g. `8.4` | PHP version |
| `stack.node_version` | e.g. `24` | Node.js version |
| `stack.db_type` | `mariadb`, `mysql` | Database engine |
| `stack.db_version` | e.g. `11.4` | Database version |
| `stack.web_root` | e.g. `/pub`, `/public` | Web root directory |
| `stack.xdebug_session` | e.g. `PHPSTORM` | Xdebug session name |

Node-first frameworks auto-detect the package manager from `package.json`, `pnpm-workspace.yaml`, or lockfiles.

---

### Safety and Reproducibility

| Field | Description |
| :--- | :--- |
| `lock.strict` | Fail `env up` when lock state is missing or mismatched |
| `lock.ignore_fields` | Fields to skip during compliance checks (e.g. `host.docker_version`) |
| `blueprint_registry.*` | Opt-in remote blueprint source with checksum + trust requirements |

---

### Remotes

Remote definitions live under `remotes.<name>`:

```yaml
remotes:
  staging:
    host: staging.example.com
    user: deploy
    path: /var/www/app
    port: 22
    capabilities:
      files: true
      media: true
      db: true
    protected: false
    auth:
      method: ssh-agent
```

Key subfields:

| Field | Description |
| :--- | :--- |
| `capabilities` | Scope flags: `files`, `media`, `db`, `deploy` |
| `protected` | Write-protect this remote |
| `auth.method` | `keychain`, `ssh-agent`, or `keyfile` |
| `auth.key_path` | Path to SSH key (for `keyfile` method) |
| `auth.strict_host_key` | Enable strict host-key verification |
| `auth.known_hosts_file` | Custom known_hosts file path |

Remote fields support `op://...` references resolved through the 1Password CLI.

→ Full guide: [Remotes and Sync](Remotes-and-Sync)

---

### Project Extensions

| Path | Purpose |
| :--- | :--- |
| `.govard/docker-compose.override.yml` | Compose overrides merged after framework includes |
| `.govard/commands/*` | Custom commands exposed via `govard custom` |
| `.govard/hooks/*` | Scripts referenced by `hooks.*.run` |

**Lifecycle hook events:**

- `pre-up` / `post-up`
- `pre-down` / `post-down`
- `pre-deploy` / `post-deploy`
- `pre-delete` / `post-delete`

> [!TIP]
> Govard fingerprints `.govard/docker-compose.override.yml`. If it changes, the next `env up` auto-re-renders the compose output.
>
> When overriding services, prefer additive merges (extra environment variables, labels, ports). Replacing full lists like `services.web.volumes` can discard required Govard-managed mounts.

---

## Config Commands

```bash
govard config get stack.php_version
govard config set stack.php_version 8.4
govard config profile --json
govard config profile apply --framework laravel --framework-version 11
```

`govard config set` writes only to `.govard.yml` (the base config).

---

## Blueprint Registry

If `blueprint_registry` is enabled:

- `provider` must be `git` or `http`
- `url` is required
- `checksum` must be a 64-character SHA-256 hex string
- `trusted` must be `true`
- Remote payloads are cached under `~/.govard/blueprint-registry/`

Govard fails fast if the checksum does not match.

---

**[← CLI Commands](CLI-Commands)** | **[Frameworks →](Frameworks)**
