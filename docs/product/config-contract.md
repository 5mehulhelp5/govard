# Govard Config Contract

## Layered Configuration

Govard loads config layers in this order:

1. `govard.yml`
2. `govard.local.yml` (legacy local override)
3. `.govard/govard.local.yml` (project extension local override, preferred)
4. `govard.<env>.yml` (legacy env override, enabled by `GOVARD_ENV`)
5. `.govard/govard.<env>.yml` (project extension env override, preferred)

Later layers override earlier layers.

## Ownership Model

- `govard.yml`:
  - Source of truth committed to repository.
  - Writable by Govard commands (`init`, `config set`, `profile apply`, etc.).
- `govard.local.yml`:
  - Developer-local overrides (legacy path).
  - Not managed by Govard write operations.
- `.govard/govard.local.yml`:
  - Project extension local overrides (preferred path).
  - Not managed by Govard write operations.
- `govard.<env>.yml`:
  - Environment-specific overrides (legacy path).
  - Not managed by Govard write operations.
- `.govard/govard.<env>.yml`:
  - Environment-specific overrides (preferred path).
  - Not managed by Govard write operations.

## Key Fields

- `recipe`: selected framework profile.
- `framework_version`: detected or user-selected framework version for profile resolution.
- `stack.*`: runtime service settings.
- `remotes.*`: remote definitions.
  - `remotes.<name>.environment`: `dev`, `staging`, or `prod`.
  - `remotes.<name>.capabilities`: operation scopes (`files`, `media`, `db`, `deploy`).
  - `remotes.<name>.protected`: explicit write-protection toggle.
  - `remotes.<name>.auth.method`: `keychain`, `ssh-agent`, or `keyfile`.
  - `remotes.<name>.auth.key_path`: explicit SSH key path override for command execution.
  - `remotes.<name>.auth.strict_host_key`: enables strict SSH host key validation.
  - `remotes.<name>.auth.known_hosts_file`: optional custom known hosts file path.
- `hooks.*`: lifecycle hook steps.
- `.govard/docker-compose.override.yml`: project compose overrides merged after framework includes.
- `.govard/commands/*`: project custom commands exposed under `govard custom`.
- `.govard/hooks/*`: project hook scripts referenced by `hooks.*.run`.

## Non-Goals

- Govard does not mutate application dependency manifests/lockfiles when applying runtime profile.
- Govard runtime profile decisions are isolated to Govard configuration.
