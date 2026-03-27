# Configuration

Govard uses layered project configuration plus framework blueprints.

## Layer Order

Govard loads config in this order:

1. `.govard.yml`
2. `.govard.<profile>.yml`
3. `.govard.local.yml`
4. `.govard/.govard.local.yml`
5. `.govard.<env>.yml`
6. `.govard/.govard.<env>.yml`

Later layers override earlier layers.

## Ownership Model

- `.govard.yml`: team-owned base config and the main writable file for Govard commands.
- `.govard.<profile>.yml`: team-shared profile override selected with `--profile`.
- `.govard.local.yml`: legacy developer-local override.
- `.govard/.govard.local.yml`: preferred developer-local override.
- `.govard.<env>.yml`: legacy environment override selected by `GOVARD_ENV`.
- `.govard/.govard.<env>.yml`: preferred environment override.

Govard write operations target the base config only. Override layers are treated as read-only inputs.

## Profiles

Use profiles when a team needs multiple runtime shapes for the same project.

```bash
govard env up --profile upgrade
govard db dump --profile perf
```

Profile behavior:

- loads `.govard.<profile>.yml`
- creates an isolated compose file under `~/.govard/compose/`
- uses separate data volumes so profile switching does not contaminate data

## Environment Override

```bash
export GOVARD_ENV=staging
govard env up
```

With `GOVARD_ENV=staging`, Govard additionally loads:

- `.govard.staging.yml`
- `.govard/.govard.staging.yml`

## Global Environment Variables

- `GOVARD_HOME_DIR`: override `~/.govard`
- `GOVARD_BLUEPRINTS_DIR`: override blueprint lookup location
- `GOVARD_IMAGE_REPOSITORY`: override managed image repository prefix
- `GOVARD_DOCKER_DIR`: override local Docker build contexts for fallback builds

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

## Key Fields

### Project identity

- `project_name`
- `framework`
- `framework_version`
- `domain`
- `extra_domains`
- `store_domains`

`extra_domains` tells Govard which additional hostnames should resolve through the local proxy.

`store_domains` is a Magento convenience map of:

```yaml
store_domains:
  brand-b.test: brand_b
  brand-c.test: brand_c
```

Or, when you need explicit runtime routing type:

```yaml
store_domains:
  brand-b.test:
    code: base
    type: website
  brand-c.test:
    code: brand_c
    type: store
```

Each key is a local hostname. Govard routes those hostnames through the local proxy automatically. Each value identifies the Magento scope code that should receive that base URL when you run `govard config auto`.

- Legacy scalar form keeps existing behavior:
  - Magento 2: scalar values are treated as store codes and use `bin/magento config:set --scope=stores`
  - Magento 1 / OpenMage: scalar values are tried as both website codes and store codes in `core_config_data`
- Object form lets you choose the scope explicitly with `type: website` or `type: store`

### Runtime stack

- `stack.services.web_server`: `nginx`, `apache`, `hybrid`
- `stack.services.search`: `opensearch`, `elasticsearch`, `none`
- `stack.services.cache`: `redis`, `valkey`, `none`
- `stack.services.queue`: `rabbitmq`, `none`
- `stack.php_version`
- `stack.node_version`
- `stack.db_type`
- `stack.db_version`
- `stack.web_root`
- `stack.xdebug_session`

### Safety and reproducibility

- `lock.strict`: fail `govard env up` when lock state is missing or mismatched
- `lock.ignore_fields`: list of fields (e.g., `host.docker_version`) to skip during compliance checks
- `blueprint_registry.*`: opt-in remote blueprint source with strict checksum and trust requirements

### Remotes

Remote definitions live under `remotes.<name>`.

Important subfields:

- `capabilities`
- `protected`
- `auth.method`
- `auth.key_path`
- `auth.strict_host_key`
- `auth.known_hosts_file`

See [Remotes and Sync](remotes-and-sync.md) for the operational side.

### Project extensions

- `.govard/docker-compose.override.yml`: compose overrides merged after framework includes
- `.govard/commands/*`: custom commands exposed through `govard custom`
- `.govard/hooks/*`: scripts referenced by `hooks.*.run`

Govard fingerprints `.govard/docker-compose.override.yml` during render. If that file changes, the next `govard env up` re-renders the managed compose output automatically.

When overriding service definitions, prefer additive merges such as extra environment variables, labels, or ports. Replacing a full list like `services.web.volumes` can discard required Govard-managed mounts.

## Config Commands

```bash
govard config get stack.php_version
govard config set stack.php_version 8.4
govard config profile --json
govard config profile apply --framework laravel --framework-version 11
```

`govard config set ...` writes only to `.govard.yml`.

## Blueprint Registry Rules

If `blueprint_registry` is enabled:

- `provider` must be `git` or `http`
- `url` is required
- `checksum` is required and must be a 64-character SHA-256 hex string
- `trusted` must be `true`
- remote payloads are cached under `~/.govard/blueprint-registry/`

Govard fails fast if the checksum does not match.

## Related Docs

- [Commands](commands.md)
- [Frameworks](frameworks.md)
- [Remotes and Sync](remotes-and-sync.md)
